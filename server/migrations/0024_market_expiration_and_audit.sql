ALTER TABLE market_listings
    DROP CONSTRAINT IF EXISTS market_listings_status_check;

ALTER TABLE market_listings
    ADD CONSTRAINT market_listings_status_check
    CHECK (status IN ('active', 'canceled', 'accepted', 'expired'));

ALTER TABLE market_listings
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '24 hours'),
    ADD COLUMN IF NOT EXISTS expired_at TIMESTAMPTZ;

ALTER TABLE market_offers
    DROP CONSTRAINT IF EXISTS market_offers_status_check;

ALTER TABLE market_offers
    ADD CONSTRAINT market_offers_status_check
    CHECK (status IN ('active', 'accepted', 'rejected', 'canceled'));

ALTER TABLE market_offers
    ADD COLUMN IF NOT EXISTS canceled_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS market_audit_records (
    id                  BIGSERIAL PRIMARY KEY,
    action              TEXT NOT NULL,
    listing_id          TEXT NOT NULL,
    offer_id            TEXT,
    actor_account_id    TEXT,
    seller_account_id   TEXT,
    bidder_account_id   TEXT,
    item_def_id         TEXT,
    stash_item_id       TEXT,
    details             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_market_audit_listing_created
    ON market_audit_records(listing_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_market_audit_actor_created
    ON market_audit_records(actor_account_id, created_at DESC, id DESC);
