# v320 Spec - Critical Hit Punch

Status: Complete
Date: 2026-06-23
Codename: critical-hit-punch

## Purpose

Give crits, blocks, misses, and immunes stronger moment-to-moment visual distinction using existing
authoritative combat events only.

## Non-goals

- No combat formula, outcome roll, protocol, or server changes.
- No new floating combat text settings or audio overhaul.

## Acceptance Criteria

- Miss, block, and immune outcomes spawn distinct client-only outcome punch markers.
- Critical hits spawn an enhanced outcome punch in addition to existing impact sparks.
- Combat text + FX routing is extracted from `main.gd` into a focused presentation module.
- Headless unit test covers spawn rules, node shape, and delta integration.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_combat_outcome_punch.gd
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=11_combat_feedback
```
