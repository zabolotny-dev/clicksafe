-- name: Save :exec
INSERT INTO messages (
    id,
    label,
    from_email,
    from_name,
    subject,
    attachment_id
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: Update :exec
UPDATE messages
SET
    label = $1,
    from_email = $2,
    from_name = $3,
    subject = $4,
    attachment_id = $5
WHERE id = $6;

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
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%') AND
    (sqlc.narg('from_email')::text IS NULL OR LOWER(from_email) ILIKE '%' || LOWER(sqlc.narg('from_email')) || '%') AND
    (sqlc.narg('from_name')::text IS NULL OR LOWER(from_name) ILIKE '%' || LOWER(sqlc.narg('from_name')) || '%') AND
    (sqlc.narg('subject')::text IS NULL OR LOWER(subject) ILIKE '%' || LOWER(sqlc.narg('subject')) || '%');
