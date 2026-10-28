-- +goose Up

CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL UNIQUE
        REFERENCES users(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    cart_id UUID NOT NULL
        REFERENCES carts(id) ON DELETE CASCADE,

    listing_id UUID NOT NULL
        REFERENCES listings(id) ON DELETE RESTRICT,

    quantity INT NOT NULL DEFAULT 1
        CHECK (quantity > 0),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT cart_items_cart_listing_unique
        UNIQUE (cart_id, listing_id)
);

CREATE INDEX idx_cart_items_listing_id
    ON cart_items(listing_id);


-- +goose Down

DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;