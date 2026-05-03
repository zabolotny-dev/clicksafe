CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    logo_path VARCHAR(255),
    attributes JSONB
);