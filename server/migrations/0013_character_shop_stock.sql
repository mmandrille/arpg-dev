-- 0013_character_shop_stock: durable generated shop stock and immutable
-- session-start shop stock snapshots for replay.

CREATE TABLE IF NOT EXISTS character_shop_stock (
    account_id        TEXT NOT NULL REFERENCES accounts(id),
    character_id      TEXT NOT NULL REFERENCES characters(id),
    shop_id           TEXT NOT NULL,
    refresh_key       TEXT NOT NULL,
    offer_index       INTEGER NOT NULL,
    offer_id          TEXT NOT NULL,
    source_depth      INTEGER NOT NULL,
    item_template_id  TEXT NOT NULL,
    rolled_payload    JSONB NOT NULL DEFAULT '{}'::jsonb,
    buy_price         INTEGER NOT NULL,
    available         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, character_id, shop_id, offer_id),
    UNIQUE (account_id, character_id, shop_id, offer_index),
    CHECK (offer_index >= 0),
    CHECK (source_depth >= 1),
    CHECK (buy_price >= 1)
);

CREATE INDEX IF NOT EXISTS idx_character_shop_stock_character
    ON character_shop_stock(account_id, character_id, shop_id, refresh_key, offer_index);

CREATE INDEX IF NOT EXISTS idx_character_shop_stock_available
    ON character_shop_stock(account_id, character_id, shop_id, available, offer_index);

CREATE TABLE IF NOT EXISTS session_start_shop_stock (
    session_id        TEXT NOT NULL REFERENCES sessions(id),
    account_id        TEXT NOT NULL REFERENCES accounts(id),
    character_id      TEXT NOT NULL REFERENCES characters(id),
    shop_id           TEXT NOT NULL,
    refresh_key       TEXT NOT NULL,
    offer_index       INTEGER NOT NULL,
    offer_id          TEXT NOT NULL,
    source_depth      INTEGER NOT NULL,
    item_template_id  TEXT NOT NULL,
    rolled_payload    JSONB NOT NULL DEFAULT '{}'::jsonb,
    buy_price         INTEGER NOT NULL,
    available         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id, shop_id, offer_id),
    UNIQUE (session_id, account_id, character_id, shop_id, offer_index),
    CHECK (offer_index >= 0),
    CHECK (source_depth >= 1),
    CHECK (buy_price >= 1)
);

CREATE INDEX IF NOT EXISTS idx_session_start_shop_stock_character
    ON session_start_shop_stock(session_id, account_id, character_id, shop_id, offer_index);
