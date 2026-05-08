CREATE TABLE IF NOT EXISTS landings (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    content_path VARCHAR(255),
    required_vars TEXT[]
);
