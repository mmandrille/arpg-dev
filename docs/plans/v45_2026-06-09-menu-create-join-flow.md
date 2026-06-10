# v45 Plan - Main Menu Create/Join Flow

Status: Ready for implementation
Goal: Replace the old Continue/New Game/Multiplayer startup shape with clear Create Game and Join Game flows backed by existing server sessions.
Architecture: The Go backend remains authoritative for accounts, characters, session creation, listed session discovery, joins, and WebSocket gameplay. Godot owns only local menu presentation, local settings, and routing to existing HTTP helpers. Solo remains a backend session, not an offline mode. No shared gameplay rules, WebSocket protocol schemas, or replay contracts should change.
Tech stack: Godot GDScript menu/settings/client-bot code, existing Go HTTP/session APIs, existing Python protocol bot listed-session proof, docs.

## Spec Review

Spec passes the planning gate.

- Baseline matches `PROGRESS.md`: v44 `skill-points-and-magic-bolt` is complete on `main`, so v45 is next.
- Scope is UI/routing plus bot proof. Backend changes are optional guards only.
- No shared contracts, protocol schema bumps, golden fixtures, or Go sim determinism changes are required.
- Server authority is preserved: `Create Game` and `Join Game` both call backend session APIs and connect the existing WebSocket.
- Client UI work requires the Godot shortcut checklist; decision below records reuse/reject.
- Bot proof is required and covered by a client-bot menu scenario plus existing protocol scenario `27_session_browser_uncapped_coop.json`.

## Baseline and shortcut decision

Baseline is v44 `skill-points-and-magic-bolt` on `main`. Reuse:

- v24 main menu, character selection, settings, pause, Return to Main Menu, and client bot menu helpers.
- v38 listed co-op create, active session list, listed join, multi-client launchers, and protocol bot proof.
- v44 skill UI state in `main.gd` teardown/input-lock paths, so menu changes do not leave skill panels active.

Godot plugin shortcut decision: **reject** external menu/lobby plugins for v45. The slice is a small navigation cleanup over existing in-repo `Control` panels, and adopting a plugin would add dependency and headless-CI surface without improving server-backed session routing. Steam/GodotSteam lobby UI remains deferred to a platform slice.

Implementation decisions:

- `Create Game` uses a local `ClientSettings.create_game_session_type` value with allowed values `coop` and `solo`; default `coop`.
- `Create Game` with `coop` calls existing `NetClient.create_listed_coop_session(character_id)`.
- `Create Game` with `solo` calls existing `NetClient.create_session("", "dungeon_levels", character_id)`.
- `Join Game` reuses the active-session browser but removes host/create from that panel; hosting belongs to root `Create Game`.
- Character panel should expose explicit modes instead of leaking `Continue` / `New Game` copy into the player-facing v45 path.
- Keep compatibility aliases for old bot button names if cheap, but new scenarios should use `create_game`, `join_game`, and session-type steps.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/plans/v45_2026-06-09-menu-create-join-flow.md` | This implementation plan |
| Modify | `docs/specs/v45_spec-menu-create-join-flow.md` | Only if implementation discovers accepted clarifications |
| Modify | `PROGRESS.md` | Lifecycle update when v45 ships |
| Modify | `client/scripts/client_settings.gd` | Persist `create_game_session_type` with default `coop` |
| Modify | `client/scripts/settings_panel.gd` | Render and emit the Create Game Type toggle |
| Modify | `client/scripts/main_menu.gd` | Replace root play buttons/signals with Create Game / Join Game |
| Modify | `client/scripts/character_select_panel.gd` | Add choose/create/forced-create modes with v45 copy |
| Modify | `client/scripts/multiplayer_sessions_panel.gd` | Reframe as Join Game list; remove host button from normal path |
| Modify | `client/scripts/main.gd` | Route create/join flows, settings sync, debug state, bot actions |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add menu/settings assertions and new action names |
| Modify | `client/tests/test_client_bot.gd` | Validate new step/action shapes and settings parser |
| Modify | `client/tests/test_coop_client.gd` | Add focused panel/routing unit coverage if helpers are testable |
| Modify | `tools/bot/scenarios/client/08_main_menu_flow.json` | Migrate old menu proof to Create Game flow |
| Create | `tools/bot/scenarios/client/20_menu_create_join_flow.json` | Focused v45 menu proof, if clearer than expanding 08 |
| Audit | `server/internal/http/session.go` | Confirm existing create/list/join contracts satisfy v45 |
| Audit | `server/internal/http/auth_session_test.go` | Add tests only if audit finds a backend guard is needed |
| Audit | `server/internal/store/repos.go` | Add query guard only if active list behavior regresses v45 |

## Task 1 - Contract and server audit

Files:
- Audit: `client/scripts/net_client.gd`
- Audit: `server/internal/http/session.go`
- Audit: `server/internal/http/auth_session_test.go`
- Audit: `server/internal/store/repos.go`
- Audit: `tools/bot/scenarios/27_session_browser_uncapped_coop.json`

- [x] Step 1.1: Confirm `NetClient.create_session`, `create_listed_coop_session`, `list_active_sessions`, and `join_listed_session` already match v45 flow needs.
- [x] Step 1.2: Confirm `POST /v0/sessions` supports both `mode: "solo"` and listed `mode: "coop"` with selected `character_id`.
- [x] Step 1.3: Confirm `GET /v0/sessions/active` hides solo, ended, and empty listed sessions and exposes no join code/account id.
- [x] Step 1.4: Confirm `POST /v0/sessions/{session_id}/join` accepts listed joins without a join code and rejects missing/dead/cross-account characters.
- [x] Step 1.5: If any guard is missing, add the smallest HTTP/store fix and matching Go test; otherwise leave server code untouched.

```bash
go test ./internal/http/...
go test ./internal/store/...
make bot scenario=27_session_browser_uncapped_coop.json
```

## Task 2 - Local settings session-type toggle

Files:
- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/settings_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 2.1: Add constants for create-game session types: `coop` and `solo`; default to `coop`.
- [x] Step 2.2: Load `create_game_session_type` from `user://settings.json`, normalize invalid/missing values to `coop`, and save it with existing settings.
- [x] Step 2.3: Add a Settings control labeled around `Create Game Type` with values `Co-op` and `Solo`; avoid changing this setting during active gameplay except as a local preference for future creates.
- [x] Step 2.4: Add `SettingsPanel` signal(s) and setter(s) so `main.gd` can sync the toggle alongside window size and text toggles.
- [x] Step 2.5: Add client unit coverage for default value, parse/normalize, save data shape, and panel sync behavior where practical.

