# v319 Spec - Boss Arena Presence

Status: Draft
Date: 2026-06-23
Codename: boss-arena-presence

## Purpose

Make boss fights feel more dramatic with a client-only ground arena tint/rim around the active boss,
including phase-aware color shifts during telegraphs, without obscuring existing telegraph decals.

## Non-goals

- No boss AI, pattern, damage, protocol, or server changes.
- No imported VFX assets, shader packages, or external plugins.
- No boss health bar art overhaul.

## Acceptance Criteria

- Live boss monsters gain a `BossArenaPresence` ground ring child at their feet.
- The ring uses telegraph color during an active telegraph phase and a default boss rim otherwise.
- Rings are removed when the boss dies or is no longer the active boss presentation target.
- Existing `BossTelegraphMarker` decals remain visible and unchanged in behavior.
- Headless unit test asserts ring creation, phase tinting, and cleanup.

## Scope and Files

- Create `client/scripts/boss_arena_presence.gd`.
- Modify `client/scripts/boss_visuals_controller.gd`.
- Extend `client/tests/test_factories.gd` boss presentation assertions.
- Add lifecycle/as-built docs when the slice ships.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=boss_telegraph_presentation
```

## Open Questions and Risks

- None. Asset/plugin decision: adopt code-native mesh/material rings; reject external assets/plugins.
