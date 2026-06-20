# v301 Spec: Melee Lunge Micro-Step

Status: Complete
Date: 2026-06-19
Codename: `melee-lunge-micro-step`
Baseline: v300 `command-retarget-grace`

## Purpose

Give local melee swings a small forward body accent so contact feels less static, while keeping the
authoritative player anchor, movement prediction, combat reach, and server outcomes unchanged.

This is the final selected Movement / Combat Fluidity autoloop slice. It should be visual-only and
short enough not to confuse the player's actual position.

## Non-goals

- Do not change server movement, pathfinding, collision, authoritative positions, attack reach,
  damage timing, cooldowns, movement prediction, or protocol schemas.
- Do not lunge for ranged basic attacks, skills, projectiles, monsters, remote players, charge/leap
  movement, or interactable hits in this slice.
- Do not move `PlayerAnchor` or `CharacterVisual`; the micro-step must stay under the visual model
  root so it composes with v299 movement smoothing.
- Do not add assets, plugins, animation clips, camera shake, global hit stop, or VFX catalogs.

## Acceptance Criteria

- Local melee basic attacks trigger a short visual-only forward offset on the player model root and
  recover back to the exact local origin.
- Ranged basic attacks do not trigger the melee lunge.
- Bot state exposes melee-lunge debug state under local player presentation, and a client bot
  scenario proves the lunge activates during a legal melee attack and settles afterward.
- `PlayerAnchor`, `CharacterVisual`, `predicted_pos`, combat reach, and server messages remain
  unchanged by the lunge.
- v296 attack-move/sticky targeting, v299 movement smoothing, and v300 command retarget grace still
  pass.

## Scope And Likely Files

- Client helper: `client/scripts/melee_lunge_presentation.gd`.
- Client animation integration: `client/scripts/animation_controller.gd`.
- Client attack-mode helper: `client/scripts/combat_reach.gd`.
- Client call-site attack mode: `client/scripts/main.gd` and `client/scripts/bot_facade.gd`, kept
  within file-size ratchets.
- Client tests: `client/tests/test_melee_lunge_presentation.gd`.
- Bot assertion plumbing: `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_wait_handlers.gd`,
  `client/scripts/bot_assertion_handlers.gd`.
- Bot proof: new `tools/bot/scenarios/client/82_melee_lunge_micro_step.json`, plus
  `78_attack_move_sticky_targeting`, `80_movement_visual_smoothing`, and
  `81_command_retarget_grace` regressions.
- Docs: v301 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=82_melee_lunge_micro_step HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=82_melee_lunge_micro_step
```

## Asset, Plugin, And Tuning Decision

- Adopt: existing `AnimationController`, `CharacterVisual/ModelRoot` hierarchy, item
  `attack_mode`, and local player bot presentation debug state.
- Borrow: v296/v299/v300 regression scenarios.
- Reject: external plugins/assets, new animation clips, moving `PlayerAnchor` or
  `CharacterVisual`, and server-side lunge movement.
- Tuning note: lunge distance and recovery duration are client-code-owned visual feel constants,
  not gameplay balance. A future broader presentation catalog can move them to data.

## ADR Alignment

- ADR-0001 D2/D3: the server remains authoritative for movement and combat outcomes.
- ADR-0007: this is client-only animation/presentation state and does not cross the wire.

## Open Questions And Risks

- No blocking questions. The lunge targets `ModelRoot` instead of the player anchor or
  `CharacterVisual` so it remains visually additive and does not interfere with v299 smoothing.
