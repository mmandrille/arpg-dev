-- 0010_character_deepest_dungeon_depth: durable max dungeon depth reached.

ALTER TABLE character_progression
    ADD COLUMN IF NOT EXISTS deepest_dungeon_depth INTEGER NOT NULL DEFAULT 0;

ALTER TABLE session_start_character_progression
    ADD COLUMN IF NOT EXISTS deepest_dungeon_depth INTEGER NOT NULL DEFAULT 0;

ALTER TABLE character_progression
    ADD CONSTRAINT character_progression_deepest_dungeon_depth_nonnegative CHECK (deepest_dungeon_depth >= 0);

ALTER TABLE session_start_character_progression
    ADD CONSTRAINT session_start_character_progression_deepest_dungeon_depth_nonnegative CHECK (deepest_dungeon_depth >= 0);
