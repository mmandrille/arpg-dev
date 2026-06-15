# Spec: `main-menu-and-character-start`

Status: Complete - `make ci` green on 2026-06-07
Branch: `feature/main-menu-and-character-start`
Slice: v24 - main menu, named character start flow, fresh-session continue, and basic window settings
Baseline: v23 `item-templates-and-rolled-drops`
Related:

- [`v20_spec-play-session-loop.md`](v20_spec-play-session-loop.md) - default town-to-dungeon play loop
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) - durable character inventory/equipment/waypoints
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item persistence baseline
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - thin client and authoritative server boundary
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - character persistence and session-scoped worlds
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

The game now has durable character-owned progression, but the interactive client still jumps
directly into a fresh session when `make play` starts. There is no player-facing place to continue
with an existing character, create a new named character, adjust basic display settings, or exit
from a normal game shell.

This slice adds the first real startup and in-game menu flow. The client opens on a main menu.
Continue lists the account's characters and starts a new `dungeon_levels` session from the selected
character's persisted inventory, equipment, rolled items, and waypoints. New Game asks for a
character name, creates a fresh character, then starts a new `dungeon_levels` session. Settings
supports window size only. During gameplay, pressing ESC opens a pause menu that blocks gameplay
input and allows Resume, Settings, Return to Main Menu, or Exit.

Worlds remain session-scoped. Leaving gameplay abandons that current world from the player's point
of view; only character-owned progression persists. Server-side session rows, inputs, and events may
still exist for replay/debug tooling, but v24 must not expose an old-world resume option in the
player-facing UI.

## 2. Non-goals

- No player-facing old-session resume. Continue always starts a fresh session/world from selected
  character progression.
- No persistent dungeon maps, monsters, corpses, floor drops, opened doors, player HP, or current
  level across sessions.
- No character deletion, rename, class selection, visual customization, stats, skills, level/XP, or
  portrait thumbnails.
- No stash, vendors, quests, gold, crafting, or progression summary beyond what is cheap to show
  from existing character data.
- No settings beyond window size in v24. Fullscreen, audio, controls remapping, accessibility
  options, graphics quality, and language selection are deferred.
- No production menu art, audio, or animated title sequence.
- No gameplay WebSocket protocol schema bump.
  the required adoption checklist.

## 3. Files to create or modify

```text
docs/specs/v24_spec-main-menu-and-character-start.md          - this slice contract
docs/plans/v24_2026-06-07-main-menu-and-character-start.md    - implementation plan
server/internal/http/character.go                             - character list/create API
server/internal/http/session.go                               - fresh session create accepts selected character_id
server/internal/http/auth_session_test.go                     - character/session API behavior
server/internal/store/models.go                               - character summary model if needed
server/internal/store/interfaces.go                           - list/create character repo methods
server/internal/store/repos.go                                - Postgres character list/create implementation
server/internal/store/store_test.go                           - character repo tests if needed
client/scenes/main.tscn                                       - mount menu UI nodes if scene changes are needed
client/scripts/main.gd                                        - boot to menu, session start routing, ESC pause
client/scripts/net_client.gd                                  - character list/create and optional end-session HTTP calls
client/scripts/main_menu.gd                                   - main menu controls
client/scripts/character_select_panel.gd                      - Continue/New Game character UI
client/scripts/settings_panel.gd                              - window-size settings UI
client/scripts/pause_menu.gd                                  - ESC pause overlay
client/scripts/client_settings.gd                             - local window size load/save helper if separated
client/scripts/bot_scenario_runner.gd                         - menu/assertion step types
client/scripts/bot_controller.gd                              - menu action dispatch/state exposure
client/tests/test_client_bot.gd                               - scenario validation for menu steps
tools/bot/scenarios/client/08_main_menu_flow.json             - client bot menu proof
PROGRESS.md                                              - lifecycle update when v24 ships
```

## 4. Data shapes

### Character list

New authenticated endpoint:

```text
GET /v0/characters
```

Response:

```json
{
  "characters": [
    {
      "character_id": "char_...",
      "name": "Mara",
      "created_at": "2026-06-07T00:00:00Z"
    }
  ]
}
```

