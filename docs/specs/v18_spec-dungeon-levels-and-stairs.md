# Spec: `dungeon-levels-and-stairs`

Status: Draft
Branch: `feature/dungeon-levels-and-stairs`
Slice: v18 — multi-level Sim, seeded stair PCG, descend/ascend intents, level HUD
Baseline: v17 `monster-chase-movement` (complete; `docs/PROGRESS.md` updated before v18 implementation)
Related:

- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — D2 multi-level Sim, D3 seeded generation, protocol table
- [`v17_spec-monster-chase-movement.md`](v17_spec-monster-chase-movement.md) — tick order extends with per-level processing
- [`v11_spec-click-to-move-and-auto-path.md`](v11_spec-click-to-move-and-auto-path.md) — navigation bounds scale for 2× dungeon floors
- [`v10_spec-click-action-and-melee-range.md`](v10_spec-click-action-and-melee-range.md) — interactable pattern reused for stairs
- [`v7_spec-gear-before-combat-scenario.md`](v7_spec-gear-before-combat-scenario.md) — world preset + session `world_id` persistence
- [`../PROGRESS.md`](../PROGRESS.md)
- ADR-0001 (authoritative server, shared rules-as-data, golden fixtures, replay determinism)
- ADR-0007 (client-only presentation; level HUD is not on the wire)

Review status: Approved after planning review. Gaps closed in this draft:

- Protocol additions require a **schema version bump to v1**; v0 files remain the legacy contract.
- Stair interactables require a non-blocking `ready` state and optional barrier schema.
- Transition output is deterministic scoped deltas, with a full new-level spawn set in the arrival
  delta so the client never has to infer remote floor state.
- Dungeon perimeter wall presentation is generated client-side from shared generation rules because
  walls are not protocol entities in the current architecture.

## 1. Purpose

Today every session runs in a **single static world**. ADR-0008 requires a **multi-level dungeon**
where the player descends through procedurally generated floors and can return upward via stairs.
This slice implements the **architectural foundation** — not the full ADR (no town, waypoints, or
character-scoped persistence yet).

After this slice:

- `Sim` holds **`levels map[int]*LevelState`**; the player entity tracks **`currentLevel`**.
- A new world preset **`dungeon_levels`** enables multi-level mode; all existing worlds stay
  **single-level** (internal level `0`) so bot scenarios `01`–`11` pass unchanged.
- Each dungeon floor is **32×20** world units (double the ~16×10 vertical-slice play area).
- On first visit to level **N**, the server generates that floor from
  **`session_seed + abs(N)`** (ADR D3): perimeter walls + procedurally placed stairs.
- **Level -1** (entry / “floor 1”): player spawn + **stairs down** only.
- **Level -2+**: **stairs up** + **stairs down**.
- New intents **`descend_intent`** and **`ascend_intent`** transition the player between levels;
  server emits **`level_changed`** `{ from_level, to_level }`.
- Protocol bumps to **v1** and adds **`level`** on `state_delta` plus **`current_level`** on
  `session_snapshot`; v0 schemas remain for historical examples/compatibility.
- Inventory and equipped state remain **session-scoped** (ADR D1 deferred).
- Golden fixture **`shared/golden/dungeon_stairs.json`** pins stair positions for a fixed seed.
- Bot scenario **`12_dungeon_levels.json`**: descend to **-2**, ascend back to **-1**; reconnect +
  replay verify both level states and player floor.
- Godot renders placeholder stair interactables and a **top-right HUD label** showing **level
  number and name** (client-only text placeholder driven from authoritative `current_level` +
  shared level-name catalog).

The proof is **LevelState refactor → seeded stair PCG → transition intents → per-level deltas →
golden fixtures → bot round-trip → client level label**, not full dungeon room generation, town,
or waypoints.

### Level numbering (ADR vs player-facing)

| ADR level | Meaning | Player-facing label (HUD) |
|-----------|---------|---------------------------|
| `0` | Town (deferred) | — |
| `-1` | First dungeon floor | “Level 1 — Entry Hall” |
| `-2` | Second dungeon floor | “Level 2 — Lower Depths” |
| `-N` | Nth dungeon floor down | “Level N — …” from rules catalog |

Bot assertions use ADR integers (`-1`, `-2`). The HUD shows **`abs(level)`** as the floor number.

## 2. Current Problems

### 2.1 Single-world Sim cannot hold multiple floors

