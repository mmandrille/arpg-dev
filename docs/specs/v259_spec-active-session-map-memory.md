# v259 Spec - Active-Session Map Memory

Status: Complete
Date: 2026-06-18
Codename: active-session-map-memory

## Purpose

Make discovery-map explored memory explicitly scoped to the current active play session. Explored
cells should persist while the same session is active, including level changes and map mode changes,
but reset when a new session starts or gameplay state is torn down. This honors the selected scope:
no reconnect/resume persistence and no cross-session character map memory.

## Non-goals

- No server, shared protocol, replay, persistence, database, or schema change.
- No reconnect/resume restoration of explored cells.
- No cross-session character map memory.
- No durable fog-of-war snapshots or server-owned explored-cell model.
- No marker, biome, LOS, or dungeon-generation behavior change.

## Acceptance Criteria

- `DiscoveryMinimap` tracks a current session key and resets explored cells when that key changes.
- Discovery state persists within a single session across map mode changes and active level changes.
- Runtime gameplay teardown clears discovery state so a fresh session cannot inherit old explored
  cells.
- Debug state exposes the current session key for focused proof.
- Existing v256-v258 minimap behavior, markers, walls, and objective pin remain intact.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/discovery_minimap.gd` - add session-scope reset and debug key.
  - `client/scripts/main.gd` - call minimap session sync from snapshots and reset on teardown.
- Client tests and bot:
  - `client/tests/test_discovery_minimap.gd` - unit proof for session-key reset and intra-session
    retention.
  - `tools/bot/scenarios/client/72_active_session_map_memory.json` - client proof that the map debug
    session key is set and explored state survives map mode changes.
- Docs:
  - `docs/plans/v259_2026-06-18-active-session-map-memory.md`
  - `docs/as-built/v259_active-session-map-memory.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets/plugins. This is lifecycle/state plumbing in the
existing minimap scripts only.

## Test and Bot Proof

- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=72_active_session_map_memory`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=72_active_session_map_memory
```

## Open Questions and Risks

- User answered on 2026-06-18 that explored-map memory should persist only during the actual active
  session and does not need reconnect/resume.
- Risk: session reset must not clear exploration on every snapshot. The reset only fires when the
  non-empty session key changes or gameplay teardown explicitly clears state.
