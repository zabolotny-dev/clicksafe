CREATE TABLE IF NOT EXISTS max_deliveries (
    id UUID PRIMARY KEY,
    target_id UUID NOT NULL UNIQUE,
    campaign_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    max_account_id UUID NOT NULL,
    adapter_account_id UUID NOT NULL,
    chat_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL,
    read_at TIMESTAMPTZ,
    replied_at TIMESTAMPTZ,
    education_sent_at TIMESTAMPTZ,
    incoming_message_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS max_adapter_processed_events (
    seq BIGINT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL
);
