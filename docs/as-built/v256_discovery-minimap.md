# v256 As-built - Discovery Minimap

Date: 2026-06-17

## What shipped

- Replaced the objective-only HUD minimap wiring with a `DiscoveryMinimap` widget.
- Added session-local explored-cell memory per active level through `DiscoveryMinimapState`.
- `TAB` toggles the minimap during gameplay. It starts hidden by default and keeps explored state
  while the client session runs.
- The minimap uses a 208x208 map area, roughly double the previous compact widget, and a transparent
  panel/background.
- The minimap draws explored floor cells, discovered wall rectangles, a centered player marker, and
  an optional elite-objective chest pin when that entity is known in client state.
- Bot debug state exposes visibility, toggle state, size, opacity, explored count, wall count, player
  marker, and objective pin data.
- Added `assert_discovery_minimap` and a dedicated `69_discovery_minimap_toggle` client scenario.
- Updated the old elite objective minimap scenario to prove the objective pin through the new
  discovery minimap after pressing `TAB`.

## Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=69_discovery_minimap_toggle
HEADLESS=1 make bot-visual scenario=45_elite_objective_minimap_pin
make maintainability
```

## Manual visual check

```bash
make bot-visual scenario=69_discovery_minimap_toggle
```

## Scope limits

- No server, shared protocol, replay, persistence, or database change shipped.
- No durable explored-map memory across sessions, reconnects, or fresh visits to the same level.
- No full-screen map overlay, route line, compass, quest pathing, marker search, or click-to-navigate.
- No imported minimap art, shader plugin, Godot addon, or asset pipeline change.
