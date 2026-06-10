# v24 Plan - Main menu and character start

Status: Ready for implementation
Goal: Add a player-facing main menu with named character creation, character selection, fresh-session start, basic window-size settings, and ESC pause menu.
Architecture: The server owns accounts, characters, character ownership checks, and session bootstrap. The client boots to UI, calls authenticated HTTP APIs to list/create/select characters, then starts a fresh `dungeon_levels` session and connects the existing WebSocket. Continue never resumes an old player world; only character-owned progression persists across fresh sessions. Existing automation paths remain explicit so visual replay, smoke, protocol bots, and dev `ARPG_SESSION_ID` flows stay usable.
Tech stack: Go HTTP/store/auth/session services, Godot 4 GDScript UI over existing `main.tscn`, Godot client bot scenarios, existing Python protocol bot regressions, local Godot `user://` settings.

## Baseline and shortcut decision

v24 builds on v23 `item-templates-and-rolled-drops`, especially v22/v23 character-owned item persistence and rolled item payload persistence. The gameplay Sim, shared rules, WebSocket message schemas, combat, loot, replay reconstruction, and inventory mutation intents should remain unchanged. The main new contracts are authenticated HTTP APIs for character list/create and selected-character fresh session creation.

Godot shortcut adoption checklist:

- **Decision:** reject new plugin adoption.
- **Reason:** this slice needs a small main menu, character picker, name prompt, settings panel, and pause overlay. Existing client UI is custom GDScript, the UI has no complex inventory/grid logic, and adding an addon would add more CI/headless integration work than it saves.
- **Borrow:** keep using in-repo Control nodes and patterns from existing `inventory_panel.gd`, `consumable_bar.gd`, and `waypoint_panel` construction. No external art pack is needed for v24.

Spec review notes resolved during planning:

- Slice number, branch, codename, and baseline match `PROGRESS.md`.
- The spec's open questions have defaults and this plan locks them: duplicate character names are allowed; Return to Main Menu should use a small end-session route if implementation cost stays low; the character picker does not need inventory/equipment summaries; window size applies immediately.
- No shared rules, golden fixtures, Go `game/`, or WebSocket schema changes are expected.
- Bot proof is mandatory because this changes the client entry flow and session start flow; add client bot scenario `08_main_menu_flow.json`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/store/interfaces.go` | Add explicit character list/create methods |
| Modify | `server/internal/store/repos.go` | Implement account-scoped character list/create queries |
| Modify | `server/internal/store/models.go` | Add character summary/request helper models only if useful |
| Add | `server/internal/http/character.go` | Register `GET /v0/characters` and `POST /v0/characters` |
| Modify | `server/internal/http/session.go` | Accept optional `character_id`; validate ownership; optionally register end-session route |
| Modify | `server/internal/http/auth_session_test.go` | Cover character APIs, ownership isolation, selected-character session start, and default fallback |
| Modify | `client/scripts/net_client.gd` | Add character list/create, selected-character session create, optional end-session helper |
| Add | `client/scripts/client_settings.gd` | Load/save/apply local window-size setting |
| Add | `client/scripts/main_menu.gd` | Main menu UI and button signals |
| Add | `client/scripts/character_select_panel.gd` | Continue list and New Game name prompt |
| Add | `client/scripts/settings_panel.gd` | Fixed window-size selector |
| Add | `client/scripts/pause_menu.gd` | ESC pause overlay controls |
| Modify | `client/scripts/main.gd` | Boot to menu, start/stop gameplay sessions, block input under overlays |
| Modify | `client/scenes/main.tscn` | Mount menu/pause/settings nodes if built as scene children |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add menu wait/assert/action step types |
| Modify | `client/scripts/bot_controller.gd` | Dispatch menu actions and expose menu state |
| Modify | `client/tests/test_client_bot.gd` | Validate new client bot step types |
| Add | `tools/bot/scenarios/client/08_main_menu_flow.json` | End-to-end client bot proof for menu/session flow |
| Modify | `PROGRESS.md` | Lifecycle update when v24 ships |

## Task 1 - Server character APIs

Files:

- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/models.go`
- Add: `server/internal/http/character.go`
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 1.1: Add `ListCharacters(ctx, accountID)` and `CreateCharacter(ctx, charID, accountID, name)` to `store.CharacterRepo`; keep `GetOrCreateDefaultCharacter` unchanged for fallback compatibility.
- [x] Step 1.2: Implement `ListCharacters` with `WHERE account_id = $1 ORDER BY created_at ASC, id ASC`.
- [x] Step 1.3: Implement `CreateCharacter` with explicit account ownership and no uniqueness constraint on character names in v24.
- [x] Step 1.4: Add `GET /v0/characters`, authenticated, returning `characters[{character_id,name,created_at}]`.
- [x] Step 1.5: Add `POST /v0/characters`, authenticated, trimming name whitespace, rejecting empty names, rejecting names longer than 24 characters, and returning the created character.
- [x] Step 1.6: Extend `createSessionRequest` with optional `CharacterID string`.
- [x] Step 1.7: For fresh session create, if `character_id` is present, load that character and reject `404 session_not_found` or `403/404 character_not_found` style errors when it does not belong to the account; if omitted, preserve the current default-character path.
- [x] Step 1.8: Load character items and waypoints for the selected character before creating the session-start snapshot.
- [x] Step 1.9: Add `POST /v0/sessions/{session_id}/end` only if it stays small: authenticated owner-only, sets `status = ended`, preserves replay/debug rows, and is idempotent for already-ended sessions.
- [x] Step 1.10: Add HTTP tests for unauthenticated character API access, create validation, account isolation, selected-character session create, cross-account rejection, omitted-character fallback, and optional end-session ownership/idempotency.

