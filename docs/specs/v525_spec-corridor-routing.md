# v525 — corridor-routing

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Codename:** corridor-routing
- **Base commit:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependency:** v524 topology motifs; specification and plan may proceed before its integration, implementation may not.

## Purpose

Improve the server-generated room graph corridors so every selected graph edge has a traversable orthogonal route. Let shared dungeon-generation rules configure the widths available to routes and deterministic route variants, while avoiding wall and obstacle collision blockers and preserving access to generated anchors.

The server remains authoritative for generated geometry, collision, reachability, doors, and encounters. Generation remains a deterministic function of the existing seeded PCG path.

## Non-goals

- Changing room graph edge selection, room topology motifs, room roles, door semantics, encounter authority, or spawn policy beyond what is required to keep routes traversable.
- Protocol or persistence changes, client-owned collision, or client-side route generation.
- New corridor wall art, enclosed tunnels, secret/destructible doors, or changes to boss-floor generation.
- New external assets, plugins, or asset pipelines.
- Performance claims or broad dungeon-generation redesign.

## Acceptance criteria

1. Shared `dungeon_generation.v0.json` and its schema define a non-empty set of positive corridor widths. Go rule loading and validation reject empty, non-positive, or otherwise invalid width configuration. Route candidates include both L bend orders and bounded orthogonal doglegs; candidate geometry and ordering are deterministic.
2. Route and width choices use only the existing seeded PCG/RNG stream and stable graph-edge order. Repeated generation with the same seed, level, and rules produces identical rooms, edges/routes, corridor zones, and walls; no `math/rand`, wall clock, or unordered-map iteration affects generation.
3. Every selected MST and loop edge is represented by a non-empty orthogonal route joining the two selected room door anchors. No edge may be silently dropped when a route candidate fails; the layout attempt must choose a valid variant or fail and allow the existing deterministic layout retry path to try again.
4. Corridor floor clearance and doorway gaps are wide enough for the configured width and player collision radius. Every corridor route is free of blocking wall/obstacle geometry, including its turns and both room connections. A route must not intersect another room perimeter in a way that closes a doorway or creates an inaccessible anchor.
5. Generated stairs, teleporters, chests, and other existing reachability targets remain reachable under the authoritative dungeon reachability validator. Tests assert edge-by-edge connectivity in addition to whole-level reachability so a disconnected selected edge cannot pass solely because another graph path exists.
6. Relevant deterministic dungeon goldens are updated only if the intended route changes alter their output. Existing bot scenario `reachable_dungeon_obstacles` passes after v524 is integrated.
7. Visual verification runs `make bot-visual scenario=wall_floor_dungeon_rollout` after v524 is available. Its generated layout/wall capture is evidence of the rendered rollout only; it does not prove every graph edge is traversable, collision-free, or performant.
8. No protocol schemas or client authority change. Any client presentation adjustment is limited to consuming the same server-authored wall and floor rectangles.

## Likely files and ownership

- `shared/rules/dungeon_generation.v0.json` and `shared/rules/dungeon_generation.v0.schema.json` — shared, schema-validated corridor width and route-variant choices.
- `server/internal/game/dungeon_profiles.go`, `server/internal/game/dungeon_generation_rules.go` — Go rule shape, validation, and wiring.
- `server/internal/game/dungeon_room_corridors.go` — deterministic route candidates, door placement, corridor zones, and blocker clearance.
- `server/internal/game/dungeon_room_perimeter_walls.go` — only if doorway-gap emission must follow per-edge widths.
- `server/internal/game/dungeon_room_corridor_sweep_test.go` and focused room-corridor tests — seed/rule fixtures, edge-level route checks, blocker clearance, and anchors.
- `shared/golden/dungeon_obstacles.json` and its validation/schema — only if output source/count contract changes.
- `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` — reuse unless focused proof shows the existing scenario cannot observe a required behavior.
- `docs/as-built/v525_corridor-routing.md`, `docs/progress/slice-lifecycle.md`, and `docs/CODEMAP.md` — handoff and ownership evidence after implementation.

## Existing implementation and asset decision

v445 introduced seeded room packing, a Prim MST plus optional loop edges, open L-shaped corridors, perimeter walls with doorway gaps, corridor zones, and full-level reachability validation. v472 added deterministic anchor-room fallback after generation failures. Current route generation chooses one of two L orders for each edge, uses a single configured corridor width, and relies on whole-level reachability; v525 strengthens that behavior without replacing the existing graph or authority model. The existing `client/scripts/dungeon_kit_wall_builder.gd` and `client/scripts/dungeon_kit_floor.gd` consume server-authored wall rectangles, and the vendored kit is under `client/assets/environment/kaykit_dungeon/`.

- **Adopt:** the already vendored KayKit Dungeon kit where corridor visuals need adjustment.
- **Borrow:** existing wall/entity presentation and server-authored rectangle/collision contracts.
- **Reject:** new external assets, plugins, and asset pipelines.

## Focused verification

- Go tests for width rule validation, route variants, per-edge route continuity, door-gap fit, blockers at straight segments and turns, anchor reachability, deterministic replay, and retry behavior.
- `go test ./internal/game/... -run 'RoomCorridor|DungeonObstacles'` from `server/`.
- `make validate-shared`.
- `make bot scenario=reachable_dungeon_obstacles`.
- `make bot-visual scenario=wall_floor_dungeon_rollout`; inspect the saved capture and state its evidence limits.
- Run `make lint-determinism` if generation code or RNG ordering changes.

The worker does not run `make ci`, `make ci-full`, commit, push, `/finish`, or modify the coordinator checkout. The coordinator owns combined `make ci` after all accepted slices are integrated.

## Security and authority notes

This is deterministic server-side generation, not an endpoint or user-input parsing change. Keep generated collision, route acceptance, and anchor reachability server-authoritative. Shared rule data remains schema-validated; do not accept client-provided route geometry or use non-deterministic randomness.

## Risks and open questions

- v524 changes room topology/motifs, so route tests and tuning must be rebased onto its integrated room geometry before implementation.
- Multiple corridor widths can overlap at turns or near unrelated room walls; validate the full movement-clearance footprint, not only the centerline.
- Existing retry behavior must remain bounded and deterministic if no route variant can satisfy clearance for a generated graph edge.
- Route geometry changes can invalidate deterministic obstacle goldens; update only the affected contract and preserve the existing seed/replay expectations.
- No blocking product question remains; exact width choices and route-variant encoding are implementation details constrained by shared-schema ownership and the criteria above.
