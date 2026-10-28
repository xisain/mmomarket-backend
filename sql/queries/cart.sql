-- name: createCart :one 
INSERT INTO carts (user_id)
VALUES ($1)
RETURNING *;

-- name: CreateCartItem :one
INSERT INTO cart_items (cart_id, listing_id, quantity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCart :one
SELECT
    cr.id,
    cr.user_id,
    COALESCE(
        JSONB_AGG(
            JSONB_BUILD_OBJECT(
                'id', ci.id,
                'listing_id', ls.id,
                'name', ls.name,
                'image_url', ls.image_url,
                'price', ls.price,
                'quantity', ci.quantity,
                'subtotal', ci.quantity * ls.price,
                'game', JSONB_BUILD_OBJECT(
                    'id', gm.id,
                    'name', gm.name
                ),
                'seller', JSONB_BUILD_OBJECT(
                    'id', sl.id,
                    'shop_name', sl.shop_name,
                    'slug', sl.slug
                )
            )
            ORDER BY ci.id
        ) FILTER (WHERE ci.id IS NOT NULL),
        '[]'::jsonb
    ) AS items,
    COALESCE(SUM(ci.quantity * ls.price), 0) AS total_price,
    COALESCE(SUM(ci.quantity), 0) AS total_quantity,
    cr.created_at,
    cr.updated_at
FROM carts cr
LEFT JOIN cart_items ci ON cr.id = ci.cart_id
LEFT JOIN listings ls ON ci.listing_id = ls.id
LEFT JOIN games gm ON ls.game_id = gm.id
LEFT JOIN sellers sl ON ls.seller_id = sl.id
WHERE cr.user_id = $1
GROUP BY cr.id;


