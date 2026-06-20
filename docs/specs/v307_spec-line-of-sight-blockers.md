# v307 Spec - Line-of-Sight Blockers

Status: Complete
Date: 2026-06-19
Codename: line-of-sight-blockers

## Purpose

Let tall dungeon obstacles block fog-of-war visibility without changing movement collision.
Wall layout entries can now explicitly carry `blocks_line_of_sight`, the server uses that metadata
when filtering fog-hidden monsters, and the Godot fog overlay uses the same layout metadata to draw
shadow masks behind tall blockers. This makes the v306 rock/column obstacle work affect gameplay
perception while keeping water, holes, and low rubble visible-through terrain.

This builds on v302-v306's obstacle-kind foundation and the existing fog-of-war wall/door
occlusion path. The slice keeps the authoritative source of visibility on the Go server and treats
the client overlay as presentation of the already-authoritative wall layout.

## Non-goals

- No new fog-of-war protocol subsystem, minimap memory changes, durable map exploration, monster AI
  awareness changes, stealth, lighting equipment, or combat balance changes.
- No true polygon, rotated, destructible, or per-piece visibility geometry; LOS uses the existing
  rectangular wall layout AABB.
- No production art, imported obstacle assets, shader packages, Godot plugins, or wall/floor
  material polish.
- No water, hole, rubble, flying-navigation, Barbarian Leap, projectile collision, movement
  collision, or door behavior changes beyond ensuring they continue to interact correctly with LOS.
- No broad generated-dungeon retuning. Generated rock and column obstacles are marked as tall
  occluders by existing kind semantics; rubble remains a low blocker.

## Acceptance Criteria

- Shared preset-world and latest v8 protocol wall schemas accept optional
  `blocks_line_of_sight: true|false` on wall layout objects.
- Existing omitted metadata remains compatible: normal `wall` layout entries still block LOS, while
  water, holes, and rubble do not.
- Generated rock and column solid obstacle groups are marked as LOS blockers in authoritative wall
  views; generated rubble remains movement/projectile blocking but not LOS blocking.
- Preset worlds can explicitly mark rock or column obstacles as LOS blockers for compact labs.
- Server fog-of-war snapshot and delta filtering hide living monsters inside light radius when the
  segment from player to monster crosses a tall rock/column blocker.
- Server fog-of-war continues to show monsters behind water, holes, and rubble inside light radius.
- Opening doors still reveal monsters through the existing closed-door barrier path.
- Client wall normalization preserves `blocks_line_of_sight`, and the fog overlay includes only
  wall/tall-blocker entries when creating shadow masks.
- A compact preset lab proves an initially hidden monster behind a tall column becomes visible after
  the player moves around the occluder.
- A protocol bot scenario proves the hidden-then-revealed visibility behavior.
- A headless visual bot scenario proves the column/rock blocker feeds the fog shadow mask.

## Scope and Likely Files

- Shared contracts/rules:
  - `shared/rules/worlds.v0.json`
  - `shared/rules/worlds.v0.schema.json`
  - `shared/protocol/session_snapshot.v8.schema.json`
  - `shared/protocol/state_delta.v8.schema.json`
- Server fog/layout:
  - `server/internal/game/obstacle_blocking.go`
  - `server/internal/game/dungeon_obstacle_variety.go`
  - `server/internal/game/dungeon_gen.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/types.go`
  - `server/internal/game/fog_of_war_test.go`
  - `server/internal/game/dungeon_obstacle_variety_test.go`
- Client fog/rendering:
  - `client/scripts/wall_renderer.gd`
  - `client/scripts/fog_of_war_overlay.gd`
  - `client/tests/test_fog_of_war_overlay.gd`
  - `client/tests/test_factories.gd`
- Bot proof:
  - `tools/bot/scenarios/102_line_of_sight_blockers.json`
  - `tools/bot/scenarios/client/77_line_of_sight_blocker_shadow.json`
- Docs:
  - `docs/plans/v307_2026-06-19-line-of-sight-blockers.md`
  - `docs/as-built/v307_line-of-sight-blockers.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject imported assets, shader packages, and Godot addons. Borrow the
existing wall renderer, fog overlay shadow-mask path, and protocol/client bot fog assertions.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'FogOfWar|ObstacleVariety|GeneratedObstacleCollisionPaths'`
- `godot --headless --path client --script res://tests/test_fog_of_war_overlay.gd`
- `make client-unit`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blockers ./scripts/bot_local.sh`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blocker_shadow ./scripts/bot_visual.sh`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=line_of_sight_blocker_shadow
```

## Open Questions and Risks

- No required questions for this run. Defaults: rock and column are tall LOS blockers; rubble is a
  low blocker; water/hole remain floor features; the rectangular layout remains authoritative.
- Risk: adding wall-view metadata touches latest protocol schemas and client layout normalization.
  Keep the field optional and preserve legacy wall behavior when the metadata is omitted.
- Risk: `types.go` and `rules.go` are near maintainability limits. Avoid `rules.go` changes in
  this slice and offset any `types.go` line growth.
