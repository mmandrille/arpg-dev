# v536 Plan — Solid dungeon props

- **Spec:** [`v536_spec-solid-dungeon-props.md`](../specs/v536_spec-solid-dungeon-props.md) · **Baseline:** `dab60eb5` + v532/v533/v535. Standalone, no branch/worktree. v537 captures after it.
- **Review gate:** pass. Server authority respected (placement and collision on the server); determinism via its own seeded stream; wire change is additive in place (project policy); goldens may move deliberately; ratchets: new logic in new files (`dungeon_props.go`, `dungeon_props_rules.go`); `rules.go` and `wall_renderer.gd` (577 lines) get minimal edits, and the client planner shrinks `dungeon_room_dressing.gd`. Asset/plugin decision: adopt the vendored KayKit assets already used by v494; reject new assets/plugins.
- **Design choice recorded:** reuse `wallObstacle` (kind `prop`) rather than a new entity type, so collision, pathfinding, reachability, wall deltas and replay need no new channel.

## Tasks

- [x] **1. Rules + schema.** `PropGenerationRules` (enabled, max_attempts, clearances, spacing, count bands via `AreaCountFormula`, catalog with footprints) under `obstacle_generation.props`; JSON data and schema; load-time validation; validate_shared green.
- [x] **2. Types + wire.** `obstacleKindProp`, `wallObstacle.propID`, `WallView.PropID`; blocking semantics (no projectile/LOS); schemas (`state_delta.v8`, `session_snapshot.v8`, worlds enum if shared).
- [x] **3. Generator.** `placeDungeonProps` with room-interior sampling, clearance checks, spacing, reachability validation, retry with fewer props; call site after doors and hazards. Tests: determinism, props-off vs props-on unchanged prefix, containment/clearance, reachability, no props when disabled.
- [x] **4. Collision proof (Go).** Sim test: player is stopped by a prop; path planning avoids it.
- [x] **5. Fallout.** Full `go test ./...`; derive or regenerate affected tests/goldens; `make lint-determinism`; the v531 audit run.
- [x] **6. Client.** Catalog by `prop_id`; `WallRenderer` renders prop walls as a MultiMesh; delete the client planner and its tests; update showme capture and `dungeon_dressing` debug state; client unit tests.
- [x] **7. Bot + capture.** Protocol scenario that walks into a prop; real-renderer capture showing props with footprints.
- [x] **8. As-built, CODEMAP, lifecycle, PROGRESS.**

## Final gate

Focused verification only; combined `make ci` once after v532–v537.
