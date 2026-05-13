CREATE TABLE IF NOT EXISTS landings (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    attachment_id UUID REFERENCES attachments(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_landings_attachment_id
    ON landings(attachment_id);
