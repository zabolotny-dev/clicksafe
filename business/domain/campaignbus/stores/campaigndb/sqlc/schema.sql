CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    type VARCHAR(64) NOT NULL DEFAULT 'EMAIL',
    message_id UUID REFERENCES messages(id) ON DELETE RESTRICT,
    landing_id UUID REFERENCES landings(id) ON DELETE RESTRICT,
    education_id UUID REFERENCES landings(id) ON DELETE RESTRICT,
    max_education_text_id UUID REFERENCES attachments(id) ON DELETE RESTRICT,
    label VARCHAR(255) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL,
    status VARCHAR(64) NOT NULL,
    date_from TIMESTAMPTZ,
    date_to TIMESTAMPTZ,
    attributes JSONB
);
