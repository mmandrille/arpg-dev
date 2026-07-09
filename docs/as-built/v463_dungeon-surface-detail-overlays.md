# v463 As-Built — Dungeon Surface Detail Overlays

Date: 2026-07-09
Spec: [`docs/specs/v463_spec-dungeon-surface-detail-overlays.md`](../specs/v463_spec-dungeon-surface-detail-overlays.md)
Plan: [`docs/plans/v463_2026-07-09-dungeon-surface-detail-overlays.md`](../plans/v463_2026-07-09-dungeon-surface-detail-overlays.md)

## What shipped

- Added `shared/assets/dungeon_surface_detail_presentation.v0.json` plus schema as the data-driven
  client presentation config for floor signs and wall artifacts.
- Added `client/scripts/dungeon_surface_detail_loader.gd` to load that config with safe defaults.
- Added `client/scripts/dungeon_surface_detail_presentation.gd` as the focused helper that:
  - segments dungeon floors into room-like cells using existing divider walls
  - places mixed-scale floor motifs (`ritual_square`, `scratched_cross`, `broken_tile`)
  - attaches modest wall artifacts (`plaque`, `crack_cluster`, `stone_inset`) to eligible wall runs
  - clears all detail roots outside dungeon levels
- Wired the helper into `main.gd` beside the existing room-tint sync so the scene coordinator stays
  thin and the feature remains client-only.
- Extended `client/tests/test_factories.gd` to prove dungeon-only attachment, mixed-size floor
  detail generation, and wall artifact attachment.
- Added `tools/bot/scenarios/client/101_dungeon_surface_detail_overlays.json` as the dedicated
  visual proof route for this slice.
- Updated `docs/CODEMAP.md` so the new presentation path is discoverable and passes codemap
  validation.

## Proof

Verified in this environment:

```bash
make validate-shared
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout
```

Dedicated v463 scenario created for normal local verification:

```bash
HEADLESS=1 make bot-client scenario=dungeon_surface_detail_overlays
make bot-visual scenario=dungeon_surface_detail_overlays
```

## Environment note

- `HEADLESS=1 make bot-client scenario=dungeon_surface_detail_overlays` could not be executed in
  this session because the current sandbox blocks access to the local Docker daemon used by the
  repo's Postgres startup path. The scenario file is present and intended as the exact follow-up
  command outside the sandbox.

## Deferred

- No external CC0 decals or texture packs were adopted in this first pass.
- No town-surface detail pass shipped.
- No lighting/fog overhaul shipped with the surface-detail work.
- No server-authored placement metadata or gameplay-visible interactables were added.