```bash
cd server && go test ./internal/store ./internal/http/... -run 'Character|Session|Auth'
```

## Task 2 - Client HTTP and settings helpers

Files:

- Modify: `client/scripts/net_client.gd`
- Add: `client/scripts/client_settings.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 2.1: Add `NetClient.list_characters() -> Array` using `GET /v0/characters`.
- [x] Step 2.2: Add `NetClient.create_character(name: String) -> Dictionary` using `POST /v0/characters`.
- [x] Step 2.3: Add optional `character_id` parameter to `NetClient.create_session(...)`, while preserving existing `create_session(resume_session_id, requested_world_id)` call sites.
- [x] Step 2.4: Add `NetClient.end_session()` if Task 1 adds the end-session route; make client callers tolerate a missing/failed end route by still returning to menu locally.
- [x] Step 2.5: Implement `ClientSettings` with supported sizes `1280x720`, `1600x900`, and `1920x1080`; load from `user://settings.json`, apply via `DisplayServer.window_set_size`, and save on selection.
- [x] Step 2.6: Make invalid/missing settings fall back to `1920x1080` without crashing.
- [x] Step 2.7: Add unit-testable validation helpers for supported size parsing or cover them through client bot unit tests.

```bash
make client-unit
```

## Task 3 - Main menu, character picker, settings, and pause UI

Files:

- Add: `client/scripts/main_menu.gd`
- Add: `client/scripts/character_select_panel.gd`
- Add: `client/scripts/settings_panel.gd`
- Add: `client/scripts/pause_menu.gd`
- Modify: `client/scenes/main.tscn`
- Modify: `client/scripts/main.gd`

- [x] Step 3.1: Add a menu layer under the existing scene using Control nodes, either directly from `main.tscn` or constructed in scripts, without nesting decorative cards inside cards.
- [x] Step 3.2: Main menu buttons: Continue, New Game, Settings, Exit.
- [x] Step 3.3: Continue opens the character picker populated from `NetClient.list_characters()`; disable or show empty-state behavior if no characters exist.
- [x] Step 3.4: New Game opens a name input; validate non-empty and max length client-side before calling `create_character`.
- [x] Step 3.5: Starting from Continue or New Game calls `create_session("", "dungeon_levels", character_id)`, connects WebSocket, hides menus, and initializes gameplay state.
- [x] Step 3.6: Settings panel shows only fixed window-size options and applies/saves immediately.
- [x] Step 3.7: Pressing ESC during gameplay opens a pause menu with Resume, Settings, Return to Main Menu, and Exit.
- [x] Step 3.8: ESC inside settings/character panels backs out to the previous menu layer where practical; root main menu must not unexpectedly quit.
- [x] Step 3.9: Return to Main Menu closes the WebSocket, optionally calls `end_session`, clears entities/inventory/teleporter/current-world state, and shows the main menu without offering resume.
- [x] Step 3.10: Exit closes any WebSocket and calls `get_tree().quit(0)`.
- [x] Step 3.11: Keep visual replay mode bypassing the menu exactly as before.
- [x] Step 3.12: Keep automation/dev paths explicit: `ARPG_BOT_CLIENT` can drive the menu for `08_main_menu_flow`, while existing client bot scenarios either bypass the menu with an env flag or use a compatibility auto-start path; `ARPG_SESSION_ID` remains dev/debug-only.

