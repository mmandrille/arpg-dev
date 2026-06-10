# v9 — Solid collision and obstacles

**Proves:** The authoritative server can block player movement against live monster bodies and
static world walls while preserving replay/resume determinism.

- Shared `worlds.v0.json` now supports static `wall` entries with axis-aligned rectangular sizes.
- `collision_lab` world places wall obstacles with a middle passage and a live monster beyond them.
- Server movement checks player circle vs live monster circles and wall AABBs; diagonal moves slide
  on one axis when possible.
- Dead monsters are non-solid, so corpses do not block loot/combat scenario flow.
- Python bot adds `move_until_player_position` and a collision lab scenario proving traversal
  through the wall gap before the final monster attack, `/state`, reconnect, and replay.
- Godot renders simple static wall boxes from shared world rules for fresh sessions and visual replay
  manifests; the server still owns all collision outcomes.
- `make ci` green on 2026-06-05.

**Explicit non-goals:** no pathfinding, navmesh, monster movement/AI, polygon collision, or wall
protocol entities. Attack range was deferred in v9 and closed by v10.
