# v315 Spec - Hit Impact Sparks

Status: Complete
Date: 2026-06-20
Codename: hit-impact-sparks

## Purpose

Improve moment-to-moment combat feel by adding small impact spark meshes when existing authoritative
damage events reach the client.

## Non-goals

- No combat math, hit validation, damage type, protocol, or server changes.
- No particle plugins, imported VFX packs, or external assets.
- No camera shake, monster death flourish, projectile trails, or audio changes.

## Acceptance Criteria

- Monster/player damage and kill events can spawn a code-native impact spark node at the target.
- Spark color follows existing damage-type presentation when available and falls back to the current
  combat-text color.
- Miss/block/immune text priority remains unchanged and does not require a spark.
- The helper is covered by a focused Godot unit test registered in `make client-unit`.

## Scope and Files

- Create `client/scripts/impact_sparks.gd`.
- Modify `client/scripts/main.gd` at existing combat-text event paths.
- Create `client/tests/test_impact_sparks.gd`.
- Register the test in `scripts/client_smoke.sh`.
- Add lifecycle/as-built docs when the slice ships.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_impact_sparks.gd
godot --headless --path client --script res://tests/test_rogue_presentation.gd
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=11_combat_feedback
```

## Open Questions and Risks

- None. Asset/plugin decision: adopt existing code-native mesh/material VFX patterns; reject
  external VFX assets/plugins for this small readability pass.
