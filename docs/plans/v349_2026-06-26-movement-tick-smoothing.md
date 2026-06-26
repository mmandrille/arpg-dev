# v349 Plan — Movement Tick Smoothing

Status: Ready for implementation
Goal: Interpolate authoritative entity positions between 10 Hz snapshots with data-driven tuning.
Architecture: `EntityTickSmoothing` lerps display position over `snapshot_interval_seconds` per entity. `EntityTickSmoothingRuntime` wires player + entity map with minimal `main.gd` touch. v299 visual offset remains orthogonal.
Tech stack: shared JSON, Godot client, client bot.

## Baseline and shortcut decision

Builds on v348. v299 anchor-offset smoothing stays; tick smoothing applies to anchor/node world positions on `entity_update`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/movement_presentation.v0.json` | Tick smoothing tuning |
| Create | `shared/assets/movement_presentation.v0.schema.json` | Schema |
| Create | `client/scripts/movement_presentation_loader.gd` | Loader singleton |
| Create | `client/scripts/entity_tick_smoothing.gd` | Per-entity lerp state |
| Create | `client/scripts/entity_tick_smoothing_runtime.gd` | Player + entity map coordinator |
| Modify | `client/scripts/main.gd` | Wire runtime (minimal) |
| Modify | `client/scripts/bot_*` | Wait/assert handlers |
| Create | `tools/bot/scenarios/client/84_entity_tick_smoothing.json` | Bot proof |
| Create | `client/tests/test_entity_tick_smoothing.gd` | Unit tests |
| Modify | `client/scripts/client_smoke.sh` | Register unit gate |

## Maintenance ratchet

- [ ] `main.gd` stays within baseline (+25 max) — delegate to runtime helper
- [ ] New files under 600 lines

## Task 1 — Shared tuning

- [x] Step 1.1: Add movement_presentation JSON + schema
```bash
make validate-shared
```

## Task 2 — Client helpers + integration

- [x] Step 2.1: Loader + smoothing classes + runtime
- [x] Step 2.2: Wire `main.gd` apply + process tick
```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_entity_tick_smoothing.gd
```

## Task 3 — Bot proof

- [x] Step 3.1: Scenario 84 + handlers
```bash
HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
```

## Task 4 — Lifecycle docs

- [ ] Update PROGRESS, lifecycle, as-built

## Final verification

- [x] `make validate-shared`
- [x] `make maintainability`
- [x] `make client-unit`
