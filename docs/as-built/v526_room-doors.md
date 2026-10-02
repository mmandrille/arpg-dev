# v526 As-Built — Room Doors

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Baseline:** `876d02c872db4b465e92f460266f83ea6245c3c8`.
- **Prerequisites:** v522–v525 and v527 are coordinator-supplied uncommitted overlays on that
  baseline. They have no separate commit SHAs in this worktree.

## Implementation

- Added schema-backed `room_corridor_pcg.doors.enabled/max_count` rules and server validation for
  non-negative bounded counts, enabled corridor generation, and the existing closed wooden-door
  barrier definition.
- Selected only normalized routed room edges whose two thresholds have the accepted route width
  and north/south orientation supported by the current horizontal barrier. Selection uses stable
  edge ordering and a dedicated seed-derived RNG stream. At most one door is added per route.
- Added ordinary closed `wooden_door` interactables at the perimeter wall center while retaining
  the wall opening. Selection is deterministic, limited to loop routes, and retains a candidate only
  when the final generated targets remain reachable. Doors are added after monsters and floor
  hazards so they do not perturb those seeded placements.
- Added deterministic/configuration tests and a bot scenario. The server test proves a closed door
  blocks movement, the existing action opens it and emits the authoritative event, and movement
  crosses afterward. The bot uses a dedicated deterministic interaction fixture, then descends.
- Updated the dungeon obstacle golden to include the newly enabled generated room door and indexed
  the new source/test/scenario in CODEMAP.
- **Security:** Door state stays server-owned. Generation emits the closed state; interaction uses
  the existing target validation, approach/path, and state-transition flow. No client-supplied door
  state or new protocol contract was added.

## Asset Decision

- **Adopt:** Existing vendored KayKit Dungeon kit where dungeon visuals already use it.
- **Borrow:** Existing wall/entity presentation, wooden-door rules, authoritative state, collider,
  and interaction behavior.
- **Reject:** New external assets, plugins, or rendering pipelines.

## Verification Evidence

| Command | Result |
|---|---|
| `make validate-shared` | PASS — 2,271 shared checks; CODEMAP validation passed. |
| Focused v526/dependency Go selection (exact command below) | PASS — focused door, golden, routing, shape, topology, and continuity tests. |
| `make bot scenario=124_room_doors` | PASS — opened generated door through the normal action and descended to level -2. |
| `make bot-visual scenario=124_room_doors` | PASS — replay matched. Godot emitted resource/ObjectDB leak warnings during shutdown. |
| `make maintainability` | PASS — file-size, extraction-coupling, and progress-dashboard ratchets. |
| `cd server && go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...` | PASS — 77 grandfathered map-range sites in 22 files. |
| `gofmt -d server/internal/game/dungeon_door_rules.go server/internal/game/dungeon_room_doors.go server/internal/game/dungeon_room_doors_test.go` | PASS — no formatting changes reported. |
| `git diff --check` | PASS. |

The exact focused Go command was:

```sh
cd server && go test ./internal/game -run 'Test(RoomThresholdDoor|GeneratedRoomThreshold|DungeonObstaclesGolden|DungeonStairsGolden|DungeonTeleportersGolden|GuardedChestGenerationGolden|RoomCorridorRouting|DungeonRoomShapes|RoomConnectionEdges|RoomWallGaps|WallContinuity)' -count=1
```

A broader exploratory command also included room-corridor seed sweeps:

```sh
cd server && go test ./internal/game -run 'RoomThresholdDoor|GeneratedRoomThreshold|Door|RoomCorridorRouting|RoomCorridor|Topology|RoomShapes|WallContinuity|DungeonObstaclesGolden|DungeonStairsGolden|DungeonTeleportersGolden|GuardedChestGenerationGolden|BossFloorGeneration' -count=1
```

It initially reported several `could not place reachable water obstacles after 48 attempts`
failures on the combined overlay when doors were inserted before monster and floor-hazard placement.
The final integrated order adds doors afterward, restricts candidates to loop routes, and drops any
candidate that fails full target reachability. The 24-seed/10-level sweep now passes with doors
enabled:
`cd server && go test ./internal/game -run 'TestRoomCorridorLayout_SeedSweepAlwaysGenerates' -count=1 -timeout 10m`.
The obstacle golden includes the generated door, and the integrated shared-data and focused door
checks pass.

No screenshot artifact was retained as proof of door appearance. Server interaction and traversal
proof is from the focused Go test and the generated-door protocol scenario.

## v526-Owned Paths

