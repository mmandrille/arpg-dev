-- 0026_character_item_weapon_sets: preserve weapon-set placement for durable
-- hand equipment. Older rows default to set 0.

ALTER TABLE character_item_instances
    ADD COLUMN IF NOT EXISTS weapon_set INTEGER NOT NULL DEFAULT 0;

ALTER TABLE session_start_item_instances
    ADD COLUMN IF NOT EXISTS weapon_set INTEGER NOT NULL DEFAULT 0;

WITH ranked AS (
    SELECT
        character_id,
        id,
        LEAST(
            ROW_NUMBER() OVER (
                PARTITION BY account_id, character_id, slot
                ORDER BY created_at ASC, id ASC
            ) - 1,
            1
        ) AS inferred_weapon_set
    FROM character_item_instances
    WHERE equipped
      AND location = 'equipped'
      AND slot IN ('main_hand', 'off_hand')
)
UPDATE character_item_instances item
SET weapon_set = ranked.inferred_weapon_set
FROM ranked
WHERE item.character_id = ranked.character_id
  AND item.id = ranked.id;

WITH ranked AS (
    SELECT
        session_id,
        character_id,
        id,
        LEAST(
            ROW_NUMBER() OVER (
                PARTITION BY session_id, account_id, character_id, slot
                ORDER BY created_at ASC, id ASC
            ) - 1,
            1
        ) AS inferred_weapon_set
    FROM session_start_item_instances
    WHERE equipped
      AND location = 'equipped'
      AND slot IN ('main_hand', 'off_hand')
)
UPDATE session_start_item_instances item
SET weapon_set = ranked.inferred_weapon_set
FROM ranked
WHERE item.session_id = ranked.session_id
  AND item.character_id = ranked.character_id
  AND item.id = ranked.id;

ALTER TABLE character_item_instances
    ADD CONSTRAINT character_item_instances_weapon_set_check
    CHECK (weapon_set >= 0 AND weapon_set <= 1) NOT VALID;

ALTER TABLE session_start_item_instances
    ADD CONSTRAINT session_start_item_instances_weapon_set_check
    CHECK (weapon_set >= 0 AND weapon_set <= 1) NOT VALID;

ALTER TABLE character_item_instances
    VALIDATE CONSTRAINT character_item_instances_weapon_set_check;

ALTER TABLE session_start_item_instances
    VALIDATE CONSTRAINT session_start_item_instances_weapon_set_check;
