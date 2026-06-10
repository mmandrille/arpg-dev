# Spec: `monster-chase-movement`

Status: Draft
Branch: `feature/monster-chase-movement`
Slice: v17 — server-authoritative monster chase movement (aggro, pathfinding, leash)
Baseline: v16 `use-consumable` (complete; `PROGRESS.md` updated before v17 implementation)
Related:

- [`v16_spec-use-consumable.md`](v16_spec-use-consumable.md)
- [`v11_spec-click-to-move-and-auto-path.md`](v11_spec-click-to-move-and-auto-path.md) — grid A\* planner reused for monsters
- [`v9_spec-solid-collision-and-obstacles.md`](v9_spec-solid-collision-and-obstacles.md) — wall/monster/player collision reused
- [`v4_spec-take-a-hit.md`](v4_spec-take-a-hit.md) — retaliation remains hit-triggered only
- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../godot-plugins-and-shortcuts.md`](../godot-plugins-and-shortcuts.md) — LimboAI deferred
- ADR-0001 (authoritative server, shared rules-as-data, golden fixtures, replay determinism)
- ADR-0007 (client-only locomotion presentation from authoritative position deltas)

## 1. Purpose

Monsters are **static obstacles** today: they block movement and retaliate when hit, but never
change position. This slice adds **server-owned chase movement** so live monsters can **aggro
the player**, **path around walls and bodies**, and **pressure positioning** — without client
AI, NavMesh, or behavior trees.

After this slice:

- Monster definitions may declare **`behavior: "chase"`** with **`aggro_radius`**, optional
  **`leash_radius`**, and optional **`move_speed`** in `shared/rules/monsters.v0.json`.
- All existing monsters default to **`behavior: "static"`** — every current bot scenario stays
  green without edits.
- Each sim tick, after player movement and before projectiles, the server runs deterministic
  **monster AI movement** for live chase monsters:
  - Player within **`aggro_radius`** and alive → select a reachable adjacent chase goal around
    the player, then plan one step using v11 **`PlanPath`** and shared **`navigation.v0.json`**.
  - Player beyond **`leash_radius`** from the monster's spawn point → switch to **return** mode
    and path one step back toward spawn.
  - At spawn (within `navigation.stop_distance`) → **idle** until player re-enters aggro.
- Monster movement uses the same **circle-vs-wall / circle-vs-entity** collision model as the
  player (v9 axis slide), with symmetric **player blocking** and **self excluded** from own
  pathfinding rasterization.
- Server emits optional presentation/debug events **`monster_aggro`** and **`monster_leashed`**
  when a monster enters chase or return mode (ADR-0007 hook for future attack anim; v17 uses
  them for bot assertions and client walk toggling).
- Golden fixture **`shared/golden/monster_chase.json`** pins chase positions on a pinned seed
  for Go and GDScript.
- Three new worlds and bot scenarios prove open-field chase, maze routing, and leash reset.
- Godot drives monster **`walk` / `idle`** from authoritative position deltas (add minimal
  monster `walk` clip via existing animation pipeline).
- Replay, reconnect resume, `/state`, and visual replay reconstruct the same monster paths
  deterministically.

The proof is **shared behavior rules → server chase sim → pathfinder reuse → collision → golden
fixtures → bot scenarios → client locomotion**, not proactive monster attacks, ranged monster
AI, or LimboAI adoption.

## 2. Current Problems

### 2.1 Monsters never move

`Sim.Tick` applies player movement and projectiles but has no monster locomotion phase. Chase
pressure, kiting, and “enemy closes distance” do not exist.

### 2.2 Pathfinding is player-only

v11 `PlanPath` and `buildBlockedFn()` rasterize obstacles for **player** auto-nav. Monsters
cannot reuse the planner without a monster-specific blocked function (exclude self, include
player as solid).

### 2.3 PROGRESS explicitly deferred monster movement

v9/v11 non-goals listed “no monster AI or monster pathfinding.” Bot scenarios assume stationary
targets (`walk_to_monster`, `action_once_until_event`). Any movement feature must default to
**static** for backward compatibility.

### 2.4 Monster presentation has no walk locomotion

`monster_anims.tres` provides `idle`, `hit`, `death` only. Position updates snap instantly in
`main.gd` with no locomotion signal for monsters.

## 3. Non-goals

- **No proactive monster attacks** (melee swing, contact damage, ranged fire). Retaliation
  remains **on successful player hit** only (v4). Monster melee attack is a follow-up slice.
- **No monster attack animation** when retaliating (still v4 non-goal).
- **No behavior trees, LimboAI, or client-side AI authority.**
- **No group aggro, assist calls, flanking, or separation steering.**
- **No monster click-to-move or player commands over monsters.**
- **No new client intents or protocol envelope bump** beyond optional new `event_type` strings
  in existing `state_delta.events`.
- **No NavMesh / Godot NavigationServer as authority.**
- **No path preview UI.**
- **No aggro on loot, doors, or other players** (solo session only).
- **No faster-than-player speed tuning beyond shared `move_speed` scalar** — no sprint bursts.
- **No polygon collision changes** — keep v9 circle/AABB model.
- **No changes to existing world presets** except adding new lab worlds.

## 4. Required Design

### 4.1 Shared monster behavior rules

Extend **`shared/rules/monsters.v0.json`** + schema:

```json
{
  "training_dummy_chase": {
    "name": "Training Dummy",
    "max_hp": 3,
    "loot_table": "basic_drop",
    "retaliation_damage": { "min": 1, "max": 1 },
    "behavior": "chase",
    "aggro_radius": 8.0,
    "leash_radius": 12.0,
    "move_speed": 1.0
  }
}
```

| Field | Required when | Meaning |
|-------|---------------|---------|
| `behavior` | optional | `"static"` (default) or `"chase"`. |
| `aggro_radius` | `behavior == "chase"` | Enter chase when `hypot(monster, player) ≤ aggro_radius`. |
| `leash_radius` | optional for chase | If set, leave chase and return to spawn when `hypot(player, spawn) > leash_radius`. Omit = no leash (chase until dead). |
| `move_speed` | optional | World units per tick (default **`1.0`**, same scale as player `moveSpeed`). v17 validates it equals `navigation.cell_size`; fractional chase speeds are deferred. |

**Validation (`tools/validate_shared.py`):**

- `behavior` must be `static` or `chase`.
- `chase` requires `aggro_radius > 0`.
- `leash_radius`, when present, must be `≥ aggro_radius`.
- `move_speed`, when present, must equal `navigation.cell_size` in v17 (one grid edge per tick —
  keeps planner steps aligned with movement deltas). Fractional speed support is a follow-up.

Existing defs (`training_dummy`, `training_dummy_reward`, etc.) omit `behavior` → treated as
**static**.

### 4.2 Internal sim state (not on wire)

Extend internal `entity` for monsters:

```text
spawnPos   Vec2    // world preset position at spawn; used for leash return
aiMode     string  // "idle" | "chase" | "return"
```

- Set `spawnPos` when spawning from world preset.
- `aiMode` is **recomputed each tick** from rules + distances (no separate persistence table).
  Replay/resume correctness comes from deterministic tick order, not stored AI snapshots.

### 4.3 Tick order (determinism contract)

Update `Sim.Tick`:

```text
1. applyInput(all client intents)
2. applyMovement(player manual + auto-nav)     // unchanged
3. advanceMonsterMovement(res)               // NEW
4. advanceProjectiles(res)
5. tick++
```

Monster movement never runs before player movement on the same tick. Projectiles see post-chase
positions.

Within `advanceMonsterMovement`, iterate live monsters in **`sortedEntityIDs`** order (stable
replay).

### 4.4 Chase AI logic (per monster, per tick)

For each live monster where `rules.Monsters[def].Behavior == "chase"`:

1. **Skip** if player dead.
2. Compute `distPlayer = hypot(monster.pos, player.pos)`.
3. Compute `distPlayerFromSpawn = hypot(player.pos, monster.spawnPos)`.
4. **Leash check:** if `leash_radius` set and `distPlayerFromSpawn > leash_radius`:
   - If previous `aiMode != "return"`, emit **`monster_leashed`** `{ entity_id, monster_def_id }`.
   - Set `aiMode = "return"`, goal = `spawnPos`.
5. **Aggro check:** else if `distPlayer ≤ aggro_radius`:
   - If previous `aiMode != "chase"`, emit **`monster_aggro`** `{ entity_id, monster_def_id }`.
   - Set `aiMode = "chase"`, goal = `player.pos`.
6. **Idle:** else if at spawn (`hypot(monster.pos, spawnPos) ≤ stop_distance`), set
   `aiMode = "idle"`, no movement.
7. **Return continuation:** else if `aiMode == "return"`, goal = `spawnPos`.
8. **Otherwise:** no movement this tick.

**Movement step** when goal is set:

- If already at goal within `stop_distance`, clear movement (idle at spawn or adjacent to player
  per collision — see below).
- `steps, ok := PlanPath(nav, monster.pos, goal, buildMonsterBlockedFn(monster.id))`
- If `!ok` or `len(steps) == 0`, no movement.
- Take **first step only** (one step per tick, same as player auto-nav consumption).
- Apply delta `step * move_speed` through **`resolveMonsterMovement(monster, delta)`** (v9 slide).
- On position change, append **`entity_update`** for the monster.

**Stop distance when chasing player:** use `navigation.stop_distance` for spawn arrival; when
chasing the player, stop planning when `distPlayer ≤ playerRadius + monsterRadius` (cannot overlap
player circle — same as v9 solid bodies). The planner **must not** use the player cell as the
goal because `PlanPath` rejects blocked goal cells. Instead add
`findMonsterChaseGoal(monster, player)`:

- Search ring cells around the player's cell in stable `ringCells` order.
- Skip blocked cells using `buildMonsterBlockedFn(monster.id)`, which includes the player.
- Prefer cells where `distance(goal, player.pos) <= playerRadius + monsterRadius + navigation.cell_size`.
- Run `PlanPath` from monster position to candidate goal and choose the first reachable candidate.
- Consume only the first returned step this tick.

### 4.5 Collision: `resolveMonsterMovement` + `buildMonsterBlockedFn`

**`buildMonsterBlockedFn(excludeMonsterID uint64)`:**

- Rasterize blocked cells using circle at cell center, same as player `buildBlockedFn`, but:
  - **Exclude** `excludeMonsterID` from live monster blocking.
  - **Include** player circle as blocking (monsters path around player).
  - Include walls, closed interactables, other live monsters.

Because the player is solid in this blocked function, monster chase must use the adjacent-goal
selection above. Return-to-spawn may use `spawnPos` directly unless another solid entity occupies
the spawn cell; in that case, no movement occurs until a reachable return goal exists.

**`resolveMonsterMovement(monster, delta)`:**

- Same axis-slide algorithm as `resolveMovement` for players.
- Block against: walls, closed doors, **player**, other live monsters (not self).
- Dead monsters non-solid (v9).

### 4.6 Protocol events

Add to **`shared/protocol/state_delta.v0.schema.json`** event validation docs by requiring
`entity_id` for the new stringly-typed event names:

| Event | Payload | When |
|-------|---------|------|
| `monster_aggro` | `entity_id`, `correlation_id` optional | First tick entering chase from idle/return |
| `monster_leashed` | `entity_id` | First tick entering return from chase |

No new intents. No snapshot schema changes (monster position already in `entity_update`).

### 4.7 Golden fixture: `shared/golden/monster_chase.json`

Cross-language contract for deterministic chase on pinned layout:

```json
{
  "version": 0,
  "seed": "cafebabecafebabe",
  "world_id": "chase_maze",
  "navigation": { "...": "from navigation.v0.json" },
  "cases": [
    {
      "name": "chase_maze_reaches_player_adjacent",
      "player_position": { "x": 10, "y": 5 },
      "monster_def_id": "training_dummy_chase",
      "idle_player_ticks": 30,
      "expected_monster_position": { "x": 9.0, "y": 5.0 },
      "expected_events": ["monster_aggro"]
    },
    {
      "name": "leash_lab_returns_to_spawn",
      "world_id": "leash_lab",
      "player_kite_steps": [{ "x": 1, "y": 0, "ticks": 20 }],
      "expected_monster_final_near_spawn": true,
      "expected_events": ["monster_aggro", "monster_leashed"]
    }
  ]
}
```

Exact coordinates finalized when lab worlds are tuned. Go `game_test` runs the sim headless and
asserts positions/events. GDScript `test_golden.gd` validates fixture/rules drift at minimum;
optional full planner parity if client mirrors chase for debug.

### 4.8 World presets

Add to **`shared/rules/worlds.v0.json`**:

#### `chase_lab`

Open arena (reuse vertical_slice cage dimensions). Player at `(2, 5)`, chase monster at
`(12, 5)`, `training_dummy_chase`. Player starts inside aggro radius — monster immediately
chases.

#### `chase_maze`

Reuse **`path_maze`** wall layout. Player at far side `(10, 5)`, monster at `(2, 5)` behind
walls. Player sends **no movement**; monster must route through the maze gap over multiple
ticks.

#### `leash_lab`

Large floor, player at `(2, 5)`, monster at `(4, 5)`, `training_dummy_chase` with
`aggro_radius: 6`, `leash_radius: 8`. Player walks east beyond leash; monster returns to spawn.

### 4.9 Client presentation

**Animation (ADR-0007):**

- Add minimal monster **`walk`** clip to `client/animations/monster_anims.tres` via
  `client/tools/build_animations.gd` (same pipeline as v3 player clips — subtle bob or slide).
- In `main.gd` entity update path for monsters: compare new server position to previous;
  call `controller.set_locomotion(position_changed && hp > 0)`.
- Terminal death and one-shot hit retain priority over walk (existing `AnimationController`).

**Optional polish (not acceptance gate):** brief position tween between ticks for smoother motion.

**No server animation state on wire.**

### 4.10 Plugin adoption

Consult [`docs/godot-plugins-and-shortcuts.md`](../godot-plugins-and-shortcuts.md).

Decision: **reject LimboAI for v17**.

- Chase is fully server-owned; client only renders `entity_update` deltas.
- LimboAI remains P3 backlog for ambient **client** BT if server AI grows complex later.

## 5. Bot scenarios

All scenarios use standard CI flow: auth → create session with `world_id` → steps → assertions →
reconnect `/state` → replay verification (existing `make bot` harness).

### 5.1 `09_chase_lab.json`

```json
{
  "id": "chase_lab",
  "world_id": "chase_lab",
  "title": "Chase lab",
  "description": "A chase monster closes distance while the player stands still.",
  "steps": [
    { "action": "wait_ticks", "ticks": 25 }
  ],
  "assertions": [
    { "type": "monster_moved", "monster_def_id": "training_dummy_chase", "min_distance": 0.5 },
    {
      "type": "monster_within_player_distance",
      "monster_def_id": "training_dummy_chase",
      "max_distance": 1.5
    },
    { "type": "event_seen", "event_type": "monster_aggro" }
  ]
}
```

### 5.2 `10_chase_maze.json`

```json
{
  "id": "chase_maze",
  "world_id": "chase_maze",
  "title": "Chase maze",
  "description": "Monster pathfinds through a wall maze toward a stationary player.",
  "steps": [
    { "action": "wait_ticks", "ticks": 40 }
  ],
  "assertions": [
    { "type": "monster_moved", "monster_def_id": "training_dummy_chase", "min_distance": 2.0 },
    {
      "type": "monster_within_player_distance",
      "monster_def_id": "training_dummy_chase",
      "max_distance": 1.5
    },
    { "type": "event_seen", "event_type": "monster_aggro" }
  ]
}
```

### 5.3 `11_leash_lab.json`

```json
{
  "id": "leash_lab",
  "world_id": "leash_lab",
  "title": "Leash lab",
  "description": "Player kites beyond leash; monster returns toward spawn.",
  "steps": [
    { "action": "wait_ticks", "ticks": 5 },
    {
      "action": "move_until_player_position",
      "x": 14,
      "y": 5,
      "tolerance": 0.5
    },
    { "action": "wait_ticks", "ticks": 15 }
  ],
  "assertions": [
    { "type": "event_seen", "event_type": "monster_leashed" },
    {
      "type": "monster_near_spawn",
      "monster_def_id": "training_dummy_chase",
      "max_distance_from_spawn": 1.0
    }
  ]
}
```

### 5.4 Regression guard: existing scenarios unchanged

`01`–`08` and client bot scenarios must pass **without modification**. Static monsters never
emit `monster_aggro`. CI runs full scenario catalog.

### 5.5 New bot harness support (`tools/bot/run.py`)

**Actions:**

| Action | Payload | Behavior |
|--------|---------|----------|
| `wait_ticks` | `ticks` (int) | Pump WebSocket deltas for N server ticks without sending intents. |

**Runtime assertions** (evaluated after steps, before `/state` reconnect):

| Type | Fields | Check |
|------|--------|-------|
| `monster_moved` | `monster_def_id`, `min_distance` | Final position ≥ `min_distance` from initial spawn snapshot. |
| `monster_within_player_distance` | `monster_def_id`, `max_distance` | `hypot(monster, player) ≤ max_distance`. |
| `monster_near_spawn` | `monster_def_id`, `max_distance_from_spawn` | `hypot(monster, spawn) ≤ threshold`. |
| `event_seen` | `event_type` | Event appeared in `state.seen_events` during scenario. |

Track **`initial_monster_positions`** in `RuntimeState` on first entity snapshot per
`monster_def_id`.

### 5.6 Client bot scenario (optional but recommended)

**`tools/bot/scenarios/client/07_monster_chase_idle.json`**

- Create session on `chase_lab` (requires server world or bot-client uses default world — if
  client bot cannot select world, document manual server preset or extend client bot with
  `world_id` env var matching v14 patterns).
- Steps: `wait_ticks` equivalent (frame waits), no player input.
- Assert via `get_bot_state()`: monster entity position moved closer to player vs initial
  snapshot.

If client bot cannot set `world_id` in v17, defer client scenario and rely on protocol bot;
note in plan.

## 6. Files to create or modify

```text
shared/rules/monsters.v0.json                 - training_dummy_chase + behavior fields
shared/rules/monsters.v0.schema.json          - behavior, aggro_radius, leash_radius, move_speed
shared/rules/worlds.v0.json                   - chase_lab, chase_maze, leash_lab
shared/protocol/state_delta.v0.schema.json    - document monster_aggro, monster_leashed events
shared/golden/monster_chase.json              - pinned chase/leash fixtures
shared/golden/monster_chase.v0.schema.json    - golden schema
server/internal/game/rules.go                 - parse/validate behavior fields
server/internal/game/sim.go                   - advanceMonsterMovement, resolveMonsterMovement,
                                                buildMonsterBlockedFn, spawnPos on entities
