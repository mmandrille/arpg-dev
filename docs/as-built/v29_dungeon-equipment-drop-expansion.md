# v29 — Dungeon equipment drop expansion

**Proves:** Real generated dungeon monsters and guarded chests can use the expanded v28 equipment
catalog through deterministic, depth-aware treasure classes.

- `shared/rules/dungeon_generation.v0.json` now declares temporary coarse loot bands for depth
  `1`, `2`, and `3+`; level `0` town still does not use dungeon loot bands.
- Depth-specific monster and guarded-chest loot tables bridge to new treasure classes, with chest
  equipment odds intentionally better than normal monster odds.
- Generated dungeon monsters and chests store their selected loot table at generation time, while
  source kill/open still owns all reward rolls in the Go sim.
- By depth `3+`, validation proves the configured dungeon/chest reward set can reach every v28
  equipment template: weapons, shield, armor pieces, belt, boots, ring, and amulet.
- `shared/golden/dungeon_equipment_drops.json` pins representative depth/source selection and
  monster/chest outcomes; `treasure_class_rolls.json` now covers varied direct equipment,
  potion, and money-like rolls.
- Protocol bot scenario `20_dungeon_equipment_drops.json` descends into generated dungeon play,
  opens a depth-band chest, picks up rolled equipment, equips it, and proves `/state`, reconnect,
  replay, and fresh-session persistence.

**Explicit non-goals:** final depth economy, item-level gates, Magic Find, affixes, unique/set
items, real gold wallet, vendors/stash/crafting/trade, combat use of armor/block/crit/hit speed,
production item/chest art, and client-side loot logic.
