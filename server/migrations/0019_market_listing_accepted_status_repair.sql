ALTER TABLE market_listings
    DROP CONSTRAINT IF EXISTS market_listings_status_check;

ALTER TABLE market_listings
    ADD CONSTRAINT market_listings_status_check
    CHECK (status IN ('active', 'canceled', 'accepted'));

ALTER TABLE market_listings
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ;
