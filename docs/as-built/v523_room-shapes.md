# v523 Handoff — Room Shapes

Date: 2026-10-02
- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
Base commit: `876d02c872db4b465e92f460266f83ea6245c3c8`
Dependency: v522 Rooms-First Dungeon Anchors is present in the assigned worktree as coordinator-supplied, uncommitted changes from the base above. No separate v522 commit SHA is available; it would be inaccurate to invent one.
Worktree: `/Users/mmandrille/git/arpg-dev-batch/v523-room-shapes` (detached)

## Result

- Added a five-entry shared room-shape catalog: rectangle, L, T, cross, and ring-shaped arena. Each entry is a fixed 3×3 walkability mask with a bounded weight.
- The schema restricts IDs, mask length, weight range, and object fields. Go validation checks the complete vocabulary, duplicate/missing IDs, mask size, active-cell minimum, connectivity, and nonzero total weight.
- Server room selection uses the existing seeded PCG stream. Room boundary walls are derived from exposed mask edges, including the arena's inner boundary. Room centers, containment, corridor doors, reachability, and v522 anchor placement account for the mask.
- Room-corridor chests now sample within connected room footprints so clearance can span adjacent active mask cells. The arena's center void remains blocked.
- Updated dungeon generation goldens for the intentional default mixed-shape layout changes. No protocol changes or new client assets were made.
- Security review: this adds no endpoint or user-controlled geometry. Shared rules remain schema-validated, and layout outcomes remain server-authoritative and deterministic.

## v522 integration state and overlap

The coordinator-supplied v522 changes were already present as unstaged/untracked worktree changes when v523 resumed. I preserved them. v523 intentionally extends overlapping `dungeon_gen.go`, `dungeon_room_corridors.go`, `CODEMAP.md`, and dungeon-layout goldens. The v522 anchor-in-room behavior remains in place. The focused chest regression initially failed after footprint containment narrowed; room-aware sampling fixed it and the full focused room suite then passed.

## Verification evidence

| Command | Result |
|---|---|
| `cd server && go test ./internal/game -run 'TestDungeonRoomShapes|TestPlaceRoomCorridorLayout|TestRoomCorridorLayout_|TestDungeonObstaclesGolden|TestDungeonStairsGolden|TestDungeonTeleportersGolden|TestDungeonTeleportersReplayGolden|TestGuardedChestGenerationGolden|TestBossFloorGenerationGolden' -count=1` | PASS (`ok`, 40.263s). Covers each forced shape, bounds, active-cell reachability, anchor containment, determinism, malformed rules, connected-cell clearance, v522 forced chest and rectangle-only compatibility, room layout reachability, and legacy/boss/golden cases. |
| `make validate-shared` | PASS: 2,269 shared checks and CODEMAP validation. |
| `cd server && go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...` | PASS: 77 grandfathered map-range sites in 22 files. |
| `./scripts/check-file-size-ratchet.sh` | PASS. |
| `python3 ./scripts/check-extraction-coupling-ratchet.py` | PASS: 0 coupled helper injections. |
| `make bot-visual scenario=wall_floor_dungeon_rollout` | PASS: visible Godot client; 1 passed, 0 failed. Godot reported 4.7.2. The scenario waits for a WebSocket, at least eight walls, at least three generated walls, and at least one generated non-perimeter wall. |
| `git diff --check` and trailing-whitespace scan of the new Go files | PASS: no diff whitespace errors or trailing-whitespace matches. |

### Initial failures and resolution

- The first focused room-suite run failed `TestRoomCorridorLayout_PreLayoutAnchorsInsideRooms` with “expected forced guarded chest.” Shape containment exposed that chest selection sampled the whole floor and rejected most points. It now samples a room footprint; the full room suite passes.
- The first `make validate-shared` installed the detached worktree's `.venv`, passed all 2,269 shared checks, then failed because the new Go shape files were not yet in `CODEMAP.md`. The CODEMAP row was updated and the final run passed.
- An initial determinism-lint invocation omitted the repository baseline and reported pre-existing grandfathered map ranges. The baseline-aware command above passes.
- An initial Go compile caught import cleanup errors from the room/anchor refactor; imports were corrected before the passing focused run.
- The visual scenario caused Godot to modify tracked asset `.glb.import` metadata. Those generated metadata edits were restored; no such files remain in the change list.

