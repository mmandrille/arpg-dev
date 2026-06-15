# v18 Plan — Dungeon levels and stairs

Status: Implemented; DB-dependent bot/CI verification blocked locally by missing Docker/Colima socket (2026-06-06)

## 1. Goal

Add the first multi-level dungeon foundation: `Sim.levels`, deterministic generated stair floors,
explicit ascend/descend intents, scoped v1 protocol deltas, protocol bot coverage, and a small
Godot level HUD.

## 2. Baseline and Approval Notes

Resolved baseline facts:

- v17 `monster-chase-movement` is complete and `PROGRESS.md` lists it as the latest shipped
  slice.
- ADR-0008 is accepted and requires `levels map[int]*LevelState`, seeded on-demand generation, and
  protocol level scoping.
- v18 intentionally defers ADR-0008 character persistence, town, waypoints, co-op routing, and full
  dungeon PCG.
- The spec was reviewed and approved after closing gaps around protocol v1, stair `ready` state,
  transition delta ordering, and dungeon wall presentation.

## 3. Design Decisions Closed During Review

- **Protocol version:** create v1 protocol schemas for new runtime messages/payloads. Keep v0 files
  as legacy contracts.
- **Single-level mode:** legacy worlds still run as one `LevelState` at `0`; v1 snapshots/deltas
  include `current_level: 0` / `level: 0`.
- **Transition output:** successful stairs emit two deterministic same-tick deltas: old-level
  `level_changed` + player remove, then new-level complete entity spawn set.
- **Stair state:** stairs are interactables with `initial_state: "ready"` and no collision barrier.
- **Reach:** use `combat.unarmed_reach` for stair reach in v18; no new reach rule unless tests show
  the shared melee fixture needs a named interactable-specific value.
- **Generation RNG:** dungeon stair placement uses a labeled per-level RNG derived from session seed
  and `abs(level)`, never `Sim.rng`.
- **Client wall rendering:** generated perimeter walls are rendered client-side from
  `dungeon_generation.v0.json`; only stair positions come over the authoritative entity stream.

## 4. File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `shared/rules/dungeon_generation.v0.json` | Floor size, spawn, stair placement, level names |
| Add | `shared/rules/dungeon_generation.v0.schema.json` | Generation catalog schema |
| Modify | `shared/rules/interactables.v0.json` | Add `stairs_down` and `stairs_up` |
| Modify | `shared/rules/interactables.v0.schema.json` | Allow `ready` transition interactables |
| Modify | `shared/rules/worlds.v0.json` | Add `dungeon_levels` |
| Modify | `shared/rules/worlds.v0.schema.json` | Add optional `mode` |
| Add | `shared/protocol/envelope.v1.schema.json` | v1 envelope enum |
| Add | `shared/protocol/messages.v1.schema.json` | `descend_intent`, `ascend_intent` |
| Add | `shared/protocol/session_snapshot.v1.schema.json` | `current_level` |
| Add | `shared/protocol/state_delta.v1.schema.json` | `level`, `ready`, `level_changed` |
| Add | `shared/golden/dungeon_stairs.v0.schema.json` | Stair golden schema |
| Add | `shared/golden/dungeon_stairs.json` | Pinned generated stair fixture |
| Modify | `tools/validate_shared.py` | Rules, protocol, and golden drift checks |
| Add | `server/internal/game/level.go` | `LevelState` |
| Add | `server/internal/game/dungeon_gen.go` | Deterministic floor generation |
| Modify | `server/internal/game/rules.go` | Load generation rules, world mode, stair definitions |
| Modify | `server/internal/game/sim.go` | Level refactor, transition intents, scoped results |
| Modify | `server/internal/game/game_test.go` | Golden, transition, regression tests |
| Modify | `server/internal/inputdecode/inputdecode.go` | TypeDescend/TypeAscend; IsClientIntent; Decode cases; game.Input fields |
| Modify | `server/internal/realtime/runner.go` | Iterate []TickResult; emit v1 snapshots/deltas; transition ordering |
| Modify | `server/internal/replay/replay.go` | Reconstruct new intents and levels |
| Add | `tools/bot/scenarios/12_dungeon_levels.json` | Round-trip stair scenario |
| Modify | `tools/bot/run.py` | Level tracking, stair action, assertions |
| Modify | `client/scripts/main.gd` | Active-level entity scope, stair clicks, dungeon walls, HUD |
| Optional | `client/scripts/level_hud.gd` | Small HUD wrapper if cleaner than inline setup |
| Modify | `client/tests/test_golden.gd` | Dungeon stair fixture/rules checks |
| Modify | `PROGRESS.md` | v18 lifecycle after implementation ships |

