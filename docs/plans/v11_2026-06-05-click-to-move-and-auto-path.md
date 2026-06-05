# Click-to-Move and Auto-Path Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Server-authoritative grid pathfinding: floor click sends `move_to_intent`; entity click auto-approaches then executes `action_intent`; WASD cancels auto-nav; maze bot proves one-click kill.

**Architecture:** Shared `navigation.v0.json` bounds grid A\*; v11 `cell_size` equals authoritative `moveSpeed` so Go sim consumes one planned grid edge per tick via existing `resolveMovement`; client sends intents and optionally predicts; golden fixtures pin exact Go planner output and client-side fixture consistency.

**Tech stack:** Go sim/tests, shared JSON schemas + golden, Python protocol bot, Godot GDScript client.

**Spec:** [`docs/specs/v11_spec-click-to-move-and-auto-path.md`](../specs/v11_spec-click-to-move-and-auto-path.md)

**Branch:** `feature/click-to-move-and-auto-path` (off current integration branch)

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/navigation.v0.json` | `cell_size`, `max_auto_steps`, `grid_bounds`, `stop_distance` |
| Create | `shared/rules/navigation.v0.schema.json` | Validate navigation rules |
| Create | `shared/golden/auto_path.json` | Pinned path cases for Go + GDScript |
| Create | `shared/golden/auto_path.v0.schema.json` | Golden schema |
| Create | `shared/protocol/examples/move_to_intent.json` | Protocol example |
| Create | `server/internal/game/pathfind.go` | Grid rasterization + A\* + direction output |
| Create | `server/internal/game/pathfind_test.go` | Unit tests for planner |
| Create | `tools/bot/scenarios/05_path_maze.json` | Single-action maze proof |
| Modify | `shared/rules/worlds.v0.json` | Add `path_maze` preset |
| Modify | `shared/protocol/messages.v0.schema.json` | Add `move_to_intent` |
| Modify | `shared/protocol/envelope.v0.schema.json` | Sync intent enum |
| Modify | `server/internal/game/rules.go` | Load `NavigationRules` |
| Modify | `server/internal/game/sim.go` | `autoNav` queue, `handleMoveTo`, auto-approach in `handleAction`, cancel on manual move |
| Modify | `server/internal/game/sim.go` (`Input`) | Add `MoveTo *MoveToIntent` |
| Modify | `server/internal/game/game_test.go` | Auto-path, cancel, maze, reject tests |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode `move_to_intent`; extend `IsClientIntent` |
| Modify | `server/internal/realtime/protocol.go` | Wire constant if present |
| Modify | `server/internal/replay/replay_test.go` | Cover `move_to_intent` round-trip |
| Modify | `server/internal/http/ws_test.go` | Adjust tests that expect `out_of_range` on far action |
| Modify | `tools/validate_shared.py` | Navigation rules + auto_path golden drift checks |
| Modify | `tools/bot/scenarios/04_door_lab.json` | Drop far `out_of_range`; use auto-approach |
| Modify | `tools/bot/run.py` | Handle `no_path` / `path_too_long` rejects; load `path_maze` world |
| Modify | `tools/bot/test_protocol.py` | Assert scenario count includes `05_path_maze` |
| Modify | `client/scripts/main.gd` | Floor click → `move_to_intent`; entity click unchanged |
| Modify | `client/scripts/smoke.gd` | Migrate any far-action assumptions |
| Modify | `client/tests/test_golden.gd` | `auto_path.json` cases |
| Modify | `docs/PROGRESS.md` | v11 row when complete |
| Modify | `docs/specs/v11_spec-click-to-move-and-auto-path.md` | Status → Complete when done |

## Plugin adoption

- [x] Consult `docs/godot-plugins-and-shortcuts.md`.
- [x] Decision: **reject** NavMesh / NavigationServer plugins — in-repo grid A\* mirrors v9 collision.

---

## Task 1: Shared contracts

- [ ] **Step 1.1:** Create `navigation.v0.json` + schema per spec §4.1 (`cell_size: 1.0`, `max_auto_steps: 100`, bounds, `stop_distance: 0.25`). Keep `cell_size` equal to server `moveSpeed` for v11.

- [ ] **Step 1.2:** Add `path_maze` world to `worlds.v0.json` — player `(0, 5)`, monster `(10, 5)`, wall segments forcing ≥2 turns (tune during Task 2 golden fill-in).

- [ ] **Step 1.3:** Add `move_to_intent` to `messages.v0.schema.json`:

```json
"move_to_intent": {
  "type": "object",
  "required": ["position"],
  "additionalProperties": false,
  "properties": {
    "position": { "$ref": "#/$defs/vec2" }
  }
}
```

- [ ] **Step 1.4:** Add `move_to_intent` to envelope enum; create `shared/protocol/examples/move_to_intent.json`.

- [ ] **Step 1.5:** Create `auto_path.json` + schema with at least one case (`path_maze_start_to_monster_approach`); leave exact `expected_step_count` / `expected_end` as placeholders until Task 2 tunes maze, then lock values.

- [ ] **Step 1.6:** Run `make validate-shared` — must pass before Task 2.

---

## Task 2: Server — pathfinding core

- [ ] **Step 2.1:** Extend `rules.go` to load `NavigationRules` from `navigation.v0.json`; fail startup on invalid bounds, non-positive `cell_size`, or `cell_size != moveSpeed` for v11.

- [ ] **Step 2.2:** Create `pathfind.go` with:

```go
// PlanPath returns one-tick unit direction steps from start to goal using 8-way A*.
// blocked(cell) comes from sim obstacle rasterization.
func PlanPath(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool)
```

Implementation notes:

- World ↔ grid: `gx = floor(x / cell_size)`, same for y.
- 8 neighbors; cost 1 per step; octile heuristic.
- Tie-break: lower `gy`, then lower `gx`.
- Convert grid path to world centroids, then emit unit `Vec2` directions. Because `cell_size == moveSpeed`, each direction maps to one call through `resolveMovement`; diagonal steps are normalized by the same sim `normalize` path before movement.

- [ ] **Step 2.3:** Add `pathfind_test.go` — open field path, blocked cell detour, unreachable goal returns `ok=false`.

- [ ] **Step 2.4:** Add sim helper `buildBlockedFn()` rasterizing walls, live monsters, closed interactables from current `Sim` state (reuse `playerRadius`, AABB/circle math from `sim.go`).

- [ ] **Step 2.5:** Add `findMeleeApproachGoal(target *entity) (Vec2, []Vec2, bool)` — search rings around target in stable order until a candidate cell is inside navigation bounds, unblocked for the player, satisfies `meleeInRange` from cell center to target, and has a valid `PlanPath` from the current player position. Return the selected goal and its path so `handleAction` does not plan against a blocked or unreachable approach cell.

- [ ] **Step 2.6:** Run:

```bash
cd server && go test ./internal/game/... -run Path -v
```

---

## Task 3: Server — auto-nav queue

- [ ] **Step 3.1:** Add to `Sim`:

```go
type autoNavState struct {
    steps         []Vec2
    pendingAction *ActionIntent
    sourceMsgID   string
    sourceCorrID  string
}
```

Field on `Sim`: `autoNav *autoNavState`.

- [ ] **Step 3.2:** Add `MoveToIntent` + wire on `Input`:

```go
MoveToIntent struct {
    Position Vec2
}
```

- [ ] **Step 3.3:** Implement `handleMoveTo`:

1. Validate position finite.
2. If within `stop_distance` of goal → ack, return.
3. `PlanPath` to goal; `no_path` / `path_too_long` reject if applicable.
4. Set `s.autoNav = {steps, nil, msgID}`; ack.

- [ ] **Step 3.4:** Refactor `handleAction`:

1. Keep immediate dispatch when `inMeleeRange`.
2. Else plan to melee approach goal; reject `no_path` / `path_too_long` with **no** movement.
3. Queue `autoNav` with `pendingAction`, `sourceMsgID`, and `sourceCorrID`; ack once at queue time.
4. Remove `out_of_range` reject branch.

- [ ] **Step 3.5:** Refactor `handleMove` — **before** setting manual `s.move`, call `clearAutoNav()`.

- [ ] **Step 3.6:** Refactor `applyMovement`:

```text
if autoNav active and s.move nil:
  pop next direction → apply via resolveMovement (one step)
  if steps empty:
    if pendingAction → execute no-ack action dispatch inline (in range check)
    clear autoNav
