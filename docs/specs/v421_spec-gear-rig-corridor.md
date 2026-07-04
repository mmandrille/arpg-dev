# v421 Spec: Gear Rig Corridor

Status: Approved
Date: 2026-07-03
Codename: gear-rig-corridor
Baseline: v420 class-skill-build-branches

## Purpose

Establish the agent-repeatable gear + skeleton corridor on **paladin**: real per-slot equipment GLBs
(replacing sword-placeholder manifest entries and runtime procedural blobs), paladin rig v2 normalization
so class scale correction is reduced, and `/showme gear` before/after proof.

## Non-goals

- No server, protocol, combat, or inventory authority changes.
- No full 68-item visual remapping (corridor + representative armor families only).
- No other hero re-rig in v421 (v422).
- No monster/companion rig (v424).
- No external AI GLB imports in v421 (gen_glb deterministic assets only).

## Acceptance criteria

- [ ] Deterministic equipment GLBs exist for head, chest, gloves, boots, belt, amulet, ring, shield
  under `client/assets/equipment/`.
- [ ] Manifest entries no longer point armor fallbacks at `rusty_sword.glb`.
- [ ] Representative `item_visuals` entries (helm, mail, boots, shield, long_sword loadout) use real assets.
- [ ] Paladin source rig applies mesh-height normalization in `rig_hero_glbs.py` (target ~1.8m) so
  `class_presentations` scale can move toward 1.0.
- [ ] `make gen-assets`, `make validate-assets`, `test_animation.gd` green for paladin sockets/clips.
- [ ] Equipment probe tests updated for GLB-mounted slots (not procedural fallback for corridor items).
- [ ] `/showme --focus gear --class-id paladin` produces improved equipped silhouette.

## Asset decision

- **Adopt:** `gen_glb.py` deterministic slot equipment meshes.
- **Borrow:** manifest + `item_visuals` + `EquipmentVisualResolver` mount path.
- **Reject:** external plugins; keeping procedural fallback only for unmapped legacy fallback IDs.

## Scope and files

| Area | Files |
|------|-------|
| Tools | `tools/assets/gen_glb.py`, `tools/assets/rig_hero_glbs.py`, tests |
| Manifest | `assets/manifests/assets.v0.json` |
| Shared | `shared/assets/item_visuals.v0.json`, `class_presentations.v0.json` |
| Client | `client/scripts/equipment_visuals.gd`, tests |
| Docs | plan, as-built, lifecycle, playbook snippet |

## Test proof

```bash
make gen-assets
make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
make client-unit
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
```
