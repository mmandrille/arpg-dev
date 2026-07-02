-- 0029_character_hired_mercenary: persist active character mercenary hire per hero.

ALTER TABLE character_progression
    ADD COLUMN IF NOT EXISTS hired_mercenary_character_id TEXT NOT NULL DEFAULT '';

ALTER TABLE session_start_character_progression
    ADD COLUMN IF NOT EXISTS hired_mercenary_character_id TEXT NOT NULL DEFAULT '';
