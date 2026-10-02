# v529 As-Built — Room Population

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).

- **Slice:** v529 `room-population`
- **Assigned base:** `876d02c872db4b465e92f460266f83ea6245c3c8` (detached worktree)
- **Dependency overlay:** coordinator-supplied v522–v525, v527, and v528 changes are present as
  uncommitted worktree edits; their individual commit SHAs were not supplied. v526 generated doors
  are absent from this worktree and require a combined integration check after merge.
- **Scope at integration:** provide canonical `{room index, role, role weight}` candidates. v530 owns
  complete-pack assignment, final per-room counts, member placement, and elite/guard relationships.

## Implementation

- Added schema-backed `monster_placement.room_role_weights` for the exact v528 role set. Defaults
  are entry 0, transition 1, combat 5, reward_objective 2, and boss_arena 0. The positive transition
  weight lets a reachable transition room receive population when all combat/reward candidates are
  blocked; zero-weight roles remain excluded from assignment.
- The initial v529 implementation preallocated the ordinary pack-size units among room budgets. During
  v530 integration this was replaced by canonical candidate rows carrying configured role weights;
  complete packs are now assigned with actual formation and obstacle constraints in view, and final
  `MonsterCount` values are written only after the whole-floor placement succeeds.
- The existing pack-size RNG draw remains in its original stream and order. The room-capacity
  predicate uses configured body/per-target clearances for obstacles, corridors, stairs,
  teleporters, and chests, and retains player minimum-spawn distance and reachability checks. The
  legacy full pack-spread placement predicate remains unchanged. v530 must use the room-capacity
  helper as its initial member eligibility filter and then enforce full pack spacing during pack
  placement.
- Added focused unit coverage for deterministic budgets and totals, role weighting and validation,
  usable-area weighting, room-layout-disabled behavior, body-clearance next to a corridor, configured
  anchor clearances, and the formerly failing transition-room capacity seed. Updated
  `docs/CODEMAP.md` for the new helper and test.

## Verification evidence

- `make validate-shared` — **PASS**, 2269 checks; CODEMAP validation passed.
- `cd server && go test ./internal/game -run 'TestDungeonRoomPopulation|TestGeneratedDungeonTargetsReachable|TestBossFloorGenerationGolden|TestDungeonMonsterGeneration|TestDungeonObstaclesGolden|TestPlaceRoomCorridorLayout_BossFloorUnaffected' -count=1` — **PASS**.
- Seed sweep — **PASS**, 24 deterministic seeds × ordinary floors −1 through −4 = 96 generations,
  zero errors. The one-off sweep harness was removed after the run.
- `cd server && go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...` — **PASS**, 77 grandfathered map-range sites in 22 files; no new v529 finding.
- `make bot scenario=28_reachable_dungeon_obstacles` — **PASS**.
- `make bot scenario=12_dungeon_levels` — **PASS**.
- `make bot scenario=14_dungeon_monsters` — **PARTIAL / timeout** on two isolated runs at 15.22s and
  15.36s against the 15.00s scenario limit. Both runs completed gameplay assertions, reconnect, and
  replay before the runner rejected elapsed time; no scenario edits were made. A first concurrent
  attempt collided during PostgreSQL schema setup and is not counted as a gameplay result.
- `cd server && go run ./cmd/determinism-lint ./internal/game/...` without the repository baseline —
  **EXPECTED FAILURE** from inherited package-wide map-range findings; the baseline-aware gate above
  passes.

## Integrated contract amendment

The first v529 worker result assigned aggregate room counts before v530 knew whether complete pack formations would fit. The v530 solver now consumes the same canonical room roles and weights, assigns the unchanged pack-size multiset across rooms, and writes final room counts after placement succeeds. This preserves the target and each pack size while avoiding false capacity estimates. The v526 door clearance check is now part of the integrated layout path and is covered by the combined focused suite.

## Integrated verification and evidence limits

