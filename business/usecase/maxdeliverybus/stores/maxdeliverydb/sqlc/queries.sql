-- name: SaveSent :exec
INSERT INTO max_deliveries (
    id, target_id, campaign_id, employee_id, max_account_id,
    adapter_account_id, chat_id, message_id, sent_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (target_id) DO UPDATE
SET adapter_account_id = EXCLUDED.adapter_account_id,
    chat_id = EXCLUDED.chat_id,
    message_id = EXCLUDED.message_id,
    sent_at = EXCLUDED.sent_at,
    updated_at = EXCLUDED.updated_at;

-- name: QueryByID :one
SELECT * FROM max_deliveries WHERE id = $1;

-- name: QueryByMessage :one
SELECT * FROM max_deliveries
WHERE adapter_account_id = $1 AND chat_id = $2 AND message_id = $3;

-- name: QueryLatestByChat :one
SELECT * FROM max_deliveries
WHERE adapter_account_id = $1 AND chat_id = $2
ORDER BY sent_at DESC, created_at DESC
LIMIT 1;

-- name: QueryLatestUnreadByChat :one
SELECT * FROM max_deliveries
WHERE adapter_account_id = $1 AND chat_id = $2 AND read_at IS NULL
ORDER BY sent_at DESC, created_at DESC
LIMIT 1;

-- name: MarkRead :execresult
UPDATE max_deliveries
SET read_at = $2, updated_at = $2
WHERE id = $1 AND read_at IS NULL;

-- name: MarkReplied :execresult
UPDATE max_deliveries
SET replied_at = $2,
    incoming_message_id = $3,
    updated_at = $2
WHERE id = $1 AND replied_at IS NULL;

-- name: MarkEducationSent :exec
UPDATE max_deliveries
SET education_sent_at = COALESCE(education_sent_at, $2),
    updated_at = $2
WHERE id = $1;

-- name: IsProcessed :one
SELECT EXISTS(SELECT 1 FROM max_adapter_processed_events WHERE seq = $1);

-- name: MarkProcessed :exec
INSERT INTO max_adapter_processed_events (seq, processed_at)
VALUES ($1, $2)
ON CONFLICT (seq) DO NOTHING;
