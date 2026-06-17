# v254 As-Built - Area-Scaled Dungeon Density

Date: 2026-06-17

## What shipped

- Replaced fixed ordinary-floor monster and obstacle counts with shared area-density formulas.
- Formula shape: `round(width * height / area_per_unit)`, clamped by rule-owned `min`/`max`.
- Range formulas use the same area-derived center plus a rule-owned `spread`.
- Entry floors now derive 20 base monsters from their 100x50 area, up from the previous fixed 18.
- Depth-4 ordinary floors keep the v252 120x70 size and now derive 30 base monsters, up from the
  previous fixed 26.
- Entry obstacle groups now derive a 6-7 range, up from the previous fixed 4-8 range.
- Depth-4 ordinary obstacle groups now derive a 9-11 range, up from the previous fixed 7-11 range.
- Monster pack-count ranges are also area-derived: entry floors resolve to 6-7 packs, while
  depth-4 ordinary floors resolve to 8-10 packs.
- Boss floors keep the compact boss-floor size and `boss_floor.monster_count` behavior.
- Extracted `tools/dungeon_density.py` so the large shared validator stayed within the file-size
  ratchet while cross-checking the new formula contract.
- Moved the obstacle golden proof into `server/internal/game/dungeon_obstacles_golden_test.go`,
  refreshed `shared/golden/dungeon_obstacles.json`, and lowered the `game_test.go` ratchet
  baseline after the extraction.
- Stabilized density-sensitive bot proofs by giving the pack-aggro protocol scenario debug
  survivability, raising the teleporter scenario elapsed cap to cover mandatory replay overhead,
  and making the client elite-objective HUD proof assert active tracker state without pinning the
  exact leader count.
- Moved the protocol account-stash persistence proof from generated dungeon traversal to the
  existing `vendor_lab` stash/loot setup so unrelated stash reconstruction no longer depends on
  denser dungeon paths.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestDungeonDensityFormulas|TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'
make bot scenario=12_dungeon_levels
make bot scenario=14_dungeon_monsters
make bot scenario=28_reachable_dungeon_obstacles
make bot scenario=13_teleporter_lab
make bot scenario=17_treasure_classes_and_guarded_chests
make bot scenario=42_pack_aggro_and_dungeon_packs
make bot scenario=68_dungeon_elite_side_objective
make bot scenario=77_elite_minion_pack_ai
VERBOSE=1 make bot scenario=account_stash_storage
SCENARIO=44_elite_objective_hud HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
make ci
```

All focused checks and full `make ci` passed on 2026-06-17 during `$autoloop`.

Manual visual proof, if desired:

```bash
make bot-visual scenario=14_dungeon_monsters
```

## Scope limits

- No new monster definitions, combat stat rebalance, loot-table changes, XP tuning, boss-floor
  population changes, room/corridor PCG, rivers, doors, destructible/secret obstacles, biome
  palettes, protocol/schema bump, or client renderer changes shipped.
