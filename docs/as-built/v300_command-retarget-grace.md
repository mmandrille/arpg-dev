# v300 As-Built: Command Retarget Grace

Date: 2026-06-19
Spec: [`docs/specs/v300_spec-command-retarget-grace.md`](../specs/v300_spec-command-retarget-grace.md)
Plan: [`docs/plans/v300_2026-06-19-command-retarget-grace.md`](../plans/v300_2026-06-19-command-retarget-grace.md)

## What Shipped

- Added `CommandRetargetGrace`, a focused client helper that stores one latest pending floor
  retarget while the local click/send throttle is cooling down.
- Rapid floor clicks replace the queued destination instead of accumulating a command list, so the
  latest click wins and stale retargets expire quickly if a longer cooldown remains.
- Fresh floor clicks during longer attack recovery dispatch immediately through the same helper
  instead of leaving a stale retarget queued, while preserving the existing attack recovery timer.
- Wired local floor-click dispatch and bot `click_floor` dispatch through the same retarget-aware
  path while leaving server movement, protocol, authoritative positions, pathfinding, combat reach,
  sticky targeting, and attack buffering unchanged.
- Exposed `command_retarget_grace` debug state from bot state and added
  `wait_command_retarget_grace` / `assert_command_retarget_grace` matching.
- Added `81_command_retarget_grace`, which sends rapid floor clicks and proves the retarget helper
  dispatches the final clicked destination.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18083 BASE_URL=http://localhost:18083 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. The local bot script printed the known post-pass
`cleanup_account.py` missing-`httpx` warning in this environment, but every scenario returned
success.

Post-loop stabilization proof on 2026-06-20 also covered the long-recovery floor-click path:

```bash
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18087 BASE_URL=http://localhost:18087 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
```

## Manual Visual Command

```bash
make bot-visual scenario=81_command_retarget_grace
```

## Deferred

- Melee lunge / micro-step remains the final selected Movement / Combat Fluidity slice.
- Retarget grace for skills, interactables, WASD movement, and preserving already-queued retargets
  through long combat recoveries remains out of scope.
