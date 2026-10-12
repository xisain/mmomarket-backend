
-- +goose Up

CREATE TABLE game_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    region_id UUID NOT NULL
        REFERENCES game_regions(id) ON DELETE RESTRICT,

    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_game_servers_region_slug
        UNIQUE (region_id, slug)
);

-- Index untuk lookup server per region
CREATE INDEX idx_game_servers_region_id
    ON game_servers(region_id);

-- +goose Down

DROP TABLE IF EXISTS game_servers;
