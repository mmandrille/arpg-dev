# v273 Plan - Crocodile Archer Model

Status: Complete
Goal: Replace the regular `dungeon_archer` body presentation with the supplied crocodile archer GLB
while keeping all ranged monster authority on the server.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/assets/monsters/archer/crocodile_archer.glb` | Runtime GLB bytes |
| Add | `client/assets/monsters/archer/crocodile_archer_0.png` and imports | Godot extracted texture/import metadata |
| Modify | `assets/manifests/assets.v0.json` | Asset id, provenance, runtime path |
| Modify | `shared/assets/monster_visuals.v0.json` and schema | Map `dungeon_archer` to the new scene/asset |
| Add | `client/scenes/monster_crocodile_archer.tscn` | Instanced monster scene |
| Add | `client/animations/monster_crocodile_archer_anims.tres` | Node-root presentation clips |
| Modify | `client/scripts/main.gd` / `monster_visuals_loader.gd` | Register visual scene key |
| Modify | `client/tests/test_animation.gd` | Scene/catalog coverage |
| Add | `docs/as-built/v273_crocodile-archer-model.md` | Closeout proof |

## Tasks

- [x] Probe the GLB with the 3D model workflow and confirm import shape.
- [x] Copy runtime bytes under `client/assets/monsters/archer/` and run Godot import.
- [x] Add `monster_crocodile_archer_v0` to the asset manifest with SHA-256 provenance.
- [x] Register the new monster scene key in shared visuals, schema, loader, and client scene lookup.
- [x] Preserve the archer bow-marker contract as a scene node so existing visual assertions still
  prove the regular ranged enemy.
- [x] Add focused animation/catalog tests.

## Verification

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

Note: `bot-visual` selects client scenarios by filename, so the visual command uses
`25_ranged_monster_ai`; the scenario id printed by the client is `client_ranged_monster_ai`.
