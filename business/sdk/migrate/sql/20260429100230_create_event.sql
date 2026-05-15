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
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    csrf_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip_address INET,
    user_agent TEXT
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
CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(64) NOT NULL,
    content_path VARCHAR(255) NOT NULL,
    required_vars TEXT[],
    uploaded_at TIMESTAMPTZ NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    attachment_id UUID REFERENCES attachments(id) ON DELETE RESTRICT,
    attributes JSONB
);
CREATE TABLE IF NOT EXISTS landings (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    html_body_id UUID REFERENCES attachments(id) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    subject VARCHAR(255),
    html_body_id UUID REFERENCES attachments(id) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS message_attachments (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    attachment_id UUID NOT NULL REFERENCES attachments(id) ON DELETE RESTRICT,
    PRIMARY KEY (message_id, attachment_id)
);
CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    message_id UUID REFERENCES messages(id) ON DELETE RESTRICT,
    landing_id UUID REFERENCES landings(id) ON DELETE RESTRICT,
    education_id UUID REFERENCES landings(id) ON DELETE RESTRICT,
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
CREATE INDEX IF NOT EXISTS idx_landings_html_body_id ON landings(html_body_id);
CREATE INDEX IF NOT EXISTS idx_messages_html_body_id ON messages(html_body_id);
CREATE INDEX IF NOT EXISTS idx_message_attachments_attachment_id ON message_attachments(attachment_id);
CREATE INDEX IF NOT EXISTS idx_organizations_attachment_id ON organizations(attachment_id);
CREATE INDEX IF NOT EXISTS idx_sessions_admin_id ON sessions(admin_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_admin_id;
DROP INDEX IF EXISTS idx_landings_html_body_id;
DROP INDEX IF EXISTS idx_messages_html_body_id;
DROP INDEX IF EXISTS idx_message_attachments_attachment_id;
DROP INDEX IF EXISTS idx_organizations_attachment_id;
DROP TABLE IF EXISTS targets;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS message_attachments;
DROP TABLE IF EXISTS landings;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
