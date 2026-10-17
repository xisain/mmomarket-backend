-- name: CreateListingCategory :one
INSERT INTO listing_categorys (name, slug, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetListingCategoryByID :one
SELECT * FROM listing_categorys WHERE id = $1 LIMIT 1;

-- name: ListListingCategories :many
SELECT * FROM listing_categorys ORDER BY name ASC, id ASC;

-- name: UpdateListingCategory :one
UPDATE listing_categorys
SET name = $2, slug = $3, description = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteListingCategory :execrows
DELETE FROM listing_categorys WHERE id = $1;
