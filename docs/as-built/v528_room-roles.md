# v528 Handoff — Dungeon Room Roles

Date: 2026-10-02
- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
Base commit: `876d02c872db4b465e92f460266f83ea6245c3c8`
Worktree: `/Users/mmandrille/git/arpg-dev-batch/v528-room-roles` (detached)

## Result

- Added schema-backed room role rules for entry, transition, combat, reward/objective, and boss arena. Go validation checks supported role order, unique IDs, weights, cardinalities, available shapes, and exact placement-to-role mappings.
- Assigns generated rooms in stable order through a purpose-separated seeded PCG stream. The fixed spawn room remains first and is assigned entry. Combat is metadata only; monster placement and composition were not changed.
- Routes up stairs and teleporters to entry rooms, down stairs to transition rooms, and eligible guarded/objective chests to reward/objective rooms. Candidate locations account for room masks, configured clearance, obstacle blocking, stair/monster separation, and live-aligned reachability.
- Reward/objective role assignment requires enough connected room interior for the configured chest clearance. Shape-aware anchor selection and teleporter-before-down-stair placement prevent the tested entry-anchor/down-stair conflict while preserving the configured separation.
- The dedicated boss-floor generation path and its boss sequence remain unchanged. The boss-arena role and boss placement mappings are validated and ready for floors that use generated rooms; current boss floors do not create room records.
- Updated seeded dungeon layout goldens for the accepted anchor/layout selection changes. No protocol, client, or external asset changes were made. The lifecycle index and `PROGRESS.md` remain untouched.

## Verification evidence

| Command | Result |
|---|---|
| `cd server && go test ./internal/game -run 'Test(DungeonRoomRoleRulesValidation|AssignDungeonRoomRolesIsDeterministicAndRuleDriven|GeneratedDungeonAnchorsFollowRoomRoles|DungeonRoomShapes_|PlaceRoomCorridorLayout_|RoomCorridorLayout_|DungeonStairsGolden|DungeonTeleporterDiscoveryAndTravel|LoadDiscoveredTeleportersAllowsFreshSessionWaypointTravel|DungeonTeleportersGolden|DungeonTeleportersReplayGolden|GuardedChestGenerationGolden|DungeonObstaclesGolden|EliteObjective|DungeonEliteObjectiveChestRequiresEliteLeader|GeneratedDungeonTargetsReachable|BossFloorGenerationGolden)$' -count=1` | PASS. Includes the 24-seed × 10-depth generation sweep and anchor-conflict seeds. |
| `make validate-shared` | PASS: 2,269 shared checks and CODEMAP validation. |
| `make lint-determinism` | PASS: 77 grandfathered map-range sites in 22 files. |
| `make bot scenario=28_reachable_dungeon_obstacles` | PASS: `OK: protocol bot (28_reachable_dungeon_obstacles)`. |
| `make bot-visual scenario=28_reachable_dungeon_obstacles` | PASS: Godot 4.7.2 replayed the 84-envelope scenario; manifest status is `passed` with `replay_match: true`. The room dressing log reports 18 instances placed from 790 safe candidates. |
| `git diff --check` | PASS. |

## Failure diagnosis and resolution

The coordinator reported that audit seed `audit-15`, level -6, could not place a teleporter and asked for entry-anchor feasibility. Anchor candidates now come from shape-aware interior positions with clearance; up stair and teleporter occupy the entry role, the teleporter is placed before the transition-room down stair, and the down stair candidate must preserve teleporter separation. The named seed sweep and anchor-conflict seed tests pass with this behavior.

The objective-chest regressions exposed two role-placement constraints. Some shaped rooms have no chest-clearance interior, so role assignment now excludes those rooms from reward/objective eligibility. Corridor clearance zones can overlap valid room interiors near a doorway; room-contained objective candidates no longer fail solely because that corridor zone overlaps, while obstacle clearance and generated reachability checks remain enforced. Both leader-kill objective tests now pass.

The resulting seeded anchor locations changed stairs, teleporter travel arrival, guarded-chest position, and obstacle output for the retained golden seeds. Those expectations were refreshed from generated results; the named goldens and shared validator pass. No monster-count or composition expectations were relaxed.

## Scope and evidence limits

- The seeded sweep covers 24 named seeds at depths -1 through -10, and the bot scenario covers one reachable-obstacle traversal. This is bounded deterministic generation and protocol evidence, not proof over every seed or a broad gameplay session.
- The visual replay validates one pinned floor and its replay match. It does not establish general shape coverage, broad gameplay quality, or performance. Godot modified 18 tracked `.glb.import` files during startup; those generated changes were restored after the run.
- No visual capture or performance measurement was needed or collected. This server-side slice does not establish visual quality or runtime performance.
- Role configuration is trusted shared data validated at schema load and Go semantic validation. Generation remains server-authoritative and deterministic; no endpoint or user-controlled placement input was introduced.
- The worker did not run CI or create a commit. The coordinator integrated the slice, reconciled lifecycle/progress, and passed combined `make ci`; `make ci-full` was not run.

