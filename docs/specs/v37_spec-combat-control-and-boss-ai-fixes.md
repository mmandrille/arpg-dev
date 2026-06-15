# Spec: `combat-control-and-boss-ai-fixes`

Status: Draft
Branch: `main`
Slice: v37 - stationary directional attacks, aggro-on-hit, and boss attack repair
Baseline: v36 `inventory-paper-doll-capacity`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, deterministic replay, thin client
- [`../adr/0007-animation-state-model.md`](../adr/0007-animation-state-model.md) - client-only attack/reaction presentation
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - dungeon levels and generated hostile floors
- [`../adr/0009-boss-floors-and-timing-mechanics.md`](../adr/0009-boss-floors-and-timing-mechanics.md) - boss floor cadence, telegraph-first attacks, and locked exits
- [`v12_spec-ranged-projectile-combat.md`](v12_spec-ranged-projectile-combat.md) - ranged weapons and projectile authority
- [`v17_spec-monster-chase-movement.md`](v17_spec-monster-chase-movement.md) - chase behavior and aggro events
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - proactive monster attacks
- [`v27_spec-hold-click-controls.md`](v27_spec-hold-click-controls.md) - sustained mouse input
- [`v35_spec-boss-floor-gate.md`](v35_spec-boss-floor-gate.md) - first boss floor and pattern events

## 1. Purpose

Several combat behaviors now block the intended ARPG feel:

1. Clicking a target or floor can move the character when the player wants to stand still and
   attack in a direction.
2. Monsters can be damaged from outside their passive aggro radius without reliably walking toward
   the attacker afterward.
3. The level `-5` boss exists and emits pattern state, but in play it is not threatening the player
   because it does not move/attack correctly.

This slice fixes those behaviors while preserving the authoritative boundary:

- Holding `SHIFT` while left-clicking performs a **stationary directional attack**. The character
  stays grounded, faces the clicked direction, and attacks without issuing `move_to_intent`.
- Directional melee attacks use an authoritative narrow cone/capsule in front of the player and
  select the nearest valid hostile in that area.
- Directional ranged attacks fire freely in the clicked direction. They do not require an enemy
  target.
- Holding `SHIFT+LMB` repeats directional attacks using the existing client attack cadence.
- Any monster damaged by a player enters aggro/chase toward that player, even if the hit came from
  outside its passive aggro radius.
- Boss monsters move toward the player and execute their authoritative pattern attacks. The level
  `-5` boss can damage a player who fails to dodge or kite.

The proof is: client `SHIFT+click` input -> new/extended authoritative directional attack contract
-> server-owned melee/projectile hit resolution -> aggro-on-hit -> boss chase + telegraphed damage
-> bot and replay coverage.

## 2. Non-goals

- No skill bar, skill tree, mana spender, or active ability catalog.
- No final attack-speed implementation. Use the existing client send cadence unless the plan finds
  a small server-side guard necessary for safety.
- No targeting assist, homing, lock-on, or target prediction for directional attacks.
- No client-side hit detection. The client can show immediate animation/projectile feedback only;
  the server owns hit selection, damage, aggro, boss attacks, HP, death, loot, and XP.
- No PvP or friendly-fire directional attacks.
- No new boss templates, boss art, arena art, audio, or production VFX.
- No boss enrage, adds, phase deck expansion, or co-op boss scaling.
- No broad monster AI rewrite beyond aggro-on-hit and boss movement/pattern repair.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v37_spec-combat-control-and-boss-ai-fixes.md       - this slice contract
docs/plans/v37_<YYYY-MM-DD>-combat-control-and-boss-ai-fixes.md - implementation plan
PROGRESS.md                                               - lifecycle update when v37 ships

shared/protocol/messages.v*.schema.json                        - directional attack intent contract
shared/protocol/envelope.v*.schema.json                        - message type allowlist if needed
shared/protocol/examples/state_delta.json                      - directional hit / boss attack examples if schema changes
shared/protocol/examples/session_snapshot.json                 - boss/chase state example only if snapshot shape changes
shared/golden/directional_attack.json                          - optional fixture for melee/ranged directional semantics
shared/golden/directional_attack.v0.schema.json                - optional fixture schema
tools/validate_shared.py                                       - validate any new shared fixture/schema

server/internal/inputdecode/inputdecode.go                     - decode directional attack payload if new message type
server/internal/inputdecode/inputdecode_test.go                - payload validation tests
server/internal/game/sim.go                                    - directional attack, aggro-on-hit, boss chase/pattern damage
server/internal/game/types.go                                  - new input/event view fields if needed
server/internal/game/game_test.go                              - stationary direction attacks, aggro-on-hit, boss damage tests
server/internal/replay/replay_test.go                          - replay parity if input schema changes
server/internal/http/ws_test.go                                - WebSocket schema/acceptance tests if needed

