-- name: Save :exec
INSERT INTO employees (
    id, 
    department_id, 
    first_name, 
    last_name, 
    email, 
    phone_number, 
    attributes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: Update :exec
UPDATE employees
SET 
    department_id = $1, 
    first_name = $2, 
    last_name = $3, 
    email = $4, 
    phone_number = $5, 
    attributes = $6
WHERE id = $7;

-- name: Delete :exec
DELETE FROM employees
WHERE id = $1;

-- name: QueryByID :one
SELECT * FROM employees
WHERE id = $1;

-- name: Query :many
SELECT * FROM employees
WHERE 
    -- Точный поиск по ID (если передан)
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) 
    AND
    -- Поиск по отделу
    (sqlc.narg('department_id')::uuid IS NULL OR department_id = sqlc.narg('department_id'))
    AND
    -- Умный поиск по имени и фамилии одной строкой (MVP вариант)
    (sqlc.narg('full_name')::text IS NULL OR LOWER(first_name || ' ' || last_name) ILIKE '%' || LOWER(sqlc.narg('full_name')) || '%')
    AND
    -- Поиск по email
    (sqlc.narg('email')::text IS NULL OR LOWER(email) ILIKE '%' || LOWER(sqlc.narg('email')) || '%')
    AND
    -- Поиск по телефону
    (sqlc.narg('phone_number')::text IS NULL OR phone_number ILIKE '%' || sqlc.narg('phone_number') || '%')
ORDER BY 
    -- Сортировка напрямую по бизнес-константам (a=ID, b=FirstName, c=LastName, d=Email, e=Phone)
    CASE WHEN @order_by::text = 'a_asc' THEN id::text END ASC,
    CASE WHEN @order_by::text = 'a_desc' THEN id::text END DESC,
    CASE WHEN @order_by::text = 'b_asc' THEN first_name END ASC,
    CASE WHEN @order_by::text = 'b_desc' THEN first_name END DESC,
    CASE WHEN @order_by::text = 'c_asc' THEN last_name END ASC,
    CASE WHEN @order_by::text = 'c_desc' THEN last_name END DESC,
    CASE WHEN @order_by::text = 'd_asc' THEN email END ASC,
    CASE WHEN @order_by::text = 'd_desc' THEN email END DESC,
    CASE WHEN @order_by::text = 'e_asc' THEN phone_number END ASC,
    CASE WHEN @order_by::text = 'e_desc' THEN phone_number END DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: Count :one
SELECT COUNT(*) FROM employees
WHERE 
    (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')) AND
    (sqlc.narg('department_id')::uuid IS NULL OR department_id = sqlc.narg('department_id')) AND
    (sqlc.narg('full_name')::text IS NULL OR LOWER(first_name || ' ' || last_name) ILIKE '%' || LOWER(sqlc.narg('full_name')) || '%') AND
    (sqlc.narg('email')::text IS NULL OR LOWER(email) ILIKE '%' || LOWER(sqlc.narg('email')) || '%') AND
    (sqlc.narg('phone_number')::text IS NULL OR phone_number ILIKE '%' || sqlc.narg('phone_number') || '%');