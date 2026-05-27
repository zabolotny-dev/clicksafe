from __future__ import annotations

import grpc


async def require_internal_token(context: grpc.aio.ServicerContext, expected: str) -> None:
    if not expected:
        return

    authorization = ""
    for key, value in context.invocation_metadata():
        if key.lower() == "authorization":
            authorization = value
            break

    if authorization != f"Bearer {expected}":
        await context.abort(grpc.StatusCode.UNAUTHENTICATED, "invalid adapter token")
