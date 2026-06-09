-- 0012_character_skills: durable skill points/ranks and immutable
-- session-start skill rank snapshots.

ALTER TABLE character_progression
    ADD COLUMN IF NOT EXISTS unspent_skill_points INTEGER NOT NULL DEFAULT 0;

ALTER TABLE session_start_character_progression
    ADD COLUMN IF NOT EXISTS unspent_skill_points INTEGER NOT NULL DEFAULT 0;

ALTER TABLE character_progression
    ADD CONSTRAINT character_progression_unspent_skill_points_nonnegative CHECK (unspent_skill_points >= 0);

ALTER TABLE session_start_character_progression
    ADD CONSTRAINT session_start_character_progression_unspent_skill_points_nonnegative CHECK (unspent_skill_points >= 0);

CREATE TABLE IF NOT EXISTS character_skill_ranks (
    account_id      TEXT NOT NULL REFERENCES accounts(id),
    character_id    TEXT NOT NULL REFERENCES characters(id),
    skill_id        TEXT NOT NULL,
    rank            INTEGER NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, character_id, skill_id),
    CHECK (rank >= 0)
);

CREATE INDEX IF NOT EXISTS idx_character_skill_ranks_character
    ON character_skill_ranks(character_id, skill_id);

CREATE TABLE IF NOT EXISTS session_start_character_skill_ranks (
    session_id      TEXT NOT NULL REFERENCES sessions(id),
    account_id      TEXT NOT NULL REFERENCES accounts(id),
    character_id    TEXT NOT NULL REFERENCES characters(id),
    skill_id        TEXT NOT NULL,
    rank            INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id, skill_id),
    CHECK (rank >= 0)
);
