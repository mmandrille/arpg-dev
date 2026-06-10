# Spec: `solid-collision-and-obstacles`

Status: Draft
Branch: `feature/solid-collision-and-obstacles`
Related: ADR-0001 D2/D3/D8, `docs/researchs/godot-plugins-and-shortcuts.md`

## 1. Purpose

Make the authoritative server prevent the player from occupying or crossing through live monsters
and static wall obstacles. Add a deterministic collision test world so protocol bots and the Godot
client can verify blocked movement and routed movement around obstacles.

This slice upgrades movement from "always apply delta" to "attempt a deterministic step and clamp
or reject the blocked portion." The server remains the source of truth; the client may predict but
must reconcile to server positions.

## 2. Non-goals

- No full click-to-move pathfinding.
- No monster AI, monster movement, or monster avoidance.
- No navmesh generation.
- No polygon/capsule collision; use simple 2D circle-vs-circle and circle-vs-AABB checks.
- No protocol schema bump for wall entities unless required by verification.
- No production art; wall visuals may be simple client primitives for this slice.

## 3. Files to create or modify

```text
shared/rules/worlds.v0.schema.json       - allow static wall obstacles in world presets
shared/rules/worlds.v0.json              - add collision_lab world with monster and walls
server/internal/game/rules.go            - parse and validate world wall obstacle data
server/internal/game/sim.go              - authoritative movement collision checks
server/internal/game/game_test.go        - deterministic movement/collision tests
tools/bot/scenarios/03_collision_lab.json - protocol scenario for blocked and routed movement
tools/bot/run.py                         - add scenario actions/assertions for movement positions
client/scripts/main.gd                   - render simple wall props from collision_lab worlds
PROGRESS.md                         - record v9 when complete
```

## 4. Data shapes

World presets gain optional static obstacle entries:

```json
{
  "type": "wall",
  "position": { "x": 4, "y": 5 },
  "size": { "x": 1, "y": 3 }
}
```

`position` is the obstacle center in the same 2D world coordinate system as entities. `size` is an
axis-aligned rectangle extent in world units. Wall obstacles are rules data only; they do not have
runtime entity IDs, HP, inventory behavior, or persistence state.

Initial collision constants for v9:

```text
player radius: 0.45
live monster radius: 0.45
wall shape: axis-aligned rectangle from world rules
```

## 5. Architecture and flow

```text
move_intent
  -> Sim stores active move direction/duration
  -> each tick attempts one unit of motion
  -> candidate position is checked against static walls and live monsters
  -> if blocked, slide one axis when possible; otherwise remain in place
  -> server emits entity_update only when position changes
  -> client reconciles predicted position to authoritative position
```

The player cannot overlap live monsters. Dead monsters are non-solid so loot pickup/combat flows do
not get blocked by corpses.

Wall obstacles are loaded from the persisted `world_id`, so fresh attach, resume, `/state`, replay,
and visual replay all reconstruct the same collision environment without adding mutable wall state.

## 6. Plugin adoption decision

The adoption checklist in `docs/researchs/godot-plugins-and-shortcuts.md` was consulted for isometric
collision resources.

Decision: **borrow/reject for v9**.

- Borrow: Isometric Collision Asset 2485 as a reference for collision UX and test layouts.
- Reject as dependency: it is a 2D/grid demo and does not own authoritative gameplay. v9 needs
  deterministic Go server collision over the existing 3D orthographic client.

## 7. Acceptance criteria

1. A movement step into a live monster stops before player/monster circles overlap.
2. A movement step into a static wall stops before the player circle intersects the wall AABB.
3. Movement around a wall using explicit waypoints succeeds in the protocol bot.
4. Dead monsters no longer block player movement.
5. Replay and reconnect reconstruct the collision world from `world_id` and remain deterministic.
6. Godot visual replay shows simple walls for the collision lab scenario.
7. `make ci` remains green.

## 8. Open questions

| # | Question | Status |
|---|----------|--------|
| 1 | Should dead monsters remain solid corpses? | Answered for v9: no, dead monsters are non-solid. |
| 2 | Should walls be emitted over protocol? | Answered for v9: no, client renders static walls from shared world rules. |
| 3 | Should pathfinding be part of collision? | Answered for v9: no, scripted waypoints only. |

## 9. Testing plan

1. `cd server && go test ./internal/game/... -run Collision`
2. `make validate-shared`
3. `make bot`
4. `make client-smoke`
5. `make ci`
