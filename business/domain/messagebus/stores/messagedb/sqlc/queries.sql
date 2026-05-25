-- name: Save :exec
INSERT INTO messages (
    id,
    type,
    label,
    from_email,
    from_name,
    subject,
    html_body_id,
    text_body_id,
    max_account_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: Update :exec
UPDATE messages
SET
    type = $1,
    label = $2,
    from_email = $3,
    from_name = $4,
    subject = $5,
    html_body_id = $6,
    text_body_id = $7,
    max_account_id = $8
WHERE id = $9;

-- name: Delete :exec
DELETE FROM messages
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM messages
WHERE id = $1;

-- name: Query :many
SELECT * FROM messages
WHERE
    -- Точный поиск по ID (если передан)
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    -- Поиск по type
    (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'))
    AND
    -- Поиск по label
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
    AND
    -- Поиск по email отправителя
    (sqlc.narg('from_email')::text IS NULL OR LOWER(from_email) ILIKE '%' || LOWER(sqlc.narg('from_email')) || '%')
    AND
    -- Поиск по имени отправителя
    (sqlc.narg('from_name')::text IS NULL OR LOWER(from_name) ILIKE '%' || LOWER(sqlc.narg('from_name')) || '%')
    AND
    -- Поиск по теме
    (sqlc.narg('subject')::text IS NULL OR LOWER(subject) ILIKE '%' || LOWER(sqlc.narg('subject')) || '%')
ORDER BY
    -- Сортировка напрямую по бизнес-константам (a=ID, b=Label, c=Email, d=FromName, e=Subject)
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC,
    CASE WHEN @order_by::text = 'b_asc' THEN label END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN label END DESC,
    CASE WHEN @order_by::text = 'c_asc' THEN from_email END ASC,
    CASE WHEN @order_by::text = 'c_desc' THEN from_email END DESC,
    CASE WHEN @order_by::text = 'd_asc' THEN from_name END ASC,
    CASE WHEN @order_by::text = 'd_desc' THEN from_name END DESC,
    CASE WHEN @order_by::text = 'e_asc' THEN subject END ASC,
    CASE WHEN @order_by::text = 'e_desc' THEN subject END DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM messages
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) AND
    (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type')) AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%') AND
    (sqlc.narg('from_email')::text IS NULL OR LOWER(from_email) ILIKE '%' || LOWER(sqlc.narg('from_email')) || '%') AND
    (sqlc.narg('from_name')::text IS NULL OR LOWER(from_name) ILIKE '%' || LOWER(sqlc.narg('from_name')) || '%') AND
    (sqlc.narg('subject')::text IS NULL OR LOWER(subject) ILIKE '%' || LOWER(sqlc.narg('subject')) || '%');

-- name: SyncAttachments :exec
WITH new_attachments AS (
    SELECT sqlc.arg('message_id')::uuid AS message_id, 
           unnest(sqlc.arg('attachment_ids')::uuid[]) AS attachment_id
),
inserted AS (
    INSERT INTO message_attachments (message_id, attachment_id)
    SELECT message_id, attachment_id FROM new_attachments
    ON CONFLICT ON CONSTRAINT message_attachments_pkey DO NOTHING
)
DELETE FROM message_attachments ma
WHERE ma.message_id = sqlc.arg('message_id')
  AND ma.attachment_id NOT IN (
      SELECT attachment_id FROM new_attachments
);

-- name: QueryAttachments :many
SELECT attachment_id FROM message_attachments WHERE message_id = $1;

-- name: QueryAttachmentsByMessageIDs :many
SELECT message_id, attachment_id FROM message_attachments
WHERE message_id = ANY(sqlc.arg('message_ids')::uuid[]);
