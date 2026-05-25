from __future__ import annotations

import asyncio
import json
import logging
import shutil
from pathlib import Path
from typing import Any
from uuid import UUID

import aiofiles
import grpc

from maxadapter.v1 import max_adapter_pb2 as pb2
from maxadapter.v1 import max_adapter_pb2_grpc as pb2_grpc

from max_adapter.config import Config
from max_adapter.db.repository import (
    AccountRecord,
    EventRecord,
    LoginAttemptRecord,
    Repository,
    rfc3339,
)
from max_adapter.services.auth import require_internal_token
from max_adapter.services.pymax_manager import PyMaxManager

logger = logging.getLogger(__name__)


ACCOUNT_STATUS_TO_PROTO = {
    "PENDING_LOGIN": pb2.ACCOUNT_STATUS_PENDING_LOGIN,
    "ACTIVE": pb2.ACCOUNT_STATUS_ACTIVE,
    "CONNECTED": pb2.ACCOUNT_STATUS_CONNECTED,
    "DISCONNECTED": pb2.ACCOUNT_STATUS_DISCONNECTED,
    "ERROR": pb2.ACCOUNT_STATUS_ERROR,
}

LOGIN_STATUS_TO_PROTO = {
    "CODE_REQUIRED": pb2.LOGIN_ATTEMPT_STATUS_CODE_REQUIRED,
    "PASSWORD_REQUIRED": pb2.LOGIN_ATTEMPT_STATUS_PASSWORD_REQUIRED,
    "COMPLETED": pb2.LOGIN_ATTEMPT_STATUS_COMPLETED,
    "FAILED": pb2.LOGIN_ATTEMPT_STATUS_FAILED,
    "EXPIRED": pb2.LOGIN_ATTEMPT_STATUS_EXPIRED,
}

EVENT_TYPE_TO_PROTO = {
    "ACCOUNT_CONNECTED": pb2.ADAPTER_EVENT_TYPE_ACCOUNT_CONNECTED,
    "ACCOUNT_DISCONNECTED": pb2.ADAPTER_EVENT_TYPE_ACCOUNT_DISCONNECTED,
    "MESSAGE_RECEIVED": pb2.ADAPTER_EVENT_TYPE_MESSAGE_RECEIVED,
    "MESSAGE_READ": pb2.ADAPTER_EVENT_TYPE_MESSAGE_READ,
    "MESSAGE_REPLIED": pb2.ADAPTER_EVENT_TYPE_MESSAGE_REPLIED,
    "RAW": pb2.ADAPTER_EVENT_TYPE_RAW,
}

RECIPIENT_KIND = {
    pb2.RECIPIENT_KIND_CHAT_ID: "CHAT_ID",
    pb2.RECIPIENT_KIND_PHONE: "PHONE",
    pb2.RECIPIENT_KIND_USER_ID: "USER_ID",
}

ATTACHMENT_KIND = {
    pb2.ATTACHMENT_KIND_PHOTO: "PHOTO",
    pb2.ATTACHMENT_KIND_FILE: "FILE",
    pb2.ATTACHMENT_KIND_VIDEO: "VIDEO",
    pb2.ATTACHMENT_KIND_AUDIO: "AUDIO",
}


