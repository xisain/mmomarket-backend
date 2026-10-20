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

-- name: ListGameCategories :many
SELECT
    g.id AS game_id,
    g.name AS game_name,
    g.slug AS game_slug,
    lc.id AS category_id,
    lc.name AS category_name,
    lc.slug AS category_slug
FROM games g
CROSS JOIN listing_categorys lc
ORDER BY g.name, lc.name;


-- name: GetCategoriesByGameSlug :many
SELECT
    lc.id AS category_id,
    lc.name AS category_name,
    lc.slug AS category_slug
FROM games g
CROSS JOIN listing_categorys lc
WHERE g.slug = sqlc.arg(game_slug)
ORDER BY lc.name;