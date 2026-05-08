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

-- name: Query :many
SELECT * FROM targets
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
AND
    (sqlc.narg('campaign_id')::uuid IS NULL OR campaign_id = sqlc.narg('campaign_id'))
AND
    (sqlc.narg('employee_id')::uuid IS NULL OR employee_id = sqlc.narg('employee_id'))
AND
    (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
AND
    (sqlc.narg('has_schedule')::boolean IS NULL
        OR (sqlc.narg('has_schedule') = true AND scheduled_at IS NOT NULL)
        OR (sqlc.narg('has_schedule') = false AND scheduled_at IS NULL))
ORDER BY created_at ASC, id ASC;

-- name: Count :one
SELECT COUNT(*) FROM targets
WHERE
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id'))
AND
    (sqlc.narg('campaign_id')::uuid IS NULL OR campaign_id = sqlc.narg('campaign_id'))
AND
    (sqlc.narg('employee_id')::uuid IS NULL OR employee_id = sqlc.narg('employee_id'))
AND
    (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
AND
    (sqlc.narg('has_schedule')::boolean IS NULL
        OR (sqlc.narg('has_schedule') = true AND scheduled_at IS NOT NULL)
        OR (sqlc.narg('has_schedule') = false AND scheduled_at IS NULL));

-- name: QueryDue :many
SELECT targets.* FROM targets
JOIN campaigns ON campaigns.id = targets.campaign_id
WHERE
    targets.status = 'PENDING'
    AND targets.scheduled_at IS NOT NULL
    AND targets.scheduled_at <= $1
    AND campaigns.status = 'ACTIVE'
ORDER BY targets.scheduled_at ASC, targets.created_at ASC, targets.id ASC;

-- name: UpdateMany :exec
UPDATE targets
SET
    token        = data.token,
    employee_id  = data.employee_id,
    campaign_id  = data.campaign_id,
    status       = data.status,
    scheduled_at = data.scheduled_at,
    created_at   = data.created_at
FROM (
    SELECT
        UNNEST(@ids::uuid[])          AS id,
        UNNEST(@tokens::text[])       AS token,
        UNNEST(@employee_ids::uuid[]) AS employee_id,
        UNNEST(@campaign_ids::uuid[]) AS campaign_id,
        UNNEST(@statuses::text[])     AS status,
        UNNEST(@scheduled_ats::timestamptz[]) AS scheduled_at,
        UNNEST(@created_ats::timestamptz[])   AS created_at
) AS data
WHERE targets.id = data.id;

-- name: QueryByToken :one
SELECT * FROM targets
WHERE token = $1;