```bash
make client-unit
make client-smoke
```

## Task 4 - Input locking and gameplay state reset

Files:

- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/inventory_panel.gd` if needed
- Modify: `client/scripts/consumable_bar.gd` if needed

- [x] Step 4.1: Introduce a single menu/input state predicate, for example `_menu_blocks_gameplay_input()`, that covers main menu, character picker, settings, and pause menu.
- [x] Step 4.2: Fold the new predicate into `_input_locked()` / `_user_input_blocked()` so WASD, world clicks, hotbar keys, inventory toggle, drag/drop, camera zoom, and facing updates cannot mutate gameplay while overlays are active.
- [x] Step 4.3: Make Resume restore gameplay input without reconnecting or resetting the current session.
- [x] Step 4.4: Add a clear gameplay teardown helper that frees entity nodes, health bars, walls/interactables/loot IDs, inventory arrays, equipped state, pending actions, and ready flags before returning to menu.
- [x] Step 4.5: Ensure fresh sessions after Return to Main Menu start from the selected character's durable progression and a new world seed, not leftover client state.
- [x] Step 4.6: Preserve smoke/autoplay behavior by bypassing or auto-starting only under explicit env flags.

```bash
make client-unit
make client-smoke
```

## Task 5 - Client bot menu proof

Files:

- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/tests/test_client_bot.gd`
- Add: `tools/bot/scenarios/client/08_main_menu_flow.json`

- [x] Step 5.1: Add client bot state fields for `main_menu_visible`, `character_panel_visible`, `settings_panel_visible`, `pause_menu_visible`, selected window size, known characters, and current `session_id`.
- [x] Step 5.2: Add menu action steps such as `click_menu_button`, `enter_character_name`, `select_character`, `select_window_size`, and `assert_session_changed`, or equivalent narrowly scoped primitives.
- [x] Step 5.3: Add validation tests in `test_client_bot.gd` for all new step types and required fields.
- [x] Step 5.4: Create `08_main_menu_flow.json`: wait for main menu, open Settings, select a window size, return, New Game with a deterministic name, wait for gameplay WebSocket open, press ESC, assert pause, Resume, Return to Main Menu, Continue, select the created character, and assert a fresh new session starts.
- [x] Step 5.5: Add assertions that pause/menu overlays block gameplay input by attempting a benign click or movement under pause and confirming no gameplay action is sent or no player position changes.
- [x] Step 5.6: Ensure existing client scenarios `01`-`07` keep passing by using the chosen bypass/auto-start path.

```bash
make client-unit
make db-up
make bot-client
```

## Task 6 - Regression gates

Files:

- Modify: existing files only if regressions expose compatibility gaps.

- [x] Step 6.1: Run the protocol bot to prove server-side session create fallback, replay, and current gameplay scenarios remain compatible.
- [x] Step 6.2: Run client smoke to prove old headless smoke still creates/resumes its explicit session path.
- [x] Step 6.3: Run Go tests for HTTP/store plus all server packages.
- [x] Step 6.4: Fix only regressions caused by v24; do not refactor unrelated gameplay systems.

```bash
make bot
make client-smoke
make test-go
```

## Task 7 - Lifecycle docs and final CI

Files:

- Modify: `PROGRESS.md`
- Modify: `docs/specs/v24_spec-main-menu-and-character-start.md`
- Modify: `docs/plans/v24_2026-06-07-main-menu-and-character-start.md`

- [x] Step 7.1: Update the v24 spec status when implementation completes.
- [x] Step 7.2: Add v24 to the `PROGRESS.md` slice numbering note and lifecycle table.
- [x] Step 7.3: Add a concise "What v24 proved" section covering main menu, named characters, fresh-session continue, window settings, ESC pause, and client bot proof.
- [x] Step 7.4: Move any newly deferred items to `Open gaps & deferred work`: delete/rename characters, richer settings, old-session resume UI if ever desired, character summaries, and production menu art/audio.
- [x] Step 7.5: Mark this plan complete only after final CI is green.

```bash
make ci
```

## Final verification

- [x] `cd server && go test ./internal/store ./internal/http/... -run 'Character|Session|Auth'`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot-client`
- [x] `make bot`
- [x] `make ci`

## Deferred scope

- Character delete, rename, class selection, visual customization, portraits, and character summaries.
- Player-facing old-world/session resume.
- Persistent dungeon maps, floor drops, corpses, opened doors, current level, player HP, or death/checkpoint state.
- Settings beyond fixed window size.
- Production menu art, audio, and animated title treatment.
- Godot UI plugin adoption.
