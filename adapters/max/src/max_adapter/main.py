from __future__ import annotations

import asyncio
import logging
import signal
from pathlib import Path

import asyncpg
import grpc

from maxadapter.v1 import max_adapter_pb2_grpc as pb2_grpc

from max_adapter.config import load_config
from max_adapter.db.repository import Repository
from max_adapter.logsetup import configure_logging
from max_adapter.security.crypto import SecretBox
from max_adapter.services.grpc_service import MaxAdapterServicer
from max_adapter.services.pymax_manager import PyMaxManager

logger = logging.getLogger(__name__)


async def serve() -> None:
    configure_logging(service="max-adapter", level=logging.INFO)
    cfg = load_config()

    logger.info("startup", extra={"status": "initializing MAX adapter"})

    pool = await asyncpg.create_pool(dsn=cfg.db_dsn, min_size=1, max_size=10)
    repo = Repository(pool)
    await repo.migrate(Path(__file__).resolve().parents[2] / "migrations")

    secrets = SecretBox(cfg.secret_key)
    manager = PyMaxManager(repo, secrets, cfg.reconnect_delay_seconds)

    server = grpc.aio.server()
    pb2_grpc.add_MaxAdapterServiceServicer_to_server(
        MaxAdapterServicer(cfg, repo, manager),
        server,
    )
    server.add_insecure_port(cfg.grpc_bind)

    await server.start()
    logger.info("startup", extra={"status": "max adapter grpc started", "host": cfg.grpc_bind})

    await manager.restore_active_accounts()

    stop_event = asyncio.Event()
    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, stop_event.set)

    await stop_event.wait()
    logger.info("shutdown", extra={"status": "max adapter stopping"})
    await manager.stop_all()
    await server.stop(grace=1.0)
    await pool.close()


def main() -> None:
    asyncio.run(serve())


if __name__ == "__main__":
    main()
