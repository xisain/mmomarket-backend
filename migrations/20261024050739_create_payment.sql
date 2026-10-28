
-- +goose Up

CREATE TABLE payment_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_payment_providers_code
        CHECK (code ~ '^[a-z][a-z0-9_]*$')
);



CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL
        REFERENCES orders(id) ON DELETE RESTRICT,

    provider_id UUID NOT NULL
        REFERENCES payment_providers(id)
        ON DELETE RESTRICT,

    -- Generated internally by MMOMARKET
    merchant_reference VARCHAR(100) NOT NULL UNIQUE,

    -- Returned by the payment gateway
    provider_payment_id VARCHAR(255),

    amount BIGINT NOT NULL
        CHECK (amount > 0),

    currency VARCHAR(3) NOT NULL DEFAULT 'IDR'
        CHECK (currency ~ '^[A-Z]{3}$'),

    status VARCHAR(32) NOT NULL DEFAULT 'pending'
        CHECK (status IN (
            'pending',
            'processing',
            'succeeded',
            'failed',
            'expired',
            'cancelled'
        )),

    payment_method VARCHAR(100),

    -- Hosted checkout URL (if provided)
    checkout_url TEXT,

    expires_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,

    failure_code VARCHAR(100),
    failure_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_payments_provider_payment
        UNIQUE (provider_id, provider_payment_id),

    CONSTRAINT chk_payments_paid_at
        CHECK (
            status <> 'succeeded'
            OR paid_at IS NOT NULL
        )
);

-- Payment history by order
CREATE INDEX idx_payments_order_id
    ON payments(order_id, created_at DESC);

-- Query payment by provider
CREATE INDEX idx_payments_provider_id
    ON payments(provider_id);

-- Expiration job
CREATE INDEX idx_payments_expiration
    ON payments(expires_at)
    WHERE status IN ('pending', 'processing');

-- One successful payment per order
CREATE UNIQUE INDEX uq_payments_order_succeeded
    ON payments(order_id)
    WHERE status = 'succeeded';


CREATE TABLE payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    payment_id UUID
        REFERENCES payments(id) ON DELETE RESTRICT,

    provider_id UUID NOT NULL
        REFERENCES payment_providers(id)
        ON DELETE RESTRICT,

    provider_event_id VARCHAR(255),

    event_type VARCHAR(100) NOT NULL,

    -- Raw verified provider event data
    payload JSONB NOT NULL,

    status VARCHAR(32) NOT NULL DEFAULT 'received'
        CHECK (status IN (
            'received',
            'processed',
            'ignored',
            'failed'
        )),

    error_message TEXT,

    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_payment_events_provider_event
    ON payment_events(
        provider_id,
        provider_event_id
    )
    WHERE provider_event_id IS NOT NULL;

CREATE INDEX idx_payment_events_payment_id
    ON payment_events(payment_id, received_at DESC);

CREATE INDEX idx_payment_events_unprocessed
    ON payment_events(received_at)
    WHERE status IN ('received', 'failed');


CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    payment_id UUID NOT NULL
        REFERENCES payments(id) ON DELETE RESTRICT,

    -- Internal refund reference
    merchant_reference VARCHAR(100) NOT NULL UNIQUE,

    -- Returned by provider
    provider_refund_id VARCHAR(255),

    amount BIGINT NOT NULL
        CHECK (amount > 0),

    currency VARCHAR(3) NOT NULL DEFAULT 'IDR'
        CHECK (currency ~ '^[A-Z]{3}$'),

    status VARCHAR(32) NOT NULL DEFAULT 'pending'
        CHECK (status IN (
            'pending',
            'processing',
            'succeeded',
            'failed',
            'cancelled'
        )),

    reason TEXT NOT NULL,
    failure_code VARCHAR(100),
    failure_message TEXT,

    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_refunds_completed_at
        CHECK (
            status <> 'succeeded'
            OR completed_at IS NOT NULL
        )
);

CREATE INDEX idx_refunds_payment_id
    ON refunds(payment_id, created_at DESC);

CREATE INDEX idx_refunds_status
    ON refunds(status, created_at);


CREATE TABLE refund_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    refund_id UUID
        REFERENCES refunds(id) ON DELETE RESTRICT,

    provider_id UUID NOT NULL
        REFERENCES payment_providers(id)
        ON DELETE RESTRICT,

    provider_event_id VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,

    payload JSONB NOT NULL,

    status VARCHAR(32) NOT NULL DEFAULT 'received'
        CHECK (status IN (
            'received',
            'processed',
            'ignored',
            'failed'
        )),

    error_message TEXT,

    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_refund_events_provider_event
    ON refund_events(
        provider_id,
        provider_event_id
    )
    WHERE provider_event_id IS NOT NULL;

CREATE INDEX idx_refund_events_refund_id
    ON refund_events(refund_id, received_at DESC);

CREATE INDEX idx_refund_events_unprocessed
    ON refund_events(received_at)
    WHERE status IN ('received', 'failed');


INSERT INTO payment_providers (code, name)
VALUES
    ('midtrans', 'Midtrans'),
    ('xendit', 'Xendit');


-- +goose Down

DROP TABLE IF EXISTS refund_events;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS payment_providers;
