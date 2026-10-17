-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    display_name
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;


-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE LOWER(email) = LOWER($1)
LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC;


-- name: UpdateUserStatus :exec
UPDATE users
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;


-- name: IsUserAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE ur.user_id = $1
      AND r.name IN ('admin', 'super_admin')
);
