# v524 Handoff — Configurable Room Topology Motifs

Date: 2026-10-02
- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
Base commit: `876d02c872db4b465e92f460266f83ea6245c3c8`
Worktree: `/Users/mmandrille/git/arpg-dev-batch/v524-topology-motifs` (detached)

## Result

- Added shared `hub_degree` and `branch_junction_count` ranges. Validation rejects malformed ranges and combinations for which no tree can satisfy the hub, non-hub branch, and loop constraints. Defaults select a hub degree of 2–3, 0–1 non-hub branch junctions, and the existing 1–2 loop edges.
- Replaced unconstrained MST-plus-random-loops with seeded degree-sequence construction. A shuffled Prüfer sequence builds the connected tree; deterministic candidate ordering and the existing gameplay PCG stream add exactly the requested simple-graph loops without exceeding the configured motif bounds.
- Made enabled hub placement mandatory for a room-layout attempt; failure to place it rejects that attempt rather than silently producing a hub-less layout.
- Added semantic topology tests for connectedness, unique non-self edges, hub degree, branch counts, cycle rank, invalid rules, infeasible motif combinations, and repeated-seed stability.
- Merged overlapping and touching room-wall door gaps before emitting wall segments so multiple connections on one room edge cannot create inverted or degenerate wall pieces.
- Fixed generated reachability checks to probe each navigation cell at the same center used by live movement. The previous lower-corner probe could accept routes that live pathfinding rejected. The retained `wall_seed_00` stair regression now passes; its seeded stairs golden was updated for the newly rejected earlier layout attempt.
- Updated CODEMAP and deterministic dungeon goldens. No protocol, client asset, or third-party dependency changes were made.

## Verification evidence

| Command | Result |
|---|---|
| `cd server && go test ./internal/game -run 'RoomCorridor|Topology|DungeonObstacles|Reachability|StairsGolden|Teleporter|GuardedChestGenerationGolden|BossFloorGeneration' -count=1` | PASS, `ok` (42.090s). Includes the retained wall-lab stair traversal and boss/legacy generation coverage. |
| `cd server && go test ./internal/game -run 'TestRoomWallGapsMergeOverlappingDoors|TestGeneratedWallLabStairsOffsetMoveGoal$' -count=1 -v` | PASS. |
| `make bot scenario=28_reachable_dungeon_obstacles` | PASS: `OK: protocol bot (28_reachable_dungeon_obstacles)`. The first run before reachability alignment timed out approaching the level-1 down stairs; after the fix the exact same scenario passed. |
| `make validate-shared` | PASS: 2,269 shared checks and CODEMAP validation. |
| `make maintainability` | PASS: file-size, extraction-coupling, and progress-dashboard ratchets. |
| `cd server && go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...` | PASS: 77 grandfathered map-range sites in 22 files. |
| `git diff --check` | PASS (final worktree review). |

## Initial failure and resolution

The first wall-lab bot run failed while moving toward down stairs at `{x:15,y:40}`; live movement stopped outside the room even though generation reachability reported a path. The retained Go navigation scenario reproduced it. Investigation showed that generated reachability checked static walls at each grid cell's lower corner, while live navigation checks at the cell center. Aligning the generated grid probe with the live pathfinder caused the invalid seeded layout to be rejected, selected a reachable deterministic layout, and changed the stairs golden. The retained Go scenario and the exact bot scenario both pass after the fix.

The overlap-gap helper also has direct horizontal and vertical regression coverage. It preserves positive wall segments when graph edges put multiple door intervals together.

## Dependency state and overlap

The coordinator supplied v522 Rooms-First and v523 Room-Shapes in this worker checkout as uncommitted changes based on the same `HEAD`; no separate dependency commit SHAs were available. The worker preserved the supplied work and extended overlapping shared files. The coordinator compared and integrated the full v522 → v523 → v524 state together:

- `server/internal/game/dungeon_profiles.go`
- `server/internal/game/dungeon_room_corridors.go`
- `server/internal/game/dungeon_room_perimeter_walls.go`
- `server/internal/game/dungeon_reachability_grid.go`
- `docs/CODEMAP.md`
- `shared/rules/dungeon_generation.v0.json` and its schema
- `shared/golden/dungeon_obstacles.json`, `dungeon_stairs.json`, `dungeon_teleporters.json`, and `guarded_chest_generation.json`

No unresolved acceptance criterion remains within focused verification. The coordinator completed integration and lifecycle/progress closeout; combined `make ci` passed. The generation and bot evidence does not establish visual quality or performance.

## Complete worktree change inventory

### v524 changes

- Modified: `docs/CODEMAP.md`; `docs/progress/slice-lifecycle.md`; `server/internal/game/dungeon_profiles.go`; `server/internal/game/dungeon_room_corridors.go`; `server/internal/game/dungeon_room_perimeter_walls.go`; `server/internal/game/dungeon_reachability_grid.go`; `server/internal/game/wall_floor_lab_nav_test.go`; `shared/rules/dungeon_generation.v0.json`; `shared/rules/dungeon_generation.v0.schema.json`; `shared/golden/dungeon_obstacles.json`; `shared/golden/dungeon_stairs.json`; `shared/golden/dungeon_teleporters.json`; `shared/golden/guarded_chest_generation.json`.
- Added: `server/internal/game/dungeon_room_topology.go`; `server/internal/game/dungeon_room_topology_test.go`; `docs/specs/v524_spec-topology-motifs.md`; `docs/plans/v524_2026-10-01-topology-motifs.md`; this handoff file.
- No v524 deletions.

### Coordinator-supplied v522/v523 changes preserved in this worktree

- Modified: `.maintainability/determinism-baseline.tsv`; `server/internal/game/dungeon_gen.go`; `server/internal/game/dungeon_room_corridor_sweep_test.go`; `server/internal/game/dungeon_room_layout.go`; `shared/rules/worlds.v0.json`; `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`.
- Deleted: `server/internal/game/dungeon_room_anchor_fallback.go`; `server/internal/game/dungeon_room_anchor_rooms.go`.
- Added: `docs/as-built/v522_rooms-first.md`; `docs/as-built/v523_room-shapes.md`; `docs/plans/v522_2026-10-01-rooms-first.md`; `docs/plans/v523_2026-10-01-room-shapes.md`; `docs/specs/v522_spec-rooms-first.md`; `docs/specs/v523_spec-room-shapes.md`; `server/internal/game/dungeon_room_anchor_placement.go`; `server/internal/game/dungeon_room_spawn.go`; `server/internal/game/dungeon_room_shapes.go`; `server/internal/game/dungeon_room_shapes_test.go`.

### Ignored artifacts observed

- `.venv/` and `arpg_tools.egg-info/` from shared validation; `tools/__pycache__/` and `tools/bot/__pycache__/`; and the local Postgres test database used by the bot scenario.
- These ignored worktree artifacts were present at handoff and removed with the detached worktree after path comparison. No visual capture was created or needed for this server-side slice.

## Handoff

At worker handoff the slice awaited integration; it is now integrated and listed in `PROGRESS.md` and the lifecycle index, with combined `make ci` passed.
