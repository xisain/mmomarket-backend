-- name: CreateListing :one
INSERT INTO listings (
    game_id,
    category_id,
    game_region_id,
    server_id,
    seller_id,
    name,
    slug,
    description,
    image_url,
    price,
    stock_type,
    stock_quantity,
    quantity_per_unit,
    status
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14
)
RETURNING *;

-- name: GetListingByID :one
SELECT * FROM listings WHERE id = $1 LIMIT 1;

-- name: GetListingBySlug :one
SELECT * FROM listings WHERE slug = $1 LIMIT 1;

-- name: ListListings :many
SELECT * FROM listings
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: ListListingsByGameAndCategory :many
SELECT * FROM listings
WHERE game_id = $1 AND category_id = $2
ORDER BY created_at DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: ListListingsByGameAndRegion :many
SELECT * FROM listings
WHERE game_id = $1 AND game_region_id = $2
ORDER BY created_at DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: ListListingsByServer :many
SELECT * FROM listings
WHERE server_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: ListListingsBySeller :many
SELECT * FROM listings
WHERE seller_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: UpdateListing :one
UPDATE listings
SET game_id = $3,
    category_id = $4,
    game_region_id = $5,
    server_id = $6,
    name = $7,
    slug = $8,
    description = $9,
    image_url = $10,
    price = $11,
    updated_at = NOW()
WHERE id = $1 AND seller_id = $2
RETURNING *;

-- name: DeleteListing :execrows
DELETE FROM listings WHERE id = $1 AND seller_id = $2;

-- name: IsSellerOwnedByUser :one
SELECT EXISTS (
    SELECT 1
    FROM sellers
    WHERE id = sqlc.arg(seller_id)
      AND owner_id = sqlc.arg(user_id)
);