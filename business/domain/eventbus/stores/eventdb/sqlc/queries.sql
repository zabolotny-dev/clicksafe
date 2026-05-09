-- name: SaveEvent :exec
INSERT INTO events (
	id, 
	campaign_id, 
	employee_id, 
	type, 
	ip_address, 
	user_agent, 
	referer, 
	occurred_at
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
);