from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    grpc_bind: str
    grpc_token: str
    db_dsn: str
    secret_key: str
    temp_dir: str
    login_ttl_seconds: int
    reconnect_delay_seconds: float


def load_config() -> Config:
    db_dsn = os.getenv("MAX_ADAPTER_DATABASE_URL")
    if not db_dsn:
        user = os.getenv("MAX_ADAPTER_DB_USER", "postgres")
        password = os.getenv("MAX_ADAPTER_DB_PASSWORD", "secret")
        host = os.getenv("MAX_ADAPTER_DB_HOST", "localhost:5432")
        name = os.getenv("MAX_ADAPTER_DB_NAME", "clicksafe_max_adapter")
        db_dsn = f"postgresql://{user}:{password}@{host}/{name}"

    return Config(
        grpc_bind=os.getenv("MAX_ADAPTER_GRPC_BIND", "0.0.0.0:9090"),
        grpc_token=os.getenv("MAX_ADAPTER_GRPC_TOKEN", ""),
        db_dsn=db_dsn,
        secret_key=os.getenv("MAX_ADAPTER_SECRET_KEY", ""),
        temp_dir=os.getenv("MAX_ADAPTER_TEMP_DIR", "/tmp/clicksafe-max-adapter"),
        login_ttl_seconds=int(os.getenv("MAX_ADAPTER_LOGIN_TTL_SECONDS", "600")),
        reconnect_delay_seconds=float(os.getenv("MAX_ADAPTER_RECONNECT_DELAY_SECONDS", "2")),
    )
