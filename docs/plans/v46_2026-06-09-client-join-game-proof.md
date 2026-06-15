# v46 Plan - Real Godot Join Game Co-op Proof

Status: Ready for implementation
Goal: Prove the player-facing Godot Join Game path against a real active listed co-op session.
Architecture: The Go backend remains authoritative for accounts, characters, listed sessions,
membership, WebSocket snapshots, and co-op player state. A Python preflight helper will create and
hold one connected listed co-op host so the existing active-session browser has a real row. The
Godot client bot then joins as a separate guest account through the v45 UI. No shared rules,
gameplay protocol, replay, or Go sim behavior should change.
Tech stack: Bash client-bot runner, Python protocol helper, Godot GDScript client/debug assertions,
existing Go HTTP/WebSocket APIs, docs.

## Spec Review

Spec passes the planning gate.

- Baseline matches `PROGRESS.md`: v45 `menu-create-join-flow` is complete on `main`, so v46 is next.
- Scope is a harness/client-bot proof over existing listed co-op behavior; no new gameplay is
  hidden in acceptance criteria.
- No shared contracts, protocol schema bumps, golden fixtures, or replay contract changes are
  required.
- Go sim determinism risk is low because server gameplay logic should not change.
- Server authority is preserved: the Godot guest uses `NetClient.join_listed_session` and the
  backend owns all membership/session outcomes.
- Client UI work requires the Godot shortcut checklist; decision below records reuse/reject.
- Bot proof is explicit: new client scenario `21_join_game_listed_session.json`, existing v45
  scenarios, and existing protocol scenario `27_session_browser_uncapped_coop.json`.

## Baseline and shortcut decision

Baseline is v45 `menu-create-join-flow` on `main`. Reuse:

- v33 true co-op WebSocket/session behavior and remote-player Godot rendering.
- v38 listed co-op session creation, active-session browser semantics, and protocol proof.
- v45 root `Create Game` / `Join Game` menu, active-session panel, character picker, and client
  debug state.
- Existing Python protocol helpers in `tools/bot/run.py` for dev login, character creation, listed
  co-op session creation, WebSocket connect, and `client_ready`.
Implementation decisions:

- Add scenario-level preflight support to `scripts/bot_client.sh`, triggered only by scenarios that
  opt in through a small JSON field such as `"preflight": {"type": "listed_coop_host"}`.
- Implement the preflight as a Python process that writes prepared session metadata, opens the host
  WebSocket, sends `client_ready`, then waits until killed by the runner cleanup path.
- Launch the Godot guest with environment variables from the preflight metadata, including expected
  session id and host account/character identifiers if useful.
