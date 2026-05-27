from __future__ import annotations

from pathlib import Path

import grpc
import pytest

from voiceadapter.v1 import voice_adapter_pb2 as pb2

from voice_adapter.config import Config
from voice_adapter.services.backend import AudioResult
from voice_adapter.services.grpc_service import VoiceAdapterServicer


@pytest.mark.asyncio
async def test_synthesize_rejects_missing_metadata(tmp_path):
    service = VoiceAdapterServicer(_config(tmp_path), FakeBackend())

    with pytest.raises(AbortError) as exc:
        await _collect(service.Synthesize(_requests(), FakeContext()))

    assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT
    assert "metadata" in exc.value.details()


@pytest.mark.asyncio
async def test_synthesize_rejects_chunks_before_metadata(tmp_path):
    service = VoiceAdapterServicer(_config(tmp_path), FakeBackend())

    with pytest.raises(AbortError) as exc:
        await _collect(
            service.Synthesize(
                _requests(pb2.SynthesizeRequest(reference_audio_chunk=pb2.ReferenceAudioChunk(data=b"x"))),
                FakeContext(),
            )
        )

    assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT


@pytest.mark.asyncio
async def test_synthesize_rejects_metadata_twice(tmp_path):
    service = VoiceAdapterServicer(_config(tmp_path), FakeBackend())
    metadata = _metadata(mode=pb2.SYNTHESIS_MODE_AUTO_VOICE)

    with pytest.raises(AbortError) as exc:
        await _collect(
            service.Synthesize(
                _requests(
                    pb2.SynthesizeRequest(metadata=metadata),
                    pb2.SynthesizeRequest(metadata=metadata),
                ),
                FakeContext(),
            )
        )

    assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT
    assert "twice" in exc.value.details()


@pytest.mark.asyncio
async def test_synthesize_rejects_clone_without_reference_audio(tmp_path):
    service = VoiceAdapterServicer(_config(tmp_path), FakeBackend())

    with pytest.raises(AbortError) as exc:
        await _collect(
            service.Synthesize(
                _requests(pb2.SynthesizeRequest(metadata=_metadata())),
                FakeContext(),
            )
        )

    assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT
    assert "reference audio" in exc.value.details()


@pytest.mark.asyncio
async def test_synthesize_streams_events_and_cleans_temp_files(tmp_path):
    backend = FakeBackend()
    service = VoiceAdapterServicer(_config(tmp_path), backend)

    events = await _collect(
        service.Synthesize(
            _requests(
                pb2.SynthesizeRequest(metadata=_metadata(filename="../voice.mp3")),
                pb2.SynthesizeRequest(reference_audio_chunk=pb2.ReferenceAudioChunk(data=b"abc")),
            ),
            FakeContext(),
        )
    )

    assert [event.WhichOneof("event") for event in events] == [
        "progress",
        "progress",
        "progress",
        "audio_chunk",
        "completed",
    ]
    assert backend.request.reference_audio.filename == "../voice.mp3"
    assert backend.reference_path.name == "voice.mp3"
    assert backend.reference_bytes == b"abc"
    assert not backend.reference_path.exists()
    assert not any(Path(tmp_path).iterdir())


@pytest.mark.asyncio
async def test_synthesize_maps_backend_errors(tmp_path):
    service = VoiceAdapterServicer(_config(tmp_path), FakeBackend(error=RuntimeError("model failed")))

    with pytest.raises(AbortError) as exc:
        await _collect(
            service.Synthesize(
                _requests(pb2.SynthesizeRequest(metadata=_metadata(mode=pb2.SYNTHESIS_MODE_AUTO_VOICE))),
                FakeContext(),
            )
        )

    assert exc.value.code() == grpc.StatusCode.FAILED_PRECONDITION
    assert "model failed" in exc.value.details()


@pytest.mark.asyncio
async def test_transcribe_writes_audio_and_cleans_temp_files(tmp_path):
    backend = FakeBackend(transcript="recognized text")
    service = VoiceAdapterServicer(_config(tmp_path), backend)

    response = await service.Transcribe(
        pb2.TranscribeRequest(
            audio=pb2.AudioDescriptor(filename="../voice.wav", mime_type="audio/wav", extension=".wav", size=3),
            audio_data=b"abc",
        ),
        FakeContext(),
    )

    assert response.text == "recognized text"
    assert backend.transcribe_path.name == "voice.wav"
    assert backend.transcribe_bytes == b"abc"
    assert not backend.transcribe_path.exists()
    assert not any(Path(tmp_path).iterdir())


@pytest.mark.asyncio
async def test_transcribe_rejects_missing_audio(tmp_path):
    service = VoiceAdapterServicer(_config(tmp_path), FakeBackend())

    with pytest.raises(AbortError) as exc:
        await service.Transcribe(pb2.TranscribeRequest(), FakeContext())

    assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT
    assert "audio data" in exc.value.details()


class FakeBackend:
    model_provider = "fake"
    model_id = "fake-model"
    device = "cpu"

    def __init__(self, error: Exception | None = None, transcript: str = "transcribed") -> None:
        self.error = error
        self.transcript = transcript
        self.request = None
        self.reference_path = None
        self.reference_bytes = b""
        self.transcribe_path = None
        self.transcribe_bytes = b""

    async def load(self) -> None:
        return None

    async def synthesize(self, request, progress=None):
        if self.error:
            raise self.error
        self.request = request
        if request.reference_audio:
            self.reference_path = request.reference_audio.path
            assert self.reference_path.exists()
            self.reference_bytes = self.reference_path.read_bytes()
        return AudioResult(
            data=b"abcdef",
            format="WAV_PCM16",
            mime_type="audio/wav",
            extension=".wav",
            sample_rate_hz=24_000,
            duration_ms=100,
            waveform_peaks=b"\x01" * 80,
        )

    async def transcribe(self, audio_path: Path) -> str:
        self.transcribe_path = audio_path
        assert self.transcribe_path.exists()
        self.transcribe_bytes = self.transcribe_path.read_bytes()
        return self.transcript

    async def close(self) -> None:
        return None


class FakeContext:
    def invocation_metadata(self):
        return ()

    async def abort(self, code, details):
        raise AbortError(code, details)


class AbortError(grpc.RpcError):
    def __init__(self, code, details) -> None:
        super().__init__()
        self._code = code
        self._details = details

    def code(self):
        return self._code

    def details(self):
        return self._details


def _config(tmp_path):
    return Config(
        grpc_bind="127.0.0.1:0",
        grpc_token="",
        temp_dir=str(tmp_path),
        model_provider="fake",
        model_id="fake-model",
        device="cpu",
        dtype="float32",
        load_asr=False,
        asr_provider="none",
        asr_model_id="",
        asr_revision="",
        asr_trust_remote_code=False,
        concurrency=1,
    )


def _metadata(mode=pb2.SYNTHESIS_MODE_VOICE_CLONE, filename="ref.wav"):
    return pb2.SynthesisMetadata(
        request_id="request-1",
        text="Hello",
        mode=mode,
        instruct="female",
        reference_audio=pb2.AudioDescriptor(
            filename=filename,
            mime_type="audio/wav",
            extension=".wav",
            size=3,
        ),
    )


async def _requests(*items):
    for item in items:
        yield item


async def _collect(events):
    return [event async for event in events]
