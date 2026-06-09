-- 0008_listed_sessions: backend-listed co-op sessions for menu discovery.

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS listed BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_sessions_active_listed
    ON sessions(listed, status, mode, updated_at DESC, id ASC)
    WHERE listed = TRUE AND status = 'active' AND mode = 'coop';

CREATE INDEX IF NOT EXISTS idx_session_members_session_connected
    ON session_members(session_id, status, connected);
