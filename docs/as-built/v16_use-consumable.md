# v16 — Use consumable

**Proves:** Consumable item use can mutate authoritative HP and inventory while the hotbar stays a
client-only input surface.

- `use_intent { item_instance_id }` is decoded, persisted with the input stream, and resolved by
  the Go sim against server-owned inventory and item rules.
- `red_potion` declares `heal: { min: 5, max: 5 }`; `shared/golden/use_consumable.json` pins heal
  amount and HP cap behavior for Go/GDScript drift checks.
- Server removes the consumed inventory row, emits `item_used` and `player_healed`, and updates the
  player entity HP; rejects include non-consumable, missing item, full HP, and dead player cases.
- `heal_lab` plus protocol bot scenario `08_heal_lab.json` proves pickup of two potions, damage,
  two uses, `/state`, reconnect resume, and replay.
- Godot adds a bottom-center `ConsumableBar` with 10 client-only slots; drag assignment and keys
  `1`-`9`/`0` send `use_intent` for the assigned inventory item.
- Client bot scenario `06_use_potion_hotbar.json` exercises hotbar assignment, double-click bag
  use, key use, and inventory removal.
- `make ci` green on 2026-06-06.

**Explicit non-goals:** no server-side hotbar persistence, stack splitting, cooldowns, buffs,
heal-over-time, production potion art, stash, vendors, or crafting.