## Dependency state and overlap

The coordinator supplied the v522 Rooms-First, v523 Room-Shapes, and v524 Topology-Motifs work as uncommitted changes on this worktree's base commit. No dependency commit SHAs are available. Those changes were preserved and listed separately in the worker inventory. v528 extends overlapping rules, generation, CODEMAP, and golden files; the coordinator compared the combined paths before integration. The worker checkout was detached at `876d02c872db4b465e92f460266f83ea6245c3c8`.

## Complete worktree change inventory

### v528 changes and shared paths extended by v528

- Modified: `docs/CODEMAP.md`; `server/internal/game/dungeon_elite_objective.go`; `server/internal/game/dungeon_gen.go`; `server/internal/game/dungeon_profiles.go`; `server/internal/game/dungeon_room_corridors.go`; `server/internal/game/dungeon_room_layout.go`; `shared/rules/dungeon_generation.v0.json`; `shared/rules/dungeon_generation.v0.schema.json`; `shared/golden/dungeon_obstacles.json`; `shared/golden/dungeon_stairs.json`; `shared/golden/dungeon_teleporters.json`; `shared/golden/guarded_chest_generation.json`.
- Added: `docs/specs/v528_spec-room-roles.md`; `docs/plans/v528_2026-10-01-room-roles.md`; `server/internal/game/dungeon_room_roles.go`; `server/internal/game/dungeon_room_roles_test.go`; this handoff file.

### Coordinator-supplied v522/v523 changes preserved

- Modified: `.maintainability/determinism-baseline.tsv`; `server/internal/game/dungeon_gen.go`; `server/internal/game/dungeon_room_corridor_sweep_test.go`; `server/internal/game/dungeon_room_layout.go`; `shared/rules/worlds.v0.json`; `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`.
- Deleted: `server/internal/game/dungeon_room_anchor_fallback.go`; `server/internal/game/dungeon_room_anchor_rooms.go`.
- Added: `docs/as-built/v522_rooms-first.md`; `docs/as-built/v523_room-shapes.md`; `docs/plans/v522_2026-10-01-rooms-first.md`; `docs/plans/v523_2026-10-01-room-shapes.md`; `docs/specs/v522_spec-rooms-first.md`; `docs/specs/v523_spec-room-shapes.md`; `server/internal/game/dungeon_room_anchor_placement.go`; `server/internal/game/dungeon_room_spawn.go`; `server/internal/game/dungeon_room_shapes.go`; `server/internal/game/dungeon_room_shapes_test.go`.

### Coordinator-supplied v524 changes preserved

- Modified: `server/internal/game/dungeon_reachability_grid.go`; `server/internal/game/dungeon_room_corridors.go`; `server/internal/game/dungeon_room_perimeter_walls.go`; `server/internal/game/wall_floor_lab_nav_test.go`; `docs/CODEMAP.md`; `shared/golden/dungeon_obstacles.json`; `shared/golden/dungeon_stairs.json`; `shared/golden/dungeon_teleporters.json`; `shared/golden/guarded_chest_generation.json`.
- Added: `docs/as-built/v524_topology-motifs.md`; `docs/plans/v524_2026-10-01-topology-motifs.md`; `docs/specs/v524_spec-topology-motifs.md`; `server/internal/game/dungeon_room_topology.go`; `server/internal/game/dungeon_room_topology_test.go`.

### Ignored artifacts observed

- `.venv/`, `arpg_tools.egg-info/`, `tools/__pycache__/`, `tools/bot/__pycache__/`, `.artifacts/bot-runs/20261002T052438Z-visual.json`, `client/.godot/`, and Godot-generated `client/**/*.uid` sidecars were observed as ignored worktree entries at handoff and removed with the detached worktree after path comparison. The bot scenarios also used a temporary PostgreSQL test database and server logs under the system temporary directory; both reported successful shutdown. Godot's tracked `.glb.import` edits were restored. The visual run retained only its manifest; no screenshot artifact was created. No temporary debug test was integrated.

## Handoff

The worker handoff required path-by-path comparison for overlapping dependency/v528 paths; the coordinator comparison was completed before cleanup. Combined `make ci` passed, lifecycle/progress closeout is recorded, and the coordinator batch commit follows.
