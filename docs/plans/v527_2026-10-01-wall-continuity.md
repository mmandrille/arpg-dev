# v527 Plan — Wall Continuity

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Goal:** Prove and correct generated wall continuity at room, corridor, perimeter, and solid
obstacle joins while retaining rectangle-based authoritative wall output.
**Baseline:** `876d02c872db4b465e92f460266f83ea6245c3c8` (v521)
**Dependencies:** v522 and v523 must be integrated into this worktree before implementation.

## Review and dependency gate

- [x] Scope retains server-owned collision/reachability and existing rectangle renderer input.
- [x] No public protocol/schema or shared tuning change is planned.
- [x] Seeded PCG, deterministic ordering, and replay-visible generation remain required.
- [x] Asset choice recorded: adopt vendored KayKit Dungeon only if visuals are needed; borrow
  existing wall/entity presentation and rectangle contract; reject new assets/plugins/pipelines.
- [x] Coordinator/owner accepts this plan.
- [x] v522 and v523 changes are present in this worktree and were reviewed before implementation.
  They were copied as uncommitted changes, so there are no separate prerequisite commit SHAs to
  report; the assigned HEAD remains the recorded v521 base.

Both review and dependency gates were satisfied before implementation started.

## File map and ownership

| Action | Path | Owner / purpose |
|---|---|---|
| Inspect, modify only if needed | `server/internal/game/dungeon_room_perimeter_walls.go` | Room edge segmentation and door gaps |
| Inspect, modify only if needed | `server/internal/game/dungeon_room_corridors.go` | Room/corridor construction and generated seam data |
| Add or extend | `server/internal/game/dungeon_wall_continuity_test.go` | Deterministic fixture and generated-layout wall continuity proofs |
| Reconcile if a new test/helper is added | `docs/CODEMAP.md` | Dungeon generation map |
| Add on implementation completion | `docs/as-built/v527_wall-continuity.md` | Evidence and limits |
| Update only at coordinator closeout | `docs/progress/slice-lifecycle.md`, `PROGRESS.md` | Batch lifecycle/status registries |

Do not edit protocol definitions, client collision, shared rules, or golden outputs unless
post-dependency evidence proves the existing rectangle contract cannot represent the correction.
Any such need is a plan change and requires review before implementation.

## Ordered tasks

### 1. Rebase the investigation on integrated prerequisites

- [x] Verify v522 and v523 changes are present in this worktree without replacing this spec or plan.
- [x] Read their changed paths and behavior; rerun only their focused checks necessary to establish
  the dependency baseline.
- [x] Trace the resulting rectangle emitters, obstacle blocking predicate, player collision
  predicate, and generated target validator. Select implementation files from that evidence.

### 2. Add deterministic geometric edge-case tests

- [x] Cover wall rectangle bounds and joins for room corners, a perimeter-adjacent room, an intended
  doorway/corridor, and a solid obstacle touching or overlapping a room/corridor seam.
- [x] Assert with the production collision radius and AABB blocking semantics that solid runs have
  no traversable leak and intentional openings preserve player clearance.
- [x] Include shape/seam context in failure messages and seed/level context in generated cases.

Smallest check:

```bash
go test ./internal/game -run 'WallContinuity'
```

### 3. Exercise deterministic generated layouts and correct defects

- [x] Select representative normal dungeon seed/level cases from current generator regressions,
  including cases changed by v522/v523.
- [x] Assert repeat generation has identical rectangle walls and outcomes.
- [x] Run generated target reachability through the existing authoritative validator; add geometry
  checks only where the grid may hide sub-cell clearance defects.
- [x] No production defect requiring a generator change was demonstrated; this slice adds proofs
  without changing generation, the rectangle contract, tuning, RNG stream, or client authority.
- [x] Production generator and boss/legacy behavior remain unchanged by v527.

Smallest check:

```bash
go test ./internal/game -run 'WallContinuity|TestDungeonRoomShapes|TestRoomCorridorLayout_PreLayoutAnchorsInsideRooms|TestRoomCorridorLayout_Deterministic|TestPlaceRoomCorridorLayout_Reachability|TestPlaceRoomCorridorLayout_RoomWallsPresent|TestPlaceRoomCorridorLayout_LegacyDividersWhenDisabled|TestPlaceRoomCorridorLayout_BossFloorUnaffected|TestPlaceRoomCorridorLayout_NoInteriorScatter' -count=1
```

Result: PASS (`ok`, 17.882s). The initial obstacle fixture failure was traced to an outside target
beyond the navigation grid, not a production wall defect; the corrected fixture proves baseline and
post-join doorway traversal.

### 4. Renderer regression and handoff evidence

- [x] Run `make bot-visual scenario=wall_floor_dungeon_rollout` after v522/v523 are available.
- [x] Record the route result and that it rendered the integrated generated wall layout.
  Treat it only as renderer/input evidence, not as server collision proof.
- [x] Write `docs/as-built/v527_wall-continuity.md` with changed paths, focused commands/results,
  dependency SHAs, deterministic and visual evidence, acceptance limits, and unresolved criteria.
- [x] Update CODEMAP for the newly added test path. Leave lifecycle and canonical progress
  status for coordinator closeout after integration and combined CI.
- [x] Report the complete changed/deleted/untracked path list and any ignored evidence to the
  coordinator; identify shared generation/test files that may conflict with sibling slices.

## Security and determinism constraints

- Collision, doorway clearance, and generated-target reachability remain server-owned. The client
  renders the server's rectangle layout and cannot authorize movement through a wall.
- Use only the existing deterministic seeded PCG path. Do not introduce `math/rand`, wall-clock
  behavior, unstable map iteration, or new random token/secret generation.
- No external dependency is warranted for rectangle geometry or reachability checks.

## Final focused gate

```bash
go test ./internal/game -run 'WallContinuity|TestDungeonRoomShapes|TestRoomCorridorLayout_PreLayoutAnchorsInsideRooms|TestRoomCorridorLayout_Deterministic|TestPlaceRoomCorridorLayout_Reachability|TestPlaceRoomCorridorLayout_RoomWallsPresent|TestPlaceRoomCorridorLayout_LegacyDividersWhenDisabled|TestPlaceRoomCorridorLayout_BossFloorUnaffected|TestPlaceRoomCorridorLayout_NoInteriorScatter' -count=1
make bot-visual scenario=wall_floor_dungeon_rollout
```

Do not run `make ci` or `make ci-full` in this slice worktree. The coordinator runs combined
`make ci` after integrating every accepted batch slice. The visual scenario does not establish
collision correctness; server tests do.
