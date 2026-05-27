from __future__ import annotations

import asyncio
import logging
import tempfile
from pathlib import Path
from typing import Any

from voice_adapter.services import audio
from voice_adapter.services.asr import AsrBackend
from voice_adapter.services.backend import AudioResult, ProgressCallback, SynthesisInput

logger = logging.getLogger(__name__)


class OmniVoiceBackend:
    model_provider = "omnivoice"

    def __init__(
        self,
        model_id: str,
        device: str,
        dtype: str,
        load_asr: bool,
        asr_model: str,
        asr_backend: AsrBackend | None = None,
    ) -> None:
        self.model_id = model_id
        self.device = device
        self.dtype = dtype
        self.load_asr = load_asr
        self.asr_model = asr_model
        self.asr_backend = asr_backend
        self.model: Any | None = None
        self.sample_rate_hz = 24_000

    async def load(self) -> None:
        import torch
        from omnivoice import OmniVoice

        device = self._resolve_device(torch)
        dtype = self._resolve_dtype(torch)

        logger.info(
            "voice model loading",
            extra={
                "model_provider": self.model_provider,
                "model_id": self.model_id,
                "requested_device": self.device,
                "resolved_device": device,
                "dtype": str(dtype),
                "internal_asr_enabled": self.load_asr,
                "external_asr_provider": self.asr_backend.provider if self.asr_backend else "",
            },
        )

        self.model = await _run_with_heartbeat(
            "voice model",
            lambda: OmniVoice.from_pretrained(
                self.model_id,
                device_map=device,
                dtype=dtype,
                load_asr=self.load_asr,
                asr_model_name=self.asr_model,
            ),
            extra={
                "model_provider": self.model_provider,
                "model_id": self.model_id,
                "device": device,
            },
        )
        self.device = device
        self.sample_rate_hz = int(getattr(self.model, "sampling_rate", 24_000))
        logger.info(
            "voice model ready",
            extra={
                "model_provider": self.model_provider,
                "model_id": self.model_id,
                "device": self.device,
                "sample_rate_hz": self.sample_rate_hz,
            },
        )
        if self.asr_backend is not None:
            await self.asr_backend.load()
        logger.info(
            "voice backend ready",
            extra={
                "model_provider": self.model_provider,
                "model_id": self.model_id,
                "device": self.device,
                "asr_provider": self.asr_backend.provider if self.asr_backend else "",
            },
        )

    async def synthesize(
        self,
        request: SynthesisInput,
        progress: ProgressCallback | None = None,
    ) -> AudioResult:
        if self.model is None:
            raise RuntimeError("voice model is not loaded")

        if progress:
            await progress("preparing", "preparing generation request", 20)

        cleanup_paths: list[Path] = []
        try:
            kwargs, cleanup_paths = await self._generation_kwargs(request, progress)
            if progress:
                await progress("generating", "running voice model", 45)
            generated = self.model.generate(**kwargs)
        finally:
            for path in cleanup_paths:
                try:
                    path.unlink(missing_ok=True)
                except OSError:
                    pass
            self._release_cuda_cache()
        if not generated:
            raise RuntimeError("voice model returned no audio")

        waveform = generated[0]
        sample_rate = self.sample_rate_hz
        wav_data, peaks, duration_ms = audio.wav_pcm16_from_samples(waveform, sample_rate)
        del generated, waveform
        self._release_cuda_cache()

        target_sample_rate = request.output.sample_rate_hz
        if target_sample_rate and target_sample_rate != sample_rate:
            wav_data = await audio.resample_wav_bytes(wav_data, target_sample_rate)
            sample_rate = target_sample_rate
            peaks = audio.wav_peaks(wav_data)
            duration_ms = audio.wav_duration_ms(wav_data)

        if progress:
            await progress("encoding", "encoding output audio", 80)

        if request.output.format == "OGG_OPUS":
            ogg_data = await audio.wav_to_ogg_opus(wav_data)
            result = AudioResult(
                data=ogg_data,
                format="OGG_OPUS",
                mime_type="audio/ogg; codecs=opus",
                extension=".ogg",
                sample_rate_hz=audio.OPUS_SAMPLE_RATE,
                duration_ms=duration_ms,
                waveform_peaks=peaks,
            )
            return result

        result = AudioResult(
            data=wav_data,
            format="WAV_PCM16",
            mime_type="audio/wav",
            extension=".wav",
            sample_rate_hz=sample_rate,
            duration_ms=duration_ms,
            waveform_peaks=peaks,
        )
        return result

    async def close(self) -> None:
        if self.asr_backend is not None:
            await self.asr_backend.close()
        self.model = None

    async def transcribe(self, audio_path: Path) -> str:
        if self.asr_backend is None:
            raise RuntimeError("ASR backend is not loaded")

        normalized_path = await audio.normalize_audio_file(audio_path)
        try:
            text = await self.asr_backend.transcribe(normalized_path)
        finally:
            try:
                normalized_path.unlink(missing_ok=True)
            except OSError:
                pass

        return text.strip()

    async def _generation_kwargs(
        self,
        request: SynthesisInput,
        progress: ProgressCallback | None = None,
    ) -> tuple[dict[str, Any], list[Path]]:
        from omnivoice import OmniVoiceGenerationConfig

        config_kwargs: dict[str, Any] = {}
        if request.generation.inference_steps is not None:
            config_kwargs["num_step"] = request.generation.inference_steps
        if request.generation.guidance_scale is not None:
            config_kwargs["guidance_scale"] = request.generation.guidance_scale
        if request.generation.denoise is not None:
            config_kwargs["denoise"] = request.generation.denoise
        if request.generation.preprocess_prompt is not None:
            config_kwargs["preprocess_prompt"] = request.generation.preprocess_prompt
        if request.generation.postprocess_output is not None:
            config_kwargs["postprocess_output"] = request.generation.postprocess_output

        kwargs = dict(request.model_options)
        cleanup_paths: list[Path] = []
        kwargs.update(
            {
                "text": request.text,
                "generation_config": OmniVoiceGenerationConfig(**config_kwargs),
            }
        )

        if request.language:
            kwargs["language"] = request.language
        if request.generation.speed is not None:
            kwargs["speed"] = request.generation.speed
        if request.generation.duration_seconds is not None:
            kwargs["duration"] = request.generation.duration_seconds
        if request.generation.audio_chunk_duration is not None:
            kwargs["audio_chunk_duration"] = request.generation.audio_chunk_duration
        if request.generation.audio_chunk_threshold is not None:
            kwargs["audio_chunk_threshold"] = request.generation.audio_chunk_threshold
        if request.instruct:
            kwargs["instruct"] = request.instruct

        if request.mode == "VOICE_CLONE":
            if request.reference_audio is None:
                raise ValueError("reference audio is required")
            ref_text = request.reference_text
            ref_audio_path = request.reference_audio.path
            preprocess_prompt = request.generation.preprocess_prompt
            if preprocess_prompt is None:
                preprocess_prompt = True
            normalized_path: Path | None = None
            if preprocess_prompt:
                if progress:
                    await progress("preprocessing", "preprocessing reference audio", 30)
                normalized_path = await self._preprocess_reference_audio(ref_audio_path)
                ref_audio_path = normalized_path
                cleanup_paths.append(normalized_path)

            if not ref_text and self.asr_backend is not None:
                if normalized_path is None:
                    # ASR needs normalized mono 24kHz WAV input.
                    normalized_path = await audio.normalize_audio_file(ref_audio_path)
                    cleanup_paths.append(normalized_path)
                if progress:
                    await progress("transcribing", "transcribing reference audio", 35)
                ref_text = await self.asr_backend.transcribe(normalized_path)
                if not ref_text:
                    raise ValueError("reference text auto-transcription returned empty text")
            # Match OmniVoice demo behavior for long references. If external ASR
            # is used, we trim before transcription so the transcript still
            # describes the prompt audio and the tokenizer does not receive a
            # 20+ second clip.
            kwargs["voice_clone_prompt"] = self.model.create_voice_clone_prompt(
                ref_audio=str(ref_audio_path),
                ref_text=ref_text or None,
                preprocess_prompt=preprocess_prompt,
            )

        return kwargs, cleanup_paths

    async def _preprocess_reference_audio(self, audio_path: Path) -> Path:
        return await asyncio.to_thread(self._preprocess_reference_audio_sync, audio_path)

    def _preprocess_reference_audio_sync(self, audio_path: Path) -> Path:
        import numpy as np
        import soundfile as sf
        from omnivoice.utils.audio import load_audio, remove_silence, trim_long_audio

        ref_wav = load_audio(str(audio_path), self.sample_rate_hz)
        ref_rms = float(np.sqrt(np.mean(ref_wav**2)))
        if 0 < ref_rms < 0.1:
            ref_wav = ref_wav * 0.1 / ref_rms

        ref_wav = trim_long_audio(ref_wav, self.sample_rate_hz, trim_threshold=20.0)
        ref_wav = remove_silence(
            ref_wav,
            self.sample_rate_hz,
            mid_sil=200,
            lead_sil=100,
            trail_sil=200,
        )
        if ref_wav.shape[-1] == 0:
            raise ValueError(
                "reference audio is empty after silence removal; try disabling prompt preprocessing"
            )

        handle = tempfile.NamedTemporaryFile(prefix="clicksafe-voice-ref-", suffix=".wav", delete=False)
        target = Path(handle.name)
        handle.close()
        sf.write(str(target), ref_wav.T, self.sample_rate_hz, subtype="PCM_16")
        return target

    def _resolve_device(self, torch) -> str:
        if self.device and self.device != "auto":
            return self.device

        try:
            from omnivoice.utils.common import get_best_device

            return str(get_best_device())
        except Exception:
            pass

        if torch.cuda.is_available():
            return "cuda:0"
        if hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
            return "mps"
        if hasattr(torch, "xpu") and torch.xpu.is_available():
            return "xpu"
        return "cpu"

    def _resolve_dtype(self, torch):
        dtype = self.dtype.strip().lower()
        if dtype in {"", "auto"}:
            return torch.float32 if self._resolve_device(torch) == "cpu" else torch.float16
        if dtype == "float16":
            return torch.float16
        if dtype == "float32":
            return torch.float32
        if dtype == "bfloat16":
            return torch.bfloat16
        raise ValueError(f"unsupported dtype: {self.dtype}")

    def _release_cuda_cache(self) -> None:
        if not str(self.device).startswith("cuda"):
            return
        try:
            import torch

            if torch.cuda.is_available():
                torch.cuda.empty_cache()
        except Exception:
            logger.debug("failed to release cuda cache", exc_info=True)


async def _run_with_heartbeat(label: str, func, extra: dict[str, Any], interval_seconds: int = 15):
    task = asyncio.create_task(asyncio.to_thread(func))
    started = asyncio.get_running_loop().time()
    try:
        while not task.done():
            await asyncio.sleep(interval_seconds)
            if task.done():
                break
            elapsed = round(asyncio.get_running_loop().time() - started)
            logger.info(
                f"{label} still loading",
                extra={
                    **extra,
                    "elapsed_seconds": elapsed,
                },
            )
        return await task
    except Exception:
        task.cancel()
        raise