The response may add cheap display-only summary fields if already available without coupling menu
UI to game simulation state, for example equipped weapon display name or inventory count. Summary
fields are optional for v24; name and id are required.

### Character create

New authenticated endpoint:

```text
POST /v0/characters
```

Request:

```json
{
  "name": "Mara"
}
```

Response:

```json
{
  "character_id": "char_...",
  "name": "Mara",
  "created_at": "2026-06-07T00:00:00Z"
}
```

Validation:

- Trim leading/trailing whitespace.
- Reject empty names.
- Enforce a small maximum length, default 24 visible characters.
- Names only need to be unique enough for display in v24; exact duplicate policy can be decided in
  the plan. Prefer allowing duplicates per account unless implementation finds that confusing for
  bot assertions.

### Session create with selected character

`POST /v0/sessions` keeps the current response shape and adds an optional request field:

```json
{
  "mode": "solo",
  "world_id": "dungeon_levels",
  "character_id": "char_..."
}
```

Rules:

- `character_id` must belong to the authenticated account.
- If `character_id` is omitted, existing default-character behavior remains for bot/smoke/dev
  compatibility.
- Fresh menu starts send `world_id: "dungeon_levels"` and the selected character id.
- `resume_session_id` remains a dev/debug path, not a player-facing menu path.

### Session exit/end

The player-facing UI treats Return to Main Menu and Exit as abandoning the current world. The
implementation may add an authenticated end-session endpoint or only close the WebSocket and stop
showing the old session in UI. If an endpoint is added, it should mark the session ended without
deleting replay/debug rows:

```text
POST /v0/sessions/{session_id}/end
```

The plan should choose the smallest implementation that satisfies the UX and keeps replay/dev
tools working.

### Local settings

Settings are client-local and do not cross the server boundary.

Logical shape, whether stored in `user://settings.json` or another Godot-local mechanism:

```json
{
  "window_size": {
    "width": 1280,
    "height": 720
  }
}
```

Supported sizes for v24 should be a fixed list, for example:

```text
1280x720
1600x900
1920x1080
```

The current project default remains `1920x1080`.

## 5. Architecture and flow

### Startup flow

```text
Godot starts main.tscn
  -> load local window-size setting
  -> authenticate with dev-login as today
  -> GET /v0/characters
  -> show main menu

Continue
  -> show character list
  -> player selects a character
  -> POST /v0/sessions { mode: "solo", world_id: "dungeon_levels", character_id }
  -> connect WebSocket
  -> gameplay starts from fresh world + persisted character progression

New Game
  -> prompt for character name
  -> POST /v0/characters { name }
  -> POST /v0/sessions { mode: "solo", world_id: "dungeon_levels", character_id }
  -> connect WebSocket
  -> gameplay starts from fresh world with a new character
```

Existing non-interactive paths must remain explicit and stable:

- Visual replay mode should still bypass the menu and load its replay playlist.
- `ARPG_AUTOPLAY` / bot flows may bypass or drive the menu depending on the scenario.
- `ARPG_SESSION_ID` can remain a dev/debug resume path, but the interactive player menu should not
  surface it.

### Pause flow

```text
gameplay running
  -> player presses ESC
  -> pause menu visible
  -> gameplay input, inventory dragging, hotbar keys, mouse world clicks, and camera zoom are blocked

Resume
  -> hide pause menu
  -> gameplay input resumes

Settings
  -> show settings panel above/inside pause flow
  -> selected window size applies immediately and persists locally

Return to Main Menu
  -> close WebSocket
  -> optionally mark session ended
  -> clear current world/client state
  -> show main menu

Exit
  -> close WebSocket if open
  -> quit Godot
```

ESC behavior:

- In gameplay with no modal menu, ESC opens pause menu.
- In settings or character panels, ESC backs out to the previous menu layer where practical.
- In the main menu root, ESC may do nothing or focus Exit; do not quit unexpectedly without a clear
  button action.

The v24 plan must run the checklist in

- `adopt`
- `borrow pattern`
- `reject`