- Prefer extending existing client-bot assertions over adding bespoke scenario code.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/specs/v46_spec-client-join-game-proof.md` | Slice spec |
| Create | `docs/plans/v46_2026-06-09-client-join-game-proof.md` | This implementation plan |
| Modify | `PROGRESS.md` | Lifecycle update when v46 ships |
| Modify | `scripts/bot_client.sh` | Scenario-specific preflight process setup, env injection, and cleanup |
| Create | `tools/bot/client_join_preflight.py` | Hold a connected listed co-op host for the Godot guest scenario |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for preflight metadata/config helpers if practical |
| Create | `tools/bot/scenarios/client/21_join_game_listed_session.json` | Godot guest Join Game proof |
| Modify | `client/scripts/bot_scenario_runner.gd` | Assertion extensions for prepared session rows/current session/remote players |
| Modify | `client/scripts/main.gd` | Debug state additions only if existing remote/session fields are insufficient |
| Modify | `client/scripts/multiplayer_sessions_panel.gd` | Row debug additions only if current rows lack mode/count/id data |
| Modify | `client/tests/test_client_bot.gd` | Static validation and runtime assertion tests |
| Modify | `client/tests/test_coop_client.gd` | Focused debug-state tests only if `main.gd` remote fields change |
| Verify | `tools/bot/scenarios/client/08_main_menu_flow.json` | Existing v45 menu proof remains green |
| Verify | `tools/bot/scenarios/client/20_menu_create_join_flow.json` | Existing v45 menu proof remains green |
| Verify | `tools/bot/scenarios/27_session_browser_uncapped_coop.json` | Existing backend listed co-op proof remains green |
| Audit | `server/internal/http/session.go` | No change expected; only touch for a real active-list/join guard |
| Audit | `server/internal/store/repos.go` | No change expected; active-list visibility already requires connected members |

## Task 1 - Preflight host helper

Files:
- Create: `tools/bot/client_join_preflight.py`
- Modify: `tools/bot/test_protocol.py`
- Reuse/Audit: `tools/bot/run.py`

- [x] Step 1.1: Add a focused Python helper that accepts `--base-url`, `--dev-token`,
  `--world-id`, `--seed`, `--email`, and `--metadata-file`.
- [x] Step 1.2: Use existing protocol-bot helpers or equivalent code to dev-login, ensure a host
  character, and create a listed co-op session for `dungeon_levels`.
- [x] Step 1.3: Open the host WebSocket with the returned `ws_url`, consume the initial
  `session_snapshot`, and send `client_ready` with a deterministic helper client version.
- [x] Step 1.4: Write metadata after the host is connected: `session_id`, `world_id`, host email,
  host character id, host local player id if known, and a ready flag.
- [x] Step 1.5: Keep pumping or holding the WebSocket until the process receives termination; close
  the WebSocket cleanly on exit.
- [x] Step 1.6: Add unit coverage for any pure config/metadata validation helpers. Avoid live
  network requirements in unit tests.

```bash
make test-py
```

## Task 2 - Client-bot runner preflight integration

Files:
- Modify: `scripts/bot_client.sh`
- Create/Use: `tools/bot/scenarios/client/21_join_game_listed_session.json`

- [x] Step 2.1: Extend client scenario parsing to accept an optional preflight block, initially
  only `{"type": "listed_coop_host"}`.
- [x] Step 2.2: For preflight scenarios, create a temporary metadata file and launch
  `tools/bot/client_join_preflight.py` before launching Godot.
- [x] Step 2.3: Wait until the metadata file reports readiness, with a bounded timeout and a clear
  failure message that includes helper logs when setup fails.
- [x] Step 2.4: Pass prepared metadata into the Godot guest process through environment variables,
  for example `ARPG_EXPECTED_JOIN_SESSION_ID` and optional host identifiers.
- [x] Step 2.5: Ensure the guest email remains distinct from the host email while preserving the
  existing scenario-email isolation behavior.
- [x] Step 2.6: Add cleanup traps so the preflight helper is terminated on success, Godot failure,
  validation failure, and shell interruption.
- [x] Step 2.7: Keep non-preflight scenarios byte-for-byte compatible in behavior.

```bash
HEADLESS=1 make bot-client scenario=20_menu_create_join_flow.json
```

## Task 3 - Client bot assertions and debug state

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/main.gd` only if existing debug state is insufficient
- Modify: `client/scripts/multiplayer_sessions_panel.gd` only if row debug state is insufficient
- Modify: `client/tests/test_client_bot.gd`
- Modify: `client/tests/test_coop_client.gd` only if `main.gd` debug state changes

- [x] Step 3.1: Extend `assert_multiplayer_session_rows` so it can require a row whose
  `session_id` equals an environment-provided expected id, plus `mode`, `listed`,
  `member_count`, and `connected_count` constraints.
- [x] Step 3.2: Extend `assert_current_session` so it can verify the current session id equals
  `ARPG_EXPECTED_JOIN_SESSION_ID` while still supporting existing `exists`, `mode`, and `listed`
  checks.
- [x] Step 3.3: Add a remote-player assertion such as `wait_remote_player_count` /
  `assert_remote_player_count`, using existing `remote_player_ids`, `party`, or `entities_debug`
  state.
- [x] Step 3.4: If the current debug state cannot distinguish local and remote player entities,
  add the smallest display-only field in `main.gd`; do not change protocol or server state.
- [x] Step 3.5: Add static validation tests for any new assertion fields or step types.
- [x] Step 3.6: Add runtime assertion tests with synthetic bot state for prepared session row,
  current session id, and remote player count.

```bash
make client-unit
```

## Task 4 - Godot guest Join Game scenario

Files:
- Create: `tools/bot/scenarios/client/21_join_game_listed_session.json`
- Keep green: `tools/bot/scenarios/client/08_main_menu_flow.json`
- Keep green: `tools/bot/scenarios/client/20_menu_create_join_flow.json`

- [x] Step 4.1: Add scenario metadata with `runner: "godot_client"`, `world_id:
  "dungeon_levels"`, a stable seed, and preflight type `listed_coop_host`.
- [x] Step 4.2: Start from `wait_main_menu` and assert root actions remain `Create Game`,
  `Join Game`, `Settings`, and `Exit`.
- [x] Step 4.3: Click `Join Game`, wait for the active-session panel, and assert the prepared
  listed session row exists with the expected session id, `mode: "coop"`, `listed: true`,
  `member_count >= 1`, and `connected_count >= 1`.
