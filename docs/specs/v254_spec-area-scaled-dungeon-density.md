# v254 Spec - Area-Scaled Dungeon Density

Status: Complete
Date: 2026-06-17
Codename: area-scaled-dungeon-density

## Purpose

Make ordinary generated dungeon floors denser by deriving monster population, monster pack count,
and obstacle group count from each floor's `width * height`. Larger floors should visibly spawn more
enemies and more obstacle groups than smaller floors, and entry floors should also receive a modest
density increase, while all tuning remains in shared dungeon-generation data.

## Non-goals

- No new monster definitions, monster AI behavior, combat stat rebalance, loot-table changes, XP
  curve changes, or boss-floor population changes.
- No full room/corridor PCG, rivers, doors, destructible/secret obstacles, biome palettes, or
  client renderer changes.
- No protocol/schema version bump; existing generated wall and monster state shapes remain valid.
- No external assets or plugins. Existing in-repo dungeon wall rendering, generated materials, and
  bot scenarios are reused.

## Acceptance Criteria

- Shared dungeon-generation rules define area-density formulas for ordinary-floor monster count,
  monster pack-count range, and obstacle-group range.
- The Go generator applies those formulas from the effective ordinary floor size before placing
  stairs, teleporters, chests, obstacles, and monsters.
- A generated entry floor derives more monsters and obstacle groups than the previous fixed base
  values.
- A generated depth-4 ordinary floor derives more monsters and obstacle groups than the entry floor
  because its effective width and height are larger.
- Formula output is clamped by rule-owned min/max caps, so future dungeon sizes can tune population
  without code edits.
- Generation remains deterministic for the same seed and level.
- Generated stairs, teleporters, chests, obstacles, and monsters remain reachable under the existing
  navigation proof.
- Boss floors keep their current compact size and `boss_floor.monster_count` behavior.

## Scope and Likely Files

- Shared rules/schema: `shared/rules/dungeon_generation.v0.json`,
  `shared/rules/dungeon_generation.v0.schema.json`.
- Server: `server/internal/game/dungeon_profiles.go`, `server/internal/game/rules.go`, and focused
  generation tests.
- Tests and tooling: `tools/validate_shared.py`, focused Go dungeon profile/generation tests, and
  existing dungeon bot scenarios.
- Docs: this spec, implementation plan, as-built note, lifecycle row, and `PROGRESS.md` close-out.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestDungeonDensityFormulas|TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'`
- `make bot scenario=12_dungeon_levels`
- `make bot scenario=14_dungeon_monsters`
- `make bot scenario=28_reachable_dungeon_obstacles`
- `make bot scenario=13_teleporter_lab`
- `make bot scenario=42_pack_aggro_and_dungeon_packs`
- `VERBOSE=1 make bot scenario=account_stash_storage`
- `SCENARIO=44_elite_objective_hud HEADLESS=1 ./scripts/bot_client_local.sh`
- `make maintainability`
- `make ci`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=14_dungeon_monsters
```

## Open Questions and Risks

- No required questions for this run.
- Risk: more monsters can lengthen bot scenarios or increase combat pressure. This slice keeps the
  first formula pass conservative and does not rebalance monster stats.
- Risk: exact dungeon goldens may shift because entry-floor population and obstacles increase.
  Tests should derive expectations from rules except for explicit deterministic layout goldens that
  intentionally own coordinates.
