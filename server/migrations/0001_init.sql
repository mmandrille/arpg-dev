-- 0001_init: foundational persistence schema for the first playable slice.
-- Entities and required fields follow spec section 4.6.

CREATE TABLE IF NOT EXISTS accounts (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS characters (
    id         TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_characters_account ON characters(account_id);

CREATE TABLE IF NOT EXISTS sessions (
    id           TEXT PRIMARY KEY,
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    character_id TEXT NOT NULL REFERENCES characters(id),
    seed         TEXT NOT NULL,
    status       TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_character ON sessions(character_id);

-- Inventory id is the protocol item_instance_id: a decimal string allocated by
-- the deterministic per-session entity counter, so replay reproduces it. That
-- counter restarts each session, so the id is unique only WITHIN a session;
-- the table is therefore keyed by (session_id, id). This is a v0 deviation from
-- spec 4.6 (character-scoped inventory): cross-session character inventory is a
-- post-v0 feature, and "survives restart" is proven by reconnecting to the same
-- session. account_id/character_id are retained for context.
CREATE TABLE IF NOT EXISTS inventory_items (
    id           TEXT NOT NULL,
    session_id   TEXT NOT NULL REFERENCES sessions(id),
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    character_id TEXT NOT NULL REFERENCES characters(id),
    item_def_id  TEXT NOT NULL,
    slot         TEXT,
    equipped     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, id)
);

CREATE INDEX IF NOT EXISTS idx_inventory_session ON inventory_items(session_id);
CREATE INDEX IF NOT EXISTS idx_inventory_character ON inventory_items(character_id);

-- Event-sourced authoritative output (spec 4.6, ADR D8.2). Ordered by
-- (tick, sequence) within a session.
CREATE TABLE IF NOT EXISTS session_events (
    id             TEXT PRIMARY KEY,
    session_id     TEXT NOT NULL REFERENCES sessions(id),
    tick           BIGINT NOT NULL,
    sequence       BIGINT NOT NULL,
    event_type     TEXT NOT NULL,
    correlation_id TEXT,
    payload        JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_session_order
    ON session_events(session_id, tick, sequence);

-- Recorded authoritative inputs (spec 4.6, ADR D8.2). The (session_id, tick,
-- sequence) order is the deterministic application order for replay.
CREATE TABLE IF NOT EXISTS session_inputs (
    id             TEXT PRIMARY KEY,
    session_id     TEXT NOT NULL REFERENCES sessions(id),
    tick           BIGINT NOT NULL,
    sequence       BIGINT NOT NULL,
    message_id     TEXT NOT NULL,
    correlation_id TEXT,
    payload        JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_inputs_session_order
    ON session_inputs(session_id, tick, sequence);

-- A message_id is unique within a session: guards duplicate-input replay.
CREATE UNIQUE INDEX IF NOT EXISTS uq_inputs_session_message
    ON session_inputs(session_id, message_id);
