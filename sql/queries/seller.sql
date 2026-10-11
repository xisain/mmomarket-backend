-- name: CreateSeller :one
INSERT INTO sellers (
    owner_id,
    shop_name,
    slug,
    description
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: CreateSellerMember :one
INSERT INTO seller_members (
    seller_id,
    user_id,
    role
)
VALUES ($1, $2, 'owner')
RETURNING *;


-- name: GetSellerByOwnerID :one
SELECT *
FROM sellers
WHERE owner_id = $1
LIMIT 1;


-- name: GetSellerByID :one
SELECT *
FROM sellers
WHERE id = $1
LIMIT 1;


-- name: ListSellers :many
SELECT *
FROM sellers
ORDER BY created_at DESC;


-- name: ApproveSeller :one
UPDATE sellers
SET status = 'approved',
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending'
RETURNING *;


-- name: RejectSeller :one
UPDATE sellers
SET status = 'rejected',
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending'
RETURNING *;


-- name: IsGlobalAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE ur.user_id = $1
      AND r.name IN ('admin', 'super_admin')
);