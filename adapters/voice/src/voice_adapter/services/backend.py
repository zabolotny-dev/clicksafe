from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Awaitable, Callable, Protocol


ProgressCallback = Callable[[str, str, int], Awaitable[None]]


@dataclass(frozen=True)
class ReferenceAudio:
    path: Path
    filename: str
    mime_type: str
    extension: str
    size: int


@dataclass(frozen=True)
class GenerationOptions:
    speed: float | None = None
    duration_seconds: float | None = None
    inference_steps: int | None = None
    guidance_scale: float | None = None
    denoise: bool | None = None
    preprocess_prompt: bool | None = None
    postprocess_output: bool | None = None
    audio_chunk_duration: float | None = None
    audio_chunk_threshold: float | None = None


@dataclass(frozen=True)
class OutputOptions:
    format: str = "WAV_PCM16"
    sample_rate_hz: int | None = None
    chunk_size: int = 64 * 1024


@dataclass(frozen=True)
class SynthesisInput:
    request_id: str
    text: str
    mode: str
    language: str | None = None
    reference_text: str | None = None
    instruct: str | None = None
    reference_audio: ReferenceAudio | None = None
    generation: GenerationOptions = field(default_factory=GenerationOptions)
    output: OutputOptions = field(default_factory=OutputOptions)
    model_options: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class AudioResult:
    data: bytes
    format: str
    mime_type: str
    extension: str
    sample_rate_hz: int
    duration_ms: int
    waveform_peaks: bytes = b""

    @property
    def total_bytes(self) -> int:
        return len(self.data)


class VoiceBackend(Protocol):
    model_provider: str
    model_id: str
    device: str

    async def load(self) -> None:
        ...

    async def synthesize(
        self,
        request: SynthesisInput,
        progress: ProgressCallback | None = None,
    ) -> AudioResult:
        ...

    async def transcribe(self, audio_path: Path) -> str:
        ...

    async def close(self) -> None:
        ...
