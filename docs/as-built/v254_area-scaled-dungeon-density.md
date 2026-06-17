# v254 As-Built - Area-Scaled Dungeon Density

Date: 2026-06-17

## What shipped

- Replaced fixed ordinary-floor monster and obstacle counts with shared area-density formulas.
- Formula shape: `round(width * height / area_per_unit)`, clamped by rule-owned `min`/`max`.
- Range formulas use the same area-derived center plus a rule-owned `spread`.
- Entry floors now derive 21 base monsters from their 100x50 area, up from the previous fixed 18.
- Depth-4 ordinary floors keep the v252 120x70 size and now derive 35 base monsters, up from the
  previous fixed 26.
- Entry obstacle groups now derive a 7-10 range, up from the previous fixed 4-8 range.
- Depth-4 ordinary obstacle groups now derive an 11-15 range, up from the previous fixed 7-11 range.
- Monster pack-count ranges are also area-derived: entry floors resolve to 6-7 packs, while
  depth-4 ordinary floors resolve to 9-11 packs.
- Boss floors keep the compact boss-floor size and `boss_floor.monster_count` behavior.
- Extracted `tools/dungeon_density.py` so the large shared validator stayed within the file-size
  ratchet while cross-checking the new formula contract.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestDungeonDensityFormulas|TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'
make bot scenario=12_dungeon_levels
make bot scenario=14_dungeon_monsters
make bot scenario=28_reachable_dungeon_obstacles
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`.

Manual visual proof, if desired:

```bash
make bot-visual scenario=14_dungeon_monsters
```

## Scope limits

- No new monster definitions, combat stat rebalance, loot-table changes, XP tuning, boss-floor
  population changes, room/corridor PCG, rivers, doors, destructible/secret obstacles, biome
  palettes, protocol/schema bump, or client renderer changes shipped.
