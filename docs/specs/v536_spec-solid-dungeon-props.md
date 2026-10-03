# v536 — Solid dungeon props

- **Status:** Complete (combined `make ci` passed 2026-10-03, 20m36s)
- **Date:** 2026-10-02
- **Codename:** `solid-dungeon-props`
- **Baseline:** `dab60eb5` plus v532, v533, v535 in the working tree.
- **Dependencies:** v535 `wider-entrances` is integrated (both slices edit `dungeon_generation.v0.json`, the dungeon goldens and the generation audit).
- **ADRs:** ADR-0001 D2 (server owns collision), ADR-0008 (procedural dungeon), ADR-0018 (kit visuals), Data-Driven Configuration Policy. [v494](../as-built/v494_dungeon-room-dressing.md) is the slice this one changes.

## Purpose

Playtest feedback: barrels and similar room decor can be walked through, which looks wrong. v494 made that decor a client-only, collision-free `MultiMeshInstance3D` planned from the session seed; the server does not know it exists, so collision cannot be added on the client without breaking server authority. Move prop placement to the server and make solid props small blocking obstacles; the client renders what the server sends.

## Scope

1. **Server placement.** New generator `placeDungeonProps` (modeled on the water/hole floor-feature generators) adds props as `wallObstacle`s of a new kind `prop` with a catalog `prop_id`, after monsters, doors and hazards so it can avoid them. It uses its own seeded RNG stream (`seed|props|level|attempt`) so earlier generation output is unchanged, and it is validated with `validateGeneratedDungeonReachability` (a failing attempt is dropped and retried; no prop count that breaks reachability ships).
2. **Placement rules (data, `obstacle_generation.props`).** Props sit inside room interiors only; keep wall clearance, corridor-mouth/door clearance, spacing between props, and the existing obstacle clearance from stairs, teleporters, chests, loot and monsters; count scales with room area and depth bands. The catalog lists each `prop_id` with weight and collision footprint. Footprints are data (no code-owned tuning).
3. **Blocking semantics.** Props block ground movement of players and monsters (all obstacle kinds do) and pathfinding; they do **not** block line of sight or projectiles (they are waist-high clutter); flying/leaping exceptions follow the existing per-kind rules.
4. **Wire.** Additive in place: wall `kind` enum gains `prop`; wall gains optional `prop_id` (`state_delta.v8` and `session_snapshot.v8`); worlds schema kind enum updated if shared.
5. **Client.** `WallRenderer` renders `prop` walls with kit assets (one MultiMesh per asset) from a catalog keyed by `prop_id` in `shared/assets/dungeon_kit_presentation.v0.json` (asset id, scale, yaw choices), and no longer plans props itself. The client planner (`DungeonRoomDressing.plan` and its tests) is deleted, not duplicated; `build` is kept. Yaw is a presentation choice derived from the wall id.

## Non-goals

Destructible or lootable props, new prop art, prop shadows/lighting changes, boss floors, town props, balance of monster density, replacing water/holes/rocks.

## Acceptance criteria

1. With props enabled, generated floors contain props only inside rooms; none overlaps a stair, teleporter, chest, door approach, corridor mouth, monster spawn or another prop, and every target stays reachable (the v531 audit still shows zero generation failures and zero reachability findings).
2. Placement is deterministic for the same seed and level, and enabling props does not change walls, anchors, doors or monsters generated before it (asserted by comparing a props-off and props-on generation).
3. A player cannot walk through a solid prop: a Go sim test moves the player into a prop and is stopped; a click-to-move path routes around it; a bot scenario proves the same through the protocol.
4. Protocol schemas, the Go wall view and the client agree on `kind: prop` and `prop_id`; `make validate-shared` passes with the schema edits.
5. The client draws each prop where the server put it (same position as its collision footprint) and no longer places any prop on its own; a real-renderer capture shows props and their footprints on a generated floor.
6. Rule-derived tests (no pinned counts or coordinates); goldens that move do so deliberately (`make regen-golden` where layout-owned).

## Surfaces and verification

`shared/rules/dungeon_generation.v0.json` + schema (props block); `shared/protocol/state_delta.v8` and `session_snapshot.v8` (wall); `shared/assets/dungeon_kit_presentation.v0.json` + schema; `server/internal/game` (`dungeon_props.go`, rules structs, `WallView`, `obstacle_blocking`/`dungeon_obstacle_variety`, `dungeon_room_layout.go` call site, goldens); client (`wall_renderer.gd`, `dungeon_room_dressing.gd`, loader, `surface_material_room_capture.gd`, tests); one bot scenario; docs. Focused: `go test ./internal/game` (full), determinism lint, `make validate-shared`, client unit tests for the renderer, one protocol scenario, one capture. No per-slice `make ci`.

## Risks

Props consume room interior and path width; the clearance rules and the reachability validation must prevent chokepoints (interacts with v535's wider entrances). Search budgets in small rooms (see v535 findings). Client MultiMesh path must keep the v494 draw-call budget. Footprints are estimates of the kit meshes; the capture checks them visually.
