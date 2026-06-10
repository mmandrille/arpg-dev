# v25 — Treasure classes and guarded chests

**Proves:** Monster and chest rewards can resolve through data-driven treasure classes with
multiple ordered drop attempts, while rare procedural chests create guarded dungeon floors.

- Shared `treasure_classes.v0.json` defines ordered attempts with success/no-drop weights and
  weighted fixed item or item-template entries.
- `dungeon_mob_drop` now bridges through `dungeon_mob_tc_1`; its primary attempt produces a rolled
  `cave_blade`, while a lower-probability secondary attempt can add `red_potion` or the money-like
  `training_badge`.
- `guarded_chest_drop` bridges through `guarded_chest_tc_1`, giving chests a primary reward and a
  lower-probability bonus attempt.
- Dungeon generation has rare `chest_placement`; successful chest floors spawn a `treasure_chest`
  and apply `monster_count_bonus` on that same level.
- Chest generation uses a labeled seed substream, so no-chest floors preserve existing stair and
  monster generation expectations.
- `treasure_chest` opens via existing `action_intent`, emits existing interactable/loot events,
  rolls loot once, and rejects repeated opens without duplicating drops.
- Bot scenario `17_treasure_classes_and_guarded_chests.json` pins a guarded chest floor, proves
  monster treasure-class loot, chest open-once behavior, pickup, `/state`, reconnect, replay, and
  fresh-session persistence.
- Bot create-session now supports an optional pinned seed only in local development; normal remote
  sessions keep server-generated OS-entropy seeds.
- Gold wallet, Magic Find, unique/set catalogs, depth-banded treasure classes, boss-floor chest
  integration, and production chest art remain deferred.

**Explicit non-goals:** no Magic Find stat or rarity modifier, no unique/set catalogs, no real gold
wallet, no boss-floor rules, no production chest art/animation/audio, and no client-side drop logic.
