-- +goose Up
-- +goose StatementBegin
ALTER TABLE campaigns
    ADD COLUMN IF NOT EXISTS type VARCHAR(64) NOT NULL DEFAULT 'EMAIL',
    ADD COLUMN IF NOT EXISTS max_education_text_id UUID REFERENCES attachments(id) ON DELETE RESTRICT;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS type VARCHAR(64) NOT NULL DEFAULT 'EMAIL',
    ADD COLUMN IF NOT EXISTS max_account_id UUID REFERENCES max_accounts(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS text_body_id UUID REFERENCES attachments(id) ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS max_deliveries (
    id UUID PRIMARY KEY,
    target_id UUID NOT NULL UNIQUE REFERENCES targets(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    max_account_id UUID NOT NULL REFERENCES max_accounts(id) ON DELETE RESTRICT,
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

CREATE INDEX IF NOT EXISTS idx_max_deliveries_adapter_message
    ON max_deliveries(adapter_account_id, chat_id, message_id);
CREATE INDEX IF NOT EXISTS idx_max_deliveries_adapter_chat
    ON max_deliveries(adapter_account_id, chat_id, sent_at DESC);

CREATE TABLE IF NOT EXISTS max_adapter_processed_events (
    seq BIGINT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS max_adapter_processed_events;
DROP INDEX IF EXISTS idx_max_deliveries_adapter_chat;
DROP INDEX IF EXISTS idx_max_deliveries_adapter_message;
DROP TABLE IF EXISTS max_deliveries;

ALTER TABLE messages
    DROP COLUMN IF EXISTS text_body_id,
    DROP COLUMN IF EXISTS max_account_id,
    DROP COLUMN IF EXISTS type;

ALTER TABLE campaigns
    DROP COLUMN IF EXISTS max_education_text_id,
    DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
