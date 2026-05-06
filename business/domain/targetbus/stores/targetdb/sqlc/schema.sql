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
