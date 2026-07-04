# v442 Plan — Equipped Gear Fit Matrix

**Spec:** [`docs/specs/v442_spec-equipped-gear-fit.md`](../specs/v442_spec-equipped-gear-fit.md)

## Tasks

- [x] 1. Land import hardening (`_flatten_node_scales`) + `equipment_display` contract (in-flight dirty changes)
- [x] 2. Fix `character_visual.gd` socket refresh: remove failed BoneAttachments, `refresh_gear_sockets()` after model swap
- [x] 3. Fix `visual_capture.gd` gear setup parity with `main.gd` (`set_character_class`, await, remount, idle pose)
- [x] 4. Add `equipped_gear_fit_probe.gd` + wire into `test_item_visuals.gd` and `client_smoke.sh`
- [x] 5. Add equipment GLB node-scale check to `validate_assets.py`
- [x] 6. Tune `item_visuals.v0.json` / `gear_sockets.v0.json` for remaining mis-mounts found by probe
- [x] 7. Update `docs/CODEMAP.md` if new files; run verification

## Verification

```bash
python3 tools/assets/import_equipment_glb.py
godot --headless --import --path client
make validate-shared validate-assets
godot --headless --path client --script res://tests/test_item_visuals.gd
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin --items helm,mail,boots,long_sword,shield
```

Autoloop batch: per-slice focused verification above; final `make ci` at batch end.
