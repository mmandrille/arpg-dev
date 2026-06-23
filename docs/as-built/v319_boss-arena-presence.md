# v319 As Built - Boss Arena Presence

Date: 2026-06-23
Spec: [`docs/specs/v319_spec-boss-arena-presence.md`](../specs/v319_spec-boss-arena-presence.md)
Plan: [`docs/plans/v319_2026-06-23-boss-arena-presence.md`](../plans/v319_2026-06-23-boss-arena-presence.md)

## What Shipped

- Added `BossArenaPresence` ground torus rings under live boss monsters.
- Rings shift to telegraph tint during active telegraph phases and use a default boss rim otherwise.
- `BossVisualsController` syncs arena presence alongside boss health bar and phase updates.
- Factory boss presentation test proves ring creation, telegraph coexistence, and cleanup.

## Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
make maintainability
```

Result: green on 2026-06-23. Full `make ci` is deferred to the enclosing `$autoloop` batch gate.

## Manual Visual Command

```bash
make bot-visual scenario=boss_telegraph_presentation
```

## Deferred

- Boss portrait art, imported arena VFX/audio, and server pattern changes remain deferred.
