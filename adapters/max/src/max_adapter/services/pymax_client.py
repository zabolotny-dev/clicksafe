from __future__ import annotations

import asyncio
import logging
import ssl
import time
from typing import Any

from pymax import SocketMaxClient
from pymax.exceptions import Error, SocketNotConnectedError, SocketSendError
from pymax.files import Audio, File, Photo, Video
from pymax.formatting import Formatting
from pymax.payloads import MessageElement, SendMessagePayload, SendMessagePayloadMessage
from pymax.static.constant import DEFAULT_TIMEOUT
from pymax.static.enum import Opcode
from pymax.types import Message

logger = logging.getLogger(__name__)

# Shorter ping interval to prevent server-side idle disconnects.
# The default PyMax value (30s) is too close to the server's idle
# timeout on api.oneme.ru, causing frequent SSL record layer failures.
_ADAPTER_PING_INTERVAL: float = 15.0


class ClickSafeSocketMaxClient(SocketMaxClient):
    """SocketMaxClient variant hardened for long-running adapter use.

    Changes vs upstream SocketMaxClient:
    - Serialized TLS writes via asyncio.Lock to avoid SSL corruption.
    - Shortened ping interval (15s) to keep idle connections alive.
    - recv_loop breaks immediately on SSL/connection errors instead
      of backing off on a broken socket.
    """

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self._socket_send_lock = asyncio.Lock()

    async def close(self) -> None:
        self.logger.info("Closing client and socket immediately")
        self._stop_event.set()
        if self._socket:
            try:
                self._socket.close()
            except Exception:
                pass
        await super().close()

    def _setup_logger(self) -> None:
        self.logger.handlers.clear()
        if not self.logger.level:
            self.logger.setLevel(logging.INFO)
        self.logger.propagate = True

    async def send_message(
        self,
        text: str,
        chat_id: int,
        notify: bool = True,
        attachment: Photo | File | Video | Audio | None = None,
        attachments: list[Photo | File | Video | Audio] | None = None,
        reply_to: int | None = None,
        use_queue: bool = False,
    ) -> Message | None:
        if reply_to is None:
            return await super().send_message(
                text=text,
                chat_id=chat_id,
                notify=notify,
                attachment=attachment,
                attachments=attachments,
                reply_to=reply_to,
                use_queue=use_queue,
            )

        self.logger.info("Sending message to chat_id=%s notify=%s", chat_id, notify)
        if attachments and attachment:
            self.logger.warning("Both attachment and attachments provided; using attachments")
            attachment = None

        attaches = []
        has_audio = False
        if attachment:
            self.logger.info("Uploading attachment for message")
            result = await self._upload_attachment(attachment)
            if not result:
                raise Error("upload_failed", "Failed to upload attachment", "Upload Error")
            attaches.append(result)
            has_audio = has_audio or self._is_audio_attachment(result)
        elif attachments:
            self.logger.info("Uploading multiple attachments for message")
            for item in attachments:
                result = await self._upload_attachment(item)
                if not result:
                    raise Error("upload_failed", "Failed to upload attachment", "Upload Error")
                attaches.append(result)
                has_audio = has_audio or self._is_audio_attachment(result)

            if not attaches:
                raise Error("upload_failed", "All attachments failed to upload", "Upload Error")

        if has_audio and use_queue:
            raise Error(
                "audio_queue_unsupported",
                "Audio messages cannot be sent through the outgoing queue.",
                "Audio Error",
            )

        raw_elements, parsed_text = Formatting.get_elements_from_markdown(text)
        elements = [
            MessageElement(type=item.type, length=item.length, from_=item.from_)
            for item in raw_elements
        ]
        payload = SendMessagePayload(
            chat_id=chat_id,
            message=SendMessagePayloadMessage(
                text=parsed_text if raw_elements else text,
                cid=int(time.time() * 1000),
                elements=elements,
                attaches=attaches,
                link=None,
            ),
            notify=notify,
        ).model_dump(by_alias=True)
        payload["message"]["link"] = {"type": "REPLY", "messageId": reply_to}

        if use_queue:
            await self._queue_message(opcode=Opcode.MSG_SEND, payload=payload)
            self.logger.debug("Message queued for sending")
            return None

        return await self._send_message_payload(payload, has_audio=has_audio)

    # ------------------------------------------------------------------
    # Override: shorter ping interval
    # ------------------------------------------------------------------
    async def _send_interactive_ping(self) -> None:
        while self.is_connected:
            try:
                await self._send_and_wait(
                    opcode=Opcode.PING,
                    payload={"interactive": True},
                    cmd=0,
                )
                self.logger.debug("Interactive ping sent successfully")
            except SocketNotConnectedError:
                self.logger.debug("Socket disconnected, exiting ping loop")
                break
            except Exception:
                self.logger.warning("Interactive ping failed")
            await asyncio.sleep(_ADAPTER_PING_INTERVAL)

    # ------------------------------------------------------------------
    # Override: recv_loop that breaks immediately on SSL errors
    # ------------------------------------------------------------------
    async def _recv_loop(self) -> None:
        if self._socket is None:
            self.logger.warning("Recv loop started without socket instance")
            return

        sock = self._socket
        loop = asyncio.get_running_loop()

        while True:
            try:
                header = await self._parse_header(loop, sock)

                if not header:
                    break

                datas = await self._recv_data(loop, header, sock)

                if not datas:
                    continue

                for data_item in datas:
                    seq = data_item.get("seq")

                    if self._handle_pending(seq % 256 if seq is not None else None, data_item):
                        continue

                    if self._incoming is not None:
                        await self._handle_incoming_queue(data_item)

                    await self._dispatch_incoming(data_item)

            except asyncio.CancelledError:
                self.logger.debug("Recv loop cancelled")
                raise
            except (ssl.SSLError, ssl.SSLEOFError):
                # SSL session is broken — no point retrying on this socket.
                self.logger.warning("SSL error in recv_loop, triggering reconnect")
                self.is_connected = False
                break
            except (ConnectionError, OSError):
                self.logger.warning("Connection error in recv_loop, triggering reconnect")
                self.is_connected = False
                break
            except Exception:
                self.logger.exception("Unexpected error in recv_loop, triggering reconnect")
                self.is_connected = False
                break

    # ------------------------------------------------------------------
    # Override: serialized TLS writes
    # ------------------------------------------------------------------
    async def _send_and_wait(
        self,
        opcode: Opcode,
        payload: dict[str, Any],
        cmd: int = 0,
        timeout: float = DEFAULT_TIMEOUT,
    ) -> dict[str, Any]:
        if not self.is_connected or self._socket is None:
            raise SocketNotConnectedError

        sock = self.sock
        msg = self._make_message(opcode, payload, cmd)
        loop = asyncio.get_running_loop()
        fut: asyncio.Future[dict[str, Any]] = loop.create_future()
        seq_key = msg["seq"] % 256

        old_fut = self._pending.get(seq_key)
        if old_fut and not old_fut.done():
            old_fut.cancel()

        self._pending[seq_key] = fut
        try:
            self.logger.debug(
                "Sending frame opcode=%s cmd=%s seq=%s",
                opcode,
                cmd,
                msg["seq"],
            )
            packet = self._pack_packet(
                msg["ver"],
                msg["cmd"],
                msg["seq"],
                msg["opcode"],
                msg["payload"],
            )
            async with self._socket_send_lock:
                await loop.run_in_executor(None, lambda: sock.sendall(packet))
            data = await asyncio.wait_for(fut, timeout=timeout)
            self.logger.debug(
                "Received frame for seq=%s opcode=%s",
                data.get("seq"),
                data.get("opcode"),
            )
            return data

        except (ssl.SSLEOFError, ssl.SSLError, ConnectionError) as conn_err:
            self.logger.warning("Connection lost during send")
            self.is_connected = False
            raise SocketNotConnectedError from conn_err
        except asyncio.TimeoutError:
            self.logger.exception(
                "Send and wait failed (opcode=%s, seq=%s)", opcode, msg["seq"]
            )
            raise SocketSendError from None
        except Exception as exc:
            self.logger.exception(
                "Send and wait failed (opcode=%s, seq=%s)", opcode, msg["seq"]
            )
            raise SocketSendError from exc

        finally:
            self._pending.pop(msg["seq"] % 256, None)
