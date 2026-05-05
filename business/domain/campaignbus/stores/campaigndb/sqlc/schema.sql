CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    message_id UUID REFERENCES messages(id) ON DELETE RESTRICT,
    label VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(64) NOT NULL,
    date_from TIMESTAMPTZ,
    date_to TIMESTAMPTZ,
    attributes JSONB
);
