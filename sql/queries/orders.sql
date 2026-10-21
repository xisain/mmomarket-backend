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
SELECT
    ord.*,
    jsonb_build_object(
        'id', sl.id,
        'shop_name', sl.shop_name,
        'slug', sl.slug
    ) AS seller,

    jsonb_build_object(
      'id',
      ls.id,
      'name',
      ls.name,
      'slug',
      ls.slug,
      'description',
      ls.description,
      'image_url',
      ls.image_url,
      'price',
      ls.price,
      'game', jsonb_build_object('id', gm.id, 'name', gm.name ),
      'category', jsonb_build_object('id', lc.id, 'name', lc.name),
      'region', jsonb_build_object('id', gr.id, 'name', gr.name),
      'server', jsonb_build_object('id',gs.id, 'name',gs.name)

    ) AS listing

FROM orders AS ord
JOIN sellers AS sl
    ON ord.seller_id = sl.id
JOIN listings AS ls
    ON ord.listing_id = ls.id
JOIN games gm 
    ON ls.game_id = gm.id
JOIN listing_categorys lc 
    ON ls.category_id = lc.id
JOIN game_regions gr 
    ON ls.game_region_id =gr.id
LEFT JOIN game_servers gs 
    ON gs.id = ls.server_id
WHERE ord.buyer_id = $1
ORDER BY ord.created_at DESC, ord.id DESC;

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

-- name: ListBuyerOrders :many
SELECT ord.id,
       ord.buyer_id,
       ord.seller_id,
       ord.listing_id,
       ord.status,
       ord.quantity,
       ord.currency,
       ord.unit_price,
       ord.total_amount,
       ord.expires_at,
       ord.paid_at,
       ord.completed_at,
       ord.created_at,
       ord.updated_at,
       jsonb_build_object(
         'id',
         sl.id,
         'owner_id',
         sl.owner_id,
         'shop_name',
         sl.shop_name,
         'slug',
         sl.slug,
         'description',
         sl.description
       ) AS seller,
       jsonb_build_object(
         'id',
         ls.id,
         'name',
         ls.name,
         'slug',
         ls.slug,
         'description',
         ls.description,
         'image_url',
         ls.image_url,
         'price',
         ls.price,
         'game',
         jsonb_build_object('id', gm.id, 'game_name', gm.name ),
         'category', JSON_BUILD_OBJECT('id', lc.id, 'categori_name', lc.name)
       ) AS listing
FROM orders ord

JOIN sellers sl ON ord.seller_id = sl.id
    JOIN listings ls ON ord.listing_id = ls.id
    JOIN games gm ON ls.game_id = gm.id
    JOIN listing_categorys lc ON ls.category_id = lc.id
WHERE ord.buyer_id = sqlc.arg(buyer_id)

ORDER BY ord.created_at DESC

LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);
