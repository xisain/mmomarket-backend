-- name: CreateOrder :one
INSERT INTO orders (
buyer_id,
seller_id,
listing_id,
quantity,
unit_price,
total_amount,
expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;


-- name: BuyerGetOrder :many
SELECT * FROM orders WHERE buyer_id = $1;

-- name: BuyerGetOrderByStatus :many
SELECT * FROM orders WHERE buyer_id = $1 AND status = $2;

-- name: BuyerGetOrderByID :one
SELECT * FROM orders WHERE id = $1 AND buyer_id = $2;

-- name: UserUpdateOrder :one
UPDATE orders SET status = $3, paid_at = NOW() ,updated_at = NOW()
WHERE id = $1 AND buyer_id = $2
RETURNING *;

-- name: SellerGetOrder :many
SELECT * FROM orders WHERE seller_id = $1;

-- name: SellerGetOrderByStatus :many
SELECT * FROM orders WHERE seller_id = $1 AND status = $2;

-- name: SellerGetOrderByID :one 
SELECT * FROM orders WHERE id = $1 AND seller_id = $2;

-- name: SellerUpdateOrder :one
UPDATE orders SET status = $3
WHERE id = $1 AND seller_id = $2
RETURNING *;