`Sim` embeds `entities`, `walls`, `move`, and `autoNav` directly. There is no place to keep
level -1 alive while the player walks on level -2, and no transition primitive.

### 2.2 World size and navigation bounds are fixed to ~16×10

`navigation.v0.json` `grid_bounds` and existing world presets assume the vertical-slice cage.
Double-size dungeon floors need expanded bounds without breaking legacy worlds.

### 2.3 No stair interactables or level-transition intents

Doors use `action_intent` + state change. ADR-0008 specifies **`descend_intent` / `ascend_intent`**
for level transitions (waypoints will reuse the same path later).

### 2.4 Client has no level context

`state_delta` is not scoped by level; the client cannot filter entities or show which floor the
player occupies.

### 2.5 PROGRESS defers multi-level work to ADR-0008

Character-scoped persistence and town are backlog items. v18 proves the Sim/protocol layer so
later slices can add town (level 0), waypoints, and Postgres character state without another
structural refactor.

## 3. Non-goals

- **No character-scoped persistence** (ADR D1) — inventory stays session-scoped.
- **No town (level 0)**, NPCs, vendors, or safe-zone combat rejection (ADR D5 partial → v19).
- **No waypoints / fast travel** (ADR D4).
- **No full dungeon PCG** — no rooms, corridors, monster density curves, or loot placement by
  depth. v18 generates **perimeter walls + stair positions** only; floors are otherwise open.
- **No co-op / per-player delta routing** — solo session only; architecture must not block it.
- **No monster population changes by depth** — optional single static dummy on -2 for visual
  proof is allowed; no new AI behavior.
- **No floor transition animation** (fade, loading screen) — instant teleport on transition.
- **No production stair art** — simple placeholder meshes/panels like v10 doors.
- **No changes to existing world presets** except adding `dungeon_levels`.
- **No Godot client-bot scenario for v18** — protocol bot is sufficient; client bot world
  selection remains unchanged.
- **No protocol backward-compat adapter** — active client/server move together for v18. v0 schema
  files stay in the repo, but runtime validation and new examples use v1 payloads.

## 4. Required Design

### 4.1 Dual-mode Sim: single-level default, multi-level opt-in

**Single-level mode** (all existing `world_id` values):

- `Sim` holds one implicit `LevelState` at key **`0`**.
- `currentLevel == 0`; v1 snapshots send `current_level: 0` and v1 deltas send `level: 0`.
- Behavior identical to pre-v18 for replay, resume, and bot scenarios `01`–`11`.

**Multi-level mode** (`world_id == "dungeon_levels"`):

- `Sim.levels map[int]*LevelState` populated on demand.
- Player starts at **`currentLevel = -1`**; level -1 generated at session create.
- Level **-2** generated on first `descend_intent` that targets it (or eagerly at create — either
  is acceptable if deterministic; **recommend lazy generation on first visit** per ADR D3).
- Visited levels remain in memory for the session lifetime (ADR D2).

**Session-global state** (not duplicated per level):

- `inventory`, `equipped`, `rng`, `tick`, `sessionID`, `seed`, `nextID` (entity IDs global per ADR D2).

**Per-level state** (`LevelState`):

