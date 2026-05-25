-- name: Save :one
INSERT INTO max_accounts (
    id, adapter_id, phone_number, label, status, max_user_id,
    last_error, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (adapter_id) DO UPDATE
SET phone_number = EXCLUDED.phone_number,
    status = EXCLUDED.status,
    max_user_id = EXCLUDED.max_user_id,
    last_error = EXCLUDED.last_error,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: UpdateLabel :one
UPDATE max_accounts
SET label = $2,
    updated_at = $3
WHERE id = $1
RETURNING *;

-- name: QueryByID :one
SELECT * FROM max_accounts
WHERE id = $1;

-- name: QueryByAdapterID :one
SELECT * FROM max_accounts
WHERE adapter_id = $1;

-- name: Delete :execresult
DELETE FROM max_accounts
WHERE id = $1;

-- name: Query :many
SELECT * FROM max_accounts
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
    AND
    (sqlc.narg('phone_number')::text IS NULL OR phone_number ILIKE '%' || sqlc.narg('phone_number') || '%')
ORDER BY created_at DESC, id DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM max_accounts
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
    AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
    AND
    (sqlc.narg('phone_number')::text IS NULL OR phone_number ILIKE '%' || sqlc.narg('phone_number') || '%');
