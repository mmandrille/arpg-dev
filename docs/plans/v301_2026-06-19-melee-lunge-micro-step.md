# v301 Plan — Melee Lunge Micro-Step

Status: Complete
Goal: Add a short visual-only model-root forward offset for local melee basic attacks.
Architecture: Add a focused helper that offsets the local player's `ModelRoot` during melee attack
presentation and tweens it back to its baseline local position. `AnimationController` triggers the
helper only when a caller identifies the attack mode as melee. `PlayerAnchor`, `CharacterVisual`,
prediction, reach, server messages, and combat outcomes remain unchanged.
Tech stack: Godot client presentation, shared item data already loaded through `ItemRulesLoader`,
Godot unit tests, Godot client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v296 attack-move/sticky targeting, v299 movement visual smoothing, and v300 command
retarget grace. Asset/plugin decision: adopt existing `AnimationController`, `CharacterVisual` /
`ModelRoot`, item `attack_mode`, and bot presentation debug patterns; borrow v296/v299/v300
regression scenarios; reject external plugins/assets, new animation clips, moving gameplay anchors,
and server-side lunge movement.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/melee_lunge_presentation.gd` | Apply melee-only model-root offset, recovery tween, replacement safety, and debug state. |
| Modify | `client/scripts/animation_controller.gd` | Trigger lunge for melee attack one-shots and expose debug state. |
| Modify | `client/scripts/combat_reach.gd` | Expose local player attack mode from equipped item data. |
| Modify | `client/scripts/main.gd` | Pass local attack mode to attack animation/presentation call sites without growing the coordinator. |
| Modify | `client/scripts/bot_facade.gd` | Pass local attack mode for direct bot monster clicks. |
| Create | `client/tests/test_melee_lunge_presentation.gd` | Cover melee lunge, recovery, and ranged no-lunge behavior. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register melee-lunge wait/assert step names. |
| Modify | `client/scripts/bot_wait_handlers.gd` | Delegate melee-lunge wait matching. |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert local player melee-lunge debug state. |
| Create | `tools/bot/scenarios/client/82_melee_lunge_micro_step.json` | Live client proof that legal melee attack lunge activates and settles. |
| Modify | `docs/specs/v301_spec-melee-lunge-micro-step.md` | Mark complete when shipped. |
| Create | `docs/as-built/v301_melee-lunge-micro-step.md` | Summarize shipped behavior and limits. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v301 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and batch handoff state. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines, and touched grandfathered files stay
inside allowance.

Hotspot files:
- [x] `client/scripts/main.gd` (grandfathered; stayed at 5813 lines)
- [x] Other touched files are under 600 lines.

Decision:
- [x] Put lunge behavior in a new focused helper.
- [x] Keep `main.gd` to same-line argument additions only, with no line-count growth.

Verification:

```bash
make maintainability
```

## Task 1 — Lunge Helper And Animation Integration

- [x] Step 1.1: Add helper for model-root lunge and recovery debug state.
- [x] Step 1.2: Trigger helper from `AnimationController` only for melee attack clips.
- [x] Step 1.3: Expose item-derived local attack mode from `CombatReach`.
- [x] Step 1.4: Pass attack mode from local attack call sites and bot facade.
- [x] Step 1.5: Add focused Godot unit tests.

```bash
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_animation.gd
```

## Task 2 — Bot Assertion And Scenario

- [x] Step 2.1: Add local melee-lunge wait/assert matching.
- [x] Step 2.2: Add `82_melee_lunge_micro_step`.
- [x] Step 2.3: Verify client bot unit coverage and live scenario.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=82_melee_lunge_micro_step HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 3 — Focused Regression Proof

- [x] Step 3.1: Rerun attack-move, movement-smoothing, and command-retarget scenarios.
- [x] Step 3.2: Run maintainability.

```bash
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification:

```bash
make bot-visual scenario=82_melee_lunge_micro_step
```

## Task 4 — Lifecycle Docs

- [x] Step 4.1: Mark spec and plan complete after focused checks pass.
- [x] Step 4.2: Add as-built and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md` for completed selected feature queue and final batch CI gate.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd`
- [x] `godot --headless --path client --script res://tests/test_animation.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=82_melee_lunge_micro_step HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18083 BASE_URL=http://localhost:18083 SCENARIO=80_movement_visual_smoothing HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `GODOT=godot ARPG_ADDR=:18084 BASE_URL=http://localhost:18084 SCENARIO=81_command_retarget_grace HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`

Autoloop batch mode: do not run `make ci` for this individual slice unless focused verification is
insufficient. The enclosing autoloop must run one final `make ci` after this selected feature queue
is committed.