- `docs/specs/v526_spec-room-doors.md`
- `docs/plans/v526_2026-10-01-room-doors.md`
- `docs/as-built/v526_room-doors.md`
- `docs/CODEMAP.md`
- `server/internal/game/dungeon_door_rules.go`
- `server/internal/game/dungeon_profiles.go`
- `server/internal/game/rules.go`
- `server/internal/game/dungeon_room_layout.go`
- `server/internal/game/dungeon_room_doors.go` (new)
- `server/internal/game/dungeon_room_doors_test.go` (new)
- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/dungeon_generation.v0.schema.json`
- `shared/rules/worlds.v0.json` (dedicated deterministic bot interaction fixture)
- `shared/golden/dungeon_obstacles.json`
- `tools/bot/scenarios/124_room_doors.json` (new)

## Combined Worktree Inventory

The worktree also contains the coordinator-supplied, uncommitted v522–v525 and v527 overlays. These
paths remain untouched by v526 except where the v526-owned paths above overlap:

- **Modified:** `.maintainability/determinism-baseline.tsv`; `server/internal/game/dungeon_gen.go`,
  `dungeon_generated_types.go`, `dungeon_reachability_grid.go`, `dungeon_room_corridor_sweep_test.go`,
  `dungeon_room_corridors.go`, `dungeon_room_perimeter_walls.go`, `wall_floor_lab_nav_test.go`;
  `shared/golden/dungeon_stairs.json`, `dungeon_teleporters.json`, `guarded_chest_generation.json`;
  `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`.
- **Deleted:** `server/internal/game/dungeon_room_anchor_fallback.go`,
  `server/internal/game/dungeon_room_anchor_rooms.go`.
- **Untracked documentation from prerequisite overlays:**
  `docs/as-built/v522_rooms-first.md`, `v523_room-shapes.md`, `v524_topology-motifs.md`,
  `v525_corridor-routing.md`, `v527_wall-continuity.md`;
  `docs/plans/v522_2026-10-01-rooms-first.md`, `v523_2026-10-01-room-shapes.md`,
  `v524_2026-10-01-topology-motifs.md`, `v525_2026-10-01-corridor-routing.md`,
  `v527_2026-10-01-wall-continuity.md`;
  `docs/specs/v522_spec-rooms-first.md`, `v523_spec-room-shapes.md`,
  `v524_spec-topology-motifs.md`, `v525_spec-corridor-routing.md`,
  `v527_spec-wall-continuity.md`.
- **Untracked prerequisite game files:**
  `server/internal/game/dungeon_room_anchor_placement.go`,
  `dungeon_room_corridor_routing.go`, `dungeon_room_corridor_routing_test.go`,
  `dungeon_room_shapes.go`, `dungeon_room_shapes_test.go`, `dungeon_room_spawn.go`,
  `dungeon_room_topology.go`, `dungeon_room_topology_test.go`, and
  `dungeon_wall_continuity_test.go`.
- **Ignored generated/runtime paths observed:** `.venv/`, `arpg_tools.egg-info/`, `client/.godot/`,
  generated Godot `client/**/*.gd.uid` and shader `.uid` files, and `tools/**/__pycache__/`.
  These ignored caches were observed at handoff and removed with the detached worktree after path comparison.

No commit, push, combined CI, PROGRESS update, or slice-lifecycle update was performed. The
coordinator completed integration, the combined `make ci` gate, and lifecycle closeout; the focused seed sweep remains recorded above.

## Coordinator Integration Verification

- `make validate-shared` — PASS, 2,271 checks plus CODEMAP validation.
- `cd server && go test ./internal/game -run 'Test(DungeonRoomRoleRulesValidation|AssignDungeonRoomRolesIsDeterministicAndRuleDriven|GeneratedDungeonAnchorsFollowRoomRoles|DungeonRoomShapes_|PlaceRoomCorridorLayout_|RoomCorridorLayout_|DungeonStairsGolden|DungeonTeleporterDiscoveryAndTravel|LoadDiscoveredTeleportersAllowsFreshSessionWaypointTravel|DungeonTeleportersGolden|DungeonTeleportersReplayGolden|GuardedChestGenerationGolden|DungeonObstaclesGolden|EliteObjective|DungeonEliteObjectiveChestRequiresEliteLeader|GeneratedDungeonTargetsReachable|BossFloorGenerationGolden|BossFloorExitsUnlockAfterBossKill|RoomCorridorRouting_|WallContinuity_|RoomThresholdDoor|GeneratedRoomThreshold)' -count=1 -timeout 10m` — PASS.
- `cd server && go test ./internal/game -run 'TestRoomCorridorLayout_SeedSweepAlwaysGenerates' -count=1 -timeout 10m` — PASS, 24 seeds across ten dungeon levels.
- `make bot scenario=room_doors` and `make bot scenario=28_reachable_dungeon_obstacles` — PASS.
- `make maintainability` and `git diff --check` — PASS after extracting generated-dungeon reachability helpers to `dungeon_generation_reachability.go` and lowering the `dungeon_gen.go` baseline to its integrated line count.
