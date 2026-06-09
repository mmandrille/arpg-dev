# Spec: `reachable-dungeon-obstacles`

Status: Draft
Branch: `main`
Slice: v40 - deterministic interior dungeon obstacles with reachability guarantees
Baseline: v39 `ui-currency-and-mana-polish`
Related:

- [`../PROGRESS.md`](../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - on-demand seeded dungeon generation
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - plugin adoption checklist for client presentation work
- [`v9_spec-solid-collision-and-obstacles.md`](v9_spec-solid-collision-and-obstacles.md) - existing wall collision contract
- [`v11_spec-click-to-move-and-auto-path.md`](v11_spec-click-to-move-and-auto-path.md) - server-owned A* navigation and auto-approach
- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md) - generated levels and stairs
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - generated hostile dungeon mobs
- [`v35_spec-boss-floor-gate.md`](v35_spec-boss-floor-gate.md) - boss floor exception

## 1. Purpose

Generated dungeon floors are currently mostly open rectangles with perimeter walls, random stairs,
a teleporter, optional chest, monsters, and loot. This slice makes normal generated dungeon floors
feel more like explorable dungeon spaces by adding deterministic interior obstacle layouts:
varied wall segments, solid blocks, and connected or bent wall groups.

The main contract is not visual density by itself. The main contract is safe procedural generation:
obstacles may divide the floor into multiple areas and force pathing around corners, but generation
must never leave required gameplay targets unreachable. The player, monsters, auto-pathing,
projectiles, stairs, teleporters, chests, and loot must continue to use the same authoritative wall
collision rules owned by the Go sim.

The first slice should prove a conservative foundation:

- Normal generated dungeon floors include interior AABB obstacles beyond the perimeter walls.
- Obstacle shapes can vary in width and length.
- Obstacle groups can connect into simple non-straight shapes, such as L or T arrangements.
- Solid block clusters can occupy floor space as larger obstacles.
- Every generated spawn or target remains reachable through open passages.
- Boss floors keep the existing v35 authored compact arena behavior unless a future slice opts in.
- Godot renders the authoritative wall layout received from the server, not a locally regenerated
  approximation.

## 2. Non-goals

- No generated doors in obstacle walls for v40. Open passages are the default connectivity solution.
- No full room/corridor PCG system.
- No polygon, rotated, curved, destructible, secret, locked, or one-way obstacles.
- No NavMesh authority or Godot-owned collision/pathfinding.
- No production dungeon art, asset-pack import, texture work, lighting pass, or sound.
- No client-side PCG duplication. The server owns generated layout and sends presentation data.
- No boss floor redesign; level `-5` and future cadence boss floors keep the v35 boss-floor contract
  by default.
- No durable dungeon map persistence across fresh sessions.
- No final density, biome, or difficulty balance pass.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v40_spec-reachable-dungeon-obstacles.md - this slice contract
docs/plans/v40_<YYYY-MM-DD>-reachable-dungeon-obstacles.md - implementation plan
docs/PROGRESS.md - lifecycle update when v40 ships

shared/rules/dungeon_generation.v0.json - obstacle generation tuning
shared/rules/dungeon_generation.v0.schema.json - schema for obstacle tuning
shared/protocol/session_snapshot.v*.schema.json - current-level wall layout in snapshots
shared/protocol/state_delta.v*.schema.json - wall layout updates on level change if needed
shared/protocol/examples/session_snapshot.json - wall layout example
shared/protocol/examples/state_delta.json - wall layout delta example if a delta op is added
shared/golden/dungeon_obstacles.json - deterministic obstacle/reachability fixture
shared/golden/dungeon_obstacles.v0.schema.json - fixture schema if needed
tools/validate_shared.py - rule/golden validation for obstacle tuning and drift

