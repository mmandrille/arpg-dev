# v256 Spec - Discovery Minimap

Status: Complete
Date: 2026-06-17
Codename: discovery-minimap

## Purpose

Replace the current objective-only "quest map" feel with a real minimap that shows the area the
player is discovering while exploring. The Godot client should own a session-local explored map for
the active floor, reveal cells around the local hero from the existing fog/light radius, draw known
walls only after they have been discovered, and let the player show or hide the map with `TAB`.

The minimap should be roughly twice the size of the old compact objective minimap and slightly
transparent so it can stay readable without fully blocking the playfield.

## Non-goals

- No server, shared protocol, replay, persistence, or database change.
- No durable explored-map memory across sessions, reconnects, or fresh visits to the same level.
- No route line, compass, quest pathing, marker search, or click-to-navigate behavior.
- No new fog gameplay, monster AI awareness, combat visibility, or authority change.
- No imported minimap art, shader plugin, Godot addon, or asset pipeline change.
- No full-screen map overlay; this slice ships only the HUD minimap toggle.

## Acceptance Criteria

- The old objective-only minimap surface is replaced by a discovery minimap on the gameplay HUD.
- `TAB` toggles the minimap visible/hidden during gameplay and does not require the quest journal.
- The minimap starts hidden by default and keeps the user's toggle state while the session runs.
- The minimap uses a map area about twice the old compact widget size, with a slightly transparent
  panel/background.
- As the local hero moves, nearby cells become permanently explored for the current active level.
- Explored state is scoped per level and resets only when the client starts a new session.
- Walls from the current authoritative wall layout appear on the minimap only where explored.
- The player dot is centered/readable, and the active elite-objective chest pin remains available
  when an objective chest is known in client entity state.
- Debug state exposes visibility, toggle state, explored cell count, map size, opacity, wall count,
  player marker, and objective pin state for unit/client bot proof.
- Existing fog overlay behavior remains unchanged.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/discovery_minimap.gd` - focused HUD widget that draws explored cells, known
    walls, player marker, optional objective pin, and debug state.
  - `client/scripts/discovery_minimap_state.gd` - pure session-local explored-cell state and
    projection helper for active level, player position, light radius, wall layout, and objective
    pin.
  - `client/scripts/main.gd` - create the minimap, toggle it on `TAB`, sync player/fog/wall/entity
    state after snapshots/deltas, and expose bot debug state.
- Client tests and bot:
  - `client/tests/test_discovery_minimap.gd` - unit coverage for default hidden state, toggle,
    explored-cell accumulation, per-level reset/scope, wall reveal, size/opacity, and objective pin.
  - `scripts/client_smoke.sh` - replace the old elite objective minimap unit gate with the new
    discovery minimap test.
  - `client/scripts/bot_assertion_handlers.gd` - add an `assert_discovery_minimap` client bot
    assertion.
  - `tools/bot/scenarios/client/69_discovery_minimap_toggle.json` - client visual proof that `TAB`
    toggles the minimap and exploration debug state is populated.
- Docs:
  - `docs/plans/v256_2026-06-17-discovery-minimap.md`
  - `docs/as-built/v256_discovery-minimap.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported minimap art, shader plugins, and Godot
addons. Borrow existing in-repo `EliteObjectiveMinimap` HUD placement/drawing conventions,
`FogOfWarOverlay` light-radius debug inputs, `WallRenderer` normalized wall-layout data, and client
bot assertion patterns.

## Test and Bot Proof

- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=69_discovery_minimap_toggle`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=69_discovery_minimap_toggle
```

## Open Questions and Risks

- No required questions. Defaults accepted on 2026-06-17: the minimap is session-local,
  client-presentational, hidden by default, doubled from the old 104px widget to about 208px, and
  slightly transparent.
- Risk: `TAB` may conflict with future control remapping. This slice keeps the direct key binding
  consistent with existing hardcoded panel keys and defers controls remapping to the settings gap.
- Risk: `client/scripts/main.gd` is a large grandfathered coordinator. The implementation should
  keep new logic in focused scripts and limit `main.gd` to creation, input, synchronization, and bot
  debug plumbing.
- Risk: this client-local explored state is presentation only. Server-authoritative fog/visibility
  remains the source of truth for gameplay and hidden monster information.
