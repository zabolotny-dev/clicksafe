from __future__ import annotations

import json
import logging
import sys
import traceback
from datetime import UTC, datetime
from typing import Any


_STANDARD_RECORD_FIELDS = frozenset(logging.makeLogRecord({}).__dict__)


class JSONFormatter(logging.Formatter):
    def __init__(self, *, service: str) -> None:
        super().__init__()
        self.service = service

    def format(self, record: logging.LogRecord) -> str:
        data: dict[str, Any] = {
            "time": datetime.fromtimestamp(record.created, UTC).isoformat(timespec="microseconds").replace("+00:00", "Z"),
            "level": record.levelname,
            "file": f"{record.filename}:{record.lineno}",
            "msg": record.getMessage(),
            "service": self.service,
        }

        for key, value in record.__dict__.items():
            if key in _STANDARD_RECORD_FIELDS or key.startswith("_"):
                continue
            data[key] = _json_safe(value)

        if record.exc_info:
            _, exc, _ = record.exc_info
            data["err"] = str(exc)
            data["trace"] = "".join(traceback.format_exception(*record.exc_info)).rstrip()

        return json.dumps(data, ensure_ascii=False, separators=(",", ":"))


def configure_logging(*, service: str, level: int = logging.INFO) -> None:
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(JSONFormatter(service=service))

    root = logging.getLogger()
    root.handlers.clear()
    root.setLevel(level)
    root.addHandler(handler)


def _json_safe(value: Any) -> Any:
    try:
        json.dumps(value)
        return value
    except TypeError:
        return str(value)
