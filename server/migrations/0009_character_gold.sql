-- 0009_character_gold: durable single-currency wallet.

ALTER TABLE character_progression
    ADD COLUMN IF NOT EXISTS gold INTEGER NOT NULL DEFAULT 0;

ALTER TABLE session_start_character_progression
    ADD COLUMN IF NOT EXISTS gold INTEGER NOT NULL DEFAULT 0;

ALTER TABLE character_progression
    ADD CONSTRAINT character_progression_gold_nonnegative CHECK (gold >= 0);

ALTER TABLE session_start_character_progression
    ADD CONSTRAINT session_start_character_progression_gold_nonnegative CHECK (gold >= 0);