client/scripts/main.gd                                         - SHIFT+click and held SHIFT+click behavior
client/scripts/sustained_click_input.gd                        - repeat mode support if reused for SHIFT+click
client/scripts/bot_controller.gd                               - synthetic client action helper if needed
client/scripts/bot_scenario_runner.gd                          - client assertions for stationary attack
client/tests/test_client_bot.gd                                - client bot validation
client/tests/test_sustained_input.gd                           - held input regression coverage

tools/bot/run.py                                               - protocol helpers/assertions for directional attack and boss damage
tools/bot/test_protocol.py                                     - bot helper unit coverage
tools/bot/scenarios/26_combat_control_and_boss_ai_fixes.json   - protocol proof
tools/bot/scenarios/client/14_shift_click_stationary_attack.json - client proof if reliable
```

Protocol note: `action_intent { target_id }` is target-based and cannot faithfully represent free
directional attacks. The plan must choose one of:

1. Add `directional_attack_intent { direction: {x, y} }`.
2. Add `attack_intent` with a discriminated target-or-direction payload.
3. Extend `action_intent` in a backward-compatible way with optional `direction`.

Default: add a new explicit `directional_attack_intent` so old click-action semantics remain
unchanged and tests can distinguish movement clicks from force-stand attacks.

## 4. Directional attack controls

### 4.1 Client input behavior

When gameplay input is allowed and the player holds `SHIFT` while pressing left mouse:

1. The client computes the ground point under the mouse.
2. The client derives a flat normalized direction from player position to that ground point.
3. The client faces the character in that direction and may play the current attack one-shot
   animation.
4. The client sends the authoritative directional attack intent.
5. The client does **not** send `move_to_intent` or start/continue floor hold-move from this input.

If the direction is degenerate because the clicked point is too close to the player, default to the
current facing direction if available, otherwise use a stable fallback direction.

### 4.2 Held SHIFT+click

Holding `SHIFT+LMB` repeats directional attacks using the existing attack send cadence.

Rules:

- Repeat while `SHIFT` remains held, LMB remains held, gameplay input is allowed, and player HP is
  above zero.
- Stop or pause repeat while menus, inventory, character panel, main menu, pause menu, or bot locks
  block input.
- Recompute direction from the current mouse ground point on each repeat.
- Do not emit floor movement intents while the SHIFT attack hold is active.
- Existing non-SHIFT hold-click behavior from v27 remains unchanged.

### 4.3 Client proof

Client debug/bot state must expose enough data to assert that synthetic `SHIFT+click`:

- sends a directional attack or equivalent debug action,
- does not send `move_to_intent`,
- leaves the player position within a small tolerance over the assertion window,
- faces or attacks toward the clicked direction.

If headless mouse modifier simulation is unreliable, the plan may expose a direct bot helper that
dispatches the same internal path as human `SHIFT+click`.

## 5. Authoritative directional attack contract

### 5.1 Direction payload

The authoritative payload should use the existing world X/Z plane naming convention:

```json
{
  "direction": { "x": 1.0, "y": 0.0 }
}
```

Server validation:

- `direction.x` and `direction.y` must be finite numbers.
- Zero-length direction is rejected with a stable reason such as `invalid_direction`, unless the
  server chooses a stable fallback and documents it in the plan.
- Server normalizes the direction before resolving hits or projectiles.
- Dead players, town safe-zone restrictions, and other existing gameplay locks still apply.

### 5.2 Directional melee

When the player's current attack mode is melee:

- The server builds a narrow deterministic attack area in front of the player.
- Default shape: short capsule or narrow cone aligned to the requested direction.
- Range uses current melee reach plus target interaction radius, consistent with target-based melee.
- Valid targets are live hostile monsters on the player's current level.
- If multiple monsters are inside the attack area, choose the nearest valid target; tie-break by
  entity id for determinism.
- If no target is inside the area, acknowledge the input and emit no damage, or emit an
  `attack_missed` event if the plan wants visible feedback. The plan must choose one behavior.

The client must not decide which monster was hit. It only sends direction.

### 5.3 Directional ranged

When the player's current attack mode is ranged:

- The server spawns a projectile from the player position moving in the requested direction.
- The projectile is not tied to a target id.
- Existing projectile collision against walls, interactables, and monsters should be reused where
  possible.
- Projectile max distance, speed, damage, hit/miss/crit/block handling, and `projectile_busy`
  behavior remain authoritative.
- A ranged directional shot can miss all enemies and expire normally.

The projectile view may omit `target_id` for free shots. Protocol schemas and client rendering must
allow that if current projectile views require a target.

## 6. Aggro on hit

Any monster damaged by a player must enter aggro/chase toward that attacking player.

Rules:

- Applies to target-based melee hits, directional melee hits, projectile impacts, and future player
  damage sources that route through the same damage helper.
- Applies even if the monster was outside passive `aggro_radius`.
- Applies to all monsters as hostile behavior. Static/training monsters that cannot move may record
  aggro without moving only if their definition explicitly cannot chase; generated dungeon monsters
  and the boss must move.
- Emits or preserves `monster_aggro` events so bots can assert the transition.
- Uses stable entity/player ids and no unordered map iteration.
- Does not aggro dead monsters.
- Aggro is contagious across nearby live chase-capable monsters on the same level. When a damaged
  monster enters aggro, close chase monsters inherit the same attacking player target; this may
  propagate through a connected nearby pack, but monsters outside the group radius stay idle.
- Aggro-on-hit also wakes other live chase-capable monsters that already have the attacking player
  inside their own `aggro_radius`, even if they are not close enough to the damaged monster to join
  through the nearby pack chain.

For co-op, the damaged monster should chase the player that caused the damage, not always the
session host/default player.

## 7. Boss movement and pattern attacks

### 7.1 Boss movement

The level `-5` boss must move toward an attacking/nearby player instead of staying inert.

Rules:

- Bosses participate in hostile movement/chase logic.
- Boss movement must preserve existing boss visual metadata and phase event rendering.
- Boss movement must not cancel the telegraph-first guarantee.
- Boss movement must be deterministic and replay-safe.
- If the boss is in an active pattern phase, the plan must specify whether it can move during that
  phase or pauses movement until recovery/cooldown. Default: boss may chase during idle/cooldown and
  telegraph/recovery, but active hit resolution uses the authoritative phase hit shape at that tick.

### 7.2 Boss pattern damage

The level `-5` boss must be able to damage a player who fails to dodge.

Acceptance:

- Boss emits `boss_phase_started` and `boss_phase_ended` as in v35.
- During an active damaging phase, if the player remains inside the authoritative hit shape/contact
  condition, the player receives `player_damaged` or `player_killed`.
- If the player exits the telegraphed danger before the active phase, the attack does not damage
  the player.
- The bot can intentionally stand still or stay in range and observe HP decrease.
- The existing v35 dodge/no-damage behavior remains covered.

### 7.3 Locked exit behavior

Boss-floor locked exit behavior remains unchanged:

- Down stairs and boss-floor teleporter remain locked/disabled while the boss is alive.
- Killing the boss unlocks them.
- This slice should not weaken the v35 progression gate.

## 8. Tests and bot proof

### 8.1 Go tests

Add focused Go tests for:

- Directional melee hits a monster in front of the player without a target id.
- Directional melee does not hit a monster behind/outside the cone/capsule.
- Directional melee tie-breaks deterministically.
- Directional ranged spawns a projectile in the requested direction and can hit a monster along
  that path.
- Directional ranged projectile can expire without hitting anything.
- Damaging a monster outside passive aggro radius causes `monster_aggro` and chase toward attacker.
- Aggro-on-hit propagates through nearby chase-capable monster groups, while monsters outside the
  group radius do not inherit the target.
- Boss moves toward the player and damages the player during an active pattern when the player does
  not dodge.
- Boss dodge/no-damage fixture from v35 remains valid.

### 8.2 Protocol bot proof

Add a scenario such as `26_combat_control_and_boss_ai_fixes.json`:

1. Use or equip a ranged weapon.
2. Fire a directional ranged shot from outside a chase monster's passive aggro radius.
3. Assert the monster is damaged.
4. Assert the monster aggros and moves toward the player.
5. Reach or directly spawn into the level `-5` boss proof setup.
6. Assert the boss moves toward the player.
7. Intentionally fail to dodge one boss active phase.
8. Assert player HP decreases from boss damage.
9. Kill the boss and assert locked exits still unlock.

### 8.3 Client bot proof

Add a client scenario if reliable:

- Synthetic `SHIFT+click` sends stationary directional attack.
- Player position remains within tolerance.
- No floor `move_to_intent` occurs from the SHIFT attack path.
- Holding `SHIFT+LMB` repeats attacks at the existing cadence.

If this is flaky in headless Godot, cover the extracted input helper with `client/tests` and use the
protocol bot for end-to-end gameplay outcomes.

## 9. Acceptance criteria

- `SHIFT+click` does not move the player.
- `SHIFT+click` melee can hit an enemy in the clicked direction.
- `SHIFT+click` ranged fires freely in the clicked direction.
- Holding `SHIFT+LMB` repeats stationary directional attacks.
- Non-SHIFT click-to-move, target-click attack, pickup, doors, stairs, teleporters, inventory, and
  hold-click movement behavior remain intact.
- Monsters damaged by a player aggro/chase that player even from outside passive aggro radius.
- The level `-5` boss moves and can damage the player through pattern attacks.
- Boss attacks remain telegraph-first and dodgeable.
- Replay remains deterministic.
- `make ci` is green.

## 10. Open implementation notes

- Prefer a new explicit `directional_attack_intent` unless the plan finds a strong reason to
  extend `action_intent`.
- Keep melee shape constants in shared rules only if they need tuning or client display. If they
  are first-pass server-only constants, document them in tests and defer shared data.
- If projectile views currently require `target_id`, make that field optional for free shots and
  update Godot rendering accordingly.
- Boss movement should reuse chase movement as much as possible rather than creating a parallel
  boss-only movement system.
- Aggro-on-hit should be implemented in one common damage path so future player damage sources
  inherit it automatically.
