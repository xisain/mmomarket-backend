
-- +goose Up

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    buyer_id UUID NOT NULL REFERENCES users(id),
    seller_id UUID NOT NULL REFERENCES sellers(id),
    listing_id UUID NOT NULL REFERENCES listings(id),

    status VARCHAR(32) NOT NULL DEFAULT 'pending_payment'
        CHECK (status IN (
            'pending_payment',
            'paid',
            'processing',
            'delivered',
            'completed',
            'cancelled',
            'disputed',
            'refunded'
        )),

    quantity INTEGER NOT NULL
        CHECK (quantity > 0),

    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',

    unit_price BIGINT NOT NULL
        CHECK (unit_price >= 0),

    total_amount BIGINT NOT NULL
        CHECK (total_amount >= 0),

    expires_at TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer_id
    ON orders(buyer_id);

CREATE INDEX idx_orders_seller_id
    ON orders(seller_id);

CREATE INDEX idx_orders_listing_id
    ON orders(listing_id);

CREATE INDEX idx_orders_status
    ON orders(status);

-- +goose Down

DROP TABLE IF EXISTS orders;
