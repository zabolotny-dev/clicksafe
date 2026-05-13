CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    subject VARCHAR(255),
    attachment_id UUID REFERENCES attachments(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_messages_attachment_id ON messages(attachment_id);