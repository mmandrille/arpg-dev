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

## Gear transform tuning notes

Use these rules when adjusting equipped item size and placement:

- `shared/assets/item_visuals.v0.json` controls the visible item's `asset_id`, slot, mount socket, and item-local transform.
- `shared/assets/gear_sockets.v0.json` controls where sockets attach to skeleton bones. Change sockets when the whole class/socket is wrong; change item visuals when one item family is wrong.
- Prefer base `local_transform` when a fit should apply to every class. Use `class_transforms.<class_id>.local_transform` only for later class-specific exceptions.
- `scale` can be non-uniform. This is useful for fallback gear: a helmet can be wider/taller/deeper with different `x`, `y`, and `z` values instead of one uniform number.
- The shared head-family fit uses `position { x: 0, y: 0.05, z: -0.11 }`, `scale { x: 1.8, y: 1.3, z: 1.9 }`.
- The shared chest-family fit uses `position { x: 0.0, y: 0.08, z: -0.12 }`, `scale { x: 2.4, y: 2.4, z: 2.4 }`.
- The shared boots-family fit uses `position { x: 0.0, y: 0.0, z: -0.1 }`, `rotation_degrees { x: 0.0, y: 270.0, z: 0.0 }`, `scale { x: 1.45, y: 1.45, z: 1.45 }`.
- In the focused paladin gear view, larger positive/negative changes are easiest to judge through fast screenshots rather than reasoning from axes alone.

Observed axis behavior from the paladin helmet/mail tuning loop:

- Helmet `y` is vertical relative to the head socket in practice: increasing it made the helmet float upward; decreasing it moved the helmet down over the face.
- Helmet `z` affected front/back fit in the render; negative `z` moved the helmet back toward the hair/head mass.
- Mail `z` affected front/back fit; negative `z` moved the armor back into the paladin torso instead of protruding forward.
- Mail `y` affected vertical placement; reducing an overly high value moved the armor lower on the chest.
- Mail `x` recenters left/right. Keep it at `0.0` unless the piece is visibly shifted off the character centerline.

Recommended iteration loop:

1. Edit only one item entry or one class override at a time.
2. Render the exact class and gear set:

```bash
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
```

3. Inspect the PNG under `.artifacts/showme/`.
4. Validate structure and runtime resolution:

```bash
python3 -m json.tool shared/assets/item_visuals.v0.json
godot --headless --path client --script res://tests/test_item_visuals.gd
```

5. If the change affects many classes or base gear, run the relevant batch suite:

```bash
make regen-screenshots SUITE=gear
```

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