```bash
make client-unit
```

## Task 3 - Root menu and panel copy cleanup

Files:
- Modify: `client/scripts/main_menu.gd`
- Modify: `client/scripts/character_select_panel.gd`
- Modify: `client/scripts/multiplayer_sessions_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 3.1: Replace `MainMenu` root signals/buttons with `create_game_pressed`, `join_game_pressed`, `settings_pressed`, and `exit_pressed`.
- [x] Step 3.2: Render root buttons as `Create Game`, `Join Game`, `Settings`, and `Exit`; remove root `Continue`, `New Game`, and `Multiplayer` from the normal UI.
- [x] Step 3.3: Add explicit character panel modes for choose-or-create and forced-create, with titles/copy that do not say `Continue` or `New Game`.
- [x] Step 3.4: In choose-or-create mode, list existing non-dead characters and provide a create-new-character affordance without requiring a root New Game button.
- [x] Step 3.5: In forced-create mode, hide the empty `No characters` list path and focus the name field; Back returns to the previous panel/root without starting a session.
- [x] Step 3.6: Reframe `MultiplayerSessionsPanel` as Join Game: title `Join Game`, actions `Refresh`, list rows, `Join Selected`, `Back`; remove `Host Listed Session` from the normal join panel.
- [x] Step 3.7: Keep dead character rows disabled and keep rename/delete affordances only if they remain stable and do not obscure the v45 start flow.
- [x] Step 3.8: Add focused unit coverage for dead row disabled behavior and any extracted panel mode helpers.

```bash
make client-unit
```

## Task 4 - Main flow routing and state teardown

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/net_client.gd` if helper naming/metadata needs cleanup only
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 4.1: Replace `_on_continue_pressed`, `_on_new_game_pressed`, and `_on_multiplayer_pressed` normal routing with `_on_create_game_pressed` and `_on_join_game_pressed`.
- [x] Step 4.2: Implement Create Game routing: list characters, forced-create if empty, choose-or-create if present, then start selected/created character through the configured session type.
- [x] Step 4.3: Implement Create Game `coop` start by calling `create_listed_coop_session(character_id)`, then `_begin_gameplay_connection(false)`.
- [x] Step 4.4: Implement Create Game `solo` start by calling `create_session("", "dungeon_levels", character_id)`, then `_begin_gameplay_connection(false)`.
- [x] Step 4.5: Implement Join Game routing: hide root menu, show active listed sessions immediately, refresh on entry, and only show character selection after a non-empty selected session id.
- [x] Step 4.6: Preserve `pending_join_session_id` semantics and join selected sessions through `join_listed_session(pending_join_session_id, character_id)`.
- [x] Step 4.7: Ensure Back behavior is correct for every flow: character panel from create returns to root; character panel from join returns to Join Game list; Join Game Back returns to root.
- [x] Step 4.8: Ensure Return to Main Menu still calls teardown/end-session paths, clears gameplay/UI state including v44 skill panels, and does not expose old-session resume.
- [x] Step 4.9: Keep visual replay, `ARPG_SESSION_ID`, `ARPG_AUTOSTART`, smoke, and protocol bot automation paths explicit and compatible.
- [x] Step 4.10: Extend `get_bot_state()` with root menu button labels or available actions, character panel mode/title, create-game session type, and join-game selected session id.

```bash
make client-unit
make client-smoke
```

