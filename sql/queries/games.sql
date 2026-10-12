-- name: CreateGame :one
INSERT INTO games (
    name,
    slug,
    description,
    publisher
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGameBySlug :one
SELECT *
FROM games
WHERE slug = $1
LIMIT 1;

-- name: GetGameById :one
SELECT *
FROM games
WHERE id = $1
LIMIT 1;

-- name: ListActiveGames :many
SELECT *
FROM games
WHERE is_active = TRUE
ORDER BY name ASC;

-- name: ListGames :many
SELECT *
FROM games
ORDER BY name ASC;