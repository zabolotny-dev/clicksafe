-- name: Save :exec
INSERT INTO targets (
    id,
    token,
    employee_id,
    campaign_id,
    status,
    scheduled_at,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: Update :exec
UPDATE targets
SET
    token = $1,
    employee_id = $2,
    campaign_id = $3,
    status = $4,
    scheduled_at = $5,
    created_at = $6
WHERE id = $7;

-- name: Delete :exec
DELETE FROM targets
WHERE id = $1;

-- name: DeleteByCampaignID :exec
DELETE FROM targets
WHERE campaign_id = $1;

-- name: QueryByID :one
SELECT * FROM targets
WHERE id = $1;

-- name: QueryByCampaignID :many
SELECT * FROM targets
WHERE campaign_id = $1
ORDER BY created_at ASC, id ASC;

-- name: QueryDue :many
SELECT * FROM targets
WHERE
    status = 'PENDING'
    AND scheduled_at IS NOT NULL
    AND scheduled_at <= $1
ORDER BY scheduled_at ASC, created_at ASC, id ASC;
