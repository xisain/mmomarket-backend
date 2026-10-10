-- +goose Up

CREATE TABLE sellers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    owner_id UUID NOT NULL REFERENCES users(id),

    shop_name VARCHAR(150) NOT NULL,
    slug VARCHAR(180) NOT NULL UNIQUE,

    description TEXT,

    status VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN (
            'pending',
            'approved',
            'rejected',
            'suspended'
        )),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE seller_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    seller_id UUID NOT NULL
        REFERENCES sellers(id) ON DELETE CASCADE,

    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    role VARCHAR(30) NOT NULL
        CHECK (role IN ('owner', 'manager', 'staff', 'booster')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (seller_id, user_id)
);

CREATE INDEX idx_sellers_owner_id
ON sellers(owner_id);

CREATE INDEX idx_seller_members_user_id
ON seller_members(user_id);

-- +goose Down

DROP TABLE seller_members;
DROP TABLE sellers;