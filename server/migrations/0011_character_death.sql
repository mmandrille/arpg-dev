-- 0011_character_death: permanent-death lifecycle flag for characters.

ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS dead BOOLEAN NOT NULL DEFAULT FALSE;
