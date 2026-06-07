-- 0004_character_item_key: item instance ids are deterministic protocol ids
-- unique within a character, not globally across all characters.

ALTER TABLE character_item_instances
    DROP CONSTRAINT IF EXISTS character_item_instances_pkey;

ALTER TABLE character_item_instances
    ADD PRIMARY KEY (character_id, id);
