# v11 — Click to move and auto path

**Proves:** The server can own deterministic click-to-move and auto-approach using shared
navigation rules while preserving replay/resume behavior.

- `move_to_intent { position }` queues server-owned floor-click movement.
- Out-of-range `action_intent` plans to a reachable melee approach cell, queues movement, and
  executes the original action on arrival with one acceptance ack.
- Shared `navigation.v0.json` defines `cell_size`, `max_auto_steps`, search bounds, and
  `stop_distance`; `auto_path.json` pins the path-maze approach fixture.
- Go A* rasterizes walls, live monsters, and closed interactables from the same collision rules
  used by movement; manual `move_intent` cancels queued auto-navigation.
- `path_maze` world plus bot scenario `05_path_maze.json` proves one entity click routes through
  a wall maze and kills a target without scripted waypoints.
- Godot empty-floor left click sends `move_to_intent`; entity click stays `action_intent`.

**Explicit non-goals (still true):** no NavMesh authority, monster AI/pathfinding, path preview UI,
door closing, inventory UI, or production navigation polish.
