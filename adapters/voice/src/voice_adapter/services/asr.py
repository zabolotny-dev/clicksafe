from __future__ import annotations

import asyncio
import logging
from pathlib import Path
from typing import Any, Protocol

logger = logging.getLogger(__name__)


class AsrBackend(Protocol):
    provider: str
    model_id: str

    async def load(self) -> None:
        ...

    async def transcribe(self, audio_path: Path) -> str:
        ...

    async def close(self) -> None:
        ...


class GigaAMAsrBackend:
    provider = "gigaam"

    def __init__(
        self,
        model_id: str,
        revision: str,
        trust_remote_code: bool,
        device: str,
    ) -> None:
        self.model_id = model_id
        self.revision = revision
        self.trust_remote_code = trust_remote_code
        self.device = device
        self.model: Any | None = None

    async def load(self) -> None:
        import torch

        self.device = _resolve_device(torch, self.device)
        logger.info(
            "asr model loading",
            extra={
                "asr_provider": self.provider,
                "asr_model_id": self.model_id,
                "asr_revision": self.revision,
                "resolved_device": self.device,
            },
        )

        from transformers import AutoModel

        self.model = await _run_with_heartbeat(
            "asr model",
            lambda: self._load_model(AutoModel, torch),
            extra={
                "asr_provider": self.provider,
                "asr_model_id": self.model_id,
                "asr_revision": self.revision,
                "device": self.device,
            },
        )
        if hasattr(self.model, "to") and self.device:
            self.model.to(self.device)
        if hasattr(self.model, "eval"):
            self.model.eval()
        logger.info(
            "asr model ready",
            extra={
                "asr_provider": self.provider,
                "asr_model_id": self.model_id,
                "asr_revision": self.revision,
                "device": self.device,
            },
        )

    def _load_model(self, auto_model, torch):
        return auto_model.from_pretrained(
            self.model_id,
            revision=self.revision,
            trust_remote_code=self.trust_remote_code,
            low_cpu_mem_usage=False,
        )

    async def transcribe(self, audio_path: Path) -> str:
        if self.model is None:
            raise RuntimeError("ASR model is not loaded")

        result = await asyncio.to_thread(self.model.transcribe, str(audio_path))
        return _transcription_text(result)

    async def close(self) -> None:
        self.model = None


class GigaAMOnnxAsrBackend:
    provider = "gigaam_onnx"

    def __init__(
        self,
        model_id: str,
        revision: str,
        device: str,
    ) -> None:
        self.model_id = _gigaam_onnx_model_id(model_id, revision)
        self.revision = revision
        self.device = device
        self.model: Any | None = None

    async def load(self) -> None:
        logger.info(
            "asr onnx model loading",
            extra={
                "asr_provider": self.provider,
                "asr_model_id": self.model_id,
                "requested_device": self.device,
            },
        )

        import onnx_asr

        self.model = await _run_with_heartbeat(
            "asr onnx model",
            lambda: onnx_asr.load_model(self.model_id),
            extra={
                "asr_provider": self.provider,
                "asr_model_id": self.model_id,
                "requested_device": self.device,
            },
        )
        logger.info(
            "asr onnx model ready",
            extra={
                "asr_provider": self.provider,
                "asr_model_id": self.model_id,
            },
        )

    async def transcribe(self, audio_path: Path) -> str:
        if self.model is None:
            raise RuntimeError("ASR model is not loaded")

        result = await asyncio.to_thread(self.model.recognize, str(audio_path))
        return _transcription_text(result)

    async def close(self) -> None:
        self.model = None


def create_asr_backend(
    *,
    enabled: bool,
    provider: str,
    model_id: str,
    revision: str,
    trust_remote_code: bool,
    device: str,
) -> AsrBackend | None:
    if not enabled:
        return None

    normalized = provider.strip().lower()
    if normalized in {"", "none", "disabled"}:
        return None
    if normalized in {"gigaam_onnx", "gigaam-onnx", "onnx", "onnx-gigaam"}:
        return GigaAMOnnxAsrBackend(
            model_id=model_id,
            revision=revision,
            device=device,
        )
    if normalized != "gigaam":
        raise ValueError(f"unsupported ASR provider: {provider}")

    return GigaAMAsrBackend(
        model_id=model_id,
        revision=revision,
        trust_remote_code=trust_remote_code,
        device=device,
    )


def _gigaam_onnx_model_id(model_id: str, revision: str) -> str:
    normalized_model = model_id.strip()
    if normalized_model and normalized_model != "ai-sage/GigaAM-v3":
        return normalized_model

    normalized_revision = revision.strip().lower().replace("_", "-")
    if normalized_revision in {"", "e2e-rnnt"}:
        return "gigaam-v3-e2e-rnnt"
    if normalized_revision in {"e2e-ctc", "rnnt", "ctc"}:
        return f"gigaam-v3-{normalized_revision}"
    return "gigaam-v3-e2e-rnnt"


def _transcription_text(result: Any) -> str:
    if result is None:
        return ""
    if isinstance(result, str):
        return result.strip()
    if isinstance(result, dict):
        for key in ("text", "transcription", "transcript"):
            value = result.get(key)
            if isinstance(value, str):
                return value.strip()
        return str(result).strip()
    if isinstance(result, (list, tuple)):
        parts = [_transcription_text(item) for item in result]
        return " ".join(part for part in parts if part).strip()
    return str(result).strip()


def _resolve_device(torch, device: str) -> str:
    if device and device != "auto":
        return device
    if torch.cuda.is_available():
        return "cuda:0"
    if hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
        return "mps"
    if hasattr(torch, "xpu") and torch.xpu.is_available():
        return "xpu"
    return "cpu"


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
