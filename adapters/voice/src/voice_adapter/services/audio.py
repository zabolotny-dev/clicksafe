from __future__ import annotations

import asyncio
import io
import os
import shutil
import struct
import tempfile
import wave
from pathlib import Path
from typing import Iterable


WAVEFORM_SAMPLES = 80
WAVEFORM_MAX_VALUE = 100
OPUS_SAMPLE_RATE = 48_000


def wav_pcm16_from_samples(samples: object, sample_rate_hz: int) -> tuple[bytes, bytes, int]:
    pcm = pcm16_from_samples(samples)
    wav_data = wav_from_pcm16(pcm, sample_rate_hz)
    duration_ms = duration_ms_from_pcm16(pcm, sample_rate_hz)
    peaks = waveform_peaks_from_pcm16(pcm)
    return wav_data, peaks, duration_ms


def pcm16_from_samples(samples: object) -> bytes:
    values = _iter_samples(samples)
    out = bytearray()
    for value in values:
        try:
            sample = float(value)
        except (TypeError, ValueError):
            sample = 0.0
        sample = max(-1.0, min(1.0, sample))
        out.extend(struct.pack("<h", int(sample * 32767)))
    return bytes(out)


def wav_from_pcm16(pcm: bytes, sample_rate_hz: int) -> bytes:
    out = io.BytesIO()
    with wave.open(out, "wb") as writer:
        writer.setnchannels(1)
        writer.setsampwidth(2)
        writer.setframerate(sample_rate_hz)
        writer.writeframes(pcm)
    return out.getvalue()


def duration_ms_from_pcm16(pcm: bytes, sample_rate_hz: int) -> int:
    if sample_rate_hz <= 0:
        return 0
    samples = len(pcm) // 2
    return max(1, round(samples * 1000 / sample_rate_hz)) if samples else 0


