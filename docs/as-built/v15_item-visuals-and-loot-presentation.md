# v15 — Item visuals and loot presentation

**Proves:** Current item presentation can be shared-data-driven without making the client
authoritative for inventory, loot, or equipment outcomes.

- `shared/assets/item_presentations.v0.json` defines client-only icon and ground-loot presentation
  metadata for every current item rule: `rusty_sword`, `training_bow`, `training_badge`,
  `quest_leaf`, and `red_potion`.
- `make validate-shared` schema-validates the presentation file and cross-checks that every item
  rule has presentation metadata and no stale presentation keys exist.
- Godot inventory slots draw distinct shape/color icons from shared presentation data instead of
  text initials, while tooltips still resolve names/stats from item rules.
- Godot ground loot renders distinct primitive silhouettes for sword, bow, badge/coin, leaf, and
  potion from the same presentation metadata; missing metadata falls back to category coloring.
- Equipped weapon GLB mounting remains unchanged through `item_visuals.v0.json` and
  `assets.v0.json`; the server/protocol are unchanged.
- Godot client bot now asserts loot and inventory presentation metadata on the inventory drop
  scenario; `test_item_visuals.gd` checks presentation coverage for every current item.
- `make ci` green on 2026-06-06.

**Explicit non-goals:** no production art, imported icon pack, texture budget, Blender export
pipeline, remote patcher, stash, vendors, crafting, consumable use, or new gameplay item stats.
