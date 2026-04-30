-- name: SaveEvent :exec
INSERT INTO events (id, campaign_id, employee_id, type, occurred_at)
VALUES ($1, $2, $3, $4, $5);