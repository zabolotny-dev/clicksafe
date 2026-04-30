-- name: Save :exec
INSERT INTO organizations (id, name, logo_url, attributes)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name, logo_url = EXCLUDED.logo_url, attributes = EXCLUDED.attributes;

-- name: GetByID :one
SELECT * FROM organizations WHERE id = $1;