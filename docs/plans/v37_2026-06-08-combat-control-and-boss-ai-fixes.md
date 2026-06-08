# v37 Plan - Combat Control and Boss AI Fixes

Status: Ready for implementation
Goal: Add force-stand directional attacks, aggro-on-hit, and boss chase/damage repair without moving combat authority out of the Go sim.
Architecture: Add one explicit client-to-server `directional_attack_intent` for attacks that have a direction but no target id. Treat zero-vector `move_intent` as an authoritative movement cancel so pressing `SHIFT` can immediately stop click-to-move, auto-approach, hold-move, and short buffered movement. Keep melee hit selection, projectile collision, aggro target selection, boss movement, boss phases, HP, loot, and replay outcomes server-owned. The Godot client only computes direction, sends intents, stops local movement presentation, and plays existing attack feedback.
Tech stack: Shared JSON protocol/examples/world presets, Go deterministic sim, Python protocol bot, Godot GDScript client/input tests.

## Baseline and shortcut decision

Baseline is v36 `inventory-paper-doll-capacity` on `main`, building on v27 sustained click controls, v12 projectile authority, v17/v21 chase combat, v33 co-op actor-scoped inputs, and v35 boss floor phase/gate work.

Godot plugin adoption decision for this slice: **reject external plugins**. The work is input routing and authoritative combat behavior, not a new UI/camera/art system. Reuse/borrow existing in-repo paths instead: `sustained_click_input.gd`, `_mouse_ground_point()`, `_aim_direction_from_mouse()`, existing animation one-shots, and current boss telegraph presentation.

Plan decisions:

- Add `directional_attack_intent { direction: { x, y } }` to the current protocol schema set with coordinated client/server updates and no compatibility shim.
- Pressing `SHIFT` by itself is force-stand: the client clears sustained move state, suppresses movement sends while held, and sends zero-vector `move_intent` once to cancel authoritative movement/auto-nav.
- Server `move_intent` with normalized zero direction cancels active movement and auto-nav, then acks.
- Any directional attack also cancels active movement and auto-nav before resolving the attack.
- Server rejects non-finite or zero-length directional attack vectors with `invalid_direction`; the client provides a stable facing fallback before sending.
- Directional melee uses a first-pass server-only narrow capsule in front of the player. No shared tuning rule is added until the shape needs client display or balance iteration.
- Directional melee with no target in the capsule acks and emits no damage/miss event.
- Directional ranged uses the existing projectile entity and swept collision, but free-shot projectile views omit `target_id`.
- Aggro-on-hit is centralized in the player-damages-monster path. It records the attacking player as the preferred chase target for co-op.
- Bosses may move during idle, cooldown, telegraph, and recovery. Bosses pause movement during active damage ticks so the telegraph remains fair and active contact is resolved from the authoritative position at that tick.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/protocol/envelope.v2.schema.json` | Allow `directional_attack_intent` in the envelope type enum. |
| Modify | `shared/protocol/messages.v2.schema.json` | Add directional attack payload schema. |
| Create | `shared/protocol/examples/directional_attack_intent.json` | Concrete message example for validation and docs. |
| Modify | `shared/protocol/examples/move_intent.json` | Document zero-vector stop movement semantics if examples validate this path cleanly. |
| Modify | `shared/protocol/examples/state_delta.json` | Include a free-shot projectile example without `target_id` if examples need schema coverage. |
| Modify | `shared/rules/worlds.v0.json` | Add `combat_control_lab` with exact bow/chase-monster positions for protocol proof. |
| Modify | `tools/validate_shared.py` | Validate new examples and any latest-schema assumptions. |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode directional attack payload and mark it as a client intent. |
| Modify | `server/internal/inputdecode/inputdecode_test.go` | Decode validation for direction and invalid payloads. |
| Modify | `server/internal/game/types.go` | Add `DirectionalAttackIntent`; keep projectile `target_id` omittable. |
| Modify | `server/internal/game/sim.go` | Stop movement semantics, directional attacks, aggro-on-hit, boss movement repair. |
| Modify | `server/internal/game/game_test.go` | Focused sim tests for stop movement, directional hit/miss/tie, projectile free shots, aggro, boss damage. |
| Modify | `server/internal/replay/replay_test.go` | Replay parity when the new intent appears in recorded inputs. |
| Modify | `server/internal/http/ws_test.go` | WebSocket acceptance/schema coverage if needed for the new intent. |
| Modify | `client/scripts/main.gd` | SHIFT force-stand, SHIFT+click directional attack, held SHIFT+LMB repeats, movement suppression. |
| Modify | `client/scripts/sustained_click_input.gd` | Add directional hold state or helper methods. |
| Modify | `client/scripts/bot_controller.gd` | Synthetic helper for stationary directional attack if mouse modifiers are unreliable. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Assertions for no movement intent and stationary attack debug state if a client scenario is added. |
| Modify | `client/tests/test_sustained_input.gd` | Regression coverage for force-stand cancellation and directional hold state. |
| Create | `client/tests/test_directional_attack_input.gd` | Focused input helper test if it keeps `main.gd` changes testable. |
| Modify | `tools/bot/run.py` | Directional attack action/helper, stop movement helper, boss/player HP assertions. |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for bot state/actions/assertions. |
| Create | `tools/bot/scenarios/26_combat_control_and_boss_ai_fixes.json` | Directional ranged and aggro-on-hit protocol proof. |
| Modify | `tools/bot/scenarios/24_boss_floor_gate.json` | Add boss movement/damage proof while preserving unlock proof. |
| Create/Modify | `tools/bot/scenarios/client/14_shift_click_stationary_attack.json` | Client proof only if reliable; otherwise keep coverage in client tests. |
| Modify | `docs/PROGRESS.md` | Lifecycle update when v37 ships. |

## Task 1 - Shared protocol and lab world

Files:
- Modify: `shared/protocol/envelope.v2.schema.json`
- Modify: `shared/protocol/messages.v2.schema.json`
- Create: `shared/protocol/examples/directional_attack_intent.json`
- Modify: `shared/protocol/examples/move_intent.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `directional_attack_intent` to message/envelope allowlists with payload `{ "direction": { "x": number, "y": number } }`; keep actor/player ids out of the payload.
```bash
make validate-shared
```

