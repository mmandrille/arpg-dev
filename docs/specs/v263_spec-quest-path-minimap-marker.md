# v263 Spec - Quest Path Minimap Marker

Status: Complete
Date: 2026-06-18
Codename: quest-path-minimap-marker

## Purpose

Give the minimap a clearer directional cue toward the current known quest objective. When an active
elite-objective pin is known to the client, the discovery minimap should draw a quest-path marker
from the centered player marker toward that objective so players can orient quickly in compact and
full-screen map modes.

## Non-goals

- No server routefinding, A* path, navmesh path, click-to-navigate, or autorun.
- No protocol/schema changes and no reveal of objectives the client does not already know.
- No persistent map/fog behavior beyond the active-session memory already shipped.
- No imported icon art, plugins, shader dependencies, or minimap legend/filter UI.

## Acceptance Criteria

- `DiscoveryMinimapState` derives a `quest_path` payload only when the active objective pin is known.
- Completed, hidden, missing-node, or unknown objectives do not expose a quest-path marker.
- The widget draws a code-native directional line/arrow from the player marker toward the objective
  pin in compact and full-screen modes.
- Bot debug state exposes whether the quest-path marker is active, plus stable normalized endpoints
  suitable for assertions.
- Existing POI marker counts and objective-pin behavior remain unchanged.

## Scope and Likely Files

- Client:
  - `client/scripts/discovery_minimap_state.gd`
  - `client/scripts/discovery_minimap.gd`
  - `client/tests/test_discovery_minimap.gd`
  - `client/scripts/bot_assertion_handlers.gd`
- Bot:
  - `tools/bot/scenarios/client/45_elite_objective_minimap_pin.json`
- Docs:
  - `docs/plans/v263_2026-06-18-quest-path-minimap-marker.md`
  - `docs/as-built/v263_quest-path-minimap-marker.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets/plugins. Use the existing code-native minimap drawing
style and objective state.

## Test and Bot Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=45_elite_objective_minimap_pin
make maintainability
```

Manual visual proof, if desired:

```bash
make bot-visual scenario=45_elite_objective_minimap_pin
```

## Open Questions and Risks

- No required questions.
- Risk: players may read a straight-line cue as a guaranteed route. Keep code/docs terminology to
  directional quest-path marker, not server-authored pathfinding.
