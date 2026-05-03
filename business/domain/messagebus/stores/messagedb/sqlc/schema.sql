CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    subject VARCHAR(255),
    content_path VARCHAR(255),
    required_vars TEXT[]
);
