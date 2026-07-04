# v435 Plan — Bone Gear Sockets

Spec: [`docs/specs/v435_spec-bone-gear-sockets.md`](../specs/v435_spec-bone-gear-sockets.md)

## Tasks

- [x] Add `shared/assets/gear_sockets.v0.json` + schema
- [x] Add `GearSocketsLoader` + update `character_visual.gd`
- [x] Wire `class_id` from `main.gd` / showme
- [x] Extend `test_animation.gd` bone socket checks
- [x] Add extended client scenario `97_bone_gear_sockets`
- [x] Run focused verification

## Verification

```bash
make validate-shared
make client-unit
SCENARIO=97_bone_gear_sockets HEADLESS=1 ./scripts/bot_client_local.sh
```
