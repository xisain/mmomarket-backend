
-- name: CreateGameRegion :one
INSERT INTO game_regions (
    game_id,
    name,
    slug
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetGameRegionByID :one
SELECT *
FROM game_regions
WHERE id = $1
LIMIT 1;


-- name: GetGameRegionBySlug :one
SELECT *
FROM game_regions
WHERE game_id = $1
  AND slug = $2
LIMIT 1;


-- name: ListRegionsByGame :many
SELECT *
FROM game_regions
WHERE game_id = $1
  AND is_active = TRUE
ORDER BY name ASC;


-- name: ListAllRegionsByGame :many
SELECT *
FROM game_regions
WHERE game_id = $1
ORDER BY name ASC;


-- name: UpdateGameRegion :one
UPDATE game_regions
SET
    name = $2,
    slug = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: SetGameRegionActive :one
UPDATE game_regions
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
