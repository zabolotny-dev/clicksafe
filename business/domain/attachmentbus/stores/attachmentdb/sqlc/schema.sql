CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(64) NOT NULL,
    content_path VARCHAR(255) NOT NULL,
    required_vars TEXT[],
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMPTZ NOT NULL
);
