# Click Action and Melee Range - Implementation Plan

Goal: Unify attack, pickup, and door activation behind `action_intent` on left click, enforce
authoritative melee reach from shared rules, and prove a visible opening door that clears a
movement barrier.

Architecture: Shared rules declare reach and interactable defs; Go sim resolves action by target
type; Godot ray-picks entities and tweens door open on `state: open`; Python bot drives `door_lab`.

Tech stack: Go sim/tests, shared JSON schemas + golden, Python protocol bot, Godot GDScript smoke.

Branch: `feature/solid-collision-and-obstacles`

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/specs/v10_spec-click-action-and-melee-range.md` | Feature contract |
| Create | `docs/plans/v10_2026-06-05-click-action-and-melee-range.md` | This checklist |
| Create | `shared/rules/interactables.v0.json` | Door defs + barrier geometry |
| Create | `shared/rules/interactables.v0.schema.json` | Validate interactable catalog |
| Create | `shared/golden/melee_reach.v0.json` | Cross-language reach cases |
| Create | `shared/golden/melee_reach.v0.schema.json` | Golden schema |
| Create | `shared/protocol/examples/action_intent.json` | Protocol example |
| Delete | `shared/protocol/examples/attack_intent.json` | Removed protocol example |
| Delete | `shared/protocol/examples/pick_up_intent.json` | Removed protocol example |
| Modify | `shared/rules/combat.v0.json` | Add `unarmed_reach` |
| Modify | `shared/rules/combat.v0.schema.json` | Require `unarmed_reach` |
| Modify | `shared/rules/items.v0.json` | Add `reach` on `rusty_sword` |
| Modify | `shared/rules/items.v0.schema.json` | Optional `reach` on weapons |
| Modify | `shared/rules/worlds.v0.json` | Add `door_lab` preset |
| Modify | `shared/rules/worlds.v0.schema.json` | Allow `interactable` entity type |
| Modify | `shared/protocol/messages.v0.schema.json` | Add `action_intent`; remove attack/pickup |
| Modify | `shared/protocol/envelope.v0.schema.json` | Sync intent enum |
| Modify | `shared/protocol/state_delta.v0.schema.json` | `interactable` entity + event |
| Modify | `shared/protocol/session_snapshot.v0.schema.json` | `interactable` entity |
| Modify | `server/internal/game/rules.go` | Parse interactables, reach, world interactables |
| Modify | `server/internal/game/sim.go` | `handleAction`, reach, door state, barrier check |
| Modify | `server/internal/game/types.go` (or views) | EntityView interactable fields |
| Modify | `server/internal/game/game_test.go` | Reach golden, door, action tests |
| Modify | `server/internal/http/ws_test.go` | WebSocket tests use `action_intent` |
| Modify | `server/internal/inputdecode/inputdecode.go` | `action_intent` decode |
| Modify | `server/internal/realtime/protocol.go` | Wire type constant |
| Modify | `server/internal/realtime/runner.go` | Intent buffering if needed |
| Modify | `server/internal/replay/replay_test.go` | Stored replay inputs use `action_intent` |
| Modify | `tools/validate_shared.py` | Interactables + reach validation |
| Create | `tools/bot/scenarios/04_door_lab.json` | Door + range end-to-end proof |
| Modify | `tools/bot/scenarios/01_vertical_slice.json` | Use action steps |
| Modify | `tools/bot/scenarios/02_gear_before_combat.json` | Use action steps |
| Modify | `tools/bot/scenarios/03_collision_lab.json` | Use action steps if any attack/pickup |
| Modify | `tools/bot/run.py` | `action_intent`, `action_entity`, `move_until_in_range`, assertions |
| Modify | `tools/bot/test_protocol.py` | Door/action tests |
| Modify | `client/scripts/main.gd` | Raycast action, door render + tween, remove E/attack intents |
| Modify | `client/scripts/smoke.gd` | `action_intent` migration |
| Modify | `client/tests/test_golden.gd` | `melee_reach` golden cases |
| Modify | `docs/PROGRESS.md` | v10 row when complete |

## Plugin Adoption Checklist

- [x] Consulted `docs/godot-plugins-and-shortcuts.md`.
- [x] Decision: **reject** door/collision plugins — server owns barrier; client uses simple box + tween.

## Task 1: Shared contracts

- [x] Step 1.1: Add `unarmed_reach` to combat rules + schema.
- [x] Step 1.2: Add optional `reach` on weapon items + schema validation.
- [x] Step 1.3: Create `interactables.v0.json` + schema (`wooden_door`).
- [x] Step 1.4: Extend worlds schema with `interactable` type; add `door_lab` preset.
- [x] Step 1.5: Add `action_intent` to protocol; remove `attack_intent` / `pick_up_intent`.
- [x] Step 1.6: Delete stale `shared/protocol/examples/attack_intent.json` and `pick_up_intent.json`.
- [x] Step 1.7: Extend state_delta/session_snapshot for `interactable` entity + `interactable_activated`; require `entity_id` for `interactable_activated` in both `state_delta.events` and `session_snapshot.recent_events`.
- [x] Step 1.8: Add `melee_reach.v0.json` golden + schema.
- [x] Step 1.9: Run `make validate-shared`.

## Task 2: Server — reach and action dispatch

- [x] Step 2.1: Parse interactables and reach in `rules.go`; validate `door_lab` world.
- [x] Step 2.2: Add interactable fields to internal `entity` + `EntityView`.
- [x] Step 2.3: Spawn interactables from world preset in `NewSimWithWorld`.
- [x] Step 2.4: Implement `playerReach()`, `targetInteractionRadius()`, `inMeleeRange()`.
- [x] Step 2.5: Implement `handleAction` dispatching to attack/pickup/open interactable.
- [x] Step 2.6: Include closed interactable barriers in `playerPositionBlocked`.
- [x] Step 2.7: Wire `action_intent` in `applyInput`; remove attack/pickup handlers.
- [x] Step 2.8: Update `inputdecode` and realtime constants.
- [x] Step 2.9: Update replay tests (`server/internal/replay/replay_test.go`) and WebSocket tests (`server/internal/http/ws_test.go`) to send/store `action_intent`.
- [x] Step 2.10: Go tests — golden reach, out_of_range, door barrier, `door_lab` closed-door passage gating, existing slice golden.

## Task 3: Bot and replay

- [x] Step 3.1: Add bot helpers: `action_entity`, `action_until_event`, `move_until_in_range`.
- [x] Step 3.2: Add reject assertion helper for `out_of_range`.
- [x] Step 3.3: Create `04_door_lab.json`; assert beyond-door loot cannot be picked/reached before activation, then can be picked up after opening and crossing the doorway.
- [x] Step 3.4: Migrate `01` / `02` / `03` scenarios to action steps.
- [x] Step 3.5: Update `test_protocol.py`.
- [x] Step 3.6: Run `make bot`.

## Task 4: Godot client

- [x] Step 4.1: Add pick colliders + `entity_id` metadata on monster, loot, interactable nodes.
- [x] Step 4.2: Implement `_pick_entity_at_mouse()` raycast; left click → `action_intent`.
- [x] Step 4.3: Face clicked targets; play attack one-shot for monster and closed-door targets only, not loot.
- [x] Step 4.4: Remove E pickup and `attack_intent` / `pick_up_intent` sends.
- [x] Step 4.5: Spawn interactable door mesh; track `state` from snapshots/deltas.
- [x] Step 4.6: Tween door rotation on `state == open` / `interactable_activated`.
- [x] Step 4.7: Update autoplay and visual replay paths to `action_intent`.
- [x] Step 4.8: Migrate `smoke.gd` to `action_intent`.
- [x] Step 4.9: Add GDScript golden tests for melee reach.
- [x] Step 4.10: Run `make client-smoke`.

## Task 5: Documentation and CI

- [x] Step 5.1: Update `docs/PROGRESS.md` lifecycle table + v10 summary when done.
- [x] Step 5.2: Run `make ci`.
- [ ] Step 5.3: Optional `make bot-visual` — confirm door swing in `door_lab` replay.

## Verification commands

```bash
make validate-shared
cd server && go test ./internal/game/...
make bot
make client-smoke
make ci
```

## Implementation order

```text
shared contracts → server sim + tests → bot migration → client pick/action/door → PROGRESS + ci
```

Do not merge partial protocol migration: `action_intent` must be wired end-to-end before removing old intents from bot/smoke.
