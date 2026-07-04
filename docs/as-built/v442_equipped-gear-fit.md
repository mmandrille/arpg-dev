# v442 As-built — Equipped Gear Fit

## Proved

- Poly Pizza equipment import strips `scale:100` glTF node transforms (`_flatten_node_scales`).
- `equipment_display.v0.json` owns global equipped/ground GLB multipliers.
- `character_visual.gd` refreshes bone sockets after class model swap; skips invalid bones instead of leaving origin attachments.
- Showme gear focus matches `main.gd` remount path and plays `idle` so `BoneAttachment3D` poses resolve before capture.
- Headless `equipped_gear_fit_probe` covers all five classes × starter + helm/mail/boots (+ paladin shield).
- `validate-assets` rejects poly.pizza equipment GLBs with non-identity node scales.

## Limits

- Matrix uses representative items per slot, not all 67 `3d_model` item defs.
- Fit probe asserts bone binding + rest-pose bands, not pixel-perfect art direction.
- Idle animation required for correct socket pose in headless/showme snapshots.
