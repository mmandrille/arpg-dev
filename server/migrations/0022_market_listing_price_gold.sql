ALTER TABLE market_listings
    ADD COLUMN IF NOT EXISTS price_gold INTEGER NOT NULL DEFAULT 0;

ALTER TABLE market_listings
    DROP CONSTRAINT IF EXISTS market_listings_price_gold_check;

ALTER TABLE market_listings
    ADD CONSTRAINT market_listings_price_gold_check
    CHECK (price_gold >= 0);
