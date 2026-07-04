# v430 Plan — Barbarian Tier-3 Mesh Swap

Spec: [`docs/specs/v430_spec-barbarian-tier3-mesh-swap.md`](../specs/v430_spec-barbarian-tier3-mesh-swap.md)

## Tasks

- [x] Archive `goliath_barbarian.glb` → `goliath_barbarian_legacy.glb`
- [x] Drop CC0 mesh at `goliath_barbarian.glb` (Poly Pizza Male Fighter by mastjie)
- [x] Set `HERO_TARGET_HEIGHTS["barbarian"]` to 1.97
- [x] `python3 tools/assets/rig_hero_glbs.py`
- [x] Update manifest provenance + `sha256`
- [x] Godot import runtime GLB
- [x] Update barbarian README
- [x] Extended client scenario `93_barbarian_tier3_visual`
- [x] Verify: validate-assets, pytest bone test, client scenario, showme
