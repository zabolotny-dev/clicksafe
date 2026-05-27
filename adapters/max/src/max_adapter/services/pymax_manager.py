from __future__ import annotations

import asyncio
import json
import logging
from pathlib import Path
from typing import Any
from uuid import UUID

from pymax import SocketMaxClient
from pymax.files import Audio, File, Photo, Video
from pymax.payloads import UserAgentPayload
from pymax.types import Message

from max_adapter.db.repository import AccountRecord, Repository
from max_adapter.security.crypto import SecretBox
from .pymax_client import ClickSafeSocketMaxClient
from .pymax_types import LoginResult, SentMessage

logger = logging.getLogger(__name__)

PyMaxAttachment = Photo | File | Video | Audio


class PyMaxManager:
    def __init__(self, repo: Repository, secrets: SecretBox, reconnect_delay: float) -> None:
        self.repo = repo
        self.secrets = secrets
        self.reconnect_delay = reconnect_delay
        self._clients: dict[UUID, ClickSafeSocketMaxClient] = {}
        self._tasks: dict[UUID, asyncio.Task[Any]] = {}
        self._lock = asyncio.Lock()

    async def restore_active_accounts(self) -> None:
        for account in await self.repo.list_accounts():
            if account.status in {"ACTIVE", "CONNECTED"}:
                try:
                    await self.start_account(account.id)
                except Exception:
                    logger.exception("failed to restore account %s", account.id)

    async def begin_login(self, phone: str, label: str, ttl_seconds: int) -> Any:
        client = self._new_client(phone=phone)
        await client.connect(client.user_agent)
        try:
            temp_token = await client.request_code(phone)
            return await self.repo.create_login_attempt(
                phone=phone,
                label=label,
                encrypted_temp_token=self.secrets.encrypt(temp_token),
                device_id=client.device_id,
                ttl_seconds=ttl_seconds,
            )
        finally:
            await self._cleanup_client(client)

    async def confirm_login(
        self,
        *,
        attempt_id: UUID,
        code: str,
        password: str = "",
    ) -> LoginResult:
        attempt = await self.repo.get_login_attempt(attempt_id)
        if attempt is None:
            return LoginResult(status="FAILED", error="login attempt not found")
        if attempt.status == "EXPIRED":
            return LoginResult(status="EXPIRED", error="login attempt expired")

        temp_token = self.secrets.decrypt(attempt.encrypted_temp_token)
        client = self._new_client(phone=attempt.phone, device_id=attempt.device_id)
        await client.connect(client.user_agent)
        try:
            response = await client._send_code(code, temp_token)
            return await self._finish_login_response(
                attempt=attempt,
                response=response,
                password=password,
                client=client,
            )
        except Exception as exc:
            await self.repo.update_login_attempt(attempt.id, status="FAILED", error=str(exc))
            return LoginResult(status="FAILED", error=str(exc))
        finally:
            await self._cleanup_client(client)

    async def confirm_password(self, *, attempt_id: UUID, password: str) -> LoginResult:
        attempt = await self.repo.get_login_attempt(attempt_id)
        if attempt is None:
            return LoginResult(status="FAILED", error="login attempt not found")
        if not attempt.password_track_id:
            return LoginResult(status="FAILED", error="password challenge not found")

        client = self._new_client(phone=attempt.phone, device_id=attempt.device_id)
        await client.connect(client.user_agent)
        try:
            token_attrs = await client._check_password(password, attempt.password_track_id)
            if not token_attrs:
                await self.repo.update_login_attempt(
                    attempt.id,
                    status="PASSWORD_REQUIRED",
                    error="incorrect password",
                )
                return LoginResult(status="PASSWORD_REQUIRED", error="incorrect password")

            token = token_attrs.get("LOGIN", {}).get("token")
            if not token:
                raise ValueError("password response did not contain login token")

            account = await self._save_account_from_token(attempt, token, client)
            await self.repo.update_login_attempt(attempt.id, status="COMPLETED")
            return LoginResult(status="COMPLETED", account=account)
        except Exception as exc:
            await self.repo.update_login_attempt(attempt.id, status="FAILED", error=str(exc))
            return LoginResult(status="FAILED", error=str(exc))
        finally:
            await self._cleanup_client(client)

    async def start_account(self, account_id: UUID) -> AccountRecord:
        async with self._lock:
            existing = self._clients.get(account_id)
            if existing and existing.is_connected:
                account = await self.repo.update_account_status(account_id, status="CONNECTED")
                if account is None:
                    raise ValueError("account not found")
                return account

            account = await self.repo.get_account(account_id)
            if account is None:
                raise ValueError("account not found")

            token = self.secrets.decrypt(account.encrypted_token)
            client = self._new_client(
                phone=account.phone,
                token=token,
                device_id=account.device_id,
                reconnect=True,
            )
            self._register_handlers(account.id, client)
            task = asyncio.create_task(client.start(), name=f"pymax-{account_id}")
            self._clients[account.id] = client
            self._tasks[account.id] = task

        try:
            await self._wait_connected(client)
            max_user_id = str(client.me.id) if client.me else account.max_user_id
            updated = await self.repo.update_account_status(
                account_id,
                status="CONNECTED",
                max_user_id=max_user_id,
            )
            await self.repo.append_event(type_="ACCOUNT_CONNECTED", account_id=account_id)
            if updated is None:
                raise ValueError("account not found")
            return updated
        except Exception as exc:
            await self.repo.update_account_status(account_id, status="ERROR", last_error=str(exc))
            raise

    async def stop_account(self, account_id: UUID) -> AccountRecord:
        async with self._lock:
            client = self._clients.pop(account_id, None)
            task = self._tasks.pop(account_id, None)

        if client:
            await client.close()
        if task:
            try:
                await asyncio.wait_for(task, timeout=5)
            except TimeoutError:
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass
            except asyncio.CancelledError:
                pass
            except Exception:
                logger.debug("pymax task raised while stopping", exc_info=True)
        elif client:
            await self._cleanup_client(client)

        updated = await self.repo.update_account_status(account_id, status="DISCONNECTED")
        await self.repo.append_event(type_="ACCOUNT_DISCONNECTED", account_id=account_id)
        if updated is None:
            raise ValueError("account not found")
        return updated

    async def delete_account(self, account_id: UUID) -> None:
        await self.stop_account(account_id)
        await self.repo.delete_account(account_id)

    async def stop_all(self) -> None:
        async with self._lock:
            clients = list(self._clients.items())
            tasks = list(self._tasks.items())
            self._clients.clear()
            self._tasks.clear()

        for account_id, client in clients:
            try:
                await client.close()
            except Exception:
                logger.debug("failed to close client %s during shutdown", account_id, exc_info=True)

        for account_id, task in tasks:
            try:
                await asyncio.wait_for(task, timeout=5)
            except asyncio.TimeoutError:
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass
            except Exception:
                pass

    async def send_message(
        self,
        *,
        account_id: UUID,
        recipient_kind: str,
        recipient_value: str,
        text: str,
        notify: bool,
        reply_to: int | None,
        attachments: list[tuple[str, Path]],
    ) -> SentMessage:
        client = await self._ensure_started(account_id)
        chat_id = await self._resolve_chat_id(client, recipient_kind, recipient_value)
        pymax_attachments = [(kind, self._build_attachment(kind, path)) for kind, path in attachments]
        attachment_batches = self._split_attachment_batches(pymax_attachments) or [[]]
        send_text_before_audio = bool(text.strip()) and self._has_audio_attachments(pymax_attachments)

        if len(attachment_batches) > 1:
            logger.info(
                "Splitting MAX message into %d sends because MAX rejects mixed/multiple attachment sends",
                len(attachment_batches),
            )

        first_message: Message | None = None
        if send_text_before_audio:
            logger.info("Sending text before audio attachment because MAX voice payloads ignore captions")
            client, first_message = await self._send_message_batch(
                account_id=account_id,
                client=client,
                chat_id=chat_id,
                text=text,
                notify=notify,
                attachments=None,
                reply_to=reply_to,
            )
            if first_message is None:
                raise ValueError("PyMax returned no message")

        for index, batch in enumerate(attachment_batches):
            client, message = await self._send_message_batch(
                account_id=account_id,
                client=client,
                chat_id=chat_id,
                text="" if send_text_before_audio else text if index == 0 else "",
                notify=notify,
                attachments=batch or None,
                reply_to=reply_to,
            )
            if message is None:
                raise ValueError("PyMax returned no message")
            if first_message is None:
                first_message = message

        if first_message is None:
            raise ValueError("PyMax returned no message")

        return SentMessage(chat_id=str(chat_id), message_id=str(first_message.id))

    async def _send_message_batch(
        self,
        *,
        account_id: UUID,
        client: ClickSafeSocketMaxClient,
        chat_id: int,
        text: str,
        notify: bool,
        attachments: list[PyMaxAttachment] | None,
        reply_to: int | None,
    ) -> tuple[ClickSafeSocketMaxClient, Message | None]:
        try:
            message = await client.send_message(
                chat_id=chat_id,
                text=text,
                notify=notify,
                attachments=attachments,
                reply_to=reply_to,
            )
        except Exception as exc:
            if reply_to is None or not self._is_reply_payload_number_error(exc):
                raise

            logger.warning(
                "Max reply send failed because PyMax serialized reply_to incorrectly; "
                "waiting for reconnect before retrying without reply link",
                exc_info=True,
            )
            await asyncio.sleep(self.reconnect_delay + 0.2)
            client = await self._ensure_started(account_id)
            await self._wait_connected(client)
            message = await client.send_message(
                chat_id=chat_id,
                text=text,
                notify=notify,
                attachments=attachments,
                reply_to=None,
            )
        return client, message

    async def _finish_login_response(
        self,
        *,
        attempt: Any,
        response: dict[str, Any],
        password: str,
        client: SocketMaxClient,
    ) -> LoginResult:
        login_attrs = response.get("tokenAttrs", {}).get("LOGIN", {})
        password_challenge = response.get("passwordChallenge")

        if password_challenge and not login_attrs:
            track_id = password_challenge.get("trackId", "")
            if not track_id:
                raise ValueError("password challenge missing track id")
            await self.repo.update_login_attempt(
                attempt.id,
                status="PASSWORD_REQUIRED",
                password_track_id=track_id,
            )
            if password:
                token_attrs = await client._check_password(password, track_id)
                if not token_attrs:
                    return LoginResult(status="PASSWORD_REQUIRED", error="incorrect password")
                token = token_attrs.get("LOGIN", {}).get("token")
                if not token:
                    raise ValueError("password response did not contain login token")
                account = await self._save_account_from_token(attempt, token, client)
                await self.repo.update_login_attempt(attempt.id, status="COMPLETED")
                return LoginResult(status="COMPLETED", account=account)
            return LoginResult(status="PASSWORD_REQUIRED", password_track_id=track_id)

        token = login_attrs.get("token")
        if not token:
            raise ValueError("login response did not contain token")

        account = await self._save_account_from_token(attempt, token, client)
        await self.repo.update_login_attempt(attempt.id, status="COMPLETED")
        return LoginResult(status="COMPLETED", account=account)

    async def _save_account_from_token(
        self,
        attempt: Any,
        token: str,
        client: SocketMaxClient,
    ) -> AccountRecord:
        return await self.repo.upsert_account(
            phone=attempt.phone,
            label=attempt.label,
            encrypted_token=self.secrets.encrypt(token),
            device_id=attempt.device_id,
            max_user_id=str(client.me.id) if client.me else "",
            status="ACTIVE",
        )

    async def _ensure_started(self, account_id: UUID) -> SocketMaxClient:
        client = self._clients.get(account_id)
        if client and client.is_connected:
            return client

        await self.start_account(account_id)
        client = self._clients.get(account_id)
        if not client:
            raise ValueError("account did not start")
        return client

    async def _resolve_chat_id(
        self,
        client: SocketMaxClient,
        recipient_kind: str,
        recipient_value: str,
    ) -> int:
        if recipient_kind == "CHAT_ID":
            return int(recipient_value)
        if recipient_kind == "USER_ID":
            if not client.me:
                raise ValueError("account profile is not loaded")
            return client.get_chat_id(client.me.id, int(recipient_value))
        if recipient_kind == "PHONE":
            if not client.me:
                raise ValueError("account profile is not loaded")
            user = await client.search_by_phone(recipient_value)
            return client.get_chat_id(client.me.id, user.id)
        raise ValueError("unsupported recipient kind")

    def _register_handlers(self, account_id: UUID, client: SocketMaxClient) -> None:
        @client.on_message()
        async def on_message(message: Message) -> None:
            payload = self._message_payload(message)
            event_type = "MESSAGE_REPLIED" if message.link else "MESSAGE_RECEIVED"
            reply_to = ""
            if message.link is not None:
                linked_msg = getattr(message.link, "message", None)
                reply_to = str(getattr(linked_msg, "id", "") or "")
            await self.repo.append_event(
                type_=event_type,
                account_id=account_id,
                chat_id=str(message.chat_id or ""),
                message_id=str(message.id),
                sender_id=str(message.sender or ""),
                text=message.text or "",
                reply_to_message_id=reply_to,
                payload=payload,
            )

        @client.on_raw_receive
        async def on_raw(data: dict[str, Any]) -> None:
            opcode = data.get("opcode")
            payload = data.get("payload") if isinstance(data.get("payload"), dict) else {}
            if opcode == 130:
                chat_id, message_id = self._read_marker_ids(payload)
                await self.repo.append_event(
                    type_="MESSAGE_READ",
                    account_id=account_id,
                    chat_id=chat_id,
                    message_id=message_id,
                    payload=self._safe_json(data),
                )
                return

            logger.debug("Ignoring raw MAX event: opcode=%s", opcode)

    @staticmethod
    def _message_payload(message: Message) -> dict[str, Any]:
        link_message_id = ""
        link_chat_id = ""
        if message.link is not None:
            link_chat_id = str(getattr(message.link, "chat_id", "") or "")
            linked_message = getattr(message.link, "message", None)
            link_message_id = str(getattr(linked_message, "id", "") or "")

        return {
            "chat_id": message.chat_id,
            "sender": message.sender,
            "id": message.id,
            "time": message.time,
            "type": str(message.type),
            "status": str(message.status) if message.status else "",
            "has_link": bool(message.link),
            "link_chat_id": link_chat_id,
            "link_message_id": link_message_id,
            "attaches": [str(a) for a in message.attaches or []],
        }

    @staticmethod
    def _read_marker_ids(payload: dict[str, Any]) -> tuple[str, str]:
        chat_id = PyMaxManager._string_id(
            payload.get("chatId")
            or payload.get("chat_id")
            or payload.get("chatID")
            or PyMaxManager._nested(payload, "chat", "id")
        )

        message_id = PyMaxManager._string_id(
            payload.get("messageId")
            or payload.get("message_id")
            or PyMaxManager._nested(payload, "message", "id")
        )
        if not message_id:
            messages = payload.get("messages")
            if isinstance(messages, list) and messages:
                first = messages[0]
                if isinstance(first, dict):
                    message_id = PyMaxManager._string_id(
                        first.get("messageId")
                        or first.get("message_id")
                        or first.get("id")
                    )

        return chat_id, message_id

    @staticmethod
    def _nested(data: dict[str, Any], *keys: str) -> Any:
        current: Any = data
        for key in keys:
            if not isinstance(current, dict):
                return None
            current = current.get(key)
        return current

    @staticmethod
    def _string_id(value: Any) -> str:
        if value is None:
            return ""
        return str(value)

    @staticmethod
    def _is_reply_payload_number_error(exc: Exception) -> bool:
        text = str(exc)
        return "Expected number" in text and "proto.payload" in text

    @staticmethod
    def _split_attachment_batches(
        attachments: list[tuple[str, PyMaxAttachment]],
    ) -> list[list[PyMaxAttachment]]:
        return [[attachment] for _, attachment in attachments]

    @staticmethod
    def _has_audio_attachments(attachments: list[tuple[str, PyMaxAttachment]]) -> bool:
        return any(kind == "AUDIO" for kind, _ in attachments)

    @staticmethod
    def _safe_json(data: dict[str, Any]) -> dict[str, Any]:
        try:
            json.dumps(data)
            return data
        except TypeError:
            return json.loads(json.dumps(data, default=str))

    @staticmethod
    def _build_attachment(kind: str, path: Path) -> Photo | File | Video | Audio:
        if kind == "PHOTO":
            return Photo(path=str(path))
        if kind == "VIDEO":
            return Video(path=str(path))
        if kind == "AUDIO":
            return Audio(path=str(path))
        return File(path=str(path))

    def _new_client(
        self,
        *,
        phone: str,
        token: str | None = None,
        device_id: UUID | None = None,
        reconnect: bool = False,
    ) -> ClickSafeSocketMaxClient:
        return ClickSafeSocketMaxClient(
            phone=phone,
            token=token,
            device_id=device_id,
            headers=UserAgentPayload(device_type="DESKTOP"),
            send_fake_telemetry=False,
            persist_session=False,
            reconnect=reconnect,
            reconnect_delay=self.reconnect_delay,
        )

    @staticmethod
    async def _cleanup_client(client: SocketMaxClient) -> None:
        try:
            await client.close()
            await client._cleanup_client()
        except Exception:
            logger.debug("failed to cleanup PyMax client", exc_info=True)

    @staticmethod
    async def _wait_connected(client: SocketMaxClient) -> None:
        for _ in range(200):
            if client.is_connected and client.me is not None:
                return
            await asyncio.sleep(0.1)
        if client.is_connected:
            return
        raise TimeoutError("Max account connection timed out")
