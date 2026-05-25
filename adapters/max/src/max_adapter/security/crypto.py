from __future__ import annotations

import base64
import hashlib

from cryptography.fernet import Fernet, InvalidToken


class SecretBox:
    def __init__(self, key: str) -> None:
        if not key:
            raise ValueError("MAX_ADAPTER_SECRET_KEY is required")

        self._fernet = Fernet(self._normalize_key(key))

    def encrypt(self, value: str) -> str:
        return self._fernet.encrypt(value.encode("utf-8")).decode("ascii")

    def decrypt(self, value: str) -> str:
        try:
            return self._fernet.decrypt(value.encode("ascii")).decode("utf-8")
        except InvalidToken as exc:
            raise ValueError("encrypted secret cannot be decrypted") from exc

    @staticmethod
    def _normalize_key(value: str) -> bytes:
        raw = value.encode("utf-8")
        try:
            Fernet(raw)
            return raw
        except Exception:
            digest = hashlib.sha256(raw).digest()
            return base64.urlsafe_b64encode(digest)
