# Spec: `menu-create-join-flow`

Status: Draft
Date: 2026-06-09
Branch: `main`
Slice: v45 - main menu create/join flow cleanup
Baseline: v44 `skill-points-and-magic-bolt`
Related:

- [`../PROGRESS.md`](../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - thin client, authoritative backend, no offline gameplay path
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - client UI shortcut checklist
- [`v24_spec-main-menu-and-character-start.md`](v24_spec-main-menu-and-character-start.md) - current main menu, character picker, settings, and pause shell
- [`v38_spec-session-browser-and-uncapped-coop-menu.md`](v38_spec-session-browser-and-uncapped-coop-menu.md) - listed co-op session creation and active-session browser
- [`v44_spec-skill-points-and-magic-bolt.md`](v44_spec-skill-points-and-magic-bolt.md) - current gameplay/client UI baseline

## 1. Purpose

The client now has character-backed starts, listed co-op session discovery, settings, and exit
controls, but the main menu still exposes older development-era concepts as separate entry points:
`Continue`, `New Game`, and `Multiplayer`. That menu shape suggests local/offline or old-session
resume semantics that the architecture does not support. All playable sessions, including solo
sessions, are backend sessions.

This slice cleans the player-facing startup flow:

- The root main menu has two primary play actions: `Create Game` and `Join Game`.
- `Settings` and `Exit` remain available as secondary root-menu actions.
- `Create Game` chooses or creates an account character, then immediately starts a fresh backend
  `dungeon_levels` session.
- The session type used by `Create Game` is controlled by a client-local Settings toggle:
  `Co-op` or `Solo`, defaulting to `Co-op`.
- `Co-op` create starts a listed co-op backend session so it can appear in the active game list.
- `Solo` create starts a backend solo session; it is not offline and does not use any local-only
  gameplay path.
- `Join Game` first shows the active listed game list, then asks the player to choose or create the
  character they will join with.
- If an account has no characters at the point where a character is required, the UI forces
  character creation before it can start or join a game.

The goal is a clearer player mental model: create a backend game, or join an existing backend game.
Character selection is still required because durable progression belongs to characters, not to
sessions.

## 2. Non-Goals

- No offline mode, local-only game, or client-authoritative gameplay path.
- No player-facing old-session resume. Returning to the menu still abandons the current world from
  the player's perspective; durable character progression persists.
- No Steam lobby, invites, friend flows, chat, ready checks, party staging room, matchmaking, or
  active-session filters/search/sorting.
- No character class selection, visual customization, portraits, or richer main-menu character
  summaries.
- No production menu art, audio, or animated title treatment.
- No WebSocket gameplay protocol schema bump.
- No new backend session model unless the implementation plan finds a small missing guard in the
  existing v24/v38 HTTP APIs.
- No Godot UI plugin adoption for v45 by default. The plan must record the plugin shortcut
  checklist result; expected decision is to reuse the existing in-repo `Control` menus.

## 3. Acceptance Criteria

1. The normal root main menu presents `Create Game`, `Join Game`, `Settings`, and `Exit`; it no
   longer presents root `Continue`, `New Game`, or `Multiplayer` buttons.
2. `Settings` includes a persistent client-local session-type toggle for `Create Game` with values
   `Co-op` and `Solo`.
3. The session-type toggle defaults to `Co-op` for a fresh local settings profile.
4. The selected session type is saved and restored through the existing client settings path.
5. `Create Game` with existing characters shows a character choice that also allows creating a new
   character.
6. `Create Game` with no characters forces the character-creation view; Back returns to the root
   menu without starting a session.
7. After a character is selected or created from `Create Game`, the client immediately starts
   gameplay in a fresh `dungeon_levels` session.
8. `Create Game` with session type `Co-op` sends a backend listed co-op create request and enters
   the returned WebSocket session.
9. `Create Game` with session type `Solo` sends a backend solo create request and enters the
   returned WebSocket session.
10. No `Create Game` path bypasses backend auth, HTTP session creation, or the existing WebSocket
    connection flow.
11. `Join Game` shows the active listed game list before any character picker is shown.
12. The active game list supports refresh, row selection, Join Selected, and Back. Back returns to
    the root menu without mutating session or character state.
13. `Join Game` with no active listed games shows a clear empty state and does not force character
    selection until a join target exists.
14. After the player selects an active listed game, the UI shows existing characters or forces
    character creation if the account has none.
15. After a character is selected or created for `Join Game`, the client joins the selected listed
    backend session and enters gameplay over the returned WebSocket.
16. Character create validation remains consistent with v24: trimmed non-empty names, maximum 24
    visible characters, and server-owned account association.
17. Dead characters, if present in the character list, remain disabled for create/join start flows.
18. Pause menu behavior remains unchanged: Resume, Settings, Return to Main Menu, and Exit still
    work during gameplay.
19. Returning to the main menu tears down gameplay state and does not expose player-facing
    old-session resume.
20. Visual replay, `ARPG_SESSION_ID`, `ARPG_AUTOSTART`, protocol bot, smoke, and existing dev
    automation paths remain explicit and compatible.
21. The Godot client debug state exposes enough menu/settings state for bot assertions: root menu
    visibility, character panel mode, join-game panel visibility, selected session id, current
    session type setting, known characters, and current session id.
22. Client bot coverage proves both no-character forced creation and existing-character selection
    for `Create Game`.
23. Client bot coverage proves `Join Game` list-first behavior and joining a prepared listed
    backend session.
24. Existing protocol bot listed co-op coverage remains green.
25. `make client-unit`, `make client-smoke`, relevant `make bot-client` scenarios, and `make ci`
    pass.

## 4. Scope And Likely Files

```text
docs/specs/v45_spec-menu-create-join-flow.md - this spec
docs/plans/v45_2026-06-09-menu-create-join-flow.md - implementation plan
docs/PROGRESS.md - lifecycle update when v45 ships

client/scripts/main_menu.gd - replace root play buttons and signals
client/scripts/settings_panel.gd - add session-type toggle
client/scripts/client_settings.gd - persist create-game session type
client/scripts/main.gd - route Create Game / Join Game / character selection flows
client/scripts/character_select_panel.gd - support choose-or-create and forced-create modes
client/scripts/multiplayer_sessions_panel.gd - reuse as Join Game active-session list, remove host action from this path
client/scripts/net_client.gd - reuse existing solo/listed create, active list, and listed join helpers
client/scripts/bot_scenario_runner.gd - add/update assertions for menu labels, settings toggle, and join flow
client/scripts/bot_controller.gd - add/update menu action dispatch if needed
client/tests/test_client_bot.gd - validate new client bot step names/shapes if needed
tools/bot/scenarios/client/08_main_menu_flow.json - update main menu flow proof
tools/bot/scenarios/client/20_menu_create_join_flow.json - optional separate client proof if cleaner than expanding 08
tools/bot/scenarios/27_session_browser_uncapped_coop.json - existing protocol proof remains the backend join/list baseline
```

Server files are not expected to change. If planning finds a required backend guard, keep it limited
to existing session/character HTTP contracts:

```text
server/internal/http/session.go - only if listed/solo create or listed join needs a small behavior guard
server/internal/http/auth_session_test.go - matching HTTP coverage for any server guard
server/internal/store/repos.go - only if active session list semantics need a small query fix
```

## 5. Flow Details

### 5.1 Root Menu

Root actions:

```text
Create Game
Join Game
Settings
Exit
```

`Create Game` and `Join Game` are the only primary play actions. `Settings` and `Exit` stay visible
but secondary in visual weight. The root menu must not use wording that implies offline play or
old-session resume.

### 5.2 Create Game

Create flow:

```text
Create Game
  -> list characters
  -> if no characters: force create character
  -> if characters exist: choose character or create new one
  -> create backend session using Settings session type
  -> connect WebSocket
  -> enter gameplay
```

When Settings session type is `Co-op`, the client sends the existing listed co-op create shape:

```json
{
  "mode": "coop",
  "listed": true,
  "world_id": "dungeon_levels",
  "character_id": "char_..."
}
```

When Settings session type is `Solo`, the client sends the existing solo create shape:

```json
{
  "mode": "solo",
  "world_id": "dungeon_levels",
  "character_id": "char_..."
}
```

Both paths are backend sessions and must connect through the existing WebSocket.

### 5.3 Join Game

Join flow:

```text
Join Game
  -> refresh and show active listed backend sessions
  -> choose listed session
  -> list characters
  -> if no characters: force create character
  -> if characters exist: choose character or create new one
  -> join selected backend session
  -> connect WebSocket
  -> enter gameplay
```

The active list reuses v38 semantics:

- `GET /v0/sessions/active` returns active listed co-op sessions only.
- Empty listed sessions are hidden.
- Join code and account ids are not exposed.
- Joining selected sessions uses `POST /v0/sessions/{session_id}/join` with `character_id`.

### 5.4 Settings

The new setting is client-local:

```text
Create Game Type: Co-op | Solo
```

Default is `Co-op`. The setting affects only future `Create Game` starts. It does not mutate an
active session and does not affect `Join Game`.

### 5.5 Character Selection

The existing character panel can be reused, but the player-facing labels should match the flow:

- Create flow title can be `Choose Character`.
- Join flow title can be `Choose Character`.
- Forced create mode should make it clear that a character is required before starting or joining.
- Existing rename/delete affordances may remain if already present and stable, but they are not the
  focus of v45.

## 6. Test And Bot Proof

Expected coverage:

- Client unit/helper coverage for persisted session-type settings, if the settings helper is
  changed.
- Client unit/helper coverage for menu action routing, if routing is extracted or already covered.
- Update `08_main_menu_flow.json` or add `20_menu_create_join_flow.json` to prove:
  - root menu labels,
  - settings session-type default and persistence,
  - no-character forced creation through `Create Game`,
  - existing-character selection through `Create Game`,
  - co-op create starts a listed backend session,
  - solo create starts a backend solo session after toggling setting,
  - join-game active list is shown before character selection,
  - selected listed session joins after character selection/create,
  - pause and Return to Main Menu still behave.
- Keep protocol bot scenario `27_session_browser_uncapped_coop.json` as the server-side listed
  session/list/join proof.
- Run `make client-unit`, `make client-smoke`, targeted `make bot-client` scenarios, and `make ci`.

If a single Godot client bot cannot reliably prepare a second-account listed session for the join
flow, the plan should either:

- use a deterministic pre-seeded backend setup step already supported by the bot harness, or
- prove the join list UI state with a prepared active-list response/unit helper and rely on protocol
  bot scenario `27_session_browser_uncapped_coop.json` for real backend joining.

## 7. Open Questions And Risks

| # | Question / Risk | Default for planning |
|---|-----------------|----------------------|
| R-1 | A full real-client `Join Game` proof may require two accounts/processes. | Prefer a prepared backend session if the harness supports it; otherwise combine client UI proof with existing protocol bot backend proof. |
| R-2 | Reusing `CharacterSelectPanel.show_continue` may leave stale `Continue` wording in the UI. | Rename labels/modes around the new flow; do not keep player-facing `Continue` on the normal path. |
| R-3 | Existing automation may still call `continue`, `new_game`, or `multiplayer` bot actions. | Preserve compatibility aliases inside bot helpers if cheap, but new scenarios should use `create_game` and `join_game`. |
| R-4 | Co-op create default may surprise solo testers by making sessions listed. | The Settings toggle must be visible, saved, and defaulted intentionally to `Co-op`; solo remains one toggle away. |
| R-5 | Existing v38 session browser includes `Host Listed Session` inside the multiplayer panel. | Remove host/create from the `Join Game` panel; creation belongs to root `Create Game`. |

