# v13 — Inventory UI

**Proves:** Human-facing inventory presentation can stay display-only while server-owned inventory
intents mutate authoritative state, persistence, replay, and resume.

- `unequip_intent` and `drop_intent` extend the protocol; `inventory_remove` lets deltas remove bag
  rows without waiting for a fresh snapshot.
- Server drop placement is deterministic, collision-free, adjacent to the player, and pinned by
  `shared/golden/inventory_drop.json` in Go and GDScript fixture checks.
- Dropping an equipped item clears `equipped.weapon`, removes the inventory row, spawns pickup-able
  loot, persists the removal, and reconstructs through replay/resume.
- `inventory_lab` plus bot scenario `07_inventory_lab.json` proves pickup, equip, unequip, drop,
  re-pickup, and re-equip over protocol, `/state`, reconnect resume, and replay.
- Godot adds a custom Diablo-dark panel toggled with `I`, one weapon slot, a bag grid, tooltips from
  item rules, double-click/drag equip, drag-to-bag unequip, drag-outside drop, and no local inventory
  authority.
- The old `Q` equip shortcut and debug hints are removed; autoplay and bot continue using explicit
  protocol `equip_intent`.
- `make ci` green on 2026-06-05.

**Explicit non-goals (still true):** no stash, vendors, crafting, stack splitting, equipment slots
beyond weapon, production item icons, Godot inventory plugins, character-scoped persistence, item
destruction, or drop targeting/range gates.
