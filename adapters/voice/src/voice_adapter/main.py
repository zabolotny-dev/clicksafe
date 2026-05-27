from __future__ import annotations

import asyncio
import logging
import signal

import grpc

from voiceadapter.v1 import voice_adapter_pb2_grpc as pb2_grpc

from voice_adapter.config import load_config
from voice_adapter.logsetup import configure_logging
from voice_adapter.services.audio import configure_audio_runtime
from voice_adapter.services.backends import create_backend
from voice_adapter.services.grpc_service import VoiceAdapterServicer

logger = logging.getLogger(__name__)


async def serve() -> None:
    configure_logging(service="voice-adapter", level=logging.INFO)
    cfg = load_config()

    logger.info("voice adapter initializing")
    ffmpeg_path = configure_audio_runtime()
    logger.info("audio runtime configured", extra={"ffmpeg": ffmpeg_path})

    backend = create_backend(cfg)
    await backend.load()

    server = grpc.aio.server()
    pb2_grpc.add_VoiceAdapterServiceServicer_to_server(
        VoiceAdapterServicer(cfg, backend),
        server,
    )
    server.add_insecure_port(cfg.grpc_bind)
    await server.start()

    logger.info(
        "voice adapter grpc started",
        extra={
            "host": cfg.grpc_bind,
            "model_provider": cfg.model_provider,
            "model_id": cfg.model_id,
            "device": backend.device,
        },
    )

    stop = asyncio.Event()
    loop = asyncio.get_running_loop()
    for sig in (signal.SIGTERM, signal.SIGINT):
        loop.add_signal_handler(sig, stop.set)

    await stop.wait()

    logger.info("shutdown", extra={"status": "voice adapter stopping"})
    await server.stop(grace=5)
    await backend.close()


def main() -> None:
    asyncio.run(serve())


if __name__ == "__main__":
    main()
