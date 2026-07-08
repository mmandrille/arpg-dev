-- 0030_account_resource_bag: account-wide instanced resource storage (shards, stones, quest items).

CREATE TABLE IF NOT EXISTS account_resource_bag_items (
    account_id          TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    bag_item_id         TEXT NOT NULL,
    source_character_id TEXT REFERENCES characters(id) ON DELETE SET NULL,
    item_def_id         TEXT NOT NULL,
    rolled_stats        JSONB NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, bag_item_id)
);

CREATE INDEX IF NOT EXISTS idx_account_resource_bag_items_account
    ON account_resource_bag_items(account_id, item_def_id);

CREATE TABLE IF NOT EXISTS session_start_account_resource_bag_items (
    session_id          TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    account_id          TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    bag_item_id         TEXT NOT NULL,
    source_character_id TEXT,
    item_def_id         TEXT NOT NULL,
    rolled_stats        JSONB NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, bag_item_id)
);

CREATE INDEX IF NOT EXISTS idx_session_start_account_resource_bag_items_session
    ON session_start_account_resource_bag_items(session_id, account_id);
