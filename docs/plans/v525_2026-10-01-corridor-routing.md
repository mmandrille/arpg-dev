# v525 Plan — corridor-routing

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01 (local project date)
- **Base commit:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Goal:** For each selected room-graph edge, produce a deterministic, collision-clear orthogonal corridor with a schema-configured width and retain whole-floor anchor reachability.
- **Prerequisite:** v524 topology motifs. Spec and plan work is valid at this base; do not begin implementation until the coordinator integrates v524 into this worktree and the merged room geometry is inspected.

## Review Gates

### Spec review — PASS

- Scope preserves the existing room graph, server authority, deterministic seeded PCG, open corridor and door model, and protocol contracts.
- Acceptance criteria require proof per selected edge as well as whole-level reachability, including corridor turns, door gaps, blocker clearance, deterministic retries, and existing anchors.
- Widths stay in schema-validated shared rules; route candidates are orthogonal and deterministically ordered.
- The required visual scenario and the limits of its evidence are explicit.
- Asset decision checked against ADR-0018, the vendored `client/assets/environment/kaykit_dungeon/` assets, and `dungeon_kit_wall_builder.gd` / `dungeon_kit_floor.gd`: adopt the vendored KayKit Dungeon kit if visuals are touched, borrow current rectangle contracts, reject new external assets/plugins/pipelines.
- No unresolved product decision blocks planning.

### Plan review — PASS

- v524 is the sole prerequisite and is now present as coordinator-supplied, uncommitted dependency changes on the recorded base; its handoff explicitly reports no separate dependency commit SHA.
- Reconciliation confirms this plan must extend the existing scalar `corridor_width` alongside v523 shape fields and v524 topology fields, while preserving room-first anchor placement, the 3x3 shape masks and inner arena walls, merged door gaps, hub/branch/loop constraints, and live-grid reachability.
- The selected graph edge set must be retained or made directly testable in an internal-only layout result so tests can verify every edge without changing wire data.
- All required correctness proof is server-side. Existing bot and visual scenarios are sufficient for runtime and renderer smoke; neither substitutes for focused edge-level tests.
- Worker final gate is focused verification. Combined `make ci` remains coordinator-owned.

## Dependency and start gate

Before code changes:

- [x] Confirm v524 and its v522/v523 prerequisites are present in this detached worktree as uncommitted changes on `876d02c872db4b465e92f460266f83ea6245c3c8`. No dependency commit SHA was supplied; record that state without inventing one.
- [x] Compare the integrated topology builder, shape-aware boundary/door helpers, room-first anchor placement, rule structs/schema, wall-gap merge, and live reachability validator with v525's likely edits.
- [x] Preserve all dependency contracts and document overlap with the coordinator in the final as-built handoff.

## File map and ownership

| Action | Path | Responsibility / overlap |
|---|---|---|
| Modify | `shared/rules/dungeon_generation.v0.json` | Replace or extend the scalar width with positive width choices; preserve all v524 topology settings. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Require a non-empty array of positive widths; keep `additionalProperties: false` and v524 schema entries. |
| Modify | `server/internal/game/dungeon_profiles.go` | Decode and validate width choices against player traversal clearance; retain existing tuning validation. |
| Modify | `server/internal/game/dungeon_room_corridors.go` | Enumerate stable orthogonal candidates, choose a configured width deterministically, reject candidates crossing blockers/room walls, and preserve every selected graph edge. |
| Modify if needed | `server/internal/game/dungeon_room_perimeter_walls.go` | Emit doorway gaps at least as wide as the selected edge width and player clearance. Avoid merging unrelated doors or weakening anchor egress. |
| Modify | `server/internal/game/dungeon_room_corridor_sweep_test.go` | Expand seeded sweep and determinism coverage for candidate routing, edges, and anchors. |
| Add/modify | focused `server/internal/game/dungeon_room_corridors*_test.go` | Table-driven fixtures for straight, aligned, L, dogleg, blocked first-choice/fallback, invalid widths, route turns, doorway gaps, and edge-level continuity. |
| Modify only if contract changes | `shared/golden/dungeon_obstacles.json`, its schema, and `tools/validate_dungeon_goldens.py` | Refresh deterministic geometry outputs only when required by changed corridor walls. |
| Reuse | `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` | Existing full-floor traversal smoke after v524 integration; modify only if a specific acceptance assertion cannot be observed. |
| Reuse | `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` | Required wall/floor real-renderer capture; no scenario edits expected. |
| Modify | `docs/CODEMAP.md` | Add new helper/test ownership if file inventory changes. |
| Add | `docs/as-built/v525_corridor-routing.md` | Final base/dependency SHA, changed paths, test outcomes, captures, evidence limits, and remaining integration risks. |
| Modify | `docs/progress/slice-lifecycle.md` | Add handoff row after implementation. Coordinator owns canonical `PROGRESS.md` current status unless the integration requires an immediate update. |

