# v429 Plan — Paladin AI Mesh Swap

Spec: [`docs/specs/v429_spec-paladin-ai-mesh-swap.md`](../specs/v429_spec-paladin-ai-mesh-swap.md)

## Tasks

- [x] Archive `knight.glb` → `knight_legacy.glb`
- [x] Drop CC0 evaluation mesh at `knight.glb` (Poly Pizza Warrior by mastjie)
- [x] `python3 tools/assets/rig_hero_glbs.py`
- [x] Update manifest provenance + `sha256`
- [x] Godot import runtime GLB
- [x] Update paladin README (swap + AI handoff notes)
- [x] Verify: validate-assets, pytest bone test, client scenario, showme

## AI handoff (future)

Replace `assets/characters/paladin/knight.glb` with a Meshy/Tripo static export (no embedded skin/animations), then re-run steps 3–6. Keep `knight_legacy.glb` or rename prior swap for A/B comparison.
