# v72 As-Built — Monster Visual Catalog

Date: 2026-06-11

## What Shipped

- Added `shared/assets/monster_visuals.v0.json` and schema as the data-driven monster
  presentation catalog for scene key, asset id, scale, height offset, and animation profile.
- Added deterministic generated GLB placeholders and Godot scenes for a quadruped predator
  and tiny flyer, while preserving the existing dummy monster visual.
- Added `dungeon_wolf` and `dungeon_bat` monster definitions with existing chase-melee
  mechanics, then included them in dungeon generation with deterministic test coverage.
- Boss templates can now define a visual model pool; the Go sim picks a deterministic boss
  visual from dummy, quadruped, or tiny flyer and emits the selected model in snapshots.
- Godot monster instancing now resolves visuals through a reusable shared-data loader instead
  of adding per-monster scene branches across gameplay code.
- Added a `showme` monster lineup focus and approved `.artifacts/showme/20260611-123159-monsters.png`
  before finalizing the slice.
- Added a compact protocol bot lab that proves the wolf and bat are present and still playable
  under unchanged authoritative combat mechanics.

## Proof

- `make validate-shared`
- `make validate-assets`
- `.venv/bin/pytest tools/assets/test_validate_assets.py -q`
- `cd server && go test ./internal/game/...`
- `make client-unit`
- `python3 skills/showme/scripts/render_focus.py --focus monsters`
- `make bot`
- `make client-smoke`
- `make ci`

## Deferred

- True flying gameplay, vertical collision, wall bypass, and flyer-specific pathing remain deferred.
- Quadruped pounce, bat dive/swarm behavior, and other new monster AI mechanics remain deferred.
- Production monster art, VFX, audio, and generalized ranged-monster equipment overlays remain deferred.
