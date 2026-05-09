-- name: Query :many
WITH matched_targets AS (
    SELECT
        t.id,
        t.token,
        t.campaign_id,
        t.employee_id,
        e.first_name,
        e.last_name,
        t.status,
        t.scheduled_at,
        t.created_at
    FROM targets t
    JOIN employees e ON e.id = t.employee_id
    WHERE
        (sqlc.narg('id')::uuid IS NULL OR t.id = sqlc.narg('id'))
        AND
        (sqlc.narg('campaign_id')::uuid IS NULL OR t.campaign_id = sqlc.narg('campaign_id'))
        AND
        (sqlc.narg('employee_id')::uuid IS NULL OR t.employee_id = sqlc.narg('employee_id'))
        AND
        (sqlc.narg('status')::text IS NULL OR t.status = sqlc.narg('status'))
        AND
        (sqlc.narg('full_name')::text IS NULL OR LOWER(e.first_name || ' ' || e.last_name) ILIKE '%' || LOWER(sqlc.narg('full_name')) || '%')
    ORDER BY
        CASE WHEN @order_by::text = 'a_asc' THEN t.id::text END ASC,
        CASE WHEN @order_by::text = 'a_desc' THEN t.id::text END DESC,
        CASE WHEN @order_by::text = 'b_asc' THEN e.first_name END ASC,
        CASE WHEN @order_by::text = 'b_desc' THEN e.first_name END DESC,
        CASE WHEN @order_by::text = 'c_asc' THEN e.last_name END ASC,
        CASE WHEN @order_by::text = 'c_desc' THEN e.last_name END DESC,
        CASE WHEN @order_by::text = 'd_asc' THEN t.status END ASC,
        CASE WHEN @order_by::text = 'd_desc' THEN t.status END DESC,
        t.id DESC
    LIMIT @limit_val OFFSET @offset_val
)
SELECT
    mt.id,
    mt.token,
    mt.campaign_id,
    mt.employee_id,
    mt.first_name,
    mt.last_name,
    mt.status,
    mt.scheduled_at,
    mt.created_at,
    ev.type AS event_type,
    ev.ip_address AS event_ip_address,
    ev.user_agent AS event_user_agent,
    ev.referer AS event_referer,
    ev.occurred_at AS event_occurred_at
FROM matched_targets mt
LEFT JOIN events ev
    ON ev.campaign_id = mt.campaign_id
    AND ev.employee_id = mt.employee_id
ORDER BY
    CASE WHEN @order_by::text = 'a_asc' THEN mt.id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN mt.id::text END DESC,
    CASE WHEN @order_by::text = 'b_asc' THEN mt.first_name END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN mt.first_name END DESC,
    CASE WHEN @order_by::text = 'c_asc' THEN mt.last_name END ASC,
    CASE WHEN @order_by::text = 'c_desc' THEN mt.last_name END DESC,
    CASE WHEN @order_by::text = 'd_asc' THEN mt.status END ASC,
    CASE WHEN @order_by::text = 'd_desc' THEN mt.status END DESC,
    mt.id DESC,
    ev.occurred_at DESC,
    ev.id DESC;

-- name: Count :one
SELECT COUNT(*) FROM targets t
JOIN employees e ON e.id = t.employee_id
WHERE
    (sqlc.narg('id')::uuid IS NULL OR t.id = sqlc.narg('id'))
    AND
    (sqlc.narg('campaign_id')::uuid IS NULL OR t.campaign_id = sqlc.narg('campaign_id'))
    AND
    (sqlc.narg('employee_id')::uuid IS NULL OR t.employee_id = sqlc.narg('employee_id'))
    AND
    (sqlc.narg('status')::text IS NULL OR t.status = sqlc.narg('status'))
    AND
    (sqlc.narg('full_name')::text IS NULL OR LOWER(e.first_name || ' ' || e.last_name) ILIKE '%' || LOWER(sqlc.narg('full_name')) || '%');
