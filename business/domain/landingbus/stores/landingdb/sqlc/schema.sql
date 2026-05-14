CREATE TABLE IF NOT EXISTS landings (
    id UUID PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    html_body_id UUID REFERENCES attachments(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_landings_html_body_id ON landings(html_body_id);
