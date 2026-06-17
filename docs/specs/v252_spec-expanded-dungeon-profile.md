# v252 Spec - Expanded Dungeon Profile

Status: Complete
Date: 2026-06-17
Codename: expanded-dungeon-profile

## Purpose

Begin the bigger dungeon-level push with a server-authored, data-driven profile for ordinary
generated dungeon floors. Deeper non-boss dungeon levels should become visibly larger and more
populated, with denser generated obstacle layouts, while preserving deterministic generation,
reachable stairs/teleporters/chests/monsters, and existing client wall rendering.

## Non-goals

- No rivers, water visuals, hazards, destructible/secret obstacles, doors in generated walls, or
  full room/corridor PCG.
- No new monster definitions, loot tables, item-level economy changes, combat rebalance, or boss
  floor layout changes.
- No protocol/schema version bump; generated walls, monsters, stairs, and level state continue to
  use existing state-delta/session snapshot shapes.
- No external art, asset pipeline, or plugin adoption. Existing generated ground and wall materials
  remain in use; biome colors/palettes are deferred to a later presentation slice.

## Acceptance Criteria

- Shared dungeon-generation rules define depth-banded floor profiles for ordinary dungeon levels.
- The Go generator applies the matching profile before placing stairs, teleporters, chests,
  obstacles, and monsters.
- At least one deeper non-boss dungeon floor is larger than the entry floor and generates more base
  monsters than the entry profile.
- Profiled floors remain deterministic for the same seed and level.
- Profiled floors keep all generated targets reachable under the existing navigation proof.
- Boss floors keep their existing compact boss-floor sizing and monster-count behavior.
- Existing protocol and client wall rendering continue to consume the generated layouts without a
  schema bump.

## Scope and Likely Files

- Shared rules/schema: `shared/rules/dungeon_generation.v0.json`,
  `shared/rules/dungeon_generation.v0.schema.json`.
- Server: `server/internal/game/dungeon_gen.go`, `server/internal/game/rules.go`, and a focused
  dungeon-profile helper/test file.
- Tests: focused Go tests for profile selection, deterministic generation, reachability, and boss
  floor exclusion.
- Bot proof: existing dungeon scenarios remain representative; run the focused dungeon scenarios
  after unit coverage.
- Docs: plan, as-built, lifecycle, and `PROGRESS.md` close-out.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'`
- `make bot scenario=12_dungeon_levels`
- `make bot scenario=28_reachable_dungeon_obstacles`
- `make maintainability`

## Open Questions and Risks

- No required questions for this run.
- Risk: tuning tests must derive expectations from rules/profile data rather than pinning accidental
  current values.
- Risk closed for this slice: the first profile starts at depth 4 so exact goldens for depths 1-3
  remain stable, and the 120x70 profile stays inside the current generated-ground footprint.
