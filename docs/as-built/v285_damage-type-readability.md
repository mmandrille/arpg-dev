# v285 As Built: Damage Type Readability

Date: 2026-06-19
Spec: [`docs/specs/v285_spec-damage-type-readability.md`](../specs/v285_spec-damage-type-readability.md)
Plan: [`docs/plans/v285_2026-06-19-damage-type-readability.md`](../plans/v285_2026-06-19-damage-type-readability.md)

## What shipped

- Floating combat text now derives known hit presentation from authoritative combat event
  `damage_type` values.
- Added readable type labels and distinct colors for physical, fire, cold, lightning, poison, and
  force damage.
- Kept `MISS`, `BLOCK`, and `IMMUNE` priority unchanged.
- Critical hits keep the critical variant while carrying the typed label when damage type metadata is
  present.
- Added `client/scripts/damage_type_combat_text.gd` so the large `main.gd` coordinator stays within
  the maintainability ratchet.
- Bot damage-number debug state now exposes `damage_type`, and damage-number scenario steps can
  match it.
- The existing fire unique client scenario now asserts a live fire damage number.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_rogue_presentation.gd
CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh
make bot-visual scenario=33_unique_burn_effect_live HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: deferred until the end of the selected autoloop queue.

## Manual visual command

```bash
make bot-visual scenario=33_unique_burn_effect_live
```

## Deferred

- New damage types, damage math, resistances, skills, balance tuning, and protocol changes remain
  deferred.
- Production art/VFX/audio for each damage type remains deferred.
