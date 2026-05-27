from __future__ import annotations

import sys
import tempfile
from pathlib import Path


def pytest_sessionstart(session):
    root = Path(__file__).resolve().parents[3]
    src = root / "adapters" / "voice" / "src"
    generated = Path(tempfile.gettempdir()) / "clicksafe-voice-adapter-py"
    proto_root = root / "adapters" / "voice" / "proto"
    proto = proto_root / "voiceadapter" / "v1" / "voice_adapter.proto"

    generated.mkdir(parents=True, exist_ok=True)
    sys.path.insert(0, str(generated))
    sys.path.insert(0, str(src))

    from grpc_tools import protoc

    rc = protoc.main(
        [
            "grpc_tools.protoc",
            "-I",
            str(proto_root),
            "--python_out",
            str(generated),
            "--grpc_python_out",
            str(generated),
            str(proto),
        ]
    )
    if rc != 0:
        raise RuntimeError(f"protoc failed with exit code {rc}")