- [x] Step 1.2: Add a directional attack example and document zero-vector `move_intent` as stop/cancel. If example validation still points at older schemas, update `tools/validate_shared.py` to validate examples against the current schema set.
```bash
make validate-shared
```

- [x] Step 1.3: Confirm projectile entity schemas allow omitted `target_id`; update examples so a free directional shot is represented without a fake `"0"` target.
```bash
make validate-shared
```

- [x] Step 1.4: Add `combat_control_lab` to `shared/rules/worlds.v0.json`: player at `{x:2,y:5}`, `training_bow` loot at `{x:3,y:5}`, one `dungeon_mob` at `{x:13,y:5}` outside passive aggro radius but within leash, plus the standard boundary walls.
```bash
make validate-shared
```

## Task 2 - Input decode and authoritative stop movement

Files:
- Modify: `server/internal/inputdecode/inputdecode.go`
- Modify: `server/internal/inputdecode/inputdecode_test.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add `DirectionalAttackIntent` to game input types and decode `directional_attack_intent` in `inputdecode`, including missing/malformed payload rejection.
```bash
cd server && go test ./internal/inputdecode
```

- [x] Step 2.2: Add `directional_attack_intent` to `IsClientIntent`, `Sim.applyInput`, dead-player rejection, and duplicate/ack handling.
```bash
cd server && go test ./internal/game/... -run 'TestDirectional|TestDeadPlayer'
```

- [x] Step 2.3: Change zero-vector `move_intent` handling to clear active `move` and `autoNav`, ack, and not queue a new move. Keep non-zero `move_intent` behavior unchanged.
```bash
cd server && go test ./internal/game/... -run TestStopMovementIntent
```

- [x] Step 2.4: Add tests proving zero-vector stop cancels click-to-move and pending action auto-approach before the player advances another tick.
```bash
cd server && go test ./internal/game/... -run TestStopMovementIntent
```

## Task 3 - Server directional attacks and aggro-on-hit

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/replay/replay_test.go`
- Modify: `server/internal/http/ws_test.go` if the WebSocket tests need explicit intent coverage

- [x] Step 3.1: Implement `handleDirectionalAttack`: validate finite/non-zero direction, normalize it, clear movement/auto-nav, then branch by current player attack mode.
```bash
cd server && go test ./internal/game/... -run TestDirectionalAttack
```

- [x] Step 3.2: Implement directional melee target selection as a deterministic narrow capsule: projection in front of the player, reach based on current melee reach plus target radius, narrow half-width, nearest target wins, entity id tie-break.
```bash
cd server && go test ./internal/game/... -run TestDirectionalMelee
```

