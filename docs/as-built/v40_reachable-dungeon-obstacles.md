# v40 — Reachable dungeon obstacles

**Proves:** Generated dungeon floors can include deterministic interior wall obstacles while the
server still guarantees all generated targets remain reachable and the client renders only
authoritative wall layouts.

- Shared dungeon generation rules now include data-driven obstacle generation tuning for group
  counts, wall segments, solid blocks, shape weights, and clearance rules.
- Protocol v3 snapshots include complete current-level `walls[]`; level transitions emit
  `wall_layout_update` before destination entity spawns so clients replace static layout before
  rendering new floor contents.
- Go dungeon generation creates deterministic non-boss interior obstacle groups on a separate
  obstacle RNG stream, retries unreachable layouts, and keeps boss floors perimeter-only.
- Reachability checks use the same grid assumptions as authoritative auto-pathing and validate
  stairs, teleporters, chests, loot, and generated monster spawns.
- Generated walls are solid for player movement, auto-pathing, monster chase, projectile sweeps,
  loot/drop placement, and travel arrival.
- Godot renders dungeon walls from snapshot/delta payloads and exposes wall counts for client bot
  assertions; preset single-level worlds still render local `worlds.v0.json` walls.
- `shared/golden/dungeon_obstacles.json` pins the `v40_obstacles` level `-2` wall order, shape
  family coverage, and reachable target positions.
- Protocol bot scenario `28_reachable_dungeon_obstacles.json` and client bot scenario
  `14_dungeon_wall_rendering.json` prove generated interior walls through `/state`, reconnect,
  replay, and headless Godot client rendering.

**Explicit non-goals:** no generated doors in obstacle walls, full room/corridor PCG, rotated or
destructible/secret obstacles, production dungeon art/lighting/sound, boss-floor obstacle generation,
durable dungeon map snapshots across fresh sessions, or final density/biome/difficulty balance.
