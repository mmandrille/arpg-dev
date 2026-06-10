# v18 — Dungeon levels and stairs

**Proves:** The authoritative Sim can hold multiple generated dungeon levels and move the player
between them with deterministic, level-scoped deltas.

- `dungeon_levels` world runs in multi-level mode; legacy worlds remain single-level at level `0`.
- `LevelState` owns per-level entities, walls, movement, auto-nav, and navigation bounds.
- `shared/rules/dungeon_generation.v0.json` drives 32x20 dungeon floors, perimeter walls, level
  names, player spawn, and deterministic stair placement.
- `descend_intent` / `ascend_intent` move the player between generated levels and emit old-level
  remove + new-level full spawn deltas with `level_changed`.
- `shared/golden/dungeon_stairs.json` pins level -1/-2 stair and loot positions.
- Godot renders generated dungeon walls, placeholder stairs, and a top-right level HUD.

**Explicit non-goals:** no character-scoped persistence, town, waypoints, full room/corridor PCG,
monster density by depth, co-op routing, or production stair art.
