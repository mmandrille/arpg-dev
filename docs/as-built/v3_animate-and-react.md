# v3 — Animate and react

**Proves:** ADR-0007 animation state model; rigged GLB → skeletal clips pipeline.

- Player: `idle` / `walk` / `attack` from client input/prediction; weapon on `hand_r` bone
- Monster: `hit` / `death` from authoritative `monster_damaged` / `monster_killed` events
- `AnimationController` priority machine: terminal > one-shot > locomotion
- Clips built by `client/tools/build_animations.gd` → committed `.tres` libraries
- **No server/protocol change** — client starts reading existing `state_delta.events`
