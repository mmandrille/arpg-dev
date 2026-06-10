# v19 — Teleporters and waypoint UI

**Proves:** Dungeon levels can expose session-scoped discovered teleporters and use them for
server-authoritative fast travel.

- Dungeon generation places one deterministic `teleporter` interactable per generated level.
- `action_intent` on a reachable teleporter discovers that level and emits
  `teleporter_discovery_update` plus `teleporter_discovered`.
- v1 snapshots include `discovered_teleporters`, listing generated/visited levels as enabled or
  disabled.
- `teleport_intent { target_level }` validates current teleporter reach and target discovery, then
  reuses v18 two-delta level transition output.
- Godot renders a placeholder teleporter and opens a left-side waypoint panel with disabled
  undiscovered rows and a scroll container for longer level lists.
- Bot scenario `13_teleporter_lab.json` covers discover -1, descend, verify -2 disabled, discover
  -2, and teleport back to -1.

**Explicit non-goals:** no character-scoped waypoint persistence, town waypoint, VFX/audio,
production teleporter art, hidden infinite level catalog, or plugin adoption.
