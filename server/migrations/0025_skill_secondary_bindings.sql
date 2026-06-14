-- 0022_skill_secondary_bindings: expand skill function-key bindings from
-- one F1-F8 row to primary + secondary rows (16 slots total).

ALTER TABLE character_skill_bindings
    DROP CONSTRAINT IF EXISTS character_skill_bindings_slot_index_check;

ALTER TABLE character_skill_bindings
    ADD CONSTRAINT character_skill_bindings_slot_index_check
    CHECK (slot_index >= 0 AND slot_index <= 15);

ALTER TABLE session_start_skill_bindings
    DROP CONSTRAINT IF EXISTS session_start_skill_bindings_slot_index_check;

ALTER TABLE session_start_skill_bindings
    ADD CONSTRAINT session_start_skill_bindings_slot_index_check
    CHECK (slot_index >= 0 AND slot_index <= 15);
