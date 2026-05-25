-- Minimal ClickSafe seed.
-- Tooling creates the demo administrator before this script.

INSERT INTO organizations (id, label, attachment_id, attributes)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'ClickSafe Demo Organization',
    NULL,
    '{"seed":"tooling","scenario":"minimal"}'::jsonb
)
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    attachment_id = EXCLUDED.attachment_id,
    attributes = EXCLUDED.attributes;