## 5. Implementation Tasks

### Task 1 — Shared Rules and Protocol Schemas

- [ ] Add `dungeon_generation.v0.json` and schema with 32x20 floor, spawn, stair placement, and
  level-name fields.
- [ ] Add `dungeon_levels` world with `mode: "multi_level"`.
- [ ] Extend interactables schema for closed blockers and ready transition interactables.
- [ ] Add `stairs_down` / `stairs_up` definitions.
- [ ] Add v1 protocol schemas copied from v0 plus level fields, new intents, `ready` state, and
  `level_changed`.
- [ ] Add representative v1 protocol examples for snapshot and transition deltas.
- [ ] Update `tools/validate_shared.py` for generation catalog, world mode, stair transition refs,
  v1 schemas, and negative integer level-name keys.

Focused check:

```bash
make validate-shared
```

### Task 2 — Server Level Model and Rule Loading

- [ ] Add `LevelState` with entities, walls, move, auto-nav, level number, and `nav *NavigationRules`.
  - Single-level worlds: `level.nav = &rules.Navigation` (pointer to global nav rules).
  - Dungeon levels: derive `NavigationRules` from `dungeon_generation` floor size and shared cell
    size/stop distance; compute `GridBounds` as `{MinX:0, MinY:0, MaxX: int(floor_width/cell), MaxY: int(floor_height/cell)}`.
  - Refactor all methods that reference `s.rules.Navigation` (pathfinding, approach, monster movement)
    to read from the active `LevelState.nav` instead.
- [ ] Refactor `Sim` to own `levels map[int]*LevelState` and `currentLevel`.
- [ ] Keep inventory, equipped, RNG, tick, seed, and entity allocation session-global.
- [ ] Add `Mode string \`json:"mode,omitempty"\`` to `WorldDef` in `rules.go`. Load and store on `Sim`
  as `multiLevel bool` (true when `world.Mode == "multi_level"`).
- [ ] Change `InteractableDef.BarrierWhenClosed` from value `InteractableBarrier` to pointer
  `*InteractableBarrier` in `rules.go`. Update `LoadRules` to not require the field for `ready`
  interactables. Existing call sites in `playerPositionBlocked`, `monsterPositionBlocked`,
  `firstProjectileHit`, and `lootDropBlocked` already gate on `state == "closed"`, so functional
  behavior is unchanged — only nil-guard the pointer dereference.
- [ ] Load dungeon generation rules in `rules.go`; expose as `Rules.DungeonGeneration`.
- [ ] Preserve single-level construction behavior through an implicit level `0`.
- [ ] Add focused tests proving legacy world snapshots and first entity IDs stay stable.

Focused check:

```bash
cd server && go test ./internal/game/... -run 'Rules|World|Snapshot'
```

### Task 3 — Deterministic Dungeon Generation

- [ ] Implement `GenerateDungeonLevel(seed, levelNum, rules)` in `dungeon_gen.go`.
- [ ] Use a per-level RNG stream derived from `SeedToUint64(session_seed + "|" + strconv.Itoa(abs(levelNum)))`. Use the existing `SeedToUint64` function from `rng.go` — do not invent a new hash.
- [ ] Generate perimeter walls from floor size/thickness.
- [ ] Generate level -1 with down stairs only.
- [ ] Generate level -2 and deeper with up and down stairs.
- [ ] Keep stairs separated by `min_separation` and inside wall margins.
- [ ] Add `shared/golden/dungeon_stairs.json` after implementation pins exact coordinates.
  - Include an optional `descend_then_ascend` case that asserts player position after ascending
    equals the `expected_stairs_down` coordinate on level -1, not the initial spawn `{4,10}`.
- [ ] Add Go and Godot golden checks against the fixture.

