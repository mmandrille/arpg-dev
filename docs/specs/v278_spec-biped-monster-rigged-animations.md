# v278 Spec: Biped Monster Rigged Animations

Status: Complete
Date: 2026-06-19
Codename: biped-monster-rigged-animations

## Purpose

Give the main melee biped (`dungeon_mob`) and ranged biped (`dungeon_archer`) real skinned monster
rigs so their walk and attack clips move bones instead of only bobbing or rotating the imported
scene root. This keeps the existing visual identities while making melee and ranged monsters read
as animated creatures in combat.

## Non-goals

- No monster stat, loot, combat, AI, projectile, protocol, or server-authority changes.
- No external assets, new plugins, Blender export dependency, or imported animation retargeting.
- No final production-quality deformation pass; deterministic rigid-region weights are acceptable.
- No generalized ranged equipment overlay system beyond preserving the existing archer marker.

## Asset Decision

- Adopt: existing committed source GLBs for `monster_dark_purple_v0` and
  `monster_crocodile_archer_v0`.
- Borrow: the deterministic GLB parsing/skinning approach from `tools/assets/rig_hero_glbs.py`
  and the client animation smoke-test pattern from v275-v277.
- Reject: external rigging tools or asset-store animation packs for this slice.

## Acceptance Criteria

- `client/assets/monsters/purple_fantasy/dark_purple_monster.glb` and
  `client/assets/monsters/archer/crocodile_archer.glb` contain a parseable `Skeleton3D` with
  biped skin joints declared in `assets/manifests/assets.v0.json`.
- `monster_dark_purple.tscn` and `monster_crocodile_archer.tscn` still instantiate the same visual
  scenes and preserve their current scale/yaw corrections.
- Both scenes expose `idle`, `walk`, `attack`, `hit`, and `death` clips.
- The client animation test proves `walk` rotates leg bones and `attack` rotates arm bones for both
  biped scenes.
- `make validate-assets` and the Godot animation smoke pass.

## Scope and Likely Files

- Assets/tooling: `tools/assets/rig_monster_glbs.py`, `tools/assets/test_rig_monster_glbs.py`,
  `assets/manifests/assets.v0.json`, runtime monster GLBs.
- Client: `client/animations/monster_dark_purple_anims.tres`,
  `client/animations/monster_crocodile_archer_anims.tres`, monster scenes if root paths need
  adjustment, `client/tests/test_animation.gd`.
- Docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, and as-built summary at finish.

## Test and Bot Proof

- `make gen-assets`
- `.venv/bin/python -m pytest tools/assets/test_rig_monster_glbs.py tools/assets/test_validate_assets.py -q`
- `make validate-assets`
- `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- `GODOT=/opt/homebrew/bin/godot make client-unit`
- Optional visual proof remains `make bot-visual scenario=41_monster_visual_catalog`.

## Open Questions and Risks

- No blocking questions.
- Risk: source GLBs may import with different Godot skeleton paths. The plan must verify the actual
  scene tree and target animation tracks accordingly.