class MaxAdapterServicer(pb2_grpc.MaxAdapterServiceServicer):
    def __init__(self, cfg: Config, repo: Repository, manager: PyMaxManager) -> None:
        self.cfg = cfg
        self.repo = repo
        self.manager = manager
        self.temp_dir = Path(cfg.temp_dir)
        self.temp_dir.mkdir(parents=True, exist_ok=True)

    async def Health(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        return pb2.HealthResponse(status="ok")

    async def BeginLogin(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        phone = request.phone.strip()
        label = request.label.strip() or phone
        if not phone:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "phone is required")

        try:
            attempt = await self.manager.begin_login(phone, label, self.cfg.login_ttl_seconds)
        except Exception as exc:
            logger.exception("begin login failed")
            await context.abort(grpc.StatusCode.FAILED_PRECONDITION, str(exc))

        return pb2.BeginLoginResponse(attempt=self._login_attempt(attempt))

    async def ConfirmLogin(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        attempt_id = await self._uuid_or_abort(request.login_attempt_id, "login_attempt_id", context)
        if not request.code:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "code is required")

        result = await self.manager.confirm_login(
            attempt_id=attempt_id,
            code=request.code,
            password=request.password,
        )
        attempt = await self.repo.get_login_attempt(attempt_id)
        return pb2.ConfirmLoginResponse(
            attempt=self._login_attempt(attempt) if attempt else pb2.LoginAttempt(),
            account=self._account(result.account),
        )

    async def ConfirmPassword(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        attempt_id = await self._uuid_or_abort(request.login_attempt_id, "login_attempt_id", context)
        if not request.password:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "password is required")

        result = await self.manager.confirm_password(
            attempt_id=attempt_id,
            password=request.password,
        )
        attempt = await self.repo.get_login_attempt(attempt_id)
        return pb2.ConfirmLoginResponse(
            attempt=self._login_attempt(attempt) if attempt else pb2.LoginAttempt(),
            account=self._account(result.account),
        )

    async def ListAccounts(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        accounts = await self.repo.list_accounts()
        return pb2.ListAccountsResponse(accounts=[self._account(account) for account in accounts])

    async def GetAccount(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        account_id = await self._uuid_or_abort(request.account_id, "account_id", context)
        account = await self.repo.get_account(account_id)
        if not account:
            await context.abort(grpc.StatusCode.NOT_FOUND, "account not found")
        return pb2.GetAccountResponse(account=self._account(account))

    async def DeleteAccount(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        account_id = await self._uuid_or_abort(request.account_id, "account_id", context)
        try:
            await self.manager.delete_account(account_id)
        except ValueError:
            await context.abort(grpc.StatusCode.NOT_FOUND, "account not found")
        return pb2.DeleteAccountResponse()

    async def StartAccount(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        account_id = await self._uuid_or_abort(request.account_id, "account_id", context)
        try:
            account = await self.manager.start_account(account_id)
        except ValueError:
            await context.abort(grpc.StatusCode.NOT_FOUND, "account not found")
        except Exception as exc:
            await context.abort(grpc.StatusCode.FAILED_PRECONDITION, str(exc))
        return self._account(account)

    async def StopAccount(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        account_id = await self._uuid_or_abort(request.account_id, "account_id", context)
        try:
            account = await self.manager.stop_account(account_id)
        except ValueError:
            await context.abort(grpc.StatusCode.NOT_FOUND, "account not found")
        return self._account(account)

    async def SendMessage(self, request_iterator, context):
        await require_internal_token(context, self.cfg.grpc_token)
        metadata = None
        request_dir: Path | None = None
        handles: dict[int, Any] = {}
        paths: dict[int, Path] = {}

        try:
            async for request in request_iterator:
                payload = request.WhichOneof("payload")
                if payload == "metadata":
                    if metadata is not None:
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "metadata sent twice")
                    metadata = request.metadata
                    account_id = await self._validate_send_metadata(metadata, context)
                    existing = await self.repo.get_send_request(metadata.client_request_id)
                    if existing and existing.status == "SENT":
                        return pb2.SendMessageResponse(
                            client_request_id=existing.client_request_id,
                            account_id=str(existing.account_id),
                            chat_id=existing.chat_id,
                            message_id=existing.message_id,
                        )
                    request_dir = self.temp_dir / metadata.client_request_id
                    request_dir.mkdir(parents=True, exist_ok=True)
                    for index, descriptor in enumerate(metadata.attachments):
                        paths[index] = request_dir / self._safe_filename(index, descriptor.filename)
                    continue

                if payload == "chunk":
                    if metadata is None or request_dir is None:
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "metadata must be sent before chunks")
                    index = request.chunk.attachment_index
                    if index < 0 or index >= len(metadata.attachments):
                        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "invalid attachment index")
                    if index not in handles:
                        handles[index] = await aiofiles.open(paths[index], "wb")
                    await handles[index].write(request.chunk.data)

            if metadata is None:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "metadata is required")

            for handle in handles.values():
                await handle.close()
            handles.clear()

            attachment_paths = []
            for index, descriptor in enumerate(metadata.attachments):
                kind = ATTACHMENT_KIND.get(descriptor.kind, "FILE")
                path = paths.get(index)
                if path is None or not path.exists():
                    await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "attachment bytes missing")
                attachment_paths.append((kind, path))

            await self.repo.save_send_request(
                client_request_id=metadata.client_request_id,
                account_id=account_id,
                chat_id="",
                message_id="",
                status="PENDING",
            )
            sent = await self.manager.send_message(
                account_id=account_id,
                recipient_kind=RECIPIENT_KIND.get(metadata.recipient.kind, ""),
                recipient_value=metadata.recipient.value,
                text=metadata.text,
                notify=metadata.notify,
                reply_to=int(metadata.reply_to_message_id) if metadata.reply_to_message_id else None,
                attachments=attachment_paths,
            )
            await self.repo.save_send_request(
                client_request_id=metadata.client_request_id,
                account_id=account_id,
                chat_id=sent.chat_id,
                message_id=sent.message_id,
                status="SENT",
            )
            return pb2.SendMessageResponse(
                client_request_id=metadata.client_request_id,
                account_id=metadata.account_id,
                chat_id=sent.chat_id,
                message_id=sent.message_id,
            )
        except grpc.RpcError:
            raise
        except Exception as exc:
            logger.exception("send message failed")
            if metadata is not None and metadata.client_request_id:
                try:
                    failed_account_id = UUID(metadata.account_id)
                except ValueError:
                    failed_account_id = None
                if failed_account_id:
                    await self.repo.save_send_request(
                        client_request_id=metadata.client_request_id,
                        account_id=failed_account_id,
                        chat_id="",
                        message_id="",
                        status="FAILED",
                        error=str(exc),
                    )
            await context.abort(grpc.StatusCode.FAILED_PRECONDITION, str(exc))
        finally:
            for handle in handles.values():
                await handle.close()
            if request_dir and request_dir.exists():
                shutil.rmtree(request_dir, ignore_errors=True)

    async def SubscribeEvents(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        consumer = request.consumer.strip()
        if not consumer:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "consumer is required")

        cursor = request.from_seq or await self.repo.get_ack(consumer)
        while True:
            events = await self.repo.list_events_after(cursor)
            if not events:
                await asyncio.sleep(1)
                continue
            for event in events:
                cursor = event.seq
                yield self._event(event)

    async def AckEvents(self, request, context):
        await require_internal_token(context, self.cfg.grpc_token)
        consumer = request.consumer.strip()
        if not consumer:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "consumer is required")
        acknowledged = await self.repo.ack_events(consumer, request.up_to_seq)
        return pb2.AckEventsResponse(acknowledged_seq=acknowledged)

    @staticmethod
    async def _uuid_or_abort(value: str, field: str, context) -> UUID:
        if not value:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"{field} is required")
        try:
            return UUID(value)
        except ValueError:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"{field} must be a UUID")

    @classmethod
    async def _validate_send_metadata(cls, metadata: pb2.SendMessageMetadata, context) -> UUID:
        account_id = await cls._uuid_or_abort(metadata.account_id, "account_id", context)
        if not metadata.client_request_id:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "client_request_id is required")
        if not metadata.recipient.value:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "recipient is required")
        if metadata.recipient.kind not in RECIPIENT_KIND:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "recipient kind is required")
        return account_id

    @staticmethod
    def _safe_filename(index: int, filename: str) -> str:
        clean = Path(filename or f"attachment-{index}").name
        return f"{index}-{clean}"

    @staticmethod
    def _account(account: AccountRecord | None) -> pb2.Account:
        if account is None:
            return pb2.Account()
        return pb2.Account(
            id=str(account.id),
            phone=account.phone,
            label=account.label,
            status=ACCOUNT_STATUS_TO_PROTO.get(account.status, pb2.ACCOUNT_STATUS_UNSPECIFIED),
            max_user_id=account.max_user_id,
            last_error=account.last_error,
            created_at=rfc3339(account.created_at),
            updated_at=rfc3339(account.updated_at),
        )

    @staticmethod
    def _login_attempt(attempt: LoginAttemptRecord | None) -> pb2.LoginAttempt:
        if attempt is None:
            return pb2.LoginAttempt()
        return pb2.LoginAttempt(
            id=str(attempt.id),
            phone=attempt.phone,
            label=attempt.label,
            status=LOGIN_STATUS_TO_PROTO.get(attempt.status, pb2.LOGIN_ATTEMPT_STATUS_UNSPECIFIED),
            error=attempt.error,
            expires_at=rfc3339(attempt.expires_at),
        )

    @staticmethod
    def _event(event: EventRecord) -> pb2.AdapterEvent:
        return pb2.AdapterEvent(
            seq=event.seq,
            type=EVENT_TYPE_TO_PROTO.get(event.type, pb2.ADAPTER_EVENT_TYPE_UNSPECIFIED),
            account_id=str(event.account_id) if event.account_id else "",
            chat_id=event.chat_id,
            message_id=event.message_id,
            sender_id=event.sender_id,
            text=event.text,
            reply_to_message_id=event.reply_to_message_id,
            payload_json=json.dumps(event.payload_json, ensure_ascii=False),
            occurred_at=rfc3339(event.occurred_at),
        )
