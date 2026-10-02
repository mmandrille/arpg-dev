# v524 Plan — Configurable room topology motifs

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Spec:** [`docs/specs/v524_spec-topology-motifs.md`](../specs/v524_spec-topology-motifs.md)
- **Base commit:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Prerequisites:** v522 `rooms-first` and v523 `room-shapes` were supplied as uncommitted coordinator changes in this detached worktree. The worktree remains at base `876d02c872db4b465e92f460266f83ea6245c3c8`; no dependency commit SHAs are available.
- **Review:** Spec and plan passed coordinator review; implementation remains gated on prerequisite integration.
- **Asset decision:** Adopt vendored KayKit Dungeon kit if room-presentation context is needed; borrow existing wall/entity presentation and rectangle contracts; reject new external assets, plugins, and pipelines.

## Goal and review

Implement configurable hub, branch, and loop motifs in the server-side room graph, with shared schema validation, seeded determinism, connectivity/reachability, and semantic graph tests. The spec fixes motif meanings and scope; it leaves no material product question open. Before implementation, re-read the dependency-integrated version of the room packing, shape, and corridor files and reconcile changes rather than replacing them.

## File map and ownership

| Planned action | Path | Ownership / integration notes |
|---|---|---|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add hub degree and branch-junction ranges. Shared with v522/v523; merge fields into their final config. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Require and bound the new fields; retain `additionalProperties: false`. |
| Modify | `server/internal/game/dungeon_profiles.go` | Extend `RoomCorridorPCGRules` and validate motif ranges and structural feasibility. Likely overlap with v522/v523. |
| Modify | `server/internal/game/dungeon_room_corridors.go` | Build connected motif-constrained edges before existing corridor routing; preserve dependency room/shape support and seeded attempt streams. |
| Modify | `server/internal/game/dungeon_room_corridors_test.go` or a focused topology test file | Semantic connectedness, degree, branch-junction, loop/cycle, determinism, invalid-rule, and reachability assertions; no coordinate goldens. |
| Modify if needed | `server/internal/game/dungeon_generation_rules.go`, `server/internal/game/rules.go` | Propagate and validate the expanded rule struct only if the existing typed path requires it. |
| Modify | `docs/CODEMAP.md` | Keep dungeon-generation file ownership and test index current. |
| Add | `docs/as-built/v524_topology-motifs.md` | Record changes, exact checks, limits, dependency baseline, and handoff manifest. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v524 handoff row after implementation evidence is ready. |

Do not modify protocol schemas, client scripts/assets, encounter population, room roles, or visual manifests. If v522/v523 add or rename shared fields/types, adapt to their integrated API and document any additional overlap in the as-built note.

## Ordered tasks

### Before prerequisites integrate

- [x] Read current baseline, repository instructions, CODEMAP, ADR-0001/0008, v445 room-corridor plan, and existing generator tests.
- [x] Draft and self-review the v524 spec for scope, criteria, determinism, shared-data ownership, server authority, and asset decision.
- [x] Draft this plan and map each criterion to a focused proof.
- [x] Receive coordinator review approval for the spec and plan; preserve the v522/v523 implementation gate.
- [x] Re-read `PROGRESS.md`, `CLAUDE.md`, `docs/CODEMAP.md`, this plan/spec, and the v522/v523 handoff files in the assigned checkout.

### After both prerequisites integrate

- [x] Reconcile the coordinator-supplied room model/config/API changes and record the dependency worktree state; verify the assigned base commit and detached checkout.
- [x] Add schema-backed topology controls and Go validation; `make validate-shared` passes.
- [x] Implement seeded connected topology construction that satisfies hub-degree and non-hub branch-junction ranges, then adds the configured distinct loop edges.
- [x] Add semantic fixture tests for connectedness, simple edges, hub and branch ranges, exact loop/cycle rank, infeasible configuration, and same-seed repeatability.
- [x] Recheck ordinary dungeon reachability and unchanged boss/legacy paths with focused Go coverage. Align generated reachability cell probes with live pathfinding after the retained stair fixture exposed a collision mismatch; update the affected deterministic stair golden.
- [x] Run `make validate-shared`, `make maintainability`, and baseline-aware determinism lint.
- [x] Run `make bot scenario=28_reachable_dungeon_obstacles`; it passes after the reachability fix.
- [x] Update CODEMAP and write as-built evidence; inspect changed/untracked paths and record exact results, evidence limits, dependency worktree state, and overlap risks.

## Acceptance-to-proof map

| Criterion | Proof |
|---|---|
| Shared and Go rules accept valid motif ranges and reject invalid/impossible bounds | Schema validator via `make validate-shared`; focused Go invalid-rule unit cases |
| Connected simple graph obeys hub, branch, and loop controls | Semantic topology-builder fixture tests; assert edge uniqueness, connectivity, degree/junction ranges, and cycle rank = edges − vertices + components |
| Same seed/rules reproduce topology and routed output | Repeat-generation equality over semantic topology and generated output for fixed seed/level; no coordinate assertions across different seeds |
| Generated floors remain reachable | Existing server reachability validator plus focused generation seeds and bot scenario `28_reachable_dungeon_obstacles` |
| Boss and disabled-PCG legacy behavior persist | Existing boss-floor and legacy-divider tests |
| Registry and handoff are complete | CODEMAP validator, maintainability check, and as-built manifest |

## Data, determinism, and authority

All new tunables live in `shared/rules/dungeon_generation.v0.json` and its schema; Go validates constraints with field-specific errors. Graph choices use the existing seeded `RNG` passed from per-attempt seed derivation; never use `math/rand`, wall-clock state, map iteration order, or a new source of unrecorded randomness. Server collision and reachability remain authoritative. No protocol changes are planned.

## Maintainability and test scope

Inspect touched Go files against `make maintainability` limits after dependency integration. Extract a narrowly named topology helper/test file if the corridor coordinator would otherwise grow beyond the ratchet; avoid refactoring unrelated room packing/routing in this slice. Do not run `make ci`, `make ci-full`, commit, push, or `/finish`; the coordinator owns combined CI and closeout after integration.

## Visual/performance evidence

No visual or smoothness claim is part of this slice, so a real-renderer capture and performance comparison are not required. Bot/server evidence proves generation compatibility and reachability only; it does not prove player-perceived layout quality or performance.

## Security boundary

This change only evaluates trusted shared game rules and per-level seeded gameplay randomness. It adds no endpoint, caller-controlled input, authentication, secret/token generation, rendering, or dependency. Keep gameplay RNG distinct from security-sensitive randomness; never generate credentials or tokens from the game PCG stream.
