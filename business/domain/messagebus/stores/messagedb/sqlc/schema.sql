CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    subject VARCHAR(255),
    html_body_id UUID REFERENCES attachments(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_messages_html_body_id ON messages(html_body_id);

CREATE TABLE IF NOT EXISTS message_attachments (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    attachment_id UUID NOT NULL REFERENCES attachments(id) ON DELETE RESTRICT,
    PRIMARY KEY (message_id, attachment_id)
);

CREATE INDEX IF NOT EXISTS idx_message_attachments_attachment_id ON message_attachments(attachment_id);
