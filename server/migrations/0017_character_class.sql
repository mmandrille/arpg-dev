-- 0017_character_class: persist authoritative character class identity.

ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS character_class TEXT NOT NULL DEFAULT 'barbarian';

UPDATE characters
   SET character_class = 'barbarian'
 WHERE character_class = '';

