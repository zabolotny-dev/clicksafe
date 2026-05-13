-- name: Save :exec
INSERT INTO attachments (
    id,
    label,
    type,
    content_path,
    required_vars,
    is_public,
    uploaded_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: Update :exec
UPDATE attachments
SET
    label = $1,
    type = $2,
    content_path = $3,
    required_vars = $4,
    is_public = $5,
    uploaded_at = $6
WHERE id = $7;

-- name: Delete :exec
DELETE FROM attachments
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM attachments
WHERE id = $1;

-- name: Query :many
SELECT * FROM attachments
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
    AND
    (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'))
ORDER BY
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC,
    CASE WHEN @order_by::text = 'b_asc' THEN label END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN label END DESC,
    CASE WHEN @order_by::text = 'c_asc' THEN type END ASC,
    CASE WHEN @order_by::text = 'c_desc' THEN type END DESC,
    CASE WHEN @order_by::text = 'd_asc' THEN uploaded_at END ASC,
    CASE WHEN @order_by::text = 'd_desc' THEN uploaded_at END DESC,
    id DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM attachments
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
    AND
    (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'));
