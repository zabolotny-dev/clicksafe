-- name: Save :exec
INSERT INTO campaigns (
    id,
    message_id,
    landing_id,
    education_id,
    label,
    domain,
    status,
    date_from,
    date_to,
    attributes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: Update :exec
UPDATE campaigns
SET
    message_id = $1,
    landing_id = $2,
    education_id = $3,
    label = $4,
    domain = $5,
    status = $6,
    date_from = $7,
    date_to = $8,
    attributes = $9
WHERE id = $10;

-- name: Delete :exec
DELETE FROM campaigns
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM campaigns
WHERE id = $1;

-- name: QueryExpired :many
SELECT * FROM campaigns
WHERE
    status = 'ACTIVE'
    AND date_to IS NOT NULL
    AND date_to <= NOW()
ORDER BY date_to ASC, id ASC;

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
