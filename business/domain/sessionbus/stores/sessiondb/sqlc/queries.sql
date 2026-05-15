-- name: Save :exec
INSERT INTO sessions (
    id,
    admin_id,
    token_hash,
    csrf_token,
    created_at,
    expires_at,
    revoked_at,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: QueryByTokenHash :one
SELECT * FROM sessions
WHERE token_hash = $1;

-- name: Revoke :exec
UPDATE sessions
SET revoked_at = $2
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeByAdminID :exec
UPDATE sessions
SET revoked_at = $2
WHERE admin_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpired :exec
DELETE FROM sessions
WHERE expires_at <= $1;
