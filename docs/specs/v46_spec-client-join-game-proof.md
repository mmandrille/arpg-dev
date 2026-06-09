# Spec: `client-join-game-proof`

Status: Draft
Date: 2026-06-09
Branch: `main`
Slice: v46 - real Godot Join Game co-op proof
Baseline: v45 `menu-create-join-flow`
Related:

- [`../PROGRESS.md`](../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - thin client, backend-owned sessions, agent-playability
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - co-op players share one authoritative Sim
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - client UI/lobby shortcut checklist
- [`v33_spec-true-coop-session.md`](v33_spec-true-coop-session.md) - server-owned co-op session baseline
- [`v38_spec-session-browser-and-uncapped-coop-menu.md`](v38_spec-session-browser-and-uncapped-coop-menu.md) - listed session discovery and uncapped co-op
- [`v45_spec-menu-create-join-flow.md`](v45_spec-menu-create-join-flow.md) - current Create Game / Join Game menu baseline

## 1. Purpose

v45 made the player-facing root menu match the backend session model: players create a backend
game or join an existing backend game. It proved `Create Game`, solo/co-op settings, and the
`Join Game` empty state through the Godot client bot, while leaving a real multi-account Godot
Join Game proof deferred.

This slice closes that gap. A deterministic client-bot scenario should prepare an active listed
co-op session with one protocol-level host kept connected, then launch a separate Godot guest
account that:

- opens `Join Game`,
- sees the prepared active listed session in the real active-session browser,
- selects and joins that listed session through the existing character picker,
- connects to the returned WebSocket session, and
- observes the shared co-op session state from the real Godot client path.

The goal is confidence in the player-facing join path before adding richer lobby, party, chat,
Steam, trade, or matchmaking work.

## 2. Non-Goals

- No new gameplay, combat, inventory, world, or replay behavior.
- No WebSocket gameplay protocol schema bump.
- No new backend session model or listed-session semantics.
- No Steam lobby, invites, friend flows, matchmaking, filters/search/sorting, chat, emotes, ready
  checks, party staging, or lobby persistence.
- No full two-window visual choreography. The default proof is one protocol-held host plus one
  Godot guest.
- No production multiplayer UI redesign, art, audio, or animated lobby treatment.
- No Godot UI or lobby plugin adoption. The plan must record the plugin shortcut checklist result;
  expected decision is to reuse the existing in-repo `Control` panels and protocol bot helpers.

## 3. Acceptance Criteria

1. A client-bot preflight path can create a unique host account, create or reuse a host character,
   create a listed co-op `dungeon_levels` session, open the host WebSocket, consume the initial
   `session_snapshot`, and send `client_ready`.
2. The preflight host remains connected while the Godot client-bot scenario runs, so the prepared
   listed session is visible through `GET /v0/sessions/active` under the current active-list rule
   that hides sessions with zero connected members.
3. The preflight path exposes the prepared `session_id` to the Godot client-bot run for assertions
   and cleanup.
4. A new Godot client-bot scenario, expected as
   `tools/bot/scenarios/client/21_join_game_listed_session.json`, opens from the root menu and
   confirms the v45 root actions are still `Create Game`, `Join Game`, `Settings`, and `Exit`.
5. The scenario clicks `Join Game` and reaches the active listed-session panel before any character
   picker is shown.
6. The active listed-session panel contains the prepared listed session with `mode: "coop"`,
   `listed: true`, `member_count >= 1`, and `connected_count >= 1`.
7. The scenario selects the prepared row and proceeds to the character picker only after a selected
   listed session id exists.
8. If the guest account has no usable character, the scenario creates one through the existing
   character panel; if a character exists, selecting it also works.
9. The guest joins the selected listed session through `POST /v0/sessions/{session_id}/join` via
   the existing `NetClient.join_listed_session` path, not by using a join code or local shortcut.
10. The guest WebSocket opens for the prepared session id and the client debug state reports
    `mode: "coop"` and `listed: true`.
11. After the guest snapshot/deltas settle, the Godot client debug state shows co-op shared state:
    at least two player entities on the current level, including a remote host player distinct from
    the local player, or an equivalent structured remote-player assertion.
12. The scenario proves Back from the Join Game list still returns to the root menu before joining.
13. The scenario cleanup closes the protocol host WebSocket and does not leave a connected listed
    session visible after the client-bot run completes.
14. Existing v45 client-bot scenarios `08_main_menu_flow.json` and `20_menu_create_join_flow.json`
    remain green.
15. Existing protocol bot scenario `27_session_browser_uncapped_coop.json` remains green.
16. No shared rules, protocol schema, golden fixture, Go sim, or replay contract changes are
    required.
17. `make client-unit`, targeted `HEADLESS=1 make bot-client scenario=21_join_game_listed_session.json`,
    and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v46_spec-client-join-game-proof.md - this spec
docs/plans/v46_2026-06-09-client-join-game-proof.md - implementation plan
docs/PROGRESS.md - lifecycle update when v46 ships

scripts/bot_client.sh - run scenario-specific preflight setup/cleanup when requested
tools/bot/client_join_preflight.py - optional focused helper for a held listed co-op host
tools/bot/run.py - reuse existing protocol helpers if that is cleaner than a new helper
tools/bot/test_protocol.py - validate any new preflight config/parser behavior

tools/bot/scenarios/client/21_join_game_listed_session.json - new Godot guest Join Game proof
tools/bot/scenarios/client/08_main_menu_flow.json - existing v45 Create Game proof stays green
tools/bot/scenarios/client/20_menu_create_join_flow.json - existing v45 menu proof stays green
tools/bot/scenarios/27_session_browser_uncapped_coop.json - backend listed co-op proof stays green

client/scripts/bot_scenario_runner.gd - new assertions only if current state checks are insufficient
client/scripts/bot_controller.gd - new action dispatch only if current menu actions are insufficient
client/scripts/main.gd - expose structured remote-player/join-session debug state only if missing
client/scripts/multiplayer_sessions_panel.gd - debug state additions only if current row data is insufficient
client/tests/test_client_bot.gd - static validation/runtime assertion coverage for any new step shape
client/tests/test_coop_client.gd - focused remote-player/debug state coverage if touched
```

Server files are not expected to change. If implementation finds a real backend guard missing, keep
the change limited to existing HTTP/session behavior and cover it with focused tests:

```text
server/internal/http/session.go - only for a small active-list/join guard if needed
server/internal/http/auth_session_test.go - matching HTTP coverage
server/internal/store/repos.go - only if active-session visibility semantics are incorrect
```

## 5. Flow Details

### 5.1 Preflight Host

The test harness should prepare the host before launching the Godot guest:

```text
host dev-login
  -> ensure host character
  -> POST /v0/sessions { mode: "coop", listed: true, world_id: "dungeon_levels", character_id }
  -> connect WebSocket using returned ws_url + access token
  -> read session_snapshot
  -> send client_ready
  -> write prepared session metadata for the Godot run
  -> keep the WebSocket alive until the scenario exits
```

The host must remain connected because current server behavior intentionally hides active listed
sessions with no connected members.

### 5.2 Godot Guest

The Godot client-bot scenario should use a distinct guest email from the host preflight account and
drive the normal v45 player path:

```text
root menu
  -> Join Game
  -> active listed sessions panel
  -> select prepared listed session
  -> choose/create guest character
  -> join selected session
  -> WebSocket open
  -> assert co-op session metadata and remote host presence
```

No join code should appear in the client active list or be required by the Godot guest path.

### 5.3 Debug Assertions

Prefer structured debug state over pixel or timing assertions. Acceptable assertion data includes:

- active session rows from `multiplayer_panel.get_debug_state()`,
- selected session id from the panel or `main.gd` join state,
- current `NetClient.session_id`, `session_mode`, and `session_listed`,
- local player id,
- player entity records containing local/remote player identity,
- party rows or remote-player entity counts if those are already exposed.

If a needed debug field is missing, add the smallest display-only debug field to the Godot client.
Do not change authoritative server outcomes for test convenience.

## 6. Test And Bot Proof

Expected coverage:

- Python unit coverage for any new preflight helper option parsing, metadata writing, or cleanup
  behavior that can be tested without a live server.
- Client unit coverage for any new client-bot assertion shape or debug-state parser.
- New Godot client-bot scenario `21_join_game_listed_session.json` proving the real Join Game path
  against a live listed session held by a protocol host.
- Existing Godot client-bot scenarios `08_main_menu_flow.json` and `20_menu_create_join_flow.json`
  remain green to protect v45 root/create behavior.
- Existing protocol bot scenario `27_session_browser_uncapped_coop.json` remains green to protect
  backend listed-session discovery and multi-peer join semantics.

Expected verification commands:

```bash
make client-unit
make bot scenario=27_session_browser_uncapped_coop.json
HEADLESS=1 make bot-client scenario=21_join_game_listed_session.json
HEADLESS=1 make bot-client scenario=08_main_menu_flow.json
HEADLESS=1 make bot-client scenario=20_menu_create_join_flow.json
make ci
```

## 7. Open Questions And Risks

No planning blockers. Defaults for implementation:

- Use one protocol-held host and one Godot guest, not two Godot processes.
- Add a focused Python preflight helper if `tools/bot/run.py` cannot be reused cleanly from
  `scripts/bot_client.sh`.
- Assert remote host presence through structured client debug state, not through screenshots.

Risks:

- The host preflight must keep the WebSocket open long enough for the active list to include the
  session; the plan should include timeout/error handling around that readiness point.
- Scenario-specific preflight support in `scripts/bot_client.sh` must clean up background processes
  on success, failure, and interruption.
- If current Godot debug state cannot distinguish local and remote players, the implementation must
  add a small display-only debug field without changing server protocol.