- The v526/v529/v530 focused room, door, population, composition, objective, and monster-generation suite passes after post-placement door clearance validation was added. The coordinator integrated this contract and combined `make ci` passed.
- Candidate weights express room preference; v530 complete-pack placement validates actual per-room fit and preserves the unchanged target and pack-size multiset.
- v529's initial scenario 14 runs completed gameplay assertions but narrowly exceeded the 15-second gate. v530 reran scenario 14 with a test-only debug progression fixture and passed; this is protocol-flow evidence, not a balance claim.
- The worker did not run combined CI or create a commit; coordinator integration, CI, and lifecycle/PROGRESS closeout are complete.

## Worktree inventory

v529-owned paths:

- `docs/specs/v529_spec-room-population.md`
- `docs/plans/v529_2026-10-01-room-population.md`
- `docs/as-built/v529_room-population.md`
- `docs/CODEMAP.md`
- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/dungeon_generation.v0.schema.json`
- `server/internal/game/dungeon_gen.go`
- `server/internal/game/dungeon_generated_types.go`
- `server/internal/game/dungeon_generation_rules.go`
- `server/internal/game/dungeon_room_layout.go`
- `server/internal/game/dungeon_room_population.go`
- `server/internal/game/dungeon_room_population_test.go`
- `server/internal/game/rules.go`

Inherited v522–v525/v527/v528 overlay paths to preserve for coordinator integration:

- Modified: `.maintainability/determinism-baseline.tsv`,
  `server/internal/game/dungeon_elite_objective.go`, `server/internal/game/dungeon_profiles.go`,
  `server/internal/game/dungeon_reachability_grid.go`,
  `server/internal/game/dungeon_room_corridor_sweep_test.go`,
  `server/internal/game/dungeon_room_corridors.go`,
  `server/internal/game/dungeon_room_perimeter_walls.go`,
  `server/internal/game/wall_floor_lab_nav_test.go`, `shared/golden/dungeon_obstacles.json`,
  `shared/golden/dungeon_stairs.json`, `shared/golden/dungeon_teleporters.json`,
  `shared/golden/guarded_chest_generation.json`, `shared/rules/worlds.v0.json`, and
  `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`.
- Deleted: `server/internal/game/dungeon_room_anchor_fallback.go` and
  `server/internal/game/dungeon_room_anchor_rooms.go`.
- Untracked: `docs/as-built/v522_rooms-first.md`, `docs/as-built/v523_room-shapes.md`,
  `docs/as-built/v524_topology-motifs.md`, `docs/as-built/v525_corridor-routing.md`,
  `docs/as-built/v527_wall-continuity.md`, `docs/as-built/v528_room-roles.md`,
  `docs/plans/v522_2026-10-01-rooms-first.md`, `docs/plans/v523_2026-10-01-room-shapes.md`,
  `docs/plans/v524_2026-10-01-topology-motifs.md`,
  `docs/plans/v525_2026-10-01-corridor-routing.md`,
  `docs/plans/v527_2026-10-01-wall-continuity.md`,
  `docs/plans/v528_2026-10-01-room-roles.md`, `docs/specs/v522_spec-rooms-first.md`,
  `docs/specs/v523_spec-room-shapes.md`, `docs/specs/v524_spec-topology-motifs.md`,
  `docs/specs/v525_spec-corridor-routing.md`, `docs/specs/v527_spec-wall-continuity.md`,
  `docs/specs/v528_spec-room-roles.md`, `server/internal/game/dungeon_room_anchor_placement.go`,
  `server/internal/game/dungeon_room_corridor_routing.go`,
  `server/internal/game/dungeon_room_corridor_routing_test.go`,
  `server/internal/game/dungeon_room_roles.go`, `server/internal/game/dungeon_room_roles_test.go`,
  `server/internal/game/dungeon_room_shapes.go`, `server/internal/game/dungeon_room_shapes_test.go`,
  `server/internal/game/dungeon_room_spawn.go`, `server/internal/game/dungeon_room_topology.go`,
  `server/internal/game/dungeon_room_topology_test.go`, and
  `server/internal/game/dungeon_wall_continuity_test.go`.

Ignored local validation artifacts: `.venv/`, `arpg_tools.egg-info/`, `tools/__pycache__/`, and
`tools/bot/__pycache__/`. The shared dungeon-generation rules/schema and generator files listed
above contain both v529 edits and inherited room-role/layout integration changes; reconcile them
carefully with v530's sibling `monster_placement.encounter_composition` additions.
