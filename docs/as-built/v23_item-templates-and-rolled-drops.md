# v23 — Item templates and rolled drops

**Proves:** Dungeon kills can produce deterministic rolled gear that remains server-authoritative
through pickup, equip, combat, persistence, reconnect, fresh sessions, and replay.

- Shared `item_templates.v0.json` defines `cave_blade`, rarity weights, bounded rollable stats,
  requirements, and reserved effect ids as data.
- Loot tables now support entries keyed by exactly one of `item_def_id` or `item_template_id`;
  legacy fixed drops and empty `no_drop` remain valid.
- `dungeon_mob` now uses `dungeon_mob_drop`, rolling a concrete `cave_blade` payload at monster
  death with the seeded Go RNG.
- Rolled item metadata is additive in protocol v1 item and loot entity views: `item_template_id`,
  `display_name`, `rarity`, `rolled_stats`, `requirements`, and `effect_ids`.
- Character item persistence stores the durable rolled payload in v22's `rolled_stats` JSON and
  reloads it through session-start snapshots without re-rolling.
- Equipped rolled weapons use rolled `damage_min` / `damage_max` for authoritative damage; rolled
  `max_hp` is display-only in v23.
- Godot inventory tooltips display instance rarity, display name, rolled damage, `max_hp`, and
  requirements; `cave_blade` reuses placeholder blade visuals.
- Bot scenario `16_rolled_drops.json` proves dungeon mob kill, rolled drop pickup, equip, damage
  use, `/state`, reconnect, replay, and fresh-session persistence.

**Explicit non-goals:** no affix grammar, procedural name generator, armor/jewelry/offhand,
stash/crafting/vendors/gold/trade, special-effect execution, item comparison UI, loot filters,
production item art, character stat requirements beyond level `1`, or Protobuf migration.
