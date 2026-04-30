CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    logo_url VARCHAR(255),
    attributes JSONB
);