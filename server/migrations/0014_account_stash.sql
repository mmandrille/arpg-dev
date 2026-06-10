-- 0014_account_stash: account-wide stash items/gold plus immutable
-- session-start account stash snapshots for replay.

CREATE TABLE IF NOT EXISTS account_stash_items (
    account_id          TEXT NOT NULL REFERENCES accounts(id),
    stash_item_id       TEXT NOT NULL,
    source_character_id TEXT REFERENCES characters(id) ON DELETE SET NULL,
    item_def_id         TEXT NOT NULL,
    rolled_stats        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, stash_item_id)
);

CREATE INDEX IF NOT EXISTS idx_account_stash_items_account_created
    ON account_stash_items(account_id, created_at, stash_item_id);

CREATE TABLE IF NOT EXISTS account_stash_gold (
    account_id TEXT PRIMARY KEY REFERENCES accounts(id),
    gold       INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (gold >= 0)
);

CREATE TABLE IF NOT EXISTS session_start_account_stash_items (
    session_id          TEXT NOT NULL REFERENCES sessions(id),
    account_id          TEXT NOT NULL REFERENCES accounts(id),
    stash_item_id       TEXT NOT NULL,
    source_character_id TEXT,
    item_def_id         TEXT NOT NULL,
    rolled_stats        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, stash_item_id)
);

CREATE INDEX IF NOT EXISTS idx_session_start_account_stash_items_account
    ON session_start_account_stash_items(session_id, account_id, created_at, stash_item_id);

CREATE TABLE IF NOT EXISTS session_start_account_stash_gold (
    session_id TEXT NOT NULL REFERENCES sessions(id),
    account_id TEXT NOT NULL REFERENCES accounts(id),
    gold       INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id),
    CHECK (gold >= 0)
);
