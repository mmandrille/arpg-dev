# v279 Spec: Quadruped Pounce Animation Support

Status: Complete
Date: 2026-06-19
Codename: quadruped-pounce-animation-support

## Purpose

Prepare the `dungeon_wolf` quadruped presentation for a future pounce attack by giving the existing
quadruped predator runtime GLB a real quadruped skin and adding bone-driven `walk`, `attack`, and
`pounce` clips. This is a visual prerequisite only; the server-owned pounce behavior ships later.

## Non-goals

- No monster stat, combat, AI, cooldown, protocol, or server-authority changes.
- No external assets, plugins, or Blender dependency.
- No final production-quality deformation; deterministic rigid-region weights are acceptable.

## Asset Decision

- Adopt: existing committed `assets/monsters/purple_fantasy/evil_fox_monster.glb`.
- Borrow: the deterministic GLB skin injection pattern from v278 and the existing quadruped scene
  correction/root structure.
- Reject: replacing the visual with a new external model or hand-authored DCC animation export.

## Acceptance Criteria

- `monster_quadruped_predator_v0` declares quadruped skin joints in `assets/manifests/assets.v0.json`
  and `make validate-assets` proves the runtime GLB contains them.
- `monster_quadruped.tscn` keeps its `ModelRoot` yaw correction and `QuadrupedModel` scale.
- The quadruped animation library exposes `idle`, `walk`, `attack`, `pounce`, `hit`, and `death`.
- The Godot animation smoke proves `walk` rotates leg bones, `attack` rotates head/spine bones, and
  `pounce` moves the model forward/up without changing `ModelRoot` yaw.

## Scope and Likely Files

- Assets/tooling: `tools/assets/rig_quadruped_monster_glbs.py`,
  `tools/assets/test_rig_quadruped_monster_glbs.py`, `assets/manifests/assets.v0.json`,
  `client/assets/monsters/purple_fantasy/evil_fox_monster.glb`, `make/shared.mk`.
- Client: `client/animations/monster_quadruped_fox_anims.tres`,
  `client/scenes/monster_quadruped.tscn`, `client/tests/test_animation.gd`.
- Docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, and as-built summary at finish.

## Test and Bot Proof

- `make gen-assets`
- `.venv/bin/python -m pytest tools/assets/test_rig_quadruped_monster_glbs.py tools/assets/test_validate_assets.py -q`
- `make validate-assets`
- `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- `GODOT=/opt/homebrew/bin/godot make client-unit`

## Open Questions and Risks

- No blocking questions.
- Risk: the source fox GLB proportions may make rigid quadruped weights imperfect. This slice only
  requires readable bone-driven motion, not final deformation.
