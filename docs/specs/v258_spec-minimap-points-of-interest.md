# v258 Spec - Minimap Points of Interest

Status: Complete
Date: 2026-06-18
Codename: minimap-points-of-interest

## Purpose

Add readable points-of-interest markers to the discovery map for entities the client has already
discovered: stairs, waypoints, town services, and the existing objective pin. The markers should
work in both compact and full-screen map modes without revealing hidden server state.

## Non-goals

- No server, shared protocol, replay, persistence, or database change.
- No markers for hidden monsters, undiscovered far stairs, unexplored services, loot, routes, or
  quest path lines.
- No marker search, filtering, labels panel, panning, click-to-navigate, or tooltip interaction.
- No imported icon art, shader plugin, Godot addon, or asset pipeline change.

## Acceptance Criteria

- The discovery map derives marker data only from known client entity records and existing objective
  state.
- Interactable markers appear only after their map cell has been explored by active-session
  discovery state.
- Stairs, waypoints, town services, and objective markers have distinct marker kinds and readable
  visual treatment.
- Markers render in both compact and full-screen map modes using the current map scale.
- Debug state exposes total marker count and per-kind counts for client unit and bot proof.
- Existing v256/v257 map mode behavior, explored cells, known walls, and objective pin compatibility
  remain intact.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/discovery_minimap_state.gd` - derive explored POI markers from known entities.
  - `client/scripts/discovery_minimap.gd` - draw marker icons and expose marker counts.
- Client tests and bot:
  - `client/tests/test_discovery_minimap.gd` - unit proof for stairs, waypoint, service, and
    objective marker derivation.
  - `client/scripts/bot_assertion_handlers.gd` - support marker count assertions.
  - `tools/bot/scenarios/client/71_minimap_points_of_interest.json` - client proof on a town-service
    lab.
- Docs:
  - `docs/plans/v258_2026-06-18-minimap-points-of-interest.md`
  - `docs/as-built/v258_minimap-points-of-interest.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external icon art, shaders, plugins, and Godot addons. Borrow the
existing `DiscoveryMinimap` code-native shape drawing and bot debug-state pattern.

## Test and Bot Proof

- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=71_minimap_points_of_interest`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=71_minimap_points_of_interest
```

## Open Questions and Risks

- No required questions.
- Risk: non-creature entities may be present in snapshots before the player has visually explored
  them. The implementation gates interactable markers by explored map cells to avoid revealing far
  points of interest.
