-- name: Save :exec
INSERT INTO departments (id, label, attributes)
VALUES ($1, $2, $3);

-- name: SaveMany :copyfrom
INSERT INTO departments (id, label, attributes)
VALUES ($1, $2, $3);

-- name: Update :exec
UPDATE departments
SET label = $1, attributes = $2
WHERE id = $3;

-- name: Delete :exec
DELETE FROM departments
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM departments
WHERE id = $1;

-- name: Query :many
SELECT * FROM departments
WHERE 
    -- Фильтр по ID: если параметр не передан (NULL), условие игнорируется
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) 
    AND
    -- Полнотекстовый поиск по label (без учета регистра)
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%')
ORDER BY 
    -- Хитрый хак для динамической сортировки (вспоминаем, что a = id, b = label)
    CASE WHEN @order_by::text = 'b_asc' THEN label END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN label END DESC,
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM departments
WHERE 
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) AND
    (sqlc.narg('label')::text IS NULL OR LOWER(label) ILIKE '%' || LOWER(sqlc.narg('label')) || '%');
