from __future__ import annotations

from dataclasses import dataclass
from max_adapter.db.repository import AccountRecord


@dataclass(slots=True)
class LoginResult:
    status: str
    account: AccountRecord | None = None
    password_track_id: str = ""
    error: str = ""


@dataclass(slots=True)
class SentMessage:
    chat_id: str
    message_id: str
