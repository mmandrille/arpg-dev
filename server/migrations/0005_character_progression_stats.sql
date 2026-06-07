-- 0005_character_progression_stats: durable character XP/stat progression and
-- immutable session-start progression snapshots.

CREATE TABLE IF NOT EXISTS character_progression (
    character_id          TEXT PRIMARY KEY REFERENCES characters(id),
    account_id            TEXT NOT NULL REFERENCES accounts(id),
    level                 INTEGER NOT NULL,
    experience            INTEGER NOT NULL,
    unspent_stat_points   INTEGER NOT NULL,
    stat_str              INTEGER NOT NULL,
    stat_dex              INTEGER NOT NULL,
    stat_vit              INTEGER NOT NULL,
    stat_magic            INTEGER NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (level >= 1),
    CHECK (experience >= 0),
    CHECK (unspent_stat_points >= 0),
    CHECK (stat_str >= 0),
    CHECK (stat_dex >= 0),
    CHECK (stat_vit >= 0),
    CHECK (stat_magic >= 0)
);

CREATE INDEX IF NOT EXISTS idx_character_progression_account
    ON character_progression(account_id, character_id);

CREATE TABLE IF NOT EXISTS session_start_character_progression (
    session_id            TEXT PRIMARY KEY REFERENCES sessions(id),
    account_id            TEXT NOT NULL REFERENCES accounts(id),
    character_id          TEXT NOT NULL REFERENCES characters(id),
    level                 INTEGER NOT NULL,
    experience            INTEGER NOT NULL,
    unspent_stat_points   INTEGER NOT NULL,
    stat_str              INTEGER NOT NULL,
    stat_dex              INTEGER NOT NULL,
    stat_vit              INTEGER NOT NULL,
    stat_magic            INTEGER NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (level >= 1),
    CHECK (experience >= 0),
    CHECK (unspent_stat_points >= 0),
    CHECK (stat_str >= 0),
    CHECK (stat_dex >= 0),
    CHECK (stat_vit >= 0),
    CHECK (stat_magic >= 0)
);
