
-- name: CreateGameServer :one
INSERT INTO game_servers (
    region_id,
    name,
    slug
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetGameServerByID :one
SELECT *
FROM game_servers
WHERE id = $1
LIMIT 1;


-- name: GetGameServerBySlug :one
SELECT *
FROM game_servers
WHERE region_id = $1
  AND slug = $2
LIMIT 1;


-- name: ListServersByRegion :many
SELECT *
FROM game_servers
WHERE region_id = $1
  AND is_active = TRUE
ORDER BY name ASC;


-- name: ListAllServersByRegion :many
SELECT *
FROM game_servers
WHERE region_id = $1
ORDER BY name ASC;


-- name: UpdateGameServer :one
UPDATE game_servers
SET
    name = $2,
    slug = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: SetGameServerActive :one
UPDATE game_servers
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
