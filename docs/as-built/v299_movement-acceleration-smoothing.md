# v299 As-Built: Movement Acceleration Smoothing

Date: 2026-06-19
Spec: [`docs/specs/v299_spec-movement-acceleration-smoothing.md`](../specs/v299_spec-movement-acceleration-smoothing.md)
Plan: [`docs/plans/v299_2026-06-19-movement-acceleration-smoothing.md`](../plans/v299_2026-06-19-movement-acceleration-smoothing.md)

## What Shipped

- Added `MovementVisualSmoothing`, a focused client helper that preserves visual continuity for
  small local player anchor movement steps by offsetting `CharacterVisual` opposite the anchor
  change, then easing the visual child back to local zero.
- Wired the helper only to local player presentation paths. `player_anchor.position`,
  `predicted_pos`, server messages, combat reach, picking, and camera targeting remain exact
  gameplay positions.
- Reset visual offsets for large spawn/teleport/correction-sized deltas so the player does not see
  misleading lag during discontinuous movement.
- Exposed `movement_visual_smoothing` debug state from bot state and added
  `wait_movement_visual_smoothing` / `assert_movement_visual_smoothing` bot matching.
- Added `80_movement_visual_smoothing`, proving the offset activates after a floor move and settles
  afterward.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=05_click_to_move HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18083 BASE_URL=http://localhost:18083 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. The local bot script printed the known post-pass
`cleanup_account.py` missing-`httpx` warning in this environment, but every scenario returned
success.

## Manual Visual Command

```bash
make bot-visual scenario=80_movement_visual_smoothing
```

## Deferred

- Command retarget grace and melee lunge remain later selected slices in this feature queue.
- Smoothing for monsters, remote players, projectiles, leap/charge motion, and broader movement
  presentation data catalogs remain out of scope.
