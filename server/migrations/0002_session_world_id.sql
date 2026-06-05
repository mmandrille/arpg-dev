-- 0002_session_world_id: persist the deterministic initial world preset.

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS world_id TEXT NOT NULL DEFAULT 'vertical_slice';
