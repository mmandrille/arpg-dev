# v279 As-Built - Quadruped Pounce Animation Support

Date: 2026-06-19
Spec: [`docs/specs/v279_spec-quadruped-pounce-animation-support.md`](../specs/v279_spec-quadruped-pounce-animation-support.md)
Plan: [`docs/plans/v279_2026-06-19-quadruped-pounce-animation-support.md`](../plans/v279_2026-06-19-quadruped-pounce-animation-support.md)

## Shipped

- Added `tools/assets/rig_quadruped_monster_glbs.py`, a deterministic quadruped skin injector for
  the existing `evil_fox_monster.glb` source asset.
- Wired `make gen-assets` to regenerate the rigged quadruped predator runtime after generated,
  hero, and biped-monster rig passes.
- Updated `monster_quadruped_predator_v0` manifest metadata with quadruped skin joints and the new
  runtime hash.
- Replaced the quadruped root-bob animation library with bone-driven `walk`, `attack`, `pounce`,
  `hit`, `death`, and `idle` clips.
- Updated `monster_quadruped.tscn` so the animation player targets the imported quadruped model,
  while preserving the existing `ModelRoot` yaw correction and model scale.
- Extended the Godot animation smoke to require quadruped bones and prove walk, attack, and pounce
  clips move the expected rig/model parts.

## Proof

```bash
.venv/bin/python -m pytest tools/assets/test_rig_quadruped_monster_glbs.py tools/assets/test_validate_assets.py -q
make validate-assets
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

All focused checks above passed on 2026-06-19.

Post-loop `$refactor` paid down the ratchet debt and the selected batch passed full `make ci` on
2026-06-19.

## Boundaries

- No server gameplay, monster stats, combat, AI, cooldown, loot, protocol, or persistence changed.
- The new `pounce` clip is presentation support only. Server-owned pounce behavior remains in the
  selected queue for v281.
