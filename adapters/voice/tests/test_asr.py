from __future__ import annotations

import sys
import types
from pathlib import Path

import pytest

from voice_adapter.services.asr import GigaAMAsrBackend, GigaAMOnnxAsrBackend, create_asr_backend


def test_create_gigaam_asr_backend():
    backend = create_asr_backend(
        enabled=True,
        provider="gigaam",
        model_id="ai-sage/GigaAM-v3",
        revision="e2e_rnnt",
        trust_remote_code=True,
        device="cpu",
    )

    assert backend is not None
    assert backend.provider == "gigaam"
    assert backend.model_id == "ai-sage/GigaAM-v3"


def test_create_gigaam_onnx_asr_backend_maps_hf_revision():
    backend = create_asr_backend(
        enabled=True,
        provider="gigaam_onnx",
        model_id="ai-sage/GigaAM-v3",
        revision="e2e_ctc",
        trust_remote_code=False,
        device="cpu",
    )

    assert isinstance(backend, GigaAMOnnxAsrBackend)
    assert backend.provider == "gigaam_onnx"
    assert backend.model_id == "gigaam-v3-e2e-ctc"


def test_create_asr_backend_disabled():
    backend = create_asr_backend(
        enabled=False,
        provider="gigaam",
        model_id="ai-sage/GigaAM-v3",
        revision="e2e_rnnt",
        trust_remote_code=True,
        device="cpu",
    )

    assert backend is None


def test_gigaam_load_model_disables_low_cpu_mem_usage():
    class FakeAutoModel:
        kwargs = None

        @classmethod
        def from_pretrained(cls, *_, **kwargs):
            cls.kwargs = kwargs
            return object()

    backend = GigaAMAsrBackend(
        model_id="ai-sage/GigaAM-v3",
        revision="e2e_rnnt",
        trust_remote_code=True,
        device="cuda:0",
    )

    backend._load_model(FakeAutoModel, object())

    assert FakeAutoModel.kwargs["low_cpu_mem_usage"] is False


@pytest.mark.asyncio
async def test_gigaam_onnx_loads_once_and_recognizes(monkeypatch, tmp_path):
    calls = []

    class FakeModel:
        def recognize(self, audio_path: str) -> dict[str, str]:
            return {"text": f"recognized {Path(audio_path).name}"}

    def load_model(model_id: str) -> FakeModel:
        calls.append(model_id)
        return FakeModel()

    monkeypatch.setitem(sys.modules, "onnx_asr", types.SimpleNamespace(load_model=load_model))

    backend = GigaAMOnnxAsrBackend(
        model_id="ai-sage/GigaAM-v3",
        revision="e2e_rnnt",
        device="cpu",
    )
    audio_path = tmp_path / "ref.wav"
    audio_path.write_bytes(b"")

    await backend.load()

    assert await backend.transcribe(audio_path) == "recognized ref.wav"
    assert calls == ["gigaam-v3-e2e-rnnt"]
