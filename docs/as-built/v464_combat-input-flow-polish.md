# v464 As-Built — Combat Input Flow Polish

Date: 2026-07-14
Spec: [`docs/specs/v464_spec-combat-input-flow-polish.md`](../specs/v464_spec-combat-input-flow-polish.md)
Plan: [`docs/plans/v464_2026-07-14-combat-input-flow-polish.md`](../plans/v464_2026-07-14-combat-input-flow-polish.md)

## What shipped

- Added debug counters to the client attack buffer and sticky target helpers so replacement,
  expiry, and cleanup behavior is directly testable.
- Proved floor retarget during local recovery invokes the caller cleanup hook before queueing the
  movement command, preventing stale pending combat intent.
- Added repeated melee-lunge coverage to prove a new local melee start replaces the prior tween and
  settles back to the model base.
- Added `103_combat_input_flow_polish`, a Godot-client visual proof that clicks a combat target,
  retargets to floor movement during recovery, and observes the retargeted movement dispatch.
- Added a small camera validity guard in `main.gd` after the visual bot exposed a freed-camera
  crash path during combat scenario teardown. The maintainability baseline documents the exception;
  broader coordinator extraction remains review/refactor work.

## Proof

Focused verification:

```bash
make validate-shared
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_combat_feel_presentation_loader.gd
make client-unit
HEADLESS=1 make bot-visual scenario=77_input_buffering
HEADLESS=1 make bot-visual scenario=82_melee_lunge_micro_step
HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish
make maintainability
```

## Manual visual commands

```bash
make bot-visual scenario=82_melee_lunge_micro_step
make bot-visual scenario=103_combat_input_flow_polish
```

## Deferred

- Authoritative basic-attack cooldown rebalance.
- Shared-rules movement-speed, acceleration, or attack-speed retuning.
- Skill, ranged, boss, and production VFX/audio combat feel passes.
