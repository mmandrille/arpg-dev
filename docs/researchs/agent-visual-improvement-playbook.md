# Agent Visual Improvement Playbook

Date: 2026-07-03  
Baseline: v421–v425 gear/rig/asset corridor

## Three tiers (AI-agent friendly)

| Tier | What | Tools | When |
|------|------|-------|------|
| 1 | JSON presentation catalogs | `shared/assets/*.v0.json` | UI icons, tints, fog, town — no new meshes |
| 2 | Deterministic GLB generation | `tools/assets/gen_glb.py`, `make gen-assets` | Equipment, weapons, placeholders — reproducible CI |
| 3 | External / AI GLB import | `$3dmodel`, manifest + provenance | Hero/monster body replacement when gen_glb is not enough |

## Equipped gear loop (Tier 2)

1. Add generator in `tools/assets/gen_glb.py` → `make gen-assets`
2. Register in `assets/manifests/assets.v0.json` (update `sha256`)
3. Map `shared/assets/item_visuals.v0.json` (`asset_id`, `mount_socket`, `local_transform`)
4. `godot --headless --path client --import` for new `.glb.import` sidecars
5. Verify: `make validate-assets`, `test_item_visuals.gd`
6. Visual: `python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin`

## Skeleton / rig loop (Tier 2–3)

1. **Heroes:** `tools/assets/rig_hero_glbs.py` + `HERO_TARGET_HEIGHTS` (~1.85m)
2. **Biped monsters:** `tools/assets/rig_monster_glbs.py` + `MONSTER_TARGET_HEIGHTS`
3. **Quadrupeds:** `tools/assets/rig_quadruped_monster_glbs.py`
4. Verify: `godot --headless --path client --script res://tests/test_animation.gd`

## AI body mesh replacement (Tier 3)

When gen_glb placeholders are not enough:

1. Drop source GLB under `assets/characters/<class>/` or `assets/monsters/<name>/`
2. Probe: `python3 skills/3dmodel/scripts/create_model_probe.py --model <path> --key <id>`
3. Copy runtime bytes to `client/assets/...`, register manifest + provenance (`license`, `sha256`)
4. Run appropriate rig tool (`rig_hero_glbs.py` / `rig_monster_glbs.py`)
5. Record adopt/borrow/reject in slice spec/plan
6. `make validate-assets` + class/monster animation smoke

**Reject:** runtime network fetch, unmanifested GLBs, skipping provenance.

## v425 scope note

This slice documents the workflow; full AI body replacement for every placeholder is deferred to future slices using Tier 3 above.
