-- name: Save :exec
INSERT INTO admins (
    id,
    first_name,
    last_name,
    login,
    password_hash,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: Update :exec
UPDATE admins
SET
    first_name = $1,
    last_name = $2,
    login = $3,
    password_hash = $4
WHERE id = $5;

-- name: QueryByID :one
SELECT * FROM admins
WHERE id = $1;

-- name: QueryByLogin :one
SELECT * FROM admins
WHERE login = $1;

-- name: Query :many
SELECT * FROM admins
WHERE
    (sqlc.narg('login')::text IS NULL OR LOWER(login) ILIKE '%' || LOWER(sqlc.narg('login')) || '%')
    AND
    (sqlc.narg('full_name')::text IS NULL OR LOWER(first_name || ' ' || last_name) ILIKE '%' || LOWER(sqlc.narg('full_name')) || '%');

-- name: Count :one
SELECT COUNT(*) FROM admins
WHERE
    (sqlc.narg('login')::text IS NULL OR LOWER(login) ILIKE '%' || LOWER(sqlc.narg('login')) || '%')
    AND
    (sqlc.narg('full_name')::text IS NULL OR LOWER(first_name || ' ' || last_name) ILIKE '%' || LOWER(sqlc.narg('full_name')) || '%');
