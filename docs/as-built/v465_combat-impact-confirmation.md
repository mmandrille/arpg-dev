# v465 As-Built — Combat Impact Confirmation

Date: 2026-07-14
Spec: [`docs/specs/v465_spec-combat-impact-confirmation.md`](../specs/v465_spec-combat-impact-confirmation.md)
Plan: [`docs/plans/v465_2026-07-14-combat-impact-confirmation.md`](../plans/v465_2026-07-14-combat-impact-confirmation.md)

## What shipped

- Extended the existing procedural combat outcome punch so normal `monster_damaged` hits and
  `monster_killed` events now produce bounded target-side confirmation.
- Kept misses, blocks, immunes, and critical hits on distinct outcome mappings, with critical still
  taking priority when event metadata marks the hit critical.
- Preserved the existing `enemy_impact_feedback` gate, so disabling enemy impact feedback also
  suppresses the new normal-hit confirmation.
- Added `104_combat_impact_confirmation`, a Godot-client visual proof that equips the combat lab
  training bow, observes authoritative hit feedback, and continues through an authoritative kill.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_combat_outcome_punch.gd
godot --headless --path client --script res://tests/test_impact_sparks.gd
HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation
```

## Manual visual command

```bash
make bot-visual scenario=104_combat_impact_confirmation
```

## Deferred

- Deterministic client bot proof for critical hits, misses, and blocks.
- Production VFX/audio assets and per-skill/per-class bespoke impact presentation.
- Combat math, cooldown, attack speed, movement speed, or damage rebalance.
