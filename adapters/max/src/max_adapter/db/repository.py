from __future__ import annotations

import json
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta
from pathlib import Path
from typing import Any
from uuid import UUID, uuid4

import asyncpg


def now_utc() -> datetime:
    return datetime.now(UTC)


def rfc3339(value: datetime | None) -> str:
    if value is None:
        return ""
    return value.astimezone(UTC).isoformat().replace("+00:00", "Z")


@dataclass(slots=True)
class AccountRecord:
    id: UUID
    phone: str
    label: str
    encrypted_token: str
    device_id: UUID
    max_user_id: str
    status: str
    last_error: str
    created_at: datetime
    updated_at: datetime


@dataclass(slots=True)
class LoginAttemptRecord:
    id: UUID
    phone: str
    label: str
    encrypted_temp_token: str
    device_id: UUID
    password_track_id: str
    status: str
    error: str
    expires_at: datetime
    created_at: datetime
    updated_at: datetime


@dataclass(slots=True)
class SendRequestRecord:
    client_request_id: str
    account_id: UUID
    chat_id: str
    message_id: str
    status: str
    error: str


@dataclass(slots=True)
class EventRecord:
    seq: int
    type: str
    account_id: UUID | None
    chat_id: str
    message_id: str
    sender_id: str
    text: str
    reply_to_message_id: str
    payload_json: dict[str, Any]
    occurred_at: datetime


