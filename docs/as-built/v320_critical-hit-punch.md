# v320 As Built - Critical Hit Punch

Date: 2026-06-23
Spec: [`docs/specs/v320_spec-critical-hit-punch.md`](../specs/v320_spec-critical-hit-punch.md)

## What Shipped

- Added `CombatOutcomePunch` rings/sparks for miss, block, immune, and crit outcomes.
- Routed combat text + sparks + outcome punch through `CombatEventPresentation`.
- Extracted presentation logic from `main.gd` to satisfy the maintainability ratchet.

## Proof

```bash
godot --headless --path client --script res://tests/test_combat_outcome_punch.gd
make client-unit
make maintainability
```

Result: green on 2026-06-23.

## Manual Visual Command

```bash
make bot-visual scenario=11_combat_feedback
```
