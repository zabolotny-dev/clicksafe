-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS max_accounts (
    id UUID PRIMARY KEY,
    adapter_id UUID NOT NULL UNIQUE,
    phone_number VARCHAR(64) NOT NULL UNIQUE,
    label VARCHAR(255) NOT NULL,
    status VARCHAR(64) NOT NULL,
    max_user_id VARCHAR(255),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_max_accounts_status ON max_accounts(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_max_accounts_status;
DROP TABLE IF EXISTS max_accounts;
-- +goose StatementEnd
