-- name: Save :exec
INSERT INTO organizations (id, name, attributes)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name, attributes = EXCLUDED.attributes;

-- name: GetByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: UpdateLogo :exec
UPDATE organizations
SET logo_url = $1
WHERE id = $2;
