# v295 As-Built: Input Buffering

Date: 2026-06-19
Spec: [`docs/specs/v295_spec-input-buffering.md`](../specs/v295_spec-input-buffering.md)
Plan: [`docs/plans/v295_2026-06-19-input-buffering.md`](../plans/v295_2026-06-19-input-buffering.md)

## What Shipped

- Added `CombatInputBuffer`, a client-only short-lived monster attack target buffer with replacement,
  expiry, and clear guards for missing targets, dead targets, non-monsters, and dead players.
- Extracted local basic-attack reach checks into `CombatReach`, including equipped weapon reach
  resolution from the client inventory/equipped state and existing local radius constants.
- Updated `main.gd` so monster clicks during local basic-attack cooldown queue briefly instead of
  being dropped, then dispatch through the existing `_send_action_intent`, facing, animation,
  cooldown, and recovery UI path when local cooldown and local reach allow it.
- Kept the server fully authoritative: no protocol, range, cooldown, damage, loot, or monster AI
  behavior changed.
- Kept sustained attack hold intact as the repeat path; the new buffer is a one-shot fallback for a
  missed click near cooldown/range legality.
- Adjusted bot mode so scripted client actions still tick local cooldowns and the attack buffer
  while real mouse/WASD input remains blocked.
- Added a `click_entity_buffered` client bot action and wait-event support for multiple acceptable
  combat event types plus dynamic target-entity matching.
- Added `77_input_buffering`, which equips the control-lab training bow, clicks the dungeon mob,
  clicks again during local recovery, and observes a later authoritative `monster_damaged` or
  `attack_missed` event against the same target.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=01_click_to_kill HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. The local bot script printed a post-pass
`cleanup_account.py` missing-`httpx` warning in this environment, but each scenario returned success.

## Manual Visual Command

```bash
make bot-visual scenario=77_input_buffering
```

## Deferred

- Attack-move and sticky targeting remain the next selected Movement / Combat Fluidity slice.
- Client-side windup/recovery presentation, hit stop, movement acceleration smoothing, retarget
  grace, and melee lunge remain later selected slices in this feature queue.
- No gameplay balance, server combat legality, or protocol changes shipped in v295.