Focused check:

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Stairs'
make client-unit
```

### Task 4 — Transition Intents and Scoped Results

- [ ] Extend input decoding for `descend_intent` and `ascend_intent`.
- [ ] Reject transition intents for dead players, non-dungeon worlds, missing stair reach,
  entry-level ascend, and invalid levels.
- [ ] Move the player entity between `LevelState` maps without changing entity ID.
- [ ] Place arrivals at the matching stair on the destination level. Assert: ascending from -2 to -1
  places player at the **down-stair position on -1**, not the original spawn point.
- [ ] Change `Tick()` to return `[]TickResult` instead of `TickResult`. Add `Level int` field to
  `TickResult`. Normal ticks return a single-element slice (`level = currentLevel`). Transition
  ticks return two elements: from-level delta first (with `level_changed` event + player remove),
  then to-level delta (with complete spawn set). Update the runner to iterate the slice and emit one
  `state_delta` envelope per element.
- [ ] Include complete destination entity spawn set in the arrival delta.
- [ ] Ensure `Tick` iterates visited levels in sorted level order and only applies player movement
  on `currentLevel`.

Focused check:

```bash
cd server && go test ./internal/game/... -run 'Transition|Level|Replay'
```

### Task 5 — Realtime, Replay, and Bot Scenario

- [ ] Add `TypeDescend = "descend_intent"` and `TypeAscend = "ascend_intent"` constants to
  `inputdecode.go`. Add both to `IsClientIntent()` (so the runner buffers and persists them).
  Add `Decode` cases for both (empty payload — unmarshal `{}` then set `in.Descend = &DescendIntent{}`).
  Add `Descend *DescendIntent` and `Ascend *AscendIntent` to `game.Input`.
- [ ] Update realtime snapshot/delta marshaling to emit v1 payloads with level fields.
  Update runner to iterate `[]TickResult` and emit one delta per element.
- [ ] Ensure reconnect resume reconstructs visited levels and current floor by replaying inputs.
- [ ] Update replay timeline output for v1 snapshot/delta payloads.
- [ ] Add `12_dungeon_levels.json`.
- [ ] Extend protocol bot runtime with `current_level`, `visited_levels`, `use_stair`,
  `assert_current_level`, and `visited_levels_contain`.
- [ ] Confirm scenarios `01`-`11` still pass unchanged.

Focused check:

```bash
make db-up
make bot
```

### Task 6 — Godot Presentation

- [ ] Load `dungeon_generation.v0.json` client-side for level names and perimeter wall dimensions.
- [ ] Render simple stair meshes from `stairs_down` / `stairs_up` interactable entities.
- [ ] On stair clicks, send `descend_intent` or `ascend_intent` instead of `action_intent`.
- [ ] Track `current_level` from snapshots and `level_changed`.
- [ ] Clear active nodes on level change and apply the destination delta's complete spawn set.
- [ ] Render dungeon perimeter walls for `dungeon_levels`; keep legacy world wall rendering intact.
- [ ] Add top-right level HUD, hidden for level `0`.
- [ ] Add/extend client golden coverage for stair fixture and level-name fallback.

Focused check:

```bash
make client-unit
make client-smoke
```

### Task 7 — Docs and Final Verification

- [ ] Update `PROGRESS.md` only after v18 implementation ships.
- [ ] Run full CI once after focused gates pass.
- [ ] Optionally run visual replay for `12_dungeon_levels.json` to inspect stairs and HUD.

Final gate:

```bash
make ci
```

Optional visual inspection:

```bash
make bot-visual scenario=12_dungeon_levels.json
```

## 6. Known Risks

- **Refactor size:** `Sim` currently stores per-world state directly. Keep mechanical movement into
  `LevelState` separate from behavior changes where possible.
- **Protocol version drift:** update validators, examples, realtime, replay, bot, and client
  together so v0/v1 contracts do not silently diverge.
- **Destination completeness:** if the arrival delta omits any destination entity, the client will
  render an incomplete floor after clearing old nodes.
- **RNG contamination:** dungeon generation must not consume combat/loot RNG, or replay and golden
  fixtures will drift.
- **Inactive level ticking:** deterministic sorted level iteration is required even when v18 has no
  meaningful inactive monsters/projectiles.
- **Navigation bounds:** legacy worlds must keep existing `navigation.v0.json` bounds while
  `dungeon_levels` uses 32x20 level-local bounds.
