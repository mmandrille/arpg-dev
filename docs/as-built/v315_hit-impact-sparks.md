# v315 As Built - Hit Impact Sparks

Date: 2026-06-20
Spec: [`docs/specs/v315_spec-hit-impact-sparks.md`](../specs/v315_spec-hit-impact-sparks.md)
Plan: [`docs/plans/v315_2026-06-20-hit-impact-sparks.md`](../plans/v315_2026-06-20-hit-impact-sparks.md)

## What Shipped

- Added `ImpactSparks`, a small code-native mesh/material helper for combat hit sparks.
- Existing monster/player damage and kill presentation paths now spawn impact sparks from the same
  authoritative combat events that drive floating combat text.
- Spark colors reuse existing damage-type combat text presentation when present, falling back to the
  current event color.
- Miss, block, and immune outcomes remain combat-text-only.
- Kept `main.gd` at its existing ratchet cap by moving VFX construction into the focused helper and
  tightening nearby routing code.

## Proof

```bash
godot --headless --path client --script res://tests/test_impact_sparks.gd
make client-unit
make maintainability
```

Result: green on 2026-06-20. Full `make ci` is deferred to the enclosing `$autoloop` batch gate.

## Manual Visual Command

```bash
make bot-visual scenario=11_combat_feedback
```

## Deferred

- Camera shake, projectile trails, kill flourish, imported VFX assets, and combat/audio changes
  remain deferred to later selected look-and-feel slices.
