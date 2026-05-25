CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    encrypted_token TEXT NOT NULL,
    device_id UUID NOT NULL,
    max_user_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL,
    label TEXT NOT NULL,
    encrypted_temp_token TEXT NOT NULL,
    device_id UUID NOT NULL,
    password_track_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    error TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_expires_at ON login_attempts(expires_at);

CREATE TABLE IF NOT EXISTS send_requests (
    client_request_id TEXT PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    chat_id TEXT NOT NULL DEFAULT '',
    message_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS event_outbox (
    seq BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    account_id UUID,
    chat_id TEXT NOT NULL DEFAULT '',
    message_id TEXT NOT NULL DEFAULT '',
    sender_id TEXT NOT NULL DEFAULT '',
    text TEXT NOT NULL DEFAULT '',
    reply_to_message_id TEXT NOT NULL DEFAULT '',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_event_outbox_occurred_at ON event_outbox(occurred_at);

CREATE TABLE IF NOT EXISTS event_offsets (
    consumer TEXT PRIMARY KEY,
    acknowledged_seq BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL
);
