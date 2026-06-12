-- 0021_character_death_level: records the dungeon level where a permanent-death
-- character body can be recovered by another same-account hero.

ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS death_level INTEGER;

ALTER TABLE characters
    DROP CONSTRAINT IF EXISTS characters_death_level_requires_dead;

ALTER TABLE characters
    ADD CONSTRAINT characters_death_level_requires_dead
    CHECK ((dead = TRUE AND death_level IS NOT NULL) OR (dead = FALSE AND death_level IS NULL));
