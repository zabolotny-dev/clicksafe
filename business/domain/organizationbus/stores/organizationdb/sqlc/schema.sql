CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    attachment_id UUID REFERENCES attachments(id) ON DELETE RESTRICT,
    attributes JSONB
);

CREATE INDEX IF NOT EXISTS idx_organizations_attachment_id ON organizations(attachment_id);