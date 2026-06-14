# v164 Plan — Session browser filters

Status: Complete
Goal: Join Game listed-session rows can be searched and sorted client-side.
Architecture: Keep `GET /v0/sessions/active` unchanged. `MultiplayerSessionsPanel` stores the
server rows, derives a filtered/sorted visible list, and emits the same join signal for selected
session ids.
Tech stack: Godot GDScript panel, bot/debug helpers, client unit tests, existing listed-session
client bot preflight.

## Baseline and shortcut decision

Builds on v38/v45/v46 listed-session browser work and the backlog item
`session-browser-filters`. Godot plugin/adoption checklist: reject external UI plugins/assets; this
is a small in-repo panel control addition.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/multiplayer_sessions_panel.gd` | Search/sort controls and filtered visible rows |
| Modify | `client/scripts/main.gd` | Bot-callable filter/sort hooks |
| Modify | `client/scripts/bot_controller.gd` | Bot actions for filter/sort |
| Modify | `client/scripts/bot_step_catalog.gd` | Step validation for filter/sort |
| Modify | `client/scripts/bot_scenario_runner.gd` | Filter assertion helper |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Wire filter assertion |
| Modify | `client/tests/test_coop_client.gd` | Panel unit coverage |
| Modify | `client/tests/test_client_bot.gd` | Bot DSL coverage |
| Modify | `tools/bot/scenarios/client/21_join_game_listed_session.json` | Exercise search/sort before join |
| Create | `docs/as-built/v164_session-browser-filters.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle closeout |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Keep `main.gd` changes to wrappers only.
- [x] Keep panel logic local to `MultiplayerSessionsPanel`; no backend changes.

## Task 1 — Panel search/sort

- [x] Add search text state and sort mode state.
- [x] Render only filtered/sorted rows while retaining server row data.
- [x] Preserve selection and join by stable `session_id`.
- [x] Add debug fields for search/sort/visible counts.

## Task 2 — Bot and tests

- [x] Add bot actions for setting multiplayer search and sort mode.
- [x] Add a bot assertion for multiplayer filter state.
- [x] Extend unit tests for panel filtering/sorting and DSL validation.
- [x] Update listed-session scenario to filter by expected session id, assert visible count, then clear.

```bash
make client-unit
make bot-client scenario=21_join_game_listed_session.json
```

## Task 3 — Lifecycle docs and CI

- [x] Mark completed plan tasks.
- [x] Update spec status/as-built/progress.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=21_join_game_listed_session.json`
- [x] `make ci`
