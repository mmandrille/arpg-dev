# v435 As Built — Bone Gear Sockets

Date: 2026-07-04

## Shipped

- `shared/assets/gear_sockets.v0.json` — default bone mounts + per-class offset overrides.
- `GearSocketsLoader` + `character_visual.gd` now attach all gear sockets via `BoneAttachment3D`.
- `class_id` wired from `main.gd` / showme before socket creation.
- Extended client scenario `97_bone_gear_sockets`.

## Verification

```bash
make validate-shared
make client-unit
SCENARIO=97_bone_gear_sockets HEADLESS=1 ./scripts/bot_client_local.sh
```
