-- 0031_session_start_character_corpses: freeze the recoverable same-account
-- corpses each member saw at session start. Live corpse rows change as bodies
-- are looted or new characters die, so replay must not read them (v477).

CREATE TABLE IF NOT EXISTS session_start_character_corpses (
    session_id          TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    account_id          TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    character_id        TEXT NOT NULL,
    corpse_character_id TEXT NOT NULL,
    ordinal             INT  NOT NULL,
    name                TEXT NOT NULL,
    level               INT  NOT NULL,
    death_level         INT  NOT NULL,
    items               JSONB NOT NULL DEFAULT '[]',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id, corpse_character_id)
);