## Evidence limits

- The visual scenario proves one fixed depth-1 client-renderer flow and its wall-count assertions only. It does not assert that every shape appears in that run, or prove per-shape navigation, human readability, gameplay quality, or performance. No screenshot/capture artifact was retained.
- The forced-shape Go fixtures independently prove every catalog shape's in-bounds room geometry and reachability for the tested seeds. They do not establish broad balance or a performance impact.
- At the v523 integration point, a rectangle-only fixture matched all 22 v522 room-wall records (IDs and order) by SHA-256. v524 later intentionally changes the room graph and therefore doorway count; the combined-state regression now asserts that rectangle masks retain the full four-sided footprint and that selecting the sole positive shape weight consumes no extra PCG draw. This keeps the v523 compatibility proof scoped to the shape change rather than v524's topology change.
- No `make ci`, `make ci-full`, commit, push, or `/finish` was run. The coordinator completed lifecycle/progress closeout; combined `make ci` passed.

## Complete worktree change inventory

### v523 changes

- Modified: `docs/CODEMAP.md` (adds the v523 geometry helper and test to v522's dungeon row); `server/internal/game/dungeon_gen.go` (room-aware chest placement); `server/internal/game/dungeon_profiles.go`; `server/internal/game/dungeon_room_corridors.go`; `server/internal/game/dungeon_room_perimeter_walls.go`; `shared/rules/dungeon_generation.v0.json`; `shared/rules/dungeon_generation.v0.schema.json`; `shared/golden/dungeon_obstacles.json`; `shared/golden/dungeon_stairs.json`; `shared/golden/dungeon_teleporters.json`; `shared/golden/guarded_chest_generation.json`.
- Added: `docs/as-built/v523_room-shapes.md`; `docs/plans/v523_2026-10-01-room-shapes.md`; `docs/specs/v523_spec-room-shapes.md`; `server/internal/game/dungeon_room_shapes.go`; `server/internal/game/dungeon_room_shapes_test.go` (includes the rectangle-only footprint/PCG regression; the v522 22-wall parity was measured at the v523 integration point).
- Deleted: none by v523.

### Coordinator-supplied v522 dependency changes in the same worktree

- Modified: `.maintainability/determinism-baseline.tsv`; `docs/CODEMAP.md`; `server/internal/game/dungeon_gen.go`; `server/internal/game/dungeon_room_corridor_sweep_test.go`; `server/internal/game/dungeon_room_corridors.go`; `server/internal/game/dungeon_room_layout.go`; `server/internal/game/wall_floor_lab_nav_test.go`; `shared/golden/dungeon_obstacles.json`; `shared/golden/dungeon_stairs.json`; `shared/golden/dungeon_teleporters.json`; `shared/golden/guarded_chest_generation.json`; `shared/rules/worlds.v0.json`; `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`.
- Deleted: `server/internal/game/dungeon_room_anchor_fallback.go`; `server/internal/game/dungeon_room_anchor_rooms.go`.
- Added: `docs/as-built/v522_rooms-first.md`; `docs/plans/v522_2026-10-01-rooms-first.md`; `docs/specs/v522_spec-rooms-first.md`; `server/internal/game/dungeon_room_anchor_placement.go`; `server/internal/game/dungeon_room_spawn.go`.

### Ignored artifacts observed

- Python validation/runtime environment: `.venv/` (including 2,477 ignored files under `.venv/lib` and 17 under `.venv/bin`) and `arpg_tools.egg-info/` (5 files).
- Godot import cache and generated IDs: `client/.godot/` (448 ignored files), ignored `client/scripts/*.uid` (125), `client/tests/*.uid` (63), plus 2 ignored `client/shaders` files and Python `tools/__pycache__/` (18).
- The visual command created/used local Postgres test database `arpg_test_v523_room_shapes_1398002240`; no screenshot or `.artifacts` output was present.
- These ignored environment/cache artifacts were present at handoff and removed with the detached worktree after path comparison.

## Handoff

The worker handoff left `PROGRESS.md` and the lifecycle registry untouched; coordinator closeout has since registered v522–v531 and combined `make ci` passed.
