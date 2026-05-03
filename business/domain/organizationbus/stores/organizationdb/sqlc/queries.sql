-- name: Save :exec
INSERT INTO organizations (id, label, logo_path, attributes)
VALUES ($1, $2, $3, $4);

-- name: Update :exec
UPDATE organizations
SET label = $1, logo_path = $2, attributes = $3
WHERE id = $4;

-- name: QueryByID :one
SELECT * FROM organizations WHERE id = $1;
