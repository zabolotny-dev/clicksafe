-- name: Save :exec
INSERT INTO organizations (id, label, attributes)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET label = EXCLUDED.label, attributes = EXCLUDED.attributes;

-- name: QueryByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: UpdateLogo :exec
UPDATE organizations
SET logo_path = $1
WHERE id = $2;
