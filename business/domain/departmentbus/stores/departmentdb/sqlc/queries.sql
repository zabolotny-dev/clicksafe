-- name: Save :exec
INSERT INTO departments (id, name, attributes)
VALUES ($1, $2, $3);

-- name: Update :exec
UPDATE departments
SET name = $1, attributes = $2
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
    -- Полнотекстовый поиск по Имени (без учета регистра)
    (sqlc.narg('name')::text IS NULL OR LOWER(name) ILIKE '%' || LOWER(sqlc.narg('name')) || '%')
ORDER BY 
    -- Хитрый хак для динамической сортировки (вспоминаем, что a = id, b = name)
    CASE WHEN @order_by::text = 'b_asc' THEN name END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN name END DESC,
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM departments
WHERE 
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) AND
    (sqlc.narg('name')::text IS NULL OR LOWER(name) ILIKE '%' || LOWER(sqlc.narg('name')) || '%');

