CREATE TABLE IF NOT EXISTS market_offers (
    id                 TEXT PRIMARY KEY,
    listing_id         TEXT NOT NULL REFERENCES market_listings(id) ON DELETE CASCADE,
    bidder_account_id  TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    status             TEXT NOT NULL CHECK (status IN ('active', 'accepted', 'rejected')),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at        TIMESTAMPTZ,
    rejected_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_market_offers_listing_status
    ON market_offers(listing_id, status, created_at ASC, id ASC);

CREATE TABLE IF NOT EXISTS market_offer_items (
    offer_id             TEXT NOT NULL REFERENCES market_offers(id) ON DELETE CASCADE,
    bidder_account_id    TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    stash_item_id         TEXT NOT NULL,
    source_character_id   TEXT,
    item_def_id           TEXT NOT NULL,
    rolled_stats          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (offer_id, stash_item_id)
);

CREATE INDEX IF NOT EXISTS idx_market_offer_items_bidder
    ON market_offer_items(bidder_account_id, stash_item_id);
