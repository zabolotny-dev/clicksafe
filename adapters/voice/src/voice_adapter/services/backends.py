from __future__ import annotations

from voice_adapter.config import Config
from voice_adapter.services.asr import create_asr_backend
from voice_adapter.services.backend import VoiceBackend
from voice_adapter.services.omnivoice_backend import OmniVoiceBackend


def create_backend(cfg: Config) -> VoiceBackend:
    provider = cfg.model_provider.strip().lower()
    if provider != "omnivoice":
        raise ValueError(f"unsupported voice model provider: {cfg.model_provider}")

    asr_backend = None
    internal_asr_model = ""
    internal_load_asr = False
    if cfg.load_asr and cfg.asr_provider.strip().lower() == "omnivoice":
        internal_load_asr = True
        internal_asr_model = cfg.asr_model_id
    else:
        asr_backend = create_asr_backend(
            enabled=cfg.load_asr,
            provider=cfg.asr_provider,
            model_id=cfg.asr_model_id,
            revision=cfg.asr_revision,
            trust_remote_code=cfg.asr_trust_remote_code,
            device=cfg.device,
        )

    return OmniVoiceBackend(
        model_id=cfg.model_id,
        device=cfg.device,
        dtype=cfg.dtype,
        load_asr=internal_load_asr,
        asr_model=internal_asr_model,
        asr_backend=asr_backend,
    )