- [x] Step 4.4: Click Back and assert the root menu returns without creating or joining a session.
- [x] Step 4.5: Reopen `Join Game`, select the prepared row, and assert the character panel is
  shown only after the selected session id is present.
- [x] Step 4.6: Create a guest character if needed using the existing character panel path, then
  submit it to join the selected listed session.
- [x] Step 4.7: Wait for WebSocket open and assert current session id equals the prepared session
  id, `mode: "coop"`, and `listed: true`.
- [x] Step 4.8: Wait for the remote host player assertion to pass.

```bash
HEADLESS=1 make bot-client scenario=21_join_game_listed_session.json
```

## Task 5 - Backend audit and existing bot proofs

Files:
- Audit: `server/internal/http/session.go`
- Audit: `server/internal/store/repos.go`
- Audit: `server/internal/http/auth_session_test.go`
- Keep green: `tools/bot/scenarios/27_session_browser_uncapped_coop.json`

- [x] Step 5.1: Confirm no server change is needed: active listed sessions remain visible only
  when at least one member is connected.
- [x] Step 5.2: Confirm the active-session response exposes row data needed by the client bot:
  `session_id`, `world_id`, `mode`, `listed`, `member_count`, and `connected_count`.
- [x] Step 5.3: Confirm listed joins continue to work without join code and do not expose join code
  through the active list.
- [x] Step 5.4: Only if the audit finds a real bug, add the smallest HTTP/store fix and matching
  Go test.
- [x] Step 5.5: Run the existing protocol proof to ensure backend listed co-op semantics still
  pass.

```bash
go test ./internal/http/...
go test ./internal/store/...
make bot scenario=27_session_browser_uncapped_coop.json
```

## Task 6 - Regression verification for v45 menu paths

Files:
- Keep green: `tools/bot/scenarios/client/08_main_menu_flow.json`
- Keep green: `tools/bot/scenarios/client/20_menu_create_join_flow.json`
- Modify: `client/tests/test_client_bot.gd` only if assertions changed compatibility

- [x] Step 6.1: Run the v45 root/create menu scenarios to ensure preflight support did not change
  normal scenario execution.
- [x] Step 6.2: Keep compatibility aliases for old bot actions only if already present; do not add
  new player-facing legacy menu behavior.
- [x] Step 6.3: Fix any assertion drift by updating tests and scenario expectations together.

```bash
HEADLESS=1 make bot-client scenario=08_main_menu_flow.json
HEADLESS=1 make bot-client scenario=20_menu_create_join_flow.json
make client-smoke
```

## Task 7 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v46_2026-06-09-client-join-game-proof.md`
- Modify: `docs/specs/v46_spec-client-join-game-proof.md` only if accepted clarifications are discovered

- [x] Step 7.1: When implementation finishes, add v46 to the slice numbering note and lifecycle
  table.
- [x] Step 7.2: Add a concise `v46 - Real Godot Join Game co-op proof` summary under "What each
  slice proved".
- [x] Step 7.3: Update the scripted scenario catalog with `join_game_listed_session`.
- [x] Step 7.4: Move any newly deferred lobby/test-harness items to Open gaps.
- [x] Step 7.5: Keep this plan's checkboxes accurate during execution.

```bash
make ci
```

## Final verification

- [x] `make test-py`
- [x] `make client-unit`
- [x] `go test ./internal/http/...` if server HTTP code changed (not needed; no server HTTP code changed, covered by `make ci` `go test ./...`)
- [x] `go test ./internal/store/...` if store active-list code changed (not needed; no store code changed, covered by `make ci` `go test ./...`)
- [x] `make bot scenario=27_session_browser_uncapped_coop.json`
- [x] `HEADLESS=1 make bot-client scenario=21_join_game_listed_session.json`
- [x] `HEADLESS=1 make bot-client scenario=08_main_menu_flow.json`
- [x] `HEADLESS=1 make bot-client scenario=20_menu_create_join_flow.json`
- [x] `make client-smoke` (via `make ci`)
- [x] `make ci`

## Deferred scope

- Offline/local-only gameplay remains out of scope.
- Steam lobbies, invites, friend flows, matchmaking, chat, ready checks, party staging, and
  filters/search/sorting remain out of scope.
- Two-window Godot visual choreography remains out of scope; the required proof is one protocol
  host plus one Godot guest.
- Server gameplay protocol/schema changes remain out of scope.
- Production multiplayer UI art/audio remains out of scope.