elif s.move active:
  existing manual move logic
```

Implementation note: extract `dispatchAction(target, input, res, ack bool)` or equivalent so
immediate actions ack exactly once, while pending auto-nav actions do not emit a second
`intent_accepted` for the original message.

- [ ] **Step 3.7:** New `move_to_intent` in `applyInput`; include in dead-player reject list.

- [ ] **Step 3.8:** New navigation intent (`move_to_intent` or `action_intent` with approach) **replaces** existing `autoNav` (re-plan from current position).

- [ ] **Step 3.9:** Update `inputdecode.go`:

```go
const TypeMoveTo = "move_to_intent"

func IsClientIntent(t string) bool {
    switch t {
    case TypeMoveIntent, TypeMoveTo, TypeAction, TypeEquip:
        return true
    ...
}
```

- [ ] **Step 3.10:** Go tests in `game_test.go`:

| Test | Asserts |
|------|---------|
| `TestAutoPathGolden` | Matches `auto_path.json` |
| `TestMoveToIntentRejectsNoPath` | Sealed goal, position unchanged |
| `TestMoveToIntentRejectsPathTooLong` | Synthetic long corridor over budget |
| `TestActionIntentAutoApproachAndAttack` | `path_maze` single intent kills monster |
| `TestManualMoveCancelsAutoNav` | WASD clears queue mid-walk |
| `TestMoveToArrivesWithinStopDistance` | Floor click stops near goal |
| `TestPendingActionDoesNotDoubleAck` | Queued action acks once at queue time and not again on arrival |
| `TestApproachGoalSkipsBlockedCells` | Approach search ignores blocked/inaccessible melee cells |

- [ ] **Step 3.11:** Lock golden `expected_step_count` / `expected_end` from passing tests.

- [ ] **Step 3.12:** Run `cd server && go test ./internal/game/...`.

---

## Task 4: Replay and WebSocket

- [ ] **Step 4.1:** Add `move_to_intent` stored-input round-trip in `replay_test.go`.

- [ ] **Step 4.2:** Update `ws_test.go` — replace assertions expecting `out_of_range` on far `action_intent` where auto-approach now applies.

- [ ] **Step 4.3:** Run `cd server && go test ./...`.

---

## Task 5: Bot and validation

- [ ] **Step 5.1:** Extend `validate_shared.py` if needed (navigation file auto-validates via glob; add golden/rules cross-check for `path_maze` world_id in cases).

- [ ] **Step 5.2:** Create `05_path_maze.json` — single `action_once_until_event` step per spec §5.1. This helper must send exactly one `action_intent`, wait through auto-navigation ticks, and fail if that one message is rejected.

- [ ] **Step 5.3:** Migrate `04_door_lab.json`:

```json
"steps": [
  {
    "action": "action_entity",
    "interactable_def_id": "wooden_door",
    "event_type": "interactable_activated"
  },
  {
    "action": "action_entity",
    "item_def_id": "training_badge",
    "event_type": "item_picked_up"
  }
]
```

- [ ] **Step 5.4:** Update `run.py` — add `action_once_until_event` for path-maze proof; keep existing repeated `action_until_event` for combat loops that need multiple attacks. Add reject helpers for `no_path` / `path_too_long` if tests need them.

- [ ] **Step 5.5:** Update `test_protocol.py` scenario discovery expectations.

- [ ] **Step 5.6:** Run `make bot`.

---

## Task 6: Godot client

- [ ] **Step 6.1:** In `_try_action_at_mouse()` — if `_pick_entity_at_mouse()` returns `""`, send `move_to_intent` with ground point from `_mouse_ground_point()` (x/z → payload x/y).

- [ ] **Step 6.2:** Keep entity click → `action_intent` (unchanged wire shape).

- [ ] **Step 6.3:** Ensure WASD path in `_handle_input` still sends `move_intent` each tick (server cancel per spec §4.4).

- [ ] **Step 6.4:** Optional: skip client-side path prediction in v11 unless trivial — reconcile to server `entity_update` only.

- [ ] **Step 6.5:** Add `auto_path.json` cases to `client/tests/test_golden.gd`. If v11 skips client prediction, validate fixture/rules consistency only: referenced `world_id` exists, `navigation` matches `navigation.v0.json`, `cell_size == 1.0`, expected end is within melee reach for the referenced target mode, and expected step count is within `max_auto_steps`. If client prediction is added, mirror the planner and assert exact path output too.

- [ ] **Step 6.6:** Update `smoke.gd` if it asserts immediate `out_of_range`.

- [ ] **Step 6.7:** Run `make client-smoke`.

---

## Task 7: Documentation and CI

- [ ] **Step 7.1:** Update `docs/PROGRESS.md` — v11 lifecycle row, summary, close “click-to-move / pathfinding” deferred gap.

- [ ] **Step 7.2:** Set spec status to Complete with date.

- [ ] **Step 7.3:** Run `make ci`.

- [ ] **Step 7.4:** Optional `make bot-visual` — confirm maze walk in replay playlist.

---

## Verification commands

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Path|AutoNav|MoveTo|Maze|Door'
make bot
make client-smoke
make ci
```

## Implementation order

```text
shared contracts → pathfind.go + tests → sim autoNav + handleMoveTo/handleAction
→ inputdecode/replay/ws → bot scenarios → client click routing → PROGRESS + ci
```

Do not merge partial work: `move_to_intent` and auto-approach `action_intent` must be wired end-to-end before bot scenario `05` is expected green.

## Spec coverage self-review

| Spec § | Task |
|--------|------|
| §4.1 navigation rules | Task 1 |
| §4.2 move_to_intent | Tasks 1, 3, 6 |
| §4.3 action auto-approach | Task 3 |
| §4.4 pathfinding + WASD cancel | Tasks 2, 3 |
| §4.5 golden auto_path | Tasks 1, 2, 3, 6 |
| §4.6 client clicks | Task 6 |
| §4.7 path_maze world | Tasks 1, 5 |
| §5 bot scenarios | Task 5 |
| §6 acceptance | Task 7 |
