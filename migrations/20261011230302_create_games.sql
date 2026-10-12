
-- +goose Up

CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    publisher VARCHAR(255),

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk pencarian/filter game aktif
CREATE INDEX idx_games_active_name
ON games(name)
WHERE is_active = TRUE;

-- +goose Down

DROP TABLE IF EXISTS games;
