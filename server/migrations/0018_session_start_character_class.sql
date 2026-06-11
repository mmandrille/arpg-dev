-- 0018_session_start_character_class: freeze class identity with session-start progression.

ALTER TABLE session_start_character_progression
    ADD COLUMN IF NOT EXISTS character_class TEXT NOT NULL DEFAULT 'barbarian';

UPDATE session_start_character_progression
   SET character_class = 'barbarian'
 WHERE character_class = '';

