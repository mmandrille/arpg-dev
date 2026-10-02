# v524 Spec — Configurable room topology motifs

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Base commit:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependencies:** v522 `rooms-first`; v523 `room-shapes` (spec and plan may proceed before integration; implementation requires both in this worktree).

## Purpose

Make ordinary dungeon room graphs express configured hub, branch, and loop motifs while retaining connected traversal, reproducible seeded generation, and the existing server-owned reachability validation. Motif controls describe graph structure, not room coordinates or visual styling.

## Scope and behavior

- Add schema-validated topology controls to `room_corridor_pcg` for the requested hub degree and non-hub branch-junction count, alongside the existing loop-edge range.
- Define a hub as the designated `isHub` room when hub-room generation is enabled; its degree is its number of distinct graph neighbors. A branch junction is a non-hub room with at least three distinct neighbors. A loop edge is an edge beyond a connected spanning tree; each adds one independent cycle.
- Generate a simple connected room graph satisfying all configured motif ranges, then route its edges through the existing corridor/door and reachability pipeline. Use the existing level/attempt-specific seeded `RNG` path and stable room/edge ordering.
- Reject infeasible candidate graphs and continue bounded generation attempts. Invalid ranges must fail shared/runtime rule validation with a useful field-specific error. Do not relax graph constraints silently.
- Preserve boss-floor exemption, legacy divider fallback when room-corridor PCG is disabled, and server authority for collision and reachability.

## Non-goals

- No changes to room packing, room shapes, corridor routing, door placement, room roles, encounter population, floor visuals, protocol payloads, or client authority.
- No new randomness source, external assets, plugins, or asset-generation pipeline.
- No promise of particular coordinates, exact route lengths, or visual quality from graph tests.

## Acceptance criteria

1. The shared dungeon-generation schema and Go rule model validate the configured hub-degree and branch-junction ranges; impossible structural bounds are rejected before generation.
2. For deterministic room fixtures and applicable rule fixtures, the topology builder produces a connected simple graph, honors configured hub degree and non-hub branch-junction ranges, and adds exactly a selected in-range number of loop edges beyond its spanning tree.
3. Repeated generation with the same seed, level, rules, and dependency-integrated baseline returns identical room topology and routed output; different seeds remain free to vary.
4. Ordinary generated floors continue to pass authoritative walkability/reachability validation; generated topologies do not create duplicate/self edges or disconnected rooms.
5. Boss floors and the disabled-PCG legacy divider path retain their existing behavior.
6. Focused Go, shared-rule, and bot checks pass. Assertions measure graph connectivity, degree/cycle/branch semantics, and reachability rather than room coordinates.

## Likely files and ownership

- Shared rule contract: `shared/rules/dungeon_generation.v0.json`, `shared/rules/dungeon_generation.v0.schema.json`.
- Server evaluator: `server/internal/game/dungeon_profiles.go`, `dungeon_room_corridors.go`, `dungeon_room_corridors_test.go`, plus generation code only if dependency integration requires a seam adjustment.
- Registries/evidence: `docs/CODEMAP.md`, `docs/progress/slice-lifecycle.md`, `docs/as-built/v524_topology-motifs.md`.
- Existing reachability and scenario contracts remain the authority; add a bot scenario only if current scenarios cannot cover generated ordinary floors with semantic graph evidence.

## Asset decision

- **Adopt:** the already vendored KayKit Dungeon kit if any existing room presentation needs inspection.
- **Borrow:** existing wall/entity presentation and rectangle contracts; this slice emits the same authoritative room/wall/corridor structures.
- **Reject:** new external assets, plugins, or pipelines. No client visual change is intended.

## Verification plan

- `cd server && go test ./internal/game/... -run 'RoomCorridor|DungeonObstacles|Reachability'`
- `make validate-shared`
- Run a focused existing dungeon-generation bot scenario (select from the current scenario catalog after prerequisite integration); compare output through protocol/semantic invariants rather than map coordinates.
- Run `make maintainability` if the final file diff touches maintainability-gated files. Do not run the batch-wide `make ci`/`make ci-full` gate in this slice worktree.

## Integration risks and security boundary

The generator, `RoomCorridorPCGRules`, and `dungeon_generation.v0.json` are likely shared with v522/v523. Rebase no branches: wait for their commits to be integrated into this detached worktree, re-read those files, then adapt the implementation without overwriting their room/shape behavior. This is a bounded algorithm change over trusted shared game rules; it adds no endpoint, authentication, user-controlled input, secret handling, or security-token generation. Preserve the seeded PCG gameplay RNG and server-side authority.
