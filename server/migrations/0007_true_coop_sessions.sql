-- 0007_true_coop_sessions: two-member co-op session membership and actor-tagged inputs.

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'solo';

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS join_code_hash TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'sessions_mode_check'
    ) THEN
        ALTER TABLE sessions
            ADD CONSTRAINT sessions_mode_check CHECK (mode IN ('solo', 'coop'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS session_members (
    session_id          TEXT NOT NULL REFERENCES sessions(id),
    account_id          TEXT NOT NULL REFERENCES accounts(id),
    character_id        TEXT NOT NULL REFERENCES characters(id),
    player_entity_id    TEXT NOT NULL DEFAULT '',
    role                TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'active',
    connected           BOOLEAN NOT NULL DEFAULT FALSE,
    current_level       INTEGER NOT NULL DEFAULT 0,
    joined_tick         BIGINT NOT NULL DEFAULT 0,
    left_tick           BIGINT,
    joined_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id),
    UNIQUE (session_id, character_id),
    CHECK (role IN ('host', 'guest')),
    CHECK (status IN ('active', 'left'))
);

CREATE INDEX IF NOT EXISTS idx_session_members_session_order
    ON session_members(session_id, joined_at, role, account_id, character_id);

CREATE INDEX IF NOT EXISTS idx_session_members_account
    ON session_members(account_id, session_id);

INSERT INTO session_members (
    session_id, account_id, character_id, role, status, connected, current_level, joined_tick, joined_at, updated_at
)
SELECT id, account_id, character_id, 'host', 'active', FALSE, 0, 0, created_at, updated_at
FROM sessions
ON CONFLICT (session_id, account_id, character_id) DO NOTHING;

ALTER TABLE session_inputs
    ADD COLUMN IF NOT EXISTS actor_account_id TEXT;

ALTER TABLE session_inputs
    ADD COLUMN IF NOT EXISTS actor_character_id TEXT;

ALTER TABLE session_inputs
    ADD COLUMN IF NOT EXISTS actor_player_entity_id TEXT;

ALTER TABLE session_start_character_progression
    DROP CONSTRAINT IF EXISTS session_start_character_progression_pkey;

ALTER TABLE session_start_character_progression
    ADD PRIMARY KEY (session_id, account_id, character_id);

ALTER TABLE session_start_item_instances
    DROP CONSTRAINT IF EXISTS session_start_item_instances_pkey;

ALTER TABLE session_start_item_instances
    ADD PRIMARY KEY (session_id, account_id, character_id, id);

ALTER TABLE session_start_waypoints
    DROP CONSTRAINT IF EXISTS session_start_waypoints_pkey;

ALTER TABLE session_start_waypoints
    ADD PRIMARY KEY (session_id, character_id, level);

ALTER TABLE session_start_hotbar_slots
    DROP CONSTRAINT IF EXISTS session_start_hotbar_slots_pkey;

ALTER TABLE session_start_hotbar_slots
    ADD PRIMARY KEY (session_id, account_id, character_id, slot_index);