- `entities map[uint64]*entity`
- `walls []wallObstacle`
- `move *activeMove` (player move queue — only active on player's current level)
- `autoNav *autoNavState`
- `nav *NavigationRules` — level-local navigation bounds (dungeon levels use 32×20-derived bounds; single-level worlds point to the same global `Rules.Navigation`)
- `levelNum int` (the map key, stored for clarity)

Refactor `Sim` methods to delegate to `s.level(s.currentLevel)` for player-facing operations;
`Tick` processes monster/projectile phases **per visited level** in stable level-number order
(e.g. sort level keys ascending: -2 before -1).

### 4.2 Shared dungeon generation rules

Add **`shared/rules/dungeon_generation.v0.json`** + schema:

```json
{
  "version": 0,
  "floor_size": { "width": 32, "height": 20 },
  "wall_thickness": 1.0,
  "player_spawn": { "x": 4, "y": 10 },
  "stair_placement": {
    "min_separation": 8.0,
    "margin_from_wall": 2.0,
    "max_attempts": 64
  },
  "level_names": {
    "-1": "Entry Hall",
    "-2": "Lower Depths",
    "-3": "Deep Vault"
  },
  "default_level_name_template": "Depth {n}"
}
```

| Field | Meaning |
|-------|---------|
| `floor_size` | Inner playable area; perimeter walls sit outside this (same convention as v9 walls). |
| `player_spawn` | Default player position on **level -1 only**; deeper levels spawn player at the matching stair tile (up stair on -2 → player appears at up stair cell on -2 when descending from -1; ascending places player at down stair on -1). |
| `stair_placement` | Seeded RNG tries random interior cells until down/up pair satisfies min separation and wall margin. |
| `level_names` | Display names for HUD; key is ADR level string. |
| `default_level_name_template` | Fallback when `-N` not listed; `{n}` = `abs(N)`. |

**Generation algorithm** (deterministic, Go-only for v18):

1. Derive `levelSeed = SeedToUint64(session_seed + "|" + strconv.Itoa(abs(levelNum)))` using the
   existing `SeedToUint64` function in `rng.go`.
2. Instantiate local RNG from `levelSeed` (separate stream from combat RNG — use dedicated PCG RNG
   or a labeled sub-sequence; **must not** consume combat `Sim.rng` rolls).
3. Build four perimeter walls from `floor_size` + `wall_thickness`.
4. **Level -1:** place one **`stairs_down`** interactable at a valid cell.
5. **Level ≤ -2:** place **`stairs_up`** and **`stairs_down`** at valid distinct cells.
6. If placement fails after `max_attempts`, fail session create / level generation with a clear
   error (test with fixed seeds to ensure success).

**Validation (`tools/validate_shared.py`):**

- Schema-valid generation catalog.
- `floor_size` ≥ 16×10 (sanity).
- Every listed `level_names` key is a negative integer string.

### 4.3 Stair interactables

Extend **`shared/rules/interactables.v0.json`**:

```json
{
  "stairs_down": {
    "name": "Stairs Down",
    "initial_state": "ready",
    "transition": "descend"
  },
  "stairs_up": {
    "name": "Stairs Up",
    "initial_state": "ready",
    "transition": "ascend"
  }
}
```

- **`transition`** is rules metadata for tooling/docs; server resolves by `interactable_def_id`.
- Stairs are **not blocking** (unlike closed doors) — walk-through triggers are not required;
  transitions use explicit intents when player is in reach.
- **`action_intent` on stairs is rejected** in v18 — use dedicated intents only (forward-compatible
  with waypoints).
- Update `interactables.v0.schema.json` and `LoadRules` so interactables can be either:
  - `initial_state: "closed"` with required `barrier_when_closed` (doors, existing behavior), or
  - `initial_state: "ready"` with no barrier and required `transition` enum
    (`"ascend" | "descend"`).
- Change `InteractableDef.BarrierWhenClosed` in Go from value type `InteractableBarrier` to pointer
  `*InteractableBarrier`. Update all call sites that dereference it to guard on `!= nil`
  (collision/projectile code already gates on `state == "closed"`, so functional behavior is
  unchanged, but the loader must not require the field for `ready` interactables).
- Runtime entity state schemas must allow `"ready"` for stair entities in snapshots/deltas.

Add **`stairs_reach`** to combat/navigation rules or reuse `combat.unarmed_reach` for stair
interaction distance (document choice in plan; default **`combat.unarmed_reach`**).

### 4.4 Level transition intents and events

**Protocol bump** — create **`shared/protocol/*v1.schema.json`** from the v0 schemas and extend
the v1 envelope/message enum:

```json
"descend_intent": { "type": "object", "additionalProperties": false, "properties": {} },
"ascend_intent":  { "type": "object", "additionalProperties": false, "properties": {} }
```

No payload in v18 (player always uses the stair on their current level). Future: optional
`target_id` for multiple stairs.

**Server handling:**

| Intent | Preconditions | Effect |
|--------|---------------|--------|
| `descend_intent` | Multi-level session; player alive; `stairs_down` exists on current level; player within reach of down stair | Ensure level `currentLevel - 1` exists (generate if needed); move player entity to that level at up-stair position (or spawn point if entering from -1 → -2 use down-stair arrival cell on -2); remove player from previous level's entity map; emit `level_changed` |
| `ascend_intent` | Multi-level session; player alive; `currentLevel < -1`; player within reach of up stair | Move player to `currentLevel + 1` at down-stair position; emit `level_changed` |

Reject reasons: `player_dead`, `not_dungeon_world`, `no_stair_in_range`, `already_at_entry`,
`invalid_level`.

**Event** — add to **`state_delta.v1.schema.json`**:

```json
{
  "event_type": "level_changed",
  "from_level": -1,
  "to_level": -2
}
```

Emit on successful transition; include in replay/resume event streams.

### 4.5 Protocol scoping

**`session_snapshot`** (`session_snapshot.v1.schema.json`):

- Add **`current_level`** (integer, required for every v1 snapshot; `0` for single-level worlds).
- **`entities`** lists entities on **`current_level` only** on attach (same as today for the
  active floor).

**`state_delta`** (`state_delta.v1.schema.json`):

- Add top-level **`level`** (integer, required for every v1 delta) — all entity changes/events in
  the delta belong to that level.
- Client **ignores** deltas where `level != current_level` except **`level_changed`** (always
  processed from player's active stream — see below).

**Transition envelope order (required):**

1. Old-level `state_delta` with `level: from_level`, `level_changed`, and `entity_remove` for the
   player.
2. New-level `state_delta` with `level: to_level` and `entity_spawn` for **every entity currently
   visible on the destination level**, including the moved player and both stair entities when
   present.

Both deltas use the same `server_tick` and deterministic ordering. This avoids a mid-session
snapshot refresh and gives the client a complete active-floor entity set after transition.

**`TickResult` extension:** Add `Level int` to `TickResult` and extend `Tick()` to return
`[]TickResult` (a slice). Normal ticks return a single-element slice. Transition ticks return two
elements in the required order (from-level first, to-level second). The runner iterates the slice
and emits one `state_delta` envelope per element. This keeps the caller API uniform and avoids
a special-case transition path in the runner.

### 4.6 Navigation bounds for dungeon floors

Option A (recommended): **`dungeon_generation.v0.json`** embeds per-floor grid bounds matching
32×20. `PlanPath` and movement use level-local bounds when `world_id == dungeon_levels`.

Legacy worlds continue using **`navigation.v0.json`** global bounds unchanged.

**Implementation mechanism:** `LevelState` holds a `nav *NavigationRules` pointer. At level
creation, dungeon floors compute `NavigationRules` from `dungeon_generation` floor size/cell size;
single-level worlds set `level.nav = &rules.Navigation`. All methods that currently receive
`s.rules.Navigation` (pathfinding, approach goals, monster movement) are refactored to accept or
read `s.level(targetLevel).nav` instead. No new global navigation file needed.

### 4.7 Golden fixture: `shared/golden/dungeon_stairs.json`

Cross-language contract for deterministic stair placement:

```json
{
  "version": 0,
  "seed": "dungeonseed0001",
  "world_id": "dungeon_levels",
  "cases": [
    {
      "name": "level_minus_1_stairs",
      "level": -1,
      "expected_stairs_down": { "x": 24, "y": 6 },
      "expected_stairs_up": null
    },
    {
      "name": "level_minus_2_stairs",
      "level": -2,
      "expected_stairs_down": { "x": 20, "y": 14 },
      "expected_stairs_up": { "x": 8, "y": 4 }
    }
  ]
}
```

Exact coordinates finalized during implementation. Go `game_test` generates levels headless and
asserts stair positions. GDScript `test_golden.gd` validates fixture/rules drift.

Optional second case: **`descend_then_ascend`** — scripted intents return player to `-1` at the
**down-stair position on level -1** (not spawn `{4,10}`). The golden fixture should assert
`player_position` after ascending equals the `expected_stairs_down` coordinate for level -1, not
the initial spawn. This confirms arrival placement logic works correctly in both directions.

### 4.8 World preset: `dungeon_levels`

Add to **`shared/rules/worlds.v0.json`**:

```json
{
  "dungeon_levels": {
    "mode": "multi_level",
    "player": { "position": { "x": 4, "y": 10 } },
    "entities": []
  }
}
```

- **`mode": "multi_level"`** — new optional field on world preset schema; absent → single-level.
  Add `Mode string \`json:"mode,omitempty"\`` to the Go `WorldDef` struct in `rules.go`.
- No static entities; generation fills walls + stairs at runtime.
- `player.position` overridden by generation rules for -1 spawn.

### 4.9 Persistence and replay

- **`world_id`** already persisted on sessions (v7) — `dungeon_levels` reconstructs correctly.
- Record **`descend_intent` / `ascend_intent`** in `session_inputs` with tick + sequence.
- **`current_level`** must be reconstructible from replay (derived from transition intents, or
  persisted in snapshot metadata on resume — **recommend derive from inputs** for determinism).
- Reconnect resume: replay inputs restore all visited `LevelState` instances and player floor.

### 4.10 Client: stairs presentation

- Render **`stairs_down`** / **`stairs_up`** as simple placeholder meshes (distinct colors, e.g.
  cyan pad = down, orange pad = up) from interactable entities in scoped deltas.
- Render dungeon perimeter walls from `shared/rules/dungeon_generation.v0.json` for
  `dungeon_levels`. This is deterministic and contains no random layout logic; stair positions
  still come from authoritative entities.
- Left-click on stair within reach sends **`descend_intent`** or **`ascend_intent`** (same reach
  rules as server); no `action_intent`.
- **`entity_map` / `entities` dict** keyed as today but active-level only. On old-level
  `level_changed`, update `current_level`, clear active nodes, then apply the next new-level delta's
  complete spawn set. Ignore later non-active deltas.

### 4.11 Client: level HUD (top-right placeholder)

Add a **read-only HUD label** in the **top-right corner** showing **floor number and name**.

**Layout:**

- Parent: existing UI `CanvasLayer` in `main.gd` (same layer as debug label / inventory).
- Control: `Label` (or thin wrapper `LevelHud.gd` if cleaner).
- Anchor: **`PRESET_TOP_RIGHT`** with margin **`12px`** top and right (mirror debug label on
  top-left at `(12, 12)`).
- Alignment: **right-aligned** text.

**Content format:**

```text
Level {n} — {name}
```

Examples:

- `Level 1 — Entry Hall` when `current_level == -1`
- `Level 2 — Lower Depths` when `current_level == -2`
- Fallback: `Level 3 — Depth 3` using `default_level_name_template` when `-3` not in catalog.

**Data sources:**

- **`current_level`** from `session_snapshot` on attach and from **`level_changed`** events /
  subsequent deltas (authoritative).
- **Display name** from **`shared/rules/dungeon_generation.v0.json`** loaded client-side (same
  JSON the server uses — not a separate wire field in v18).

**Visibility:**

- **Hidden** (or empty string) when `current_level == 0` / single-level worlds — legacy sessions
  show no level HUD.
- Visible whenever `current_level < 0`.

**Style (placeholder):**

- Font size **14–16**, color **`#c9a227`** (match inventory caption gold from v13) or **`#f4ead8`**
  cream for readability on dark background.
- No background panel required in v18; optional subtle shadow if needed for contrast.

**Non-goals for HUD:** no minimap, no depth ladder, no animation on transition, no server-driven
display name field (rules JSON is sufficient for v18).

### 4.12 Tick order (multi-level)

For each tick:

```text
1. applyInput(all intents — transition intents may change currentLevel mid-tick)
2. For each visited level L in sorted level order:
     a. applyMovement on L if L == currentLevel
     b. advanceMonsterMovement on L
     c. advanceProjectiles on L
3. tick++
```

Player movement and auto-nav only run on the player's **current** level. Idle levels still tick
monsters/projectiles if any exist (v18 likely has none on inactive floors).

## 5. Bot scenario

### 5.1 `12_dungeon_levels.json`

```json
{
  "id": "dungeon_levels",
  "world_id": "dungeon_levels",
  "title": "Dungeon levels",
  "description": "Descend to level -2 via stairs, then ascend back to level -1.",
  "steps": [
    {
      "action": "use_stair",
      "interactable_def_id": "stairs_down",
      "intent": "descend_intent",
      "event_type": "level_changed"
    },
    {
      "action": "assert_current_level",
      "level": -2
    },
    {
      "action": "use_stair",
      "interactable_def_id": "stairs_up",
      "intent": "ascend_intent",
      "event_type": "level_changed"
    },
    {
      "action": "assert_current_level",
      "level": -1
    }
  ],
  "assertions": [
    { "type": "current_level", "level": -1 },
    { "type": "event_seen", "event_type": "level_changed" },
    { "type": "visited_levels_contain", "levels": [-1, -2] }
  ]
}
```

### 5.2 Bot harness additions (`tools/bot/run.py`)

**RuntimeState** tracks:

- `current_level: int` (default `0`)
- `visited_levels: set[int]`
- Update from snapshot `current_level` and `level_changed` events.

**Actions:**

| Action | Behavior |
|--------|----------|
| `use_stair` | Find interactable by `interactable_def_id` on current level; `walk_toward` until in reach; send `descend_intent` or `ascend_intent` from step; wait for `event_type`. |
| `assert_current_level` | Fail fast if `current_level != level`. |

**Assertions:**

| Type | Check |
|------|-------|
| `current_level` | Final floor matches. |
| `visited_levels_contain` | All listed levels were generated/visited during session. |
| `event_seen` | `level_changed` observed (count ≥ 2 for round trip). |

### 5.3 Regression guard

Scenarios **`01`–`11`** and client bot scenarios pass **without modification**. Single-level worlds
never emit `level_changed`; HUD hidden.

## 6. Files to create or modify

```text
shared/rules/dungeon_generation.v0.json           - floor size, stair placement, level names
shared/rules/dungeon_generation.v0.schema.json
shared/rules/interactables.v0.json                - stairs_down, stairs_up
shared/rules/interactables.v0.schema.json         - optional transition metadata
shared/rules/worlds.v0.json                       - dungeon_levels preset
shared/rules/worlds.v0.schema.json                - mode: multi_level
shared/protocol/envelope.v1.schema.json            - v1 message enum
shared/protocol/messages.v1.schema.json            - descend_intent, ascend_intent
shared/protocol/session_snapshot.v1.schema.json    - current_level
shared/protocol/state_delta.v1.schema.json         - level, ready state, level_changed event
shared/protocol/examples/*                         - v1 examples for dungeon transition
shared/golden/dungeon_stairs.json
shared/golden/dungeon_stairs.v0.schema.json
server/internal/game/level.go                     - LevelState type (new)
server/internal/game/sim.go                       - levels map, refactor, transitions, PCG hook
server/internal/game/dungeon_gen.go              - seeded stair + wall generation (new)
server/internal/game/rules.go                     - load dungeon_generation + world mode; WorldDef.Mode; *InteractableBarrier
server/internal/game/game_test.go                 - golden + transition unit tests
server/internal/inputdecode/inputdecode.go        - TypeDescend / TypeAscend constants; IsClientIntent; Decode cases
server/internal/realtime/hub.go                   - snapshot/delta level scoping
server/internal/realtime/runner.go                - iterate []TickResult slice; emit per-element deltas
server/internal/replay/replay.go                  - decode new intents
tools/bot/scenarios/12_dungeon_levels.json
tools/bot/run.py                                  - stair actions, level tracking
tools/validate_shared.py                          - generation + golden validation
client/scripts/main.gd                            - level-scoped entities, stair clicks, dungeon walls, LevelHud
client/scripts/level_hud.gd                       - top-right label (new, optional wrapper)
client/tests/test_golden.gd                       - dungeon_stairs fixture checks
docs/PROGRESS.md                                  - v18 lifecycle when complete
docs/plans/v18_2026-06-06-dungeon-levels-and-stairs.md  - implementation plan (separate artifact)
```

## 7. Acceptance criteria

1. **`make validate-shared`** passes; generation catalog, stairs, and protocol schemas valid.
2. **`make test-go`** — level transition unit tests + `dungeon_stairs.json` golden green.
3. **`make client-unit`** — `test_golden.gd` dungeon stairs fixture green.
4. **`make bot`** — scenarios `01`–`12` pass including reconnect + replay for `dungeon_levels`.
5. **Single-level regression** — `vertical_slice` behavior unchanged; no level HUD shown.
6. **Determinism** — same seed + inputs → identical stair positions, level map, and event order on
   replay.
7. **Double size** — dungeon floor playable area is 32×20; navigation respects expanded bounds.
8. **Stair PCG** — level -1 has down only; level -2 has up + down; positions differ per seed.
9. **Level HUD** — in `dungeon_levels` session, top-right label shows `Level N — Name` matching
   rules catalog; updates on descend/ascend; hidden on legacy worlds.
10. **Protocol v1** — v1 schemas validate new messages and deltas; v0 legacy schemas remain present
    and existing examples are either migrated or clearly versioned.

## 8. Verification commands

```bash
make validate-shared
make test-go
make client-unit
make db-up && make server          # terminal 1
make bot                           # terminal 2 — includes 12_dungeon_levels.json
make ci
make play                          # manual — walk to stairs, confirm HUD updates
make bot-visual scenario=12_dungeon_levels.json  # optional visual replay
```

## 9. Follow-up slices (not v18)

| Topic | ADR ref | Notes |
|-------|---------|-------|
| Town level 0 + rift entrance | D5 | Static `worlds.v0.json` hub |
| Character-scoped persistence | D1 | `characters` table |
| Waypoints | D4 | Reuse transition machinery |
| Full dungeon PCG | D3 | Rooms, corridors, monster/loot curves |
| Floor transition polish | — | Fade, audio, stair art |
