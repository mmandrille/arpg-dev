-- 0003_character_progression: durable default-character items, waypoints, and
-- immutable session-start progression snapshots.

CREATE TABLE IF NOT EXISTS character_item_instances (
    id           TEXT NOT NULL,
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    character_id TEXT NOT NULL REFERENCES characters(id),
    item_def_id  TEXT NOT NULL,
    location     TEXT NOT NULL DEFAULT 'inventory',
    slot         TEXT,
    equipped     BOOLEAN NOT NULL DEFAULT FALSE,
    rolled_stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (character_id, id),
    CHECK (location IN ('inventory', 'equipped', 'stash'))
);

CREATE INDEX IF NOT EXISTS idx_character_items_character
    ON character_item_instances(character_id, created_at, id);

CREATE TABLE IF NOT EXISTS character_waypoints (
    character_id   TEXT NOT NULL REFERENCES characters(id),
    level          INTEGER NOT NULL,
    discovered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (character_id, level)
);

CREATE TABLE IF NOT EXISTS session_start_item_instances (
    session_id   TEXT NOT NULL REFERENCES sessions(id),
    id           TEXT NOT NULL,
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    character_id TEXT NOT NULL REFERENCES characters(id),
    item_def_id  TEXT NOT NULL,
    location     TEXT NOT NULL,
    slot         TEXT,
    equipped     BOOLEAN NOT NULL DEFAULT FALSE,
    rolled_stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, id),
    CHECK (location IN ('inventory', 'equipped', 'stash'))
);

CREATE TABLE IF NOT EXISTS session_start_waypoints (
    session_id     TEXT NOT NULL REFERENCES sessions(id),
    character_id   TEXT NOT NULL REFERENCES characters(id),
    level          INTEGER NOT NULL,
    discovered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, level)
);

-- Best-effort preservation for existing development rows. The old deterministic
-- item id is session-local, so mint a globally durable id while preserving the
-- item definition/equipment state for the owning character.
INSERT INTO character_item_instances (
    id, account_id, character_id, item_def_id, location, slot, equipped, rolled_stats, created_at, updated_at
)
SELECT
    inventory_items.session_id || ':' || inventory_items.id,
    inventory_items.account_id,
    inventory_items.character_id,
    inventory_items.item_def_id,
    CASE WHEN inventory_items.equipped THEN 'equipped' ELSE 'inventory' END,
    inventory_items.slot,
    inventory_items.equipped,
    '{}'::jsonb,
    inventory_items.created_at,
    now()
FROM inventory_items
ON CONFLICT (character_id, id) DO NOTHING;
