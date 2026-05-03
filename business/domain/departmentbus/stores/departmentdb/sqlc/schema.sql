CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    attributes JSONB
);
