# v39 — UI currency and mana polish

**Proves:** Character gold and player mana can extend the existing authoritative economy,
progression, protocol, and Godot UI paths without making coins occupy inventory space.

- Gold is now durable character progression state, persisted through the store and surfaced in
  snapshots, deltas, `/state`, replay, and the Godot HUD/inventory presentation.
- Currency loot uses `item_def_id: "gold"` with rolled amounts; pickup removes the loot, emits
  `gold_picked_up`, and updates the character wallet instead of adding a bag item.
- Generated dungeon gold rolls scale by depth and monster rarity, while static/town rewards use the
  base gold range.
- Player entity views now carry `mana` / `max_mana`; blue potions declare `mana_restore`, restore
  mana authoritatively, and reject use while mana is already full.
- Character-derived armor now comes from DEX instead of VIT, with shared golden and client formula
  coverage updated together.
- Godot UI adds the mana bar, gold counter, larger interface text, mutually exclusive panels,
  character rename affordance, top-right debug text setting, and sword hand-pose polish.

**Explicit non-goals:** no gold sinks, vendors, stash, trade, crafting, multiple currencies, mana
regeneration, spells, skill bar, active ability catalog, production UI art, or broader UI redesign.
