-- 0015_character_skill_bindings: durable character-owned skill function-key
-- bindings and immutable session-start snapshots for resume/replay.

CREATE TABLE IF NOT EXISTS character_skill_bindings (
    character_id TEXT NOT NULL REFERENCES characters(id),
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    slot_index   INTEGER NOT NULL,
    skill_id     TEXT NOT NULL DEFAULT '',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (character_id, slot_index),
    CHECK (slot_index >= 0 AND slot_index <= 7)
);

CREATE INDEX IF NOT EXISTS idx_character_skill_bindings_account
    ON character_skill_bindings(account_id, character_id);

CREATE TABLE IF NOT EXISTS character_skill_preferences (
    character_id          TEXT PRIMARY KEY REFERENCES characters(id),
    account_id            TEXT NOT NULL REFERENCES accounts(id),
    right_click_skill_id  TEXT NOT NULL DEFAULT '',
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_character_skill_preferences_account
    ON character_skill_preferences(account_id, character_id);

CREATE TABLE IF NOT EXISTS session_start_skill_bindings (
    session_id   TEXT NOT NULL REFERENCES sessions(id),
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    character_id TEXT NOT NULL REFERENCES characters(id),
    slot_index   INTEGER NOT NULL,
    skill_id     TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id, slot_index),
    CHECK (slot_index >= 0 AND slot_index <= 7)
);

CREATE TABLE IF NOT EXISTS session_start_skill_preferences (
    session_id             TEXT NOT NULL REFERENCES sessions(id),
    account_id             TEXT NOT NULL REFERENCES accounts(id),
    character_id           TEXT NOT NULL REFERENCES characters(id),
    right_click_skill_id   TEXT NOT NULL DEFAULT '',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id)
);
