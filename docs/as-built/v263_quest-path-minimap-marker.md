# v263 As-Built - Quest Path Minimap Marker

Date: 2026-06-18
Spec: [`docs/specs/v263_spec-quest-path-minimap-marker.md`](../specs/v263_spec-quest-path-minimap-marker.md)
Plan: [`docs/plans/v263_2026-06-18-quest-path-minimap-marker.md`](../plans/v263_2026-06-18-quest-path-minimap-marker.md)

## Shipped Behavior

- `DiscoveryMinimapState` now derives a `quest_path` dictionary from the existing active objective
  pin.
- The quest-path marker is active only when the active objective is known and offset from the player.
  Hidden, completed, missing-node, or unknown objectives keep the marker inactive.
- `DiscoveryMinimap` draws a code-native directional line and arrow from the centered player marker
  toward the known objective in compact and full-screen modes.
- Bot debug state exposes `has_quest_path`, normalized start/end coordinates, and angle radians.
- Existing POI marker counts and objective-pin behavior remain unchanged.

## Boundaries

- No server routefinding, pathfinding, navmesh, click-to-navigate, protocol, persistence, or database
  behavior changed.
- The cue is directional; it is not a guaranteed navigable path.
- No imported minimap art, shader plugin, Godot addon, legend, marker filter, or tooltip UI shipped.

## Verification

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=45_elite_objective_minimap_pin
make maintainability
make ci
```

All commands passed on 2026-06-18. The selected autoloop batch `make ci` gate completed in 10m40s.

Manual visual check:

```bash
make bot-visual scenario=45_elite_objective_minimap_pin
```
