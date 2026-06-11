CREATE TABLE IF NOT EXISTS market_listings (
    id                  TEXT PRIMARY KEY,
    seller_account_id   TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    stash_item_id        TEXT NOT NULL,
    source_character_id  TEXT,
    item_def_id          TEXT NOT NULL,
    rolled_stats         JSONB NOT NULL DEFAULT '{}'::jsonb,
    status              TEXT NOT NULL CHECK (status IN ('active', 'canceled')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    canceled_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_market_listings_active
    ON market_listings(created_at DESC, id ASC)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_market_listings_seller_status
    ON market_listings(seller_account_id, status, created_at DESC);
