-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events(
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    occurred_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    logo_url VARCHAR(255),
    attributes JSONB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