the UI is small, existing client UI is custom GDScript, and avoiding a dependency keeps CI/headless
automation simpler. Borrow visual/layout patterns from lightweight Godot menu examples only if they
help without adding runtime dependencies.

## 6. Acceptance criteria

1. Interactive `make play` opens the main menu before creating or connecting to a gameplay
   WebSocket session.
2. `GET /v0/characters` returns all characters owned by the authenticated account and no characters
   from other accounts.
3. New Game prompts for a non-empty character name, creates a named character, then starts a fresh
   `dungeon_levels` session for that character.
4. Continue opens a character picker and starts a fresh `dungeon_levels` session for the selected
   character.
5. Continue never resumes a previous world/session from the player-facing UI. Character inventory,
   equipment, rolled item payloads, and waypoints persist; dungeon map, monsters, floor drops,
   player HP, and current level do not.
6. `POST /v0/sessions` rejects a `character_id` owned by another account.
7. Omitting `character_id` from `POST /v0/sessions` keeps existing default-character behavior for
   bot/smoke/dev compatibility.
8. Settings shows a fixed list of window sizes and applies the selected size to the Godot window.
9. Window-size choice persists locally across client restarts.
10. Pressing ESC during gameplay opens a pause menu.
11. While pause/settings/menu overlays are visible, gameplay input is blocked: no world click
    actions, WASD movement, hotbar use, inventory toggle, or camera zoom should affect gameplay.
12. Resume from the pause menu hides the overlay and restores gameplay input.
13. Return to Main Menu closes the active gameplay connection/state and shows the main menu without
    offering old-session resume.
14. Exit from the main menu or pause menu quits the Godot application cleanly.
15. Visual replay and existing bot/smoke flows still run through their explicit automation paths.
16. Client bot scenario `08_main_menu_flow.json` proves main menu, new named character, continue
    with selected character, ESC pause, settings window-size action, resume, and return to menu.
17. `make ci` green.

## 7. Testing plan

1. `cd server && go test ./internal/store/... ./internal/http/... -run 'Character|Session|Auth'`
   - character list/create, account isolation, selected-character session create.
2. `make client-unit`
   - menu scenario validation, local settings helper, input lock state if covered by unit tests.
3. `make bot-client`
   - includes `08_main_menu_flow.json` plus existing client scenarios.
4. `make client-smoke`
   - proves existing headless smoke path remains compatible.
5. `make bot`
   - protocol regression scenarios stay green.
6. `make ci`
   - final gate.
7. Manual: `make play`, create a named character, exit to menu, continue with that character, change
   window size, press ESC during gameplay, return to menu, and confirm a fresh world starts while
   character-owned gear/waypoints persist.

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | New Game creates a new named character. | The user wants to pick from characters and enter a name at creation time. |
| 2 | Continue starts a fresh session/world from selected character state. | World/session state is not durable player progression; only character state persists. |
| 3 | No player-facing old-session resume. | Keeps the loop simple and aligns with v20/v22 session-scoped world decisions. |
| 4 | Settings includes window size only in v24. | Delivers useful display control without expanding into audio/controls/graphics systems. |
| 5 | ESC opens an in-game pause menu. | Standard game affordance and necessary shell around settings/return/exit. |
| 6 | Existing default-character session create remains valid. | Protects current bot, smoke, replay, and dev automation paths. |
| 7 | Server remains authoritative for character ownership and session bootstrap. | Maintains ADR-0001 thin-client boundary. |

## 9. Open questions

| # | Question | Default if unanswered |
|---|----------|----------------------|
| Q-1 | Should duplicate character names be allowed on the same account? | Allow duplicates in v24; display character creation time if needed for disambiguation. |
| Q-2 | Should Return to Main Menu mark the current session ended server-side? | Prefer adding a small end-session route if it is cheap; otherwise close the WebSocket and keep old sessions debug-only. |
| Q-3 | Should the menu show inventory/equipped summaries in the character picker? | No for v24 unless already trivial from existing store data. |
| Q-4 | Should the selected window size apply instantly or after confirmation? | Apply instantly and persist on selection. |
