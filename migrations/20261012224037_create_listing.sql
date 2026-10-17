-- +goose Up

CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    game_id UUID NOT NULL
        REFERENCES games(id) ON DELETE RESTRICT,

    category_id UUID NOT NULL
        REFERENCES listing_categorys(id) ON DELETE RESTRICT,

    game_region_id UUID
        REFERENCES game_regions(id) ON DELETE RESTRICT,

    server_id UUID
        REFERENCES game_servers(id) ON DELETE RESTRICT,

    seller_id UUID NOT NULL
        REFERENCES sellers(id) ON DELETE RESTRICT,

    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    image_url TEXT,

    price NUMERIC(12, 2) NOT NULL
        CHECK (price >= 0),

    stock_type VARCHAR(20) NOT NULL DEFAULT 'finite'
        CHECK (stock_type IN ('finite', 'unlimited')),

    stock_quantity BIGINT,

    quantity_per_unit BIGINT NOT NULL DEFAULT 1
        CHECK (quantity_per_unit > 0),

    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (
            status IN ('draft', 'active', 'sold', 'archived')
        ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_listing_stock CHECK (
        (
            stock_type = 'finite'
            AND stock_quantity IS NOT NULL
            AND stock_quantity >= 0
        )
        OR
        (
            stock_type = 'unlimited'
            AND stock_quantity IS NULL
        )
    ),

    CONSTRAINT server_requires_region CHECK (
        server_id IS NULL
        OR game_region_id IS NOT NULL
    )
);

CREATE INDEX idx_listings_game_category
    ON listings(game_id, category_id);

CREATE INDEX idx_listings_game_region
    ON listings(game_id, game_region_id);

CREATE INDEX idx_listings_server
    ON listings(server_id);

CREATE INDEX idx_listings_seller
    ON listings(seller_id);

CREATE INDEX idx_listings_status
    ON listings(status);

-- +goose Down

DROP TABLE IF EXISTS listings;