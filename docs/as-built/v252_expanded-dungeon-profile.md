# v252 As-Built - Expanded Dungeon Profile

Date: 2026-06-17

## What shipped

- Added `floor_profiles` to shared dungeon-generation rules and schema.
- Added the first ordinary-floor depth profile for depth 4+ non-boss floors.
- Expanded profiled ordinary floors from 100x50 to 120x70.
- Increased profiled ordinary-floor base monster count from 18 to 26.
- Increased profiled pack-count and obstacle-group ranges through shared rules.
- Applied the matching profile before stairs, teleporters, chests, obstacles, and monsters are
  generated.
- Updated dungeon navigation bounds to use the profiled ordinary-floor size.
- Kept boss floors on their compact boss-floor rules.
- Kept depths 1-3 on the base profile so existing exact dungeon stair/obstacle goldens remain
  stable.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'
make bot scenario=12_dungeon_levels
make bot scenario=28_reachable_dungeon_obstacles
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`.

Full `make ci` was attempted on 2026-06-17. After fixing profile-aware Go reachability tests,
`cd server && go test ./...` passes. The remaining full-CI blocker is the protocol bot catalog:
`teleporter_lab`, `dungeon_monsters`, `boss_floor_gate`, `account_stash_storage`,
`pack_aggro_and_dungeon_packs`, and `ranger_piercing_and_pinning_shots`.

## Scope limits

- No rivers, water visuals, hazards, doors, destructible obstacles, secret obstacles, room/corridor
  PCG, biome colors, new monster definitions, loot-table changes, boss-floor changes, protocol bump,
  client renderer changes, external assets, or final balance tuning shipped.