- [x] Step 3.3: Add melee tests for hit in front, no hit behind/outside capsule, deterministic tie-break, no movement after force-stand attack, and ack/no event when the swing hits nothing.
```bash
cd server && go test ./internal/game/... -run TestDirectionalMelee
```

- [x] Step 3.4: Refactor projectile spawning so target-based ranged attacks and directional ranged shots share the same authoritative projectile creation/collision path; set `targetID` only for target-based shots.
```bash
cd server && go test ./internal/game/... -run TestDirectionalRanged
```

- [x] Step 3.5: Add ranged tests for free projectile direction, hit along the path, wall/interactable blocking reuse, projectile expiration without hit, `projectile_busy`, and omitted projectile `target_id` in the entity view.
```bash
cd server && go test ./internal/game/... -run TestDirectionalRanged
```

- [x] Step 3.6: Centralize player-to-monster damage mutation in one helper used by target melee, directional melee, and projectile impacts. After positive HP damage, call aggro-on-hit before retaliation/kill follow-up.
```bash
cd server && go test ./internal/game/... -run 'TestAggroOnHit|TestRangedProjectile'
```

- [x] Step 3.7: Add `aiTargetPlayerID` or equivalent to monster entities. For chase-capable monsters and bosses, aggro-on-hit sets the attacker as preferred target; target selection falls back to nearest living player only if the preferred player is gone/dead/off-level.
```bash
cd server && go test ./internal/game/... -run 'TestAggroOnHit|TestCoop'
```

- [x] Step 3.8: Emit `monster_aggro` with stable ids when a damaged live monster newly enters chase/aggro. Static monsters may emit aggro without movement; generated dungeon mobs and bosses must move.
```bash
cd server && go test ./internal/game/... -run TestAggroOnHit
```

- [x] Step 3.9: Add replay parity coverage for directional attack inputs and aggro/boss state after reconstruction.
```bash
cd server && go test ./internal/replay/...
```

- [x] Step 3.10: Make aggro-on-hit contagious through nearby live chase-capable monsters on the same level. Close monsters inherit the same attacking player target, propagation can chain through a nearby group, and monsters outside the group radius stay idle.
```bash
cd server && go test ./internal/game/... -run 'TestAggroOnHit|TestDirectionalRanged'
```

## Task 4 - Boss movement and damaging pattern repair

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `tools/bot/scenarios/24_boss_floor_gate.json`

- [x] Step 4.1: Let boss monsters participate in hostile movement instead of skipping them in `advanceMonsterMovement`; reuse base monster chase speed/nav/blocking.
```bash
cd server && go test ./internal/game/... -run TestBoss
```

- [x] Step 4.2: Apply boss movement only during idle, cooldown, telegraph, and recovery. Do not move bosses during active ticks.
```bash
cd server && go test ./internal/game/... -run TestBoss
```

- [x] Step 4.3: Ensure boss target selection uses aggro-on-hit preferred player when set, otherwise nearest living player on the boss level. Keep sorted player/entity tie-breaks.
```bash
cd server && go test ./internal/game/... -run 'TestBoss|TestCoop'
```

- [x] Step 4.4: Add a focused boss test where the player fails to dodge/contact-break and receives `player_damaged` during the active phase.
```bash
cd server && go test ./internal/game/... -run TestBossDamagesStationaryPlayer
```

- [x] Step 4.5: Keep the v35 dodge/no-damage and locked-exit/unlock tests green; adjust only if they relied on the boss being inert rather than telegraph-first.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloor|TestBossPattern'
```

## Task 5 - Python protocol bot scenarios

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/26_combat_control_and_boss_ai_fixes.json`
- Modify: `tools/bot/scenarios/24_boss_floor_gate.json`

- [x] Step 5.1: Add a `directional_attack` bot step that sends `directional_attack_intent` with either an explicit direction or a direction computed from player to a selected entity.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.2: Add bot assertions for player HP decrease from a start snapshot, boss/chase monster moved by entity id, and optional projectile target omission if useful.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.3: Create `26_combat_control_and_boss_ai_fixes.json`: use `combat_control_lab`, pick up/equip `training_bow`, fire a directional ranged shot from outside `dungeon_mob` passive aggro radius, assert damage, `monster_aggro`, and movement toward the player.
```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=combat_control_and_boss_ai_fixes
```

