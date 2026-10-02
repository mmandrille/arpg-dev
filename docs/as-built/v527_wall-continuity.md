# v527 Handoff — Wall Continuity

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Base commit:** `876d02c872db4b465e92f460266f83ea6245c3c8` (v521)
**Dependencies:** v522 Rooms-First and v523 Room Shapes are present as coordinator-supplied,
uncommitted worktree changes. No separate prerequisite SHAs are available in this worktree.
**Worktree:** `/Users/mmandrille/git/arpg-dev-batch/v527-wall-continuity` (detached)

## Result

- Added `server/internal/game/dungeon_wall_continuity_test.go`; no production generation or collision
  code needed correction.
- For each of the five v523 room masks, the test samples every boundary edge endpoint and midpoint
  with the production circle/AABB collision predicate, verifies active cells are reachable from
  the room center, and verifies inactive cells cannot be reached through the enclosed room boundary.
- For every room mask, a two-room fixture checks both intended door centers remain clear and the
  room centers remain path-reachable through the generated corridor.
- A rectangle fixture joins a solid `rock` obstacle to the outside face of a perimeter wall beside
  an intended doorway. It first proves the doorway is navigable without the obstacle, then checks
  the wall/obstacle contact remains blocked and the doorway remains navigable after the join.
- Three deterministic generated seed/level cases (`-1`, `-3`, and `-7`) compare complete output
  across repeated generation, rerun the authoritative target reachability validator, and verify all
  room centers remain mutually reachable.
- Updated the dungeon generation row in `docs/CODEMAP.md` with the new test path.

## Verification

```bash
cd server && go test ./internal/game -run 'WallContinuity|TestDungeonRoomShapes|TestRoomCorridorLayout_PreLayoutAnchorsInsideRooms|TestRoomCorridorLayout_Deterministic|TestPlaceRoomCorridorLayout_Reachability|TestPlaceRoomCorridorLayout_RoomWallsPresent|TestPlaceRoomCorridorLayout_LegacyDividersWhenDisabled|TestPlaceRoomCorridorLayout_BossFloorUnaffected|TestPlaceRoomCorridorLayout_NoInteriorScatter' -count=1
```

PASS (`ok`, 17.882s). This focused set includes v522/v523 room-shape and room-layout regressions,
the v522 rectangle-only wall golden, boss/legacy behavior, new geometry fixtures, and representative
seed determinism/reachability.

```bash
make bot-visual scenario=wall_floor_dungeon_rollout
```

PASS: visible Godot 4.7.2 client; one scenario passed, zero failed. It built and started the local
server, then the client scenario observed the generated depth-1 wall layout and its existing wall
count assertions. No screenshot artifact was retained. This is renderer/input evidence only; it does
not prove collision continuity, every room mask in a live run, broad gameplay quality, or performance.

Additional checks:

- `python3 tools/validate_codemap.py` — PASS (`codemap ok`).
- `git diff --check` — PASS.
- No `make ci` or `make ci-full` was run.

One initial test-fixture attempt placed its outside target beyond the configured navigation grid, so
the baseline doorway was unreachable with or without the obstacle. The fixture was moved inside
floor bounds and now explicitly asserts baseline reachability before adding the obstacle. This was a
test-fixture issue, not a production geometry defect.

## Acceptance limits

The wall samples use representative points on each emitted edge plus navigation checks of every
active/inactive shape cell; they are not an exhaustive continuous-space proof over all possible
wall coordinates. Generated-floor determinism and target reachability cover the three named cases,
not the full 600-seed dependency sweep. The visual scenario covers one fixed generated depth-1 route
and retains no image.

## v527-owned change inventory

- Added: `docs/as-built/v527_wall-continuity.md`, `server/internal/game/dungeon_wall_continuity_test.go`.
- Modified: `docs/CODEMAP.md` (v527 addition is the new test path in the dungeon test column).
- The previously created v527 spec and plan remain untracked handoff artifacts:
  `docs/specs/v527_spec-wall-continuity.md`, `docs/plans/v527_2026-10-01-wall-continuity.md`.
- No files were deleted by v527. No production code, protocol, schema, golden, shared rules, or bot
  scenario file was changed by v527.

## Coordinator-supplied dependency inventory in the same worktree

These v522/v523 changes were present before v527 implementation and are intentionally left intact.

- Modified: `.maintainability/determinism-baseline.tsv`, `docs/CODEMAP.md`,
  `server/internal/game/dungeon_gen.go`, `server/internal/game/dungeon_profiles.go`,
  `server/internal/game/dungeon_room_corridor_sweep_test.go`,
  `server/internal/game/dungeon_room_corridors.go`, `server/internal/game/dungeon_room_layout.go`,
  `server/internal/game/dungeon_room_perimeter_walls.go`,
  `server/internal/game/wall_floor_lab_nav_test.go`, `shared/golden/dungeon_obstacles.json`,
  `shared/golden/dungeon_stairs.json`, `shared/golden/dungeon_teleporters.json`,
  `shared/golden/guarded_chest_generation.json`, `shared/rules/dungeon_generation.v0.json`,
  `shared/rules/dungeon_generation.v0.schema.json`, `shared/rules/worlds.v0.json`,
  `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`.
- Deleted: `server/internal/game/dungeon_room_anchor_fallback.go`,
  `server/internal/game/dungeon_room_anchor_rooms.go`.
- Added: `docs/as-built/v522_rooms-first.md`, `docs/as-built/v523_room-shapes.md`,
  `docs/plans/v522_2026-10-01-rooms-first.md`, `docs/plans/v523_2026-10-01-room-shapes.md`,
  `docs/specs/v522_spec-rooms-first.md`, `docs/specs/v523_spec-room-shapes.md`,
  `server/internal/game/dungeon_room_anchor_placement.go`, `server/internal/game/dungeon_room_shapes.go`,
  `server/internal/game/dungeon_room_shapes_test.go`, `server/internal/game/dungeon_room_spawn.go`.

No v527 edit conflicts with the dependency generation implementation or goldens. `docs/CODEMAP.md`
is shared with both prerequisites; preserve the combined row during integration.

## Ignored evidence and local artifacts

- Godot's import step dirtied tracked `.glb.import` files; they were restored because they were clean
  before the v527 visual command. No tracked asset metadata change remains.
- The visual command created/used `.venv/`, `arpg_tools.egg-info/`, and Godot's ignored
  `.godot/` import cache and `.uid` sidecars (333 under `client/scripts/` and 157 under
  `client/tests/` at handoff). These ignored caches were removed with the detached worktree after path comparison. No `.artifacts` screenshot was produced or
  retained.
- Local test database created by the visual scenario: `arpg_test_v527_wall_continuity_98498186`.
  The test database remains available locally; no database cleanup was performed.

## Closeout ownership

The worker left `PROGRESS.md` and `docs/progress/slice-lifecycle.md` unchanged. The coordinator subsequently closed the integrated batch, updated both registries, and passed combined `make ci`.