server/internal/game/dungeon_gen.go - deterministic obstacle generation and retry/connectivity loop
server/internal/game/pathfind.go - reuse or expose connectivity checks over generated blockers
server/internal/game/rules.go - obstacle rule parsing and validation
server/internal/game/types.go - snapshot/delta wall layout view types
server/internal/game/sim.go - populate and expose current-level walls
server/internal/game/*_test.go - generation, reachability, replay, and collision tests
server/internal/replay/* - replay parity if snapshot/delta shape changes
server/internal/http/*_test.go - /state parity if wall layout is exposed there

client/scripts/main.gd - render authoritative generated wall layout on snapshot and level change
client/tests/* - focused wall-layout application/debug tests if helper extraction is useful

tools/bot/run.py - scenario assertions/helpers for wall count, reachability, and traversal if needed
tools/bot/scenarios/28_reachable_dungeon_obstacles.json - protocol bot proof
tools/bot/scenarios/client/14_dungeon_wall_rendering.json - client rendering/debug proof if reliable
```

Protocol note: v40 should use one clean coordinated protocol shape for authoritative wall layout.
The current client can render only perimeter walls for generated dungeon levels because it reads
`dungeon_generation.v0.json` locally. Interior generated walls must come from the server so the
client remains presentation-only.

## 4. Data shapes

### 4.1 Dungeon obstacle rules

Add a data-driven obstacle generation block to `shared/rules/dungeon_generation.v0.json`.
The exact field names can be finalized in the plan, but the contract should cover:

```json
{
  "obstacle_generation": {
    "enabled": true,
    "max_attempts": 64,
    "target_group_count": { "min": 4, "max": 8 },
    "wall_segment": {
      "min_length": 3,
      "max_length": 14,
      "thickness": 1.0
    },
    "solid_block": {
      "min_size": { "x": 2, "y": 2 },
      "max_size": { "x": 6, "y": 5 }
    },
    "shape_weights": {
      "line": 4,
      "l": 3,
      "t": 2,
      "block": 2
    },
    "clearance": {
      "player_spawn": 3.0,
      "stairs": 2.0,
      "teleporter": 2.0,
      "chest": 2.0,
      "monster": 1.5,
      "loot": 1.5
    }
  }
}
```

Rules validation must reject impossible or unsafe settings, including non-positive dimensions,
empty shape weights, negative clearances, and attempt counts that cannot be retried.

### 4.2 Wall layout view

Expose current-level walls as authoritative layout data in `session_snapshot` and level-change
deltas. The plan may choose exact field names, but the preferred shape is a top-level array:

```json
{
  "walls": [
    {
      "id": "wall_-1_0001",
      "position": { "x": 12.5, "y": 7.0 },
      "size": { "x": 9.0, "y": 1.0 },
      "source": "generated"
    }
  ]
}
```

Notes:

- `position` and `size` use the existing 2D world-space convention for wall AABBs.
- `id` only needs stable ordering/debug identity; walls are static and do not need to become
  gameplay entities.
- `source` may be `perimeter`, `generated`, or `preset` if useful for debugging. It is optional
  if the plan keeps the schema smaller.
- Snapshots must include the complete current-level wall layout.
- On level changes, either the first delta for the destination level includes a complete wall
  layout update, or the server sends a fresh snapshot-equivalent layout before entity rendering.
- Reconnect and replay timeline output must expose the same wall layout as live play.

## 5. Architecture and flow

### 5.1 Generation order

The generator should keep deterministic order and avoid consuming the main sim RNG. The preferred
normal-floor flow is:

```text
seed + level
  -> perimeter walls
  -> stairs
  -> teleporter
  -> optional chest
  -> guaranteed loot
  -> candidate interior obstacle groups
  -> connectivity validation
  -> monster placement in reachable free cells
  -> final reachability validation
  -> populate LevelState
```

If monster placement works better before obstacles, the plan may choose that order, but the final
validation must still prove all generated monsters and target objects are reachable. Any new RNG
streams must be labeled and stable, for example `seed|obstacles|abs(level)`.

### 5.2 Connectivity contract

For each non-boss generated dungeon level, generation must validate reachability from the player
spawn or arrival marker to every generated target:

- stairs up,
- stairs down,
- teleporter,
- guarded chest when present,
- guaranteed floor loot,
- every generated monster spawn,
- any other generated interactable or pickup added by existing rules.

Reachability should be checked through the same rasterization/cell-size assumptions used by
authoritative auto-pathing. A floor that fails connectivity must retry obstacle generation with the
same deterministic RNG stream. If all attempts fail, generation must return an explicit error rather
than silently creating an unreachable floor.

### 5.3 Obstacle shapes

The first implementation must support these shape families:

- `line`: one horizontal or vertical AABB wall segment.
- `l`: two connected perpendicular AABB segments.
- `t`: three connected AABB segments.
- `block`: one larger AABB solid block.

The generator may merge overlapping/touching AABBs only if the merged result remains deterministic
and does not erase the intended shape diversity from tests. Otherwise, keep the generated AABBs as
separate wall obstacles.

### 5.4 Collision, navigation, and combat

Generated walls must reuse existing `wallObstacle` collision behavior:

- Player movement slides or blocks against generated walls like existing world walls.
- Monster movement and chase pathing route around generated walls.
- Auto-approach and click-to-move route around generated walls.
- Projectile sweeps collide with generated walls.
- Loot drop placement avoids generated walls.
- Travel arrival placement avoids generated walls.

No gameplay authority moves to Godot. Godot only renders wall meshes from server-provided layout
and keeps using existing input paths.

### 5.5 Client presentation and plugin decision

The implementation plan must record the Godot shortcut decision required by `AGENTS.md`.
Expected decision for this slice:

```text
Reject plugin adoption. v40 needs simple server-provided AABB wall boxes rendered through the
existing in-repo Godot wall path. No UI framework, camera plugin, asset pack, or external collision
tool should be adopted for the first proof.
```

## 6. Determinism and replay

The generator must be deterministic for the same seed, level, and rules. Tests should avoid locking
incidental tuning values except where a golden fixture intentionally owns a seed-specific layout.

Required determinism rules:

- No wall-clock time in `server/internal/game`.
- No unseeded randomness.
- No map iteration for generation order unless keys are sorted first.
- New generated wall IDs or ordering must be stable under replay.
- Reconnect snapshots, `/state`, and replay timelines must expose the same wall layout for the same
  reconstructed level.
- Existing boss-floor golden behavior must not drift unless the plan explicitly updates the boss
  floor fixture for a reason unrelated to v40 obstacles.

## 7. Acceptance criteria

1. `make validate-shared` validates the new dungeon obstacle rules, protocol examples, and any new
   golden fixtures.
2. A pinned non-boss generated level contains perimeter walls plus at least one generated interior
   obstacle group.
3. The pinned obstacle fixture proves at least two shape families among line, L/T group, and block.
4. Repeating `GenerateDungeonLevel(seed, level, rules)` produces the same wall layout and generated
   target positions.
5. Generated stairs up/down, teleporter, chest when present, guaranteed loot, and every generated
   monster spawn are reachable from the player spawn or arrival marker.
6. A generated floor that cannot satisfy reachability under test-injected bad tuning fails generation
   clearly instead of returning an unreachable level.
7. Server movement, auto-pathing, monster chase, projectile collision, travel arrival, and loot/drop
   placement treat generated walls as solid.
8. `session_snapshot` and level-change handling expose the authoritative current-level wall layout
   to the client and debug/replay paths.
9. Godot renders generated interior walls from server-provided layout after fresh snapshot,
   reconnect, visual replay, and level change.
10. Protocol bot scenario `28_reachable_dungeon_obstacles.json` descends into a pinned obstacle
    floor, proves interior walls exist, routes around them to a generated target, interacts with or
    picks up something behind the obstacle structure, then verifies `/state`, reconnect, and replay.
11. Client bot coverage proves the real Godot client applies and renders non-perimeter generated
    wall layout, unless the implementation plan records a concrete headless limitation and covers
    the behavior with client unit tests instead.
12. Boss floor generation and `24_boss_floor_gate.json` remain green and keep boss floors unchanged
    by default.
13. `make ci` is green before the slice is finished.

## 8. Open questions

| # | Question | Status |
|---|----------|--------|
| Q-1 | Should generated doors appear inside obstacle walls now? | Answered: no, use open passages in v40. |
| Q-2 | How rich should shapes be in the first slice? | Answered: orthogonal AABB segments, blocks, and L/T connected groups. |
| Q-3 | Should every monster/loot position be reachable, or only critical progression objects? | Answered: every generated spawn/target must remain reachable. |
| Q-4 | Should boss floors get obstacles? | Answered: no, preserve v35 boss floors by default. |

## 9. Testing plan

1. Shared validation:

   ```bash
   make validate-shared
   ```

2. Go generation and sim tests:

   ```bash
   cd server && go test ./internal/game/... -run 'TestDungeon.*Obstacle|TestGenerated.*Reachable|TestBossFloor|TestProjectile|TestPath'
   ```

3. Full Go suite:

   ```bash
   make test-go
   ```

4. Protocol bot proof:

   ```bash
   make bot scenario=reachable_dungeon_obstacles
   make bot
   ```

5. Client unit and smoke:

   ```bash
   make client-unit
   make client-smoke
   make bot-client scenario=14_dungeon_wall_rendering.json
   ```

   If the headless client bot cannot reliably prove wall rendering, the plan must replace this with
   focused client unit coverage plus `make client-smoke`.

6. Final verification:

   ```bash
   make ci
   ```