def waveform_peaks_from_pcm16(pcm: bytes) -> bytes:
    if not pcm:
        return b""

    samples = len(pcm) // 2
    peaks = [0] * WAVEFORM_SAMPLES
    for index, (sample,) in enumerate(struct.iter_unpack("<h", pcm[: samples * 2])):
        bucket = min(WAVEFORM_SAMPLES - 1, index * WAVEFORM_SAMPLES // samples)
        peaks[bucket] = max(peaks[bucket], abs(sample))

    max_peak = max(peaks) or 1
    return bytes(
        max(
            1,
            min(WAVEFORM_MAX_VALUE, round(peak * WAVEFORM_MAX_VALUE / max_peak)),
        )
        for peak in peaks
    )


async def normalize_audio_file(source: Path, sample_rate_hz: int = 24_000) -> Path:
    target = _new_temp_path(".wav")
    returncode, _, stderr = await _run_ffmpeg(
        "-nostdin",
        "-hide_banner",
        "-loglevel",
        "error",
        "-y",
        "-i",
        str(source),
        "-vn",
        "-ac",
        "1",
        "-ar",
        str(sample_rate_hz),
        "-c:a",
        "pcm_s16le",
        str(target),
    )
    if returncode != 0:
        target.unlink(missing_ok=True)
        message = stderr.decode("utf-8", errors="replace").strip()
        raise ValueError(f"ffmpeg failed to normalize reference audio: {message}")
    return target


async def resample_wav_bytes(wav_data: bytes, sample_rate_hz: int) -> bytes:
    returncode, stdout, stderr = await _run_ffmpeg(
        "-nostdin",
        "-hide_banner",
        "-loglevel",
        "error",
        "-i",
        "pipe:0",
        "-f",
        "wav",
        "-ac",
        "1",
        "-ar",
        str(sample_rate_hz),
        "pipe:1",
        input_data=wav_data,
    )
    if returncode != 0:
        message = stderr.decode("utf-8", errors="replace").strip()
        raise ValueError(f"ffmpeg failed to resample audio: {message}")
    return stdout


async def wav_to_ogg_opus(wav_data: bytes) -> bytes:
    returncode, stdout, stderr = await _run_ffmpeg(
        "-nostdin",
        "-hide_banner",
        "-loglevel",
        "error",
        "-i",
        "pipe:0",
        "-vn",
        "-ac",
        "1",
        "-ar",
        str(OPUS_SAMPLE_RATE),
        "-c:a",
        "libopus",
        "-b:a",
        "32k",
        "-application",
        "voip",
        "-f",
        "ogg",
        "pipe:1",
        input_data=wav_data,
    )
    if returncode != 0:
        message = stderr.decode("utf-8", errors="replace").strip()
        raise ValueError(f"ffmpeg failed to convert audio to Ogg Opus: {message}")
    return stdout


def wav_duration_ms(wav_data: bytes) -> int:
    with wave.open(io.BytesIO(wav_data), "rb") as reader:
        frames = reader.getnframes()
        sample_rate = reader.getframerate()
    if sample_rate <= 0:
        return 0
    return max(1, round(frames * 1000 / sample_rate)) if frames else 0


def wav_peaks(wav_data: bytes) -> bytes:
    with wave.open(io.BytesIO(wav_data), "rb") as reader:
        frames = reader.readframes(reader.getnframes())
        channels = reader.getnchannels()
        sample_width = reader.getsampwidth()
    return waveform_peaks_from_pcm16(_pcm16_mono(frames, sample_width, channels))


def resolve_ffmpeg_path() -> str:
    try:
        import imageio_ffmpeg
    except ImportError:
        raise RuntimeError("imageio-ffmpeg is required for audio conversion") from None

    return imageio_ffmpeg.get_ffmpeg_exe()


def configure_audio_runtime() -> str:
    ffmpeg_path = resolve_ffmpeg_path()
    ffmpeg_path = _ensure_ffmpeg_on_path(ffmpeg_path)

    # pydub is pulled in by OmniVoice and probes for a system ffmpeg binary at
    # import time. Point it at the imageio-ffmpeg executable before OmniVoice is
    # imported so we keep using bundled ffmpeg without apt-installing ffmpeg.
    os.environ.setdefault("FFMPEG_BINARY", ffmpeg_path)
    try:
        from pydub import AudioSegment

        AudioSegment.converter = ffmpeg_path
        AudioSegment.ffmpeg = ffmpeg_path
    except ImportError:
        pass

    return ffmpeg_path


def _ensure_ffmpeg_on_path(ffmpeg_path: str) -> str:
    existing = shutil.which("ffmpeg")
    if existing:
        return existing

    bin_dir = Path(tempfile.gettempdir()) / "clicksafe-voice-bin"
    bin_dir.mkdir(parents=True, exist_ok=True)
    link_path = bin_dir / "ffmpeg"

    if not link_path.exists():
        try:
            link_path.symlink_to(ffmpeg_path)
        except OSError:
            shutil.copy2(ffmpeg_path, link_path)
            link_path.chmod(0o755)

    os.environ["PATH"] = f"{bin_dir}{os.pathsep}{os.environ.get('PATH', '')}"
    return str(link_path)


async def _run_ffmpeg(*args: str, input_data: bytes | None = None) -> tuple[int, bytes, bytes]:
    ffmpeg_path = resolve_ffmpeg_path()
    proc = await asyncio.create_subprocess_exec(
        ffmpeg_path,
        *args,
        stdin=asyncio.subprocess.PIPE if input_data is not None else None,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
    )
    stdout, stderr = await proc.communicate(input_data)
    return proc.returncode or 0, stdout, stderr


def _new_temp_path(suffix: str) -> Path:
    handle = tempfile.NamedTemporaryFile(prefix="clicksafe-voice-", suffix=suffix, delete=False)
    path = Path(handle.name)
    handle.close()
    return path


def _iter_samples(samples: object) -> Iterable[object]:
    if hasattr(samples, "tolist"):
        values = samples.tolist()
    else:
        values = samples

    if isinstance(values, (bytes, bytearray)):
        for sample in values:
            yield (sample - 128) / 128
        return

    for value in values:  # type: ignore[union-attr]
        yield value


def _pcm16_mono(frames: bytes, sample_width: int, channels: int) -> bytes:
    if sample_width <= 0 or channels <= 0:
        return b""

    frame_width = sample_width * channels
    if frame_width <= 0:
        return b""

    out = bytearray()
    for index in range(0, len(frames) - frame_width + 1, frame_width):
        sample = frames[index : index + sample_width]
        if sample_width == 2:
            out.extend(sample)
            continue
        if sample_width == 1:
            value = (sample[0] - 128) << 8
            out.extend(struct.pack("<h", value))
            continue
        value = int.from_bytes(sample, "little", signed=True)
        shift = max(0, (sample_width - 2) * 8)
        out.extend(struct.pack("<h", max(-32768, min(32767, value >> shift))))
    return bytes(out)