class Repository:
    def __init__(self, pool: asyncpg.Pool) -> None:
        self.pool = pool

    async def migrate(self, migrations_dir: Path) -> None:
        async with self.pool.acquire() as conn:
            async with conn.transaction():
                for path in sorted(migrations_dir.glob("*.sql")):
                    await conn.execute(path.read_text())

    async def create_login_attempt(
        self,
        *,
        phone: str,
        label: str,
        encrypted_temp_token: str,
        device_id: UUID,
        ttl_seconds: int,
    ) -> LoginAttemptRecord:
        ts = now_utc()
        row = await self.pool.fetchrow(
            """
            INSERT INTO login_attempts (
                id, phone, label, encrypted_temp_token, device_id, status,
                expires_at, created_at, updated_at
            )
            VALUES ($1, $2, $3, $4, $5, 'CODE_REQUIRED', $6, $7, $7)
            RETURNING *
            """,
            uuid4(),
            phone,
            label,
            encrypted_temp_token,
            device_id,
            ts + timedelta(seconds=ttl_seconds),
            ts,
        )
        return self._login_attempt(row)

    async def get_login_attempt(self, attempt_id: UUID) -> LoginAttemptRecord | None:
        row = await self.pool.fetchrow("SELECT * FROM login_attempts WHERE id = $1", attempt_id)
        return self._login_attempt(row) if row else None

    async def update_login_attempt(
        self,
        attempt_id: UUID,
        *,
        status: str,
        error: str = "",
        password_track_id: str = "",
    ) -> LoginAttemptRecord:
        row = await self.pool.fetchrow(
            """
            UPDATE login_attempts
            SET status = $2,
                error = $3,
                password_track_id = CASE WHEN $4 = '' THEN password_track_id ELSE $4 END,
                updated_at = $5
            WHERE id = $1
            RETURNING *
            """,
            attempt_id,
            status,
            error,
            password_track_id,
            now_utc(),
        )
        return self._login_attempt(row)

    async def upsert_account(
        self,
        *,
        phone: str,
        label: str,
        encrypted_token: str,
        device_id: UUID,
        max_user_id: str,
        status: str,
        last_error: str = "",
    ) -> AccountRecord:
        ts = now_utc()
        row = await self.pool.fetchrow(
            """
            INSERT INTO accounts (
                id, phone, label, encrypted_token, device_id, max_user_id,
                status, last_error, created_at, updated_at
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
            ON CONFLICT (phone) DO UPDATE
            SET label = EXCLUDED.label,
                encrypted_token = EXCLUDED.encrypted_token,
                device_id = EXCLUDED.device_id,
                max_user_id = EXCLUDED.max_user_id,
                status = EXCLUDED.status,
                last_error = EXCLUDED.last_error,
                updated_at = EXCLUDED.updated_at
            RETURNING *
            """,
            uuid4(),
            phone,
            label,
            encrypted_token,
            device_id,
            max_user_id,
            status,
            last_error,
            ts,
        )
        return self._account(row)

    async def list_accounts(self) -> list[AccountRecord]:
        rows = await self.pool.fetch("SELECT * FROM accounts ORDER BY created_at DESC")
        return [self._account(row) for row in rows]

    async def get_account(self, account_id: UUID) -> AccountRecord | None:
        row = await self.pool.fetchrow("SELECT * FROM accounts WHERE id = $1", account_id)
        return self._account(row) if row else None

    async def delete_account(self, account_id: UUID) -> None:
        await self.pool.execute("DELETE FROM accounts WHERE id = $1", account_id)

    async def update_account_status(
        self,
        account_id: UUID,
        *,
        status: str,
        last_error: str = "",
        max_user_id: str | None = None,
    ) -> AccountRecord | None:
        row = await self.pool.fetchrow(
            """
            UPDATE accounts
            SET status = $2,
                last_error = $3,
                max_user_id = COALESCE($4, max_user_id),
                updated_at = $5
            WHERE id = $1
            RETURNING *
            """,
            account_id,
            status,
            last_error,
            max_user_id,
            now_utc(),
        )
        return self._account(row) if row else None

    async def get_send_request(self, client_request_id: str) -> SendRequestRecord | None:
        row = await self.pool.fetchrow(
            "SELECT * FROM send_requests WHERE client_request_id = $1",
            client_request_id,
        )
        return self._send_request(row) if row else None

    async def save_send_request(
        self,
        *,
        client_request_id: str,
        account_id: UUID,
        chat_id: str,
        message_id: str,
        status: str,
        error: str = "",
    ) -> SendRequestRecord:
        ts = now_utc()
        row = await self.pool.fetchrow(
            """
            INSERT INTO send_requests (
                client_request_id, account_id, chat_id, message_id,
                status, error, created_at, updated_at
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
            ON CONFLICT (client_request_id) DO UPDATE
            SET chat_id = EXCLUDED.chat_id,
                message_id = EXCLUDED.message_id,
                status = EXCLUDED.status,
                error = EXCLUDED.error,
                updated_at = EXCLUDED.updated_at
            RETURNING *
            """,
            client_request_id,
            account_id,
            chat_id,
            message_id,
            status,
            error,
            ts,
        )
        return self._send_request(row)

    async def append_event(
        self,
        *,
        type_: str,
        account_id: UUID | None,
        chat_id: str = "",
        message_id: str = "",
        sender_id: str = "",
        text: str = "",
        reply_to_message_id: str = "",
        payload: dict[str, Any] | None = None,
    ) -> EventRecord:
        row = await self.pool.fetchrow(
            """
            INSERT INTO event_outbox (
                type, account_id, chat_id, message_id, sender_id,
                text, reply_to_message_id, payload_json, occurred_at
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
            RETURNING *
            """,
            type_,
            account_id,
            chat_id,
            message_id,
            sender_id,
            text,
            reply_to_message_id,
            json.dumps(payload or {}),
            now_utc(),
        )
        return self._event(row)

    async def list_events_after(self, seq: int, limit: int = 100) -> list[EventRecord]:
        rows = await self.pool.fetch(
            """
            SELECT * FROM event_outbox
            WHERE seq > $1
            ORDER BY seq ASC
            LIMIT $2
            """,
            seq,
            limit,
        )
        return [self._event(row) for row in rows]

    async def ack_events(self, consumer: str, up_to_seq: int) -> int:
        ts = now_utc()
        row = await self.pool.fetchrow(
            """
            INSERT INTO event_offsets (consumer, acknowledged_seq, updated_at)
            VALUES ($1, $2, $3)
            ON CONFLICT (consumer) DO UPDATE
            SET acknowledged_seq = GREATEST(event_offsets.acknowledged_seq, EXCLUDED.acknowledged_seq),
                updated_at = EXCLUDED.updated_at
            RETURNING acknowledged_seq
            """,
            consumer,
            up_to_seq,
            ts,
        )
        return int(row["acknowledged_seq"])

    async def get_ack(self, consumer: str) -> int:
        row = await self.pool.fetchrow(
            "SELECT acknowledged_seq FROM event_offsets WHERE consumer = $1",
            consumer,
        )
        return int(row["acknowledged_seq"]) if row else 0

    @staticmethod
    def _account(row: asyncpg.Record) -> AccountRecord:
        return AccountRecord(
            id=row["id"],
            phone=row["phone"],
            label=row["label"],
            encrypted_token=row["encrypted_token"],
            device_id=row["device_id"],
            max_user_id=row["max_user_id"],
            status=row["status"],
            last_error=row["last_error"],
            created_at=row["created_at"],
            updated_at=row["updated_at"],
        )

    @staticmethod
    def _login_attempt(row: asyncpg.Record) -> LoginAttemptRecord:
        return LoginAttemptRecord(
            id=row["id"],
            phone=row["phone"],
            label=row["label"],
            encrypted_temp_token=row["encrypted_temp_token"],
            device_id=row["device_id"],
            password_track_id=row["password_track_id"],
            status=row["status"],
            error=row["error"],
            expires_at=row["expires_at"],
            created_at=row["created_at"],
            updated_at=row["updated_at"],
        )

    @staticmethod
    def _send_request(row: asyncpg.Record) -> SendRequestRecord:
        return SendRequestRecord(
            client_request_id=row["client_request_id"],
            account_id=row["account_id"],
            chat_id=row["chat_id"],
            message_id=row["message_id"],
            status=row["status"],
            error=row["error"],
        )

    @staticmethod
    def _event(row: asyncpg.Record) -> EventRecord:
        payload = row["payload_json"]
        if isinstance(payload, str):
            payload = json.loads(payload)
        return EventRecord(
            seq=row["seq"],
            type=row["type"],
            account_id=row["account_id"],
            chat_id=row["chat_id"],
            message_id=row["message_id"],
            sender_id=row["sender_id"],
            text=row["text"],
            reply_to_message_id=row["reply_to_message_id"],
            payload_json=payload or {},
            occurred_at=row["occurred_at"],
        )
