# v447 Plan — Class Gear Retune

**Spec:** [`docs/specs/v447_spec-class-gear-retune.md`](../specs/v447_spec-class-gear-retune.md)

Status: Complete  
Goal: Fix sword grip, shield facing, armor scale, and dual-boot visibility on post-v444 class rigs.  
Architecture: Data-driven retune in `item_visuals` / `gear_sockets`; remove double-downscale on Tier-3 armor GLBs; mirror boots to `foot_r` in resolver; tighten headless probe min-scale bands.

## Baseline and shortcut decision

Reuses v439 `class_transforms`, v442 probe/showme path, v444 class bodies. Adopt in-repo Poly Pizza GLBs with transform tuning only.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/assets/item_visuals.v0.json` | Base + per-class transforms for showme/starter items |
| Modify | `shared/assets/gear_sockets.v0.json` | `boots_right_socket`, per-class boot offsets |
| Modify | `client/scripts/equipment_visuals.gd` | Dual-boot mount + off-hand z mirror fix |
| Modify | `client/tests/equipped_gear_fit_probe.gd` | Min global scale + boots mirror assertion |
| Modify | `client/tests/item_visual_scale_probe.gd` | Shield scale ceiling for normalized GLB |
| Add | `docs/as-built/v447_class-gear-retune.md` | Proof summary |

## Task 1 — Retune item_visuals base transforms

- [x] Bump helm/mail/boots scale toward 1.0 (Tier-3 imports already normalized).
- [x] Fix `long_sword` / starter sword grip offset and rotation.
- [x] Fix `shield` / `starter_paladin_shield` rotation to vertical.
- [x] Add per-class `class_transforms` for failing class×item pairs.

## Task 2 — Gear sockets + dual boots

- [x] Add `boots_right_socket` on `foot_r` in `gear_sockets.v0.json`.
- [x] Mirror boots mount in `equipment_visuals.gd` (left primary, right mirrored).

## Task 3 — Probe tightening

- [x] Add `MIN_GLOBAL_SCALE` bands for head/chest/boots in `equipped_gear_fit_probe.gd`.
- [x] Assert boots mirror node exists on both feet.

## Task 4 — Visual spot-check

- [x] `render_focus.py --focus gear --class-id paladin`
- [x] `render_focus.py --focus gear --class-id rogue`

## Task 5 — Lifecycle docs

- [x] `docs/as-built/v447_class-gear-retune.md`
- [x] `PROGRESS.md` + slice lifecycle row

## Final verification

- [x] `tools/validate_shared.py`
- [x] `godot --headless --path client --script res://tests/test_item_visuals.gd`
- [ ] `make ci` (autoloop batch gate)
