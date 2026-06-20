# v301 As-Built: Melee Lunge Micro-Step

Date: 2026-06-19
Spec: [`docs/specs/v301_spec-melee-lunge-micro-step.md`](../specs/v301_spec-melee-lunge-micro-step.md)
Plan: [`docs/plans/v301_2026-06-19-melee-lunge-micro-step.md`](../plans/v301_2026-06-19-melee-lunge-micro-step.md)

## What Shipped

- Added `MeleeLungePresentation`, a focused client helper that offsets the local player's
  `ModelRoot` forward during melee attack presentation and tweens it back to its baseline local
  position.
- Extended `AnimationController` to trigger the helper only when the attack animation caller marks
  the attack mode as `melee`, and to expose `melee_lunge` debug state.
- Added `CombatReach.local_player_attack_mode()` so local basic-attack presentation uses equipped
  item `attack_mode` data. Ranged weapons such as bows do not trigger the lunge.
- Kept `PlayerAnchor`, `CharacterVisual`, `predicted_pos`, combat reach, server messages, damage,
  cooldowns, and pathing unchanged. The lunge composes with v299 smoothing because it moves only
  `ModelRoot`, not the gameplay anchor or smoothed visual parent.
- Added `82_melee_lunge_micro_step`, which attack-moves into melee range, waits for the lunge to
  activate, observes an authoritative combat event, and waits for the lunge to settle.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=82_melee_lunge_micro_step HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18083 BASE_URL=http://localhost:18083 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18084 BASE_URL=http://localhost:18084 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. `test_animation.gd` still prints the existing Godot
ObjectDB/resource warnings after its PASS line but exits 0. The local bot script printed the known
post-pass `cleanup_account.py` missing-`httpx` warning in this environment, but every scenario
returned success.

## Manual Visual Command

```bash
make bot-visual scenario=82_melee_lunge_micro_step
```

## Deferred

- The selected Movement / Combat Fluidity feature queue is complete.
- Lunge variants for ranged attacks, skills, monsters, remote players, leap/charge motion, and
  production VFX/audio catalogs remain out of scope.
