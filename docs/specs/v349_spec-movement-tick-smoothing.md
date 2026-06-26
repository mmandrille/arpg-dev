# v349 Spec: Movement Tick Smoothing

Status: Complete  
Date: 2026-06-26  
Codename: `movement-tick-smoothing`  
Baseline: v348 `forward-plus-renderer`

## Purpose

Interpolate authoritative entity positions between 10 Hz `entity_update` snapshots so players and
visible combat entities glide at render rate instead of stepping every sim tick.

Distinct from v299 anchor-offset smoothing: v299 preserves local `CharacterVisual` continuity for
small anchor reconciliation steps; v349 time-lerps full world positions between server snapshots
for the local player anchor and remote monsters/players/companions.

## Non-goals

- No server, protocol, shared golden, or sim tick-rate changes.
- No changes to movement authority, prediction math, pathfinding, or collision.
- No projectile smoothing (projectiles keep existing presentation path).
- No interactable/loot placement smoothing.
- No leap/charge/teleport smoothing (large corrections snap).

## Acceptance criteria

- Data-driven tuning in `shared/assets/movement_presentation.v0.json` (`snapshot_interval_seconds`,
  `snap_distance`, `enabled`).
- Local player anchor and visible `monster`/`player`/`companion` nodes interpolate between
  consecutive authoritative positions over ~one tick interval (default 0.1s).
- Snap threshold bypasses interpolation for teleports/large corrections.
- Bot debug state exposes tick-smoothing activity; client bot scenario proves active-then-settled
  during a floor move.
- v299 `80_movement_visual_smoothing` regression still passes.

## Scope and likely files

| Area | Files |
|------|-------|
| Shared tuning | `shared/assets/movement_presentation.v0.json`, schema |
| Client helper | `entity_tick_smoothing.gd`, `entity_tick_smoothing_runtime.gd`, loader |
| Integration | `client/scripts/main.gd` (minimal wiring) |
| Tests | `client/tests/test_entity_tick_smoothing.gd` |
| Bot | `84_entity_tick_smoothing.json`, assertion/wait handlers |
| Docs | as-built, lifecycle, `PROGRESS.md` |

## Test and bot proof

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
make maintainability
```

## Asset decision

- Adopt: existing entity map + bot debug patterns.
- Borrow: v299 movement visual smoothing (orthogonal; both may be active).
- Reject: external plugins; server-side interpolation.
