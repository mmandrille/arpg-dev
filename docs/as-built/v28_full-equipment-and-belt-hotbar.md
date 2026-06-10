# v28 — Full equipment and belt hotbar

**Proves:** The single weapon slot can be replaced by server-authoritative paper-doll equipment while
keeping replay, persistence, bots, and the Godot UI in sync.

- Wire `equipped` now exposes `head`, `amulet`, `chest`, `gloves`, `belt`, `boots`,
  `ring_left`, `ring_right`, `main_hand`, and `off_hand`; legacy `weapon` was migrated to
  `main_hand` across schemas, fixtures, bots, smoke, and client code.
- Go sim enforces slot compatibility, logical ring slots, one-hand plus shield coexistence,
  two-handed sword/bow occupancy, and offhand blocking when `main_hand` holds a two-handed item.
- `use_hotbar_intent { slot_index }` resolves the assigned item server-side, while direct
  `use_intent { item_instance_id }` remains valid for bag use.
- Character hotbar layout persists in Postgres, session-start hotbar snapshots preserve replay
  determinism, and stale item removal clears every referencing hotbar slot.
- Base hotbar capacity is 2; belts roll `hotbar_slots` and expand capacity up to 10. Disabled slots
  retain assignments, no-op client-side when pressed, and reject server-side if explicitly used.
- `equipment_lab`, `equipment_lab_tc_1`, and `shared/golden/full_equipment.json` cover every v28
  equipment category, shield display rolls, belt capacity, and hotbar re-enable behavior.
- Godot inventory now renders named paper-doll slots and sends protocol-backed equip/unequip/hotbar
  intents; the consumable bar is snapshot/delta driven and updates capacity from authoritative
  equipment deltas.
- Protocol bot scenario `19_full_equipment.json` proves full slot coverage, hand occupancy, pinned
  belt capacity 10, disabled-slot persistence, reconnect/replay, and fresh-session persistence.
- Client bot scenario `10_full_equipment.json` proves named loot pickup, paper-doll equip, disabled
  hotbar assignment, belt expansion, and enabled hotbar use through the Godot UI path.

**Explicit non-goals:** armor mitigation, block chance execution, affix grammar, comparison UI,
stash/vendors, production icons/art, offhand abilities/dual-wield, and deeper dungeon drop economy.
