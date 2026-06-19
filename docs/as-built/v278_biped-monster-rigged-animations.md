# v278 As-Built - Biped Monster Rigged Animations

Date: 2026-06-19
Spec: [`docs/specs/v278_spec-biped-monster-rigged-animations.md`](../specs/v278_spec-biped-monster-rigged-animations.md)
Plan: [`docs/plans/v278_2026-06-19-biped-monster-rigged-animations.md`](../plans/v278_2026-06-19-biped-monster-rigged-animations.md)

## Shipped

- Added `tools/assets/rig_monster_glbs.py`, a deterministic wrapper that rigs the existing
  dark-purple melee biped and crocodile archer source GLBs with the shared humanoid skin joints.
- Wired `make gen-assets` to regenerate those rigged monster runtime GLBs after the generated and
  hero rig passes.
- Preserved the user-provided tiny flyer bat runtime by removing the obsolete generated tiny-flyer
  overwrite from `tools/assets/gen_glb.py`.
- Updated `assets/manifests/assets.v0.json` with biped required skin joints and new runtime hashes
  for `monster_dark_purple_v0` and `monster_crocodile_archer_v0`.
- Replaced root-bob biped animation clips with bone-driven walk/attack/hit/death tracks and kept
  the existing scene scale/yaw corrections and archer marker intact.
- Extended the Godot animation smoke to require biped monster skeletons and prove `walk` rotates
  leg bones while `attack` rotates arm bones.

## Proof

```bash
.venv/bin/python -m pytest tools/assets/test_rig_monster_glbs.py tools/assets/test_validate_assets.py -q
make validate-assets
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

All focused checks above passed on 2026-06-19.

Post-loop `$refactor` paid down the ratchet debt and the selected batch passed full `make ci` on
2026-06-19.

## Boundaries

- No server gameplay, monster stats, combat, loot, protocol, persistence, AI, projectile behavior,
  or authority changed.
- The biped rigs use deterministic rigid-region weights. Final artist-quality deformation remains
  future art/tech-art polish.
