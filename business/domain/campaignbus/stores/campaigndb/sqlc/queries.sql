-- name: Save :exec
INSERT INTO campaigns (
    id,
    message_id,
    label,
    status,
    date_from,
    date_to,
    attributes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: Update :exec
UPDATE campaigns
SET
    message_id = $1,
    label = $2,
    status = $3,
    date_from = $4,
    date_to = $5,
    attributes = $6
WHERE id = $7;

-- name: Delete :exec
DELETE FROM campaigns
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM campaigns
WHERE id = $1;

-- name: Query :many
SELECT * FROM campaigns
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
    AND
    (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
    AND
    (sqlc.narg('date_from')::timestamptz IS NULL OR date_to > sqlc.narg('date_from'))
    AND
    (sqlc.narg('date_to')::timestamptz IS NULL OR date_from < sqlc.narg('date_to'))
ORDER BY
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC,
    CASE WHEN @order_by::text = 'b_asc' THEN label END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN label END DESC,
    CASE WHEN @order_by::text = 'c_asc' THEN status END ASC,
    CASE WHEN @order_by::text = 'c_desc' THEN status END DESC,
    CASE WHEN @order_by::text = 'd_asc' THEN date_from END ASC,
    CASE WHEN @order_by::text = 'd_desc' THEN date_from END DESC,
    CASE WHEN @order_by::text = 'e_asc' THEN date_to END ASC,
    CASE WHEN @order_by::text = 'e_desc' THEN date_to END DESC,
    id DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM campaigns
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%') AND
    (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('date_from')::timestamptz IS NULL OR date_to > sqlc.narg('date_from')) AND
    (sqlc.narg('date_to')::timestamptz IS NULL OR date_from < sqlc.narg('date_to'));
