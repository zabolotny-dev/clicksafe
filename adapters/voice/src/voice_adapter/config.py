from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    grpc_bind: str
    grpc_token: str
    temp_dir: str
    model_provider: str
    model_id: str
    device: str
    dtype: str
    load_asr: bool
    asr_provider: str
    asr_model_id: str
    asr_revision: str
    asr_trust_remote_code: bool
    concurrency: int


def load_config() -> Config:
    return Config(
        grpc_bind=os.getenv("VOICE_ADAPTER_GRPC_BIND", "0.0.0.0:9091"),
        grpc_token=os.getenv("VOICE_ADAPTER_GRPC_TOKEN", ""),
        temp_dir=os.getenv("VOICE_ADAPTER_TEMP_DIR", "/tmp/clicksafe-voice-adapter"),
        model_provider=os.getenv("VOICE_ADAPTER_MODEL_PROVIDER", "omnivoice"),
        model_id=os.getenv("VOICE_ADAPTER_MODEL_ID", "k2-fsa/OmniVoice"),
        device=os.getenv("VOICE_ADAPTER_DEVICE", "auto"),
        dtype=os.getenv("VOICE_ADAPTER_DTYPE", "auto"),
        load_asr=_bool_env("VOICE_ADAPTER_LOAD_ASR", True),
        asr_provider=os.getenv("VOICE_ADAPTER_ASR_PROVIDER", "gigaam_onnx"),
        asr_model_id=os.getenv(
            "VOICE_ADAPTER_ASR_MODEL_ID",
            os.getenv("VOICE_ADAPTER_ASR_MODEL", "gigaam-v3-e2e-rnnt"),
        ),
        asr_revision=os.getenv("VOICE_ADAPTER_ASR_REVISION", ""),
        asr_trust_remote_code=_bool_env("VOICE_ADAPTER_ASR_TRUST_REMOTE_CODE", False),
        concurrency=max(1, _int_env("VOICE_ADAPTER_CONCURRENCY", 1)),
    )


def _bool_env(name: str, default: bool) -> bool:
    value = os.getenv(name)
    if value is None:
        return default
    return value.strip().lower() in {"1", "true", "yes", "y", "on"}


def _int_env(name: str, default: int) -> int:
    value = os.getenv(name)
    if value is None:
        return default
    try:
        return int(value)
    except ValueError:
        return default
