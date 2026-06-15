# v17 — Monster chase movement

**Proves:** opt-in server-authoritative monster chase with aggro, leash return, and v11 path reuse.

- Shared `behavior: "chase"` on `training_dummy_chase` with `aggro_radius`, `leash_radius`, and
  `move_speed == navigation.cell_size`; all legacy monsters default to static.
- `Sim.Tick` runs `advanceMonsterMovement` after player movement and before projectiles; monsters
  replan each tick, path around walls/player/other monsters, and emit edge-only `monster_aggro` /
  `monster_leashed` events plus `entity_update` position deltas.
- Golden `shared/golden/monster_chase.json` pins maze chase and leash return on seed
  `cafebabecafebabe`.
- Worlds `chase_lab`, `chase_maze`, and `leash_lab` plus bot scenarios `09`–`11` prove open-field
  chase, maze routing, and leash reset through `/state`, reconnect, and replay.
- Godot drives monster `walk`/`idle` from authoritative position deltas; `monster_anims.tres` adds
  a minimal walk clip.
- `make ci` green on 2026-06-06.

or NavMesh authority. Nearby group aggro was added later in v37.
