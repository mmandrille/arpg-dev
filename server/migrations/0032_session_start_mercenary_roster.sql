-- 0032_session_start_mercenary_roster: freeze the member account's alternate
-- character hire candidates so live setup and replay use the same roster.

CREATE TABLE IF NOT EXISTS session_start_mercenary_roster (
    session_id             TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    account_id             TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    character_id           TEXT NOT NULL,
    mercenary_character_id TEXT NOT NULL,
    name                   TEXT NOT NULL,
    character_class        TEXT NOT NULL,
    dead                   BOOLEAN NOT NULL,
    progression            JSONB NOT NULL,
    items                  JSONB NOT NULL DEFAULT '[]',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, character_id, mercenary_character_id),
    FOREIGN KEY (session_id, account_id, character_id)
        REFERENCES session_start_character_progression(session_id, account_id, character_id)
        ON DELETE CASCADE
);
