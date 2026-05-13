-- name: Save :exec
INSERT INTO landings (
    id,
    label,
    attachment_id
) VALUES (
    $1, $2, $3
);

-- name: Update :exec
UPDATE landings
SET
    label = $1,
    attachment_id = $2
WHERE id = $3;

-- name: Delete :exec
DELETE FROM landings
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM landings
WHERE id = $1;

-- name: Query :many
SELECT * FROM landings
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
ORDER BY
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC,
    CASE WHEN @order_by::text = 'b_asc' THEN label END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN label END DESC,
    id DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM landings
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%');
