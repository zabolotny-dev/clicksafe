CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    status VARCHAR(64) NOT NULL
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

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    occurred_at TIMESTAMP NOT NULL
);
