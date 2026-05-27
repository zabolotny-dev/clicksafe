from __future__ import annotations

import asyncio
import logging
import shutil
from pathlib import Path
from uuid import uuid4

import grpc
from google.protobuf.json_format import MessageToDict

from voiceadapter.v1 import voice_adapter_pb2 as pb2
from voiceadapter.v1 import voice_adapter_pb2_grpc as pb2_grpc

from voice_adapter.config import Config
from voice_adapter.services.auth import require_internal_token
from voice_adapter.services.backend import (
    GenerationOptions,
    OutputOptions,
    ReferenceAudio,
    SynthesisInput,
    VoiceBackend,
)

logger = logging.getLogger(__name__)


MODE_TO_NAME = {
    pb2.SYNTHESIS_MODE_VOICE_CLONE: "VOICE_CLONE",
    pb2.SYNTHESIS_MODE_VOICE_DESIGN: "VOICE_DESIGN",
    pb2.SYNTHESIS_MODE_AUTO_VOICE: "AUTO_VOICE",
}

FORMAT_TO_NAME = {
    pb2.AUDIO_OUTPUT_FORMAT_WAV_PCM16: "WAV_PCM16",
    pb2.AUDIO_OUTPUT_FORMAT_OGG_OPUS: "OGG_OPUS",
}

FORMAT_TO_PROTO = {
    "WAV_PCM16": pb2.AUDIO_OUTPUT_FORMAT_WAV_PCM16,
    "OGG_OPUS": pb2.AUDIO_OUTPUT_FORMAT_OGG_OPUS,
}


