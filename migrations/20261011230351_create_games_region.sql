
-- +goose Up

CREATE TABLE game_regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    game_id UUID NOT NULL
        REFERENCES games(id) ON DELETE RESTRICT,

    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_game_regions_game_slug
        UNIQUE (game_id, slug)
);

-- Index foreign key untuk lookup region per game
CREATE INDEX idx_game_regions_game_id
    ON game_regions(game_id);

-- +goose Down

DROP TABLE IF EXISTS game_regions;
