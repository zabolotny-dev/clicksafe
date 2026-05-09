-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events(
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    occurred_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    logo_path VARCHAR(255),
    attributes JSONB
);

CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    attributes JSONB
);

CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY,
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone_number VARCHAR(64) UNIQUE,
    attributes JSONB
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    subject VARCHAR(255),
    content_path VARCHAR(255),
    required_vars TEXT[]
);

CREATE TABLE IF NOT EXISTS landings (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    content_path VARCHAR(255),
    required_vars TEXT[]
);

CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    message_id UUID REFERENCES messages(id) ON DELETE RESTRICT,
    landing_id UUID REFERENCES landings(id) ON DELETE RESTRICT,
    domain VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(64) NOT NULL,
    date_from TIMESTAMPTZ,
    date_to TIMESTAMPTZ,
    attributes JSONB
);

CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY,
    token VARCHAR(64) NOT NULL UNIQUE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    status VARCHAR(64) NOT NULL,
    scheduled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,

    UNIQUE (employee_id, campaign_id)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS landings;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS targets;
-- +goose StatementEnd