class VoiceAdapterServicer(pb2_grpc.VoiceAdapterServiceServicer):
    def __init__(self, cfg: Config, backend: VoiceBackend) -> None:
        self.cfg = cfg
        self.backend = backend
        self.temp_dir = Path(cfg.temp_dir)
        self.temp_dir.mkdir(parents=True, exist_ok=True)
        self.semaphore = asyncio.Semaphore(cfg.concurrency)

    async def Health(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        return pb2.HealthResponse(
            status="ok",
            model_provider=self.backend.model_provider,
            model_id=self.backend.model_id,
            device=self.backend.device,
        )

    async def Transcribe(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)

        if not request.audio_data:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "audio data is required")

        request_id = str(uuid4())
        request_dir = self.temp_dir / request_id
        request_dir.mkdir(parents=True, exist_ok=True)
        audio_path = request_dir / self._safe_filename(request.audio.filename)

        try:
            audio_path.write_bytes(request.audio_data)
            text = await self.backend.transcribe(audio_path)
            if not text:
                await context.abort(grpc.StatusCode.FAILED_PRECONDITION, "ASR returned empty text")
            return pb2.TranscribeResponse(text=text)
        except grpc.RpcError:
            raise
        except ValueError as exc:
            logger.info("transcribe invalid", extra={"error": str(exc)})
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
        except Exception as exc:
            logger.exception("transcribe failed")
            await context.abort(grpc.StatusCode.FAILED_PRECONDITION, str(exc))
        finally:
            if request_dir.exists():
                shutil.rmtree(request_dir, ignore_errors=True)

    async def Synthesize(self, request_iterator, context):
        await require_internal_token(context, self.cfg.grpc_token)

        metadata = None
        request_dir: Path | None = None
        reference_path: Path | None = None
        reference_size = 0

        try:
            async for request in request_iterator:
                payload = request.WhichOneof("payload")
                if payload == "metadata":
                    if metadata is not None:
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "metadata sent twice")
                    metadata = request.metadata
                    request_id = metadata.request_id.strip() or str(uuid4())
                    metadata.request_id = request_id
                    request_dir = self.temp_dir / request_id
                    request_dir.mkdir(parents=True, exist_ok=True)
                    if metadata.mode == pb2.SYNTHESIS_MODE_VOICE_CLONE:
                        reference_path = request_dir / self._safe_filename(metadata.reference_audio.filename)
                    continue

                if payload == "reference_audio_chunk":
                    if metadata is None:
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "metadata must be sent before chunks")
                    if metadata.mode != pb2.SYNTHESIS_MODE_VOICE_CLONE:
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "reference audio chunks are only valid for voice clone")
                    if request_dir is None or reference_path is None:
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "reference audio metadata is required")
                    data = request.reference_audio_chunk.data
                    with reference_path.open("ab") as file:
                        file.write(data)
                    reference_size += len(data)
                    continue

                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "request payload is required")

            if metadata is None:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "metadata is required")

            synthesis = await self._to_synthesis_input(metadata, reference_path, reference_size, context)

            logger.info(
                "synthesize request",
                extra={
                    "request_id": synthesis.request_id,
                    "mode": synthesis.mode,
                    "model_provider": self.backend.model_provider,
                    "model_id": self.backend.model_id,
                    "device": self.backend.device,
                    "output_format": synthesis.output.format,
                },
            )

            yield self._progress("accepted", "request accepted", 5)
            async with self.semaphore:
                yield self._progress("synthesizing", "running voice synthesis", 25)
                result = await self.backend.synthesize(synthesis)

            yield self._progress("streaming", "streaming generated audio", 90)
            offset = 0
            chunk_size = max(1, synthesis.output.chunk_size)
            while offset < len(result.data):
                chunk = result.data[offset : offset + chunk_size]
                yield pb2.SynthesizeEvent(
                    audio_chunk=pb2.AudioChunk(
                        data=chunk,
                        offset=offset,
                    )
                )
                offset += len(chunk)

            logger.info(
                "synthesize complete",
                extra={
                    "request_id": synthesis.request_id,
                    "mode": synthesis.mode,
                    "output_format": result.format,
                    "duration_ms": result.duration_ms,
                    "total_bytes": result.total_bytes,
                },
            )
            yield pb2.SynthesizeEvent(
                completed=pb2.SynthesisCompleted(
                    request_id=synthesis.request_id,
                    format=FORMAT_TO_PROTO.get(result.format, pb2.AUDIO_OUTPUT_FORMAT_UNSPECIFIED),
                    mime_type=result.mime_type,
                    extension=result.extension,
                    sample_rate_hz=result.sample_rate_hz,
                    duration_ms=result.duration_ms,
                    total_bytes=result.total_bytes,
                    waveform_peaks=result.waveform_peaks,
                )
            )
        except grpc.RpcError:
            raise
        except ValueError as exc:
            logger.info("synthesize invalid", extra={"error": str(exc)})
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
        except Exception as exc:
            logger.exception("synthesize failed")
            await context.abort(grpc.StatusCode.FAILED_PRECONDITION, str(exc))
        finally:
            if request_dir and request_dir.exists():
                shutil.rmtree(request_dir, ignore_errors=True)

    async def _to_synthesis_input(
        self,
        metadata: pb2.SynthesisMetadata,
        reference_path: Path | None,
        reference_size: int,
        context,
    ) -> SynthesisInput:
        text = metadata.text.strip()
        if not text:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "text is required")

        mode = MODE_TO_NAME.get(metadata.mode)
        if not mode:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "mode is required")

        reference_audio = None
        if mode == "VOICE_CLONE":
            if reference_path is None or reference_size == 0 or not reference_path.exists():
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "reference audio is required")
            reference_audio = ReferenceAudio(
                path=reference_path,
                filename=metadata.reference_audio.filename,
                mime_type=metadata.reference_audio.mime_type,
                extension=metadata.reference_audio.extension,
                size=metadata.reference_audio.size or reference_size,
            )
        elif mode == "VOICE_DESIGN" and not metadata.instruct.strip():
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "instruct is required for voice design")

        return SynthesisInput(
            request_id=metadata.request_id.strip() or str(uuid4()),
            text=text,
            mode=mode,
            language=self._language(metadata.language),
            reference_text=metadata.reference_text.strip() or None,
            instruct=metadata.instruct.strip() or None,
            reference_audio=reference_audio,
            generation=self._generation_options(metadata.generation_config),
            output=self._output_options(metadata.output_config),
            model_options=MessageToDict(metadata.model_options),
        )

    @staticmethod
    def _generation_options(config: pb2.GenerationConfig) -> GenerationOptions:
        return GenerationOptions(
            speed=config.speed if config.HasField("speed") else None,
            duration_seconds=config.duration_seconds if config.HasField("duration_seconds") else None,
            inference_steps=config.inference_steps if config.HasField("inference_steps") else None,
            guidance_scale=config.guidance_scale if config.HasField("guidance_scale") else None,
            denoise=config.denoise if config.HasField("denoise") else None,
            preprocess_prompt=config.preprocess_prompt if config.HasField("preprocess_prompt") else None,
            postprocess_output=config.postprocess_output if config.HasField("postprocess_output") else None,
            audio_chunk_duration=config.audio_chunk_duration if config.HasField("audio_chunk_duration") else None,
            audio_chunk_threshold=config.audio_chunk_threshold if config.HasField("audio_chunk_threshold") else None,
        )

    @staticmethod
    def _output_options(config: pb2.OutputConfig) -> OutputOptions:
        output_format = FORMAT_TO_NAME.get(config.format, "WAV_PCM16")
        return OutputOptions(
            format=output_format,
            sample_rate_hz=config.sample_rate_hz if config.sample_rate_hz > 0 else None,
            chunk_size=config.chunk_size if config.chunk_size > 0 else 64 * 1024,
        )

    @staticmethod
    def _progress(stage: str, message: str, percent: int) -> pb2.SynthesizeEvent:
        return pb2.SynthesizeEvent(
            progress=pb2.SynthesisProgress(
                stage=stage,
                message=message,
                percent=percent,
            )
        )

    @staticmethod
    def _safe_filename(filename: str) -> str:
        clean = Path(filename or "reference-audio").name
        return clean or "reference-audio"

    @staticmethod
    def _language(value: str) -> str | None:
        language = value.strip()
        if not language or language.lower() == "auto":
            return None
        return language
