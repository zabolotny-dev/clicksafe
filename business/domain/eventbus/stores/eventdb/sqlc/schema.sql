CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    occurred_at TIMESTAMP NOT NULL
);