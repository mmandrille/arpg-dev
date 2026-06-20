# v316 As Built - Monster Death Flourish

Date: 2026-06-20
Spec: [`docs/specs/v316_spec-monster-death-flourish.md`](../specs/v316_spec-monster-death-flourish.md)
Plan: [`docs/plans/v316_2026-06-20-monster-death-flourish.md`](../plans/v316_2026-06-20-monster-death-flourish.md)

## What Shipped

- `ModelReactionController.enter_death()` now adds a code-native `DeathFlourish` marker with small
  emitted shard meshes.
- Terminal reset clears the flourish so reused/reset visuals return to their baseline state.
- Existing hit reaction, death rotation, color darkening, and impact feedback counters remain intact.
- The existing animation/reaction unit test now asserts flourish creation and cleanup.

## Proof

```bash
godot --headless --path client --script res://tests/test_animation.gd
make client-unit
make maintainability
```

Result: green on 2026-06-20. Full `make ci` is deferred to the enclosing `$autoloop` batch gate.

## Manual Visual Command

```bash
make bot-visual scenario=11_combat_feedback
```

## Deferred

- Loot reveal timing, corpse persistence, imported death VFX/audio, and server/entity lifecycle
  changes remain deferred.
