-- 0006_character_hotbar_and_equipment: durable character hotbar layout and
-- immutable session-start hotbar snapshots for replay.

UPDATE character_item_instances
SET slot = 'main_hand'
WHERE slot = 'weapon';

UPDATE session_start_item_instances
SET slot = 'main_hand'
WHERE slot = 'weapon';

CREATE TABLE IF NOT EXISTS character_hotbar_slots (
    character_id      TEXT NOT NULL REFERENCES characters(id),
    account_id        TEXT NOT NULL REFERENCES accounts(id),
    slot_index        INTEGER NOT NULL,
    item_instance_id  TEXT,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (character_id, slot_index),
    CHECK (slot_index >= 0 AND slot_index <= 9)
);

CREATE INDEX IF NOT EXISTS idx_character_hotbar_account
    ON character_hotbar_slots(account_id, character_id);

CREATE TABLE IF NOT EXISTS session_start_hotbar_slots (
    session_id        TEXT NOT NULL REFERENCES sessions(id),
    account_id        TEXT NOT NULL REFERENCES accounts(id),
    character_id      TEXT NOT NULL REFERENCES characters(id),
    slot_index        INTEGER NOT NULL,
    item_instance_id  TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, slot_index),
    CHECK (slot_index >= 0 AND slot_index <= 9)
);