server/internal/game/game_test.go             - golden + unit tests (collision, leash, static default)
server/internal/game/pathfind.go              - no change expected (reuse PlanPath)
tools/bot/scenarios/09_chase_lab.json
tools/bot/scenarios/10_chase_maze.json
tools/bot/scenarios/11_leash_lab.json
tools/bot/run.py                              - wait_ticks, new assertions, initial position tracking
tools/validate_shared.py                      - monster behavior validation + golden drift
client/scripts/main.gd                        - monster locomotion from position delta
client/animations/monster_anims.tres          - add walk clip
client/tools/build_animations.gd              - generate monster walk if needed
client/tests/test_golden.gd                  - monster_chase golden checks
PROGRESS.md                              - v17 lifecycle when complete
docs/plans/v17_2026-06-06-monster-chase-movement.md  - implementation plan (separate artifact)
```

## 7. Acceptance criteria

1. **`make validate-shared`** passes; chase defs and worlds schema-valid.
2. **`make test-go`** — monster chase unit tests + `monster_chase.json` golden green.
3. **`make client-unit`** — `test_golden.gd` monster chase fixture green.
4. **`make bot`** — scenarios `01`–`11` pass including reconnect + replay for new labs.
5. **Static default** — all pre-v17 monsters behave as today (no movement).
6. **Determinism** — same seed + inputs → identical monster positions and event order on replay.
7. **Visual** — `make bot-visual scenario=10_chase_maze.json` shows monster routing through maze
   (manual eyeball optional; not CI gate).
8. **Client locomotion** — moving chase monster plays `walk` while translating; idle at spawn
   plays `idle`.

## 8. Verification commands

```bash
make validate-shared
cd server && go test ./internal/game/... -run MonsterChase
make client-unit
make db-up && make server    # terminal 1
make bot                     # terminal 2 — includes 09–11
make ci
make bot-visual scenario=10_chase_maze.json   # optional
```

## 9. Open questions (resolve during implementation)

| # | Question | Proposed default |
|---|----------|------------------|
| 1 | Replan every tick vs cache path until blocked? | Resolved: **replan every tick** (simplest determinism; ≤1 step consumed). |
| 2 | Should chase monsters block each other in tight corridors? | Resolved: **yes** — same solid-body rules as player vs monster. |
| 3 | Emit `monster_aggro` every chase tick or edge only? | Resolved: **edge only** (idle/return → chase). |
| 4 | `move_speed < 1.0` partial steps? | Resolved: **defer** — v17 requires `move_speed == navigation.cell_size`. |
| 5 | Client bot `world_id` selection? | Resolved: existing v14 client bot already supports scenario `world_id`; add a v17 client scenario only if low cost after protocol bot is green. |

## 10. Follow-up slices (not v17)

| Slice idea | Builds on |
|------------|-----------|
| **Monster melee attack** — proactive swing when in range, `player_damaged` without player attacking first | v17 chase + v4 events |
| **Monster ranged AI** | v12 projectiles |
| **Jump / gap traversal** | [`docs/ideas/keyboard-jump-and-gaps.md`](../ideas/keyboard-jump-and-gaps.md) |
| **LimboAI client presentation layer** | Only if server BT becomes too heavy |

## 11. Architecture diagram

```text
                    shared/rules/monsters.v0.json
                    (behavior, aggro_radius, leash_radius)
                                  │
                                  ▼
┌──────────────┐   intents    ┌─────────────────────────────────────┐
│ Godot client │─────────────►│ Sim.Tick                             │
│ (no monster  │              │  1. applyInput                       │
│  AI)         │◄─────────────│  2. applyMovement (player)           │
└──────────────┘ state_delta  │  3. advanceMonsterMovement ◄── NEW   │
       │                      │       PlanPath (v11)                   │
       │ position delta       │       resolveMonsterMovement (v9)      │
       ▼                      │       events: monster_aggro/leashed  │
 monster walk/idle            │  4. advanceProjectiles               │
 (ADR-0007)                    └─────────────────────────────────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    ▼             ▼             ▼
            monster_chase.json   bot 09-11    replay/resume
            (golden)             scenarios
```
