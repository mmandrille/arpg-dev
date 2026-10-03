# v536 As-built — Solid dungeon props

- **Status:** Complete — focused slice verification plus the combined `make ci` on the integrated v532–v537 state (CI OK, 11 stages, 20m36s, 2026-10-03). `make ci-full` was not run.
- **Date:** 2026-10-02
- **Spec:** [`v536_spec-solid-dungeon-props.md`](../specs/v536_spec-solid-dungeon-props.md) · **Plan:** [`v536_2026-10-02-solid-dungeon-props.md`](../plans/v536_2026-10-02-solid-dungeon-props.md)
- **Base:** `dab60eb5` plus v532, v533 and v535 in the tree.

## What changed

**Server (authority).** `server/internal/game/dungeon_props.go` places props as `wallObstacle`s of a new kind `prop` carrying a catalog `prop_id`, after monsters, doors and hazards. It draws only from its own seeded stream (`seed|props|level|attempt`), so everything generated earlier is unchanged (a test compares props-off and props-on output). Props go only inside room interiors (shape-aware) with wall, door/corridor-mouth and spacing clearance plus the existing stair/teleporter/chest/loot/monster clearance. Each attempt validates reachability of every target and room on a prepared grid (`dungeon_reachability_grid.go` `withObstacles`, so a props attempt does not rebuild the whole floor grid); failed attempts ask for fewer props, and placing none is valid. Props block ground movement and pathfinding like every obstacle, but not line of sight or projectiles. Count and catalog are data (`obstacle_generation.props` in `dungeon_generation.v0.json`: area-per-prop count bands by depth, per-prop weight and collision footprint).

**Wire (additive, in place).** Wall `kind` gains `prop` and wall gains `prop_id` in `state_delta.v8` and `session_snapshot.v8`; `WallView.PropID` in Go.

**Client.** `DungeonRoomDressing.plan` and its helpers are deleted; `placements_from_walls` maps the server's prop walls to a model, scale and yaw (yaw from a stable hash of the wall id) using the catalog keyed by `prop_id` in `dungeon_kit_presentation.v0.json` (shape changed from a weighted list with density/clearance fields to `props: {prop_id: {asset_id, scale, yaw_degrees}}`). `WallRenderer` draws prop walls as one MultiMesh per asset, gives them an empty anchor node, and keeps them out of room corners and floor-tile planning (a prop does not remove a floor tile). `render_wall_layout` and `main.gd` lose the `floor_key`/anchors plumbing and the deferred dressing refresh. The `dungeon-room` showme capture feeds sample prop walls.

**Cross-checks.** New `tools/validate_dungeon_props.py` (called from `validate_shared`, +2 checks) fails if a server prop id has no client presentation or vice versa. A client test asserts each model's rendered extent matches its server footprint within a small tolerance.

## Tests and proof

- `dungeon_props_test.go`: placement inside rooms and clear of targets/doors/corridors/each other; reachability holds; props-off vs props-on generation identical apart from the props; deterministic; **a player walking into a prop is stopped** (sim); rule validation rejects bad configs.
- Protocol scenario `126_dungeon_props` passes (live payload schema gate validates `kind: prop`/`prop_id`).
- Client: `test_dungeon_room_dressing` (placements follow server walls, catalog parity, batching, no floor loss, teardown), `test_factories`, `test_dungeon_kit`, `test_wall_occlusion_fade`, `test_coop_client` pass.
- Goldens: `shared/golden/dungeon_obstacles.json` regenerated (props add walls). `TestPlaceRoomCorridorLayout_NoInteriorScatter` ignores props.
- `make validate-shared` (2,273 checks), `make validate-assets`, `make lint-determinism` pass.

## Limits

- **Real-floor visual:** the capture in `assets/v536/room.png` is the fixture room with sample prop walls (same render path), not a real generated floor; real-floor frames are in v537.
- **Footprints are conservative squares** (1.2–1.4 units) because props rotate with a client-chosen yaw; the table's and stack's true extents are smaller on one axis, so collision is slightly larger than the model there.
- **Generation cost:** props add measurable time at deep levels. The `game` Go package approaches Go's default 10-minute timeout (596 s isolated at the time of this run); see the closing note in the v537 as-built for the timeout decision.
- Not proven: gameplay feel (how often props crowd a doorway); the clearance rules keep props off corridor mouths and doors but were checked statistically (sweep), not exhaustively; `make ci-full` not run.