- [x] Step 5.4: Update `24_boss_floor_gate.json` so it also proves boss movement and intentional failed-dodge damage before killing the boss and asserting exits still unlock.
```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=boss_floor_gate
```

- [x] Step 5.5: Run the full protocol bot catalog after individual scenarios pass.
```bash
make bot
```

## Task 6 - Godot force-stand and directional input

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/sustained_click_input.gd`
- Create: `client/scripts/directional_attack_input.gd`
- Modify: `client/tests/test_sustained_input.gd`
- Create: `client/tests/test_directional_attack_input.gd`
- Deferred: `tools/bot/scenarios/client/14_shift_click_stationary_attack.json`

- [x] Step 6.1: Add a shift modifier helper. On `SHIFT` key press while gameplay input is allowed, clear sustained-click state, send zero-vector `move_intent`, and stop local click-to-move/hold-move presentation.
```bash
make client-unit
```

- [x] Step 6.2: While `SHIFT` is held, suppress WASD movement sends, floor `move_to_intent`, and hold-move repeats. Releasing `SHIFT` must not start normal movement until the user gives a new movement input.
```bash
make client-unit
```

- [x] Step 6.3: On `SHIFT+LMB`, compute mouse ground direction from player to cursor, fall back to last facing direction when degenerate, face the character, play the existing attack one-shot, and send `directional_attack_intent`.
```bash
make client-unit
```

- [x] Step 6.4: Add held `SHIFT+LMB` repeat at existing `SEND_INTERVAL`: recompute direction each repeat, keep the player stationary, and stop repeat if `SHIFT`, LMB, gameplay input, or player HP state no longer allows it.
```bash
make client-unit
```

- [x] Step 6.5: Preserve non-SHIFT behavior for click-to-move, target-click attack/auto-approach, loot, doors, stairs, teleporters, inventory, and existing hold-click movement.
```bash
make client-unit
```

- [x] Step 6.6: Add client test coverage for force-stand cancellation, directional hold start/stop/cadence, and no `move_to_intent` from the SHIFT attack path. If full scene input is brittle, expose a direct helper used by both human input and `bot_controller.gd`.
```bash
make client-unit
```

- [x] Step 6.7: Deferred `tools/bot/scenarios/client/14_shift_click_stationary_attack.json`: headless modifier/mouse proof is not reliable with the current bot controller fallback, so v37 relies on `client/tests/test_directional_attack_input.gd`, `client/tests/test_sustained_input.gd`, and protocol bot proof instead.
```bash
HEADLESS=1 make bot-client scenario=14_shift_click_stationary_attack.json
```

## Task 7 - Lifecycle docs and CI

Files:
- Modify: `docs/PROGRESS.md`
- Modify: `docs/specs/v37_spec-combat-control-and-boss-ai-fixes.md` only if implementation forces an as-built clarification
- Modify: `docs/plans/v37_2026-06-08-combat-control-and-boss-ai-fixes.md` if task scope changes during implementation

- [x] Step 7.1: Update `docs/PROGRESS.md` with v37 lifecycle row, latest completed slice, what v37 proved, and any newly deferred combat/boss/client-control items.
```bash
rg -n "v37|combat-control-and-boss-ai-fixes|Latest completed slice|Open gaps" docs/PROGRESS.md
```

- [x] Step 7.2: Run final local CI.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/inputdecode`
- [x] `cd server && go test ./internal/game/...`
- [x] `cd server && go test ./internal/replay/...`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=combat_control_and_boss_ai_fixes`
- [x] `ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=boss_floor_gate`
- [x] `make bot`
- [x] Not run: `HEADLESS=1 make bot-client scenario=14_shift_click_stationary_attack.json` because the client scenario is deferred in Step 6.7; `make client-unit` and protocol bot coverage are the v37 gates.
- [x] `make ci`

## Deferred scope

- No final attack-speed/cooldown gameplay, skill bar, mana system, active ability catalog, homing/target prediction, client hit detection, PvP/friendly fire, new boss templates, production boss/combat VFX/audio/art, boss enrage/adds, co-op boss scaling, broad monster AI rewrite, or Protobuf migration.
- Directional melee shape constants remain server-only first-pass values unless implementation discovers the client must render them or shared tuning is needed.
- A Godot client bot scenario for modifier mouse input may be deferred if headless input remains unreliable; client unit/helper tests plus protocol bot coverage remain mandatory.
