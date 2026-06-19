# v273 As-Built - Crocodile Archer Model

Date: 2026-06-18
Spec: [`docs/specs/v273_spec-crocodile-archer-model.md`](../specs/v273_spec-crocodile-archer-model.md)
Plan: [`docs/plans/v273_2026-06-18-crocodile-archer-model.md`](../plans/v273_2026-06-18-crocodile-archer-model.md)

## Shipped

- `dungeon_archer` now resolves through `shared/assets/monster_visuals.v0.json` to
  `monster_crocodile_archer_v0` and `monster_crocodile_archer`.
- The supplied GLB is registered in the asset manifest with runtime bytes under
  `client/assets/monsters/archer/` and the Godot-extracted texture/import metadata.
- The new Godot scene uses the static GLB plus node-root idle, walk, hit, death, and attack clips.
- `main.gd` now uses a visual-scene dictionary for monster scene lookup, shrinking the touched
  coordinator while adding the new scene.
- The existing archer bow-marker client contract remains present so ranged monster visual bots still
  assert the regular ranged enemy.

## Proof

```bash
python3 skills/3dmodel/scripts/create_model_probe.py --model assets/monsters/archer/crocodile_archer.glb --key crocodile_archer --yaw-degrees 0
make validate-shared
make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
make client-unit
make maintainability
HEADLESS=1 make bot-visual scenario=25_ranged_monster_ai
make bot scenario=ranged_monster_ai
```

All commands above passed on 2026-06-18.

## Boundaries

- No server monster definition, combat, AI, navigation, loot, or projectile behavior changed.
- The GLB has no skin or embedded animations, so this slice uses scene-level presentation clips.
- Final rigged attack animation and generalized ranged equipment overlays remain future art work.