## Task 5 - Client bot actions and assertions

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 5.1: Add or document new menu button action values: `create_game`, `join_game`, `refresh_sessions`, `join_selected_session`, `back`, `start`, `resume`, `return_to_main_menu`, `settings`, and `exit`.
- [x] Step 5.2: Keep compatibility aliases for `continue`, `new_game`, `multiplayer`, `host_listed_session`, and `join_first_listed_session` only if cheap and non-confusing; new scenarios should not use old names.
- [x] Step 5.3: Add a bot action for selecting `Create Game Type` in settings, for example `select_create_game_type` with `session_type: "coop" | "solo"`.
- [x] Step 5.4: Add assertions for menu labels/available actions, character panel mode, selected create-game session type, current session `mode`, and current session `listed` flag.
- [x] Step 5.5: Extend static scenario validation tests for the new action/assertion step types and required fields.
- [x] Step 5.6: Extend runtime assertion tests for the new state fields without requiring a live server.

```bash
make client-unit
```

## Task 6 - Bot scenarios

Files:
- Modify: `tools/bot/scenarios/client/08_main_menu_flow.json`
- Create: `tools/bot/scenarios/client/20_menu_create_join_flow.json`
- Keep: `tools/bot/scenarios/27_session_browser_uncapped_coop.json`
- Modify: `tools/bot/test_protocol.py` only if protocol bot helper assertions change

- [x] Step 6.1: Update `08_main_menu_flow.json` to cover the renamed root menu and preserve the existing settings, pause, Return to Main Menu, and fresh-session-change proof.
- [x] Step 6.2: Add `20_menu_create_join_flow.json` for v45-specific coverage, unless the updated `08_main_menu_flow.json` remains clearer as the single menu proof.
- [x] Step 6.3: Cover no-character forced creation from `Create Game`, default `Co-op` session type, listed co-op session creation, WebSocket open, and debug assertion that the session is `mode: "coop"` and `listed: true`.
- [x] Step 6.4: Return to main menu, switch Settings create-game type to `Solo`, create/select a character, start gameplay, and assert the session is `mode: "solo"` and not listed.
- [x] Step 6.5: Cover existing-character selection by returning to root and starting from an already-created character without entering a new name.
- [x] Step 6.6: Cover `Join Game` list-first behavior and empty-state behavior when no active listed sessions exist.
- [x] Step 6.7: For real Join Game against an active listed session, prefer preparing a second-account listed session through an existing harness/server setup if feasible. If not feasible in one Godot process, document the deferral in the plan during execution and rely on `27_session_browser_uncapped_coop.json` for backend join/list proof plus client unit assertions for selected-row routing.
- [x] Step 6.8: Keep protocol bot scenario `27_session_browser_uncapped_coop.json` green to prove server listed discovery and joins remain correct.

Execution note: the live Godot bot harness consistently sees the legacy default `Hero` character for a fresh local account, so `08_main_menu_flow.json` and `20_menu_create_join_flow.json` prove the choose-or-create path and create-new affordance in live backend sessions. The forced-create mode itself is covered by `client/tests/test_coop_client.gd` panel/routing unit coverage. A real multi-account Godot Join Game proof remains deferred; `27_session_browser_uncapped_coop.json` remains the backend listed join proof.

```bash
make bot scenario=27_session_browser_uncapped_coop.json
HEADLESS=1 make bot-client scenario=08_main_menu_flow.json
HEADLESS=1 make bot-client scenario=20_menu_create_join_flow.json
```

## Task 7 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v45_spec-menu-create-join-flow.md` only if accepted clarifications were discovered
- Modify: `docs/plans/v45_2026-06-09-menu-create-join-flow.md`

- [x] Step 7.1: Update `PROGRESS.md` current status to latest completed slice v45 when implementation is finished.
- [x] Step 7.2: Add v45 to the slice numbering note and lifecycle table with spec and plan links.
- [x] Step 7.3: Add a concise `v45 - Menu create/join flow` summary under "What each slice proved".
- [x] Step 7.4: Update the bot scenario summary list with the v45 client proof.
- [x] Step 7.5: Move any newly deferred menu/lobby items to Open gaps, especially if real multi-account Godot Join Game proof remains deferred.
- [x] Step 7.6: Keep this plan's checkboxes updated as tasks are completed during execution.

```bash
make ci
```

## Final verification

- [x] `go test ./internal/http/...` if server guards changed
- [x] `go test ./internal/store/...` if active-list query guards changed
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot scenario=27_session_browser_uncapped_coop.json`
- [x] `HEADLESS=1 make bot-client scenario=08_main_menu_flow.json`
- [x] `HEADLESS=1 make bot-client scenario=20_menu_create_join_flow.json` if the new scenario is created
- [x] `make ci`

## Deferred scope

- Offline/local-only gameplay remains out of scope.
- Old-session player-facing resume remains out of scope.
- Steam lobby/invites, matchmaking, chat, ready checks, and filters/search/sorting remain out of scope.
- Production menu art/audio remains out of scope.
- Server protocol/schema changes remain out of scope unless Task 1 finds a small required HTTP guard.
