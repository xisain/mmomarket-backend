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


-- name: UpdateUserStatus :exec
UPDATE users
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;