# v316 Spec - Monster Death Flourish

Status: Complete
Date: 2026-06-20
Codename: monster-death-flourish

## Purpose

Make monster/player death reactions read more clearly by adding a small code-native death flourish
to the existing terminal model reaction path.

## Non-goals

- No death rules, loot timing, entity cleanup, protocol, server, or replay changes.
- No imported VFX/audio assets or particle plugins.
- No loot reveal timing changes.

## Acceptance Criteria

- `ModelReactionController.enter_death()` adds a `DeathFlourish` visual marker to the reaction root.
- The flourish is asserted in the existing headless animation/model-reaction test.
- Hit reactions and terminal reset behavior remain unchanged.

## Scope and Files

- Modify `client/scripts/model_reaction_controller.gd`.
- Modify `client/tests/test_animation.gd`.
- Add lifecycle/as-built docs when the slice ships.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_animation.gd
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=11_combat_feedback
```

## Open Questions and Risks

- None. Asset/plugin decision: adopt existing code-native reaction mesh/materials; reject external
  assets/plugins for this terminal reaction polish.