Do not change protocol schemas, client collision/authority, room graph selection, boss floors, or `main.gd`.

## Maintenance and data ownership

- Width options and any new generation budget belong in shared rules with schema validation; do not add corridor width literals as tuning in Go.
- Use existing seeded `RNG`, stable edge/candidate ordering, and bounded retry attempts. Never use `math/rand`, wall-clock state, or unordered map traversal in generation.
- Keep route geometry and edge validation in the dungeon-generation owner; do not grow unrelated coordinators. Check touched files against `.maintainability/file-size-baseline.tsv` and run `make maintainability` if implementation moves or adds code beyond focused existing functions.
- Preserve protocol/golden contracts unless generated geometry requires a targeted dungeon-obstacle golden update. Any intentional golden change must pass both Go golden tests and shared validation.

## Ordered tasks

### 1. Rebase the design on v524

- [x] Verify prerequisite integration, record its SHA, and inspect the new room motif/edge data.
- [x] Keep v524 shape generation unchanged; document any shared-file overlap before editing.
- [x] Add schema-backed positive `corridor_widths` options and Go validation. Choose widths only through seeded PCG.

Smallest check:

```bash
make validate-shared
```

### 2. Route every selected edge

- [x] Make route candidates explicit and stable: both L-bend orders plus bounded orthogonal doglegs that leave room for movement clearance.
- [x] For each selected MST/loop edge, select a width and try candidates in seeded order. Reject any centerline/collision footprint crossing a wall, obstacle, unrelated room perimeter, or blocked turn; never silently omit an edge.
- [x] Ensure the route joins its actual room door anchors and that emitted wall gaps use a compatible width. If no candidate is valid, fail the current layout attempt so the existing bounded deterministic layout retry can proceed.
- [x] Keep route, graph edge, corridor zones, and perimeter gaps as internal generation data; do not change serialized protocol data.

Smallest check:

```bash
go test ./internal/game/... -run 'RoomCorridor'
```

### 3. Prove edge and anchor behavior

- [x] Add focused edge-level assertions: every selected edge has a continuous orthogonal floor route from door to door, its full movement-clearance footprint is unblocked, and doorway gaps remain passable.
- [x] Cover routes blocked at the first bend/candidate and prove the next valid seeded candidate is selected; prove impossible layouts retry/fail boundedly instead of dropping edges.
- [x] Cover width validation and multiple configured widths; use rule-derived bounds rather than duplicating live tuning values.
- [x] Preserve full-level reachability for stairs, teleporters, chests, spawn, and the other current targets.
- [x] Extend repeated-generation checks to include route/edge output; run the existing generation seed sweep and targeted golden tests. Refresh only affected golden rows.

Smallest check:

```bash
go test ./internal/game/... -run 'RoomCorridor|DungeonObstacles'
```

### 4. Protocol bot, visual capture, and focused final gate

- [x] Run the existing generated-wall traversal scenario against the integrated state.
- [x] Run the required renderer scenarios; preserve the replay manifest and state evidence limits.
- [x] Validate shared contracts, generation determinism, maintainability, and whitespace.
- [x] Update CODEMAP for the new routing helper/test and write as-built handoff evidence.
- [x] Leave `docs/progress/slice-lifecycle.md` unchanged; the coordinator owns lifecycle/progress closeout during `/finish`.

```bash
make bot scenario=reachable_dungeon_obstacles
make bot-visual scenario=wall_floor_dungeon_rollout
make validate-shared
make lint-determinism
make maintainability
git diff --check
```

Run only the checks relevant to actual changed paths, while always running both named bot scenarios after dependency integration. Do not run `make ci` or `make ci-full`; the coordinator owns combined batch CI.

## Handoff and integration

The worker stops after focused verification with the detached worktree available. Do not commit, push, invoke `/finish`, modify `main`, archive the worktree, or merge the worker into another checkout. Report:

- base SHA and integrated v524 SHA;
- every changed/deleted/untracked path and ignored evidence path;
- exact focused commands and outcomes;
- visual capture location plus what it does and does not prove;
- unresolved acceptance criteria, test gaps, and likely overlaps with v524/v526+ slices.

The coordinator compares this worktree against the integrated state, resolves shared rule/schema conflicts, updates lifecycle/progress registries during `/finish`, and runs combined `make ci` after all accepted slices are integrated.
