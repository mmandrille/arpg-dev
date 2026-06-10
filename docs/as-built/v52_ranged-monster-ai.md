# v52 — Ranged monster AI

**Proves:** Generated dungeon floors now include a server-authoritative ranged monster variant
with a visible client-side bow marker.

- Shared monster rules add `dungeon_archer` with ranged attack fields, while dungeon generation
  rules define a deterministic melee-heavy spawn pool plus a minimum archer guarantee on normal
  generated floors.
- The Go sim keeps melee mobs unchanged, gives ranged chase monsters range/line-of-sight-aware
  standoff goals, and spawns monster-owned projectile entities only when the shot is clear.
- Monster-owned projectiles advance through the existing authoritative projectile lifecycle and
  can damage only their living target player on the same level, resolving miss/block/damage/death
  through the existing monster combat stat path.
- Dungeon monster definition rolls use their own seeded RNG stream so adding archer composition
  does not perturb established dungeon layout and placement fixtures.
- Godot attaches a small procedural bow marker to `dungeon_archer` nodes from authoritative
  `monster_def_id` metadata and exposes `has_bow_marker` through client bot debug state.
- Protocol bot scenario `38_ranged_monster_ai.json` proves generated archer presence and
  archer-sourced ranged player damage; client bot scenario `25_ranged_monster_ai.json` proves the
  live bow-marker presentation and ranged damage path.
