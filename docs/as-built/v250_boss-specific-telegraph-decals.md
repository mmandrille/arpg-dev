# v250 As-Built - Boss-Specific Telegraph Decals

Date: 2026-06-17

## What shipped

- Replaced the generic boss telegraph marker mesh with shape-specific code-native decals.
- Added marker shapes for line, cone, summon-circle, melee-contact, and generic circle telegraphs.
- Kept decal radius and color driven by existing server telegraph metadata.
- Preserved boss tinting, phase bar behavior, and marker cleanup.
- Exposed `telegraph_marker_shape` in entity presentation debug state and bot entity reaction
  assertions.
- Added `66_boss_telegraph_decals.json`, which observes Cave Warden line, summon-circle, and cone
  marker shapes using the existing boss pattern sequence.

## Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=66_boss_telegraph_decals.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` remains deferred until the feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=66_boss_telegraph_decals.json
```

## Scope limits

- No server/protocol changes, boss balance changes, timing changes, new boss patterns, imported VFX
  art, audio, particles, or exact line/cone aim projection shipped.
