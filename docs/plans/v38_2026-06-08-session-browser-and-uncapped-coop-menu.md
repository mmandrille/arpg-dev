# v38 Plan — Multiplayer Menu Session Browser

Status: Ready for implementation
Goal: Make co-op joinable from the main menu through active listed sessions, remove the hard two-player cap, and add local/remote multi-client launchers.
Architecture: The Go server remains authoritative for session membership, player entity creation, realtime fanout, persistence, and replay. Listed sessions are a platform/API affordance layered on top of existing `mode: "coop"` sessions; joining a listed session skips manual join-code entry but still validates authenticated character ownership. The current multi-player `Sim` model is reused and audited for N members, with deterministic ordering kept host first, then joined tick/account/character. Godot stays a thin menu/input/render client that lists, creates, joins, then connects the existing WebSocket.
Tech stack: Go HTTP/store/realtime/replay, Postgres migrations, Godot GDScript Control UI, Python protocol bot and launcher scripts, Make targets.

## Baseline and shortcut decision

Baseline is v37 `combat-control-and-boss-ai-fixes` on `main`; `make ci` was green on 2026-06-08.

This plan builds directly on:

- v24 main menu and character picker for player-facing startup.
- v33 authoritative co-op sessions, session members, actor-tagged inputs, and recipient-scoped snapshots.
- v34 remote player presentation for non-local player entities.
- v37 current gameplay/control baseline.

Godot plugin adoption checklist result:

| Candidate | Decision | Reason |
|-----------|----------|--------|
| Godot high-level multiplayer / ENet / `@rpc` | Reject | Game state remains Go-authoritative over the existing WebSocket protocol. |
| Steam lobby templates / GodotSteam | Reject for v38 | Useful later for Steam invites, but this slice only needs backend-listed sessions. |
| Godot lobby/menu plugin | Reject | Current Control-node menu is enough, and a plugin would add dependency surface without solving server discovery. |
| Existing in-repo menu panels | Reuse/borrow | Matches v24 patterns and keeps headless client bot coverage straightforward. |
| Character profile hotkey panel | Reuse/borrow | Small display-only Control panel; existing HUD/panel patterns cover name, level, and area without a plugin. |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/plans/v38_2026-06-08-session-browser-and-uncapped-coop-menu.md` | Implementation plan |
| Modify | `docs/specs/v38_spec-session-browser-and-uncapped-coop-menu.md` | Only if implementation discovers small contract clarifications |
| Modify | `docs/PROGRESS.md` | Lifecycle update when v38 ships |
| Modify | `README.md` | `make play`, `make play N`, and `make play-remote N` usage |
| Create | `server/migrations/0008_listed_sessions.sql` | Add listed-session storage/indexes |
| Modify | `server/internal/store/models.go` | `Session.Listed`, `SessionSummary`, comments without two-player wording |
| Modify | `server/internal/store/interfaces.go` | list active sessions repository method |
| Modify | `server/internal/store/repos.go` | listed create/load, active list query, remove party cap |
| Modify | `server/internal/store/store_test.go` | listed list, uncapped membership, duplicate guards |
| Modify | `server/internal/http/session.go` | listed create/join/list API and response fields |
| Modify | `server/internal/http/auth_session_test.go` | HTTP API behavior and no cap |
| Modify | `server/internal/http/ws_test.go` | three-client WebSocket proof |
| Modify | `server/internal/realtime/session_loop.go` | N-member display names/connectivity/fanout audit |
| Modify | `server/internal/replay/replay.go` | N-member replay audit if tests reveal drift |
| Modify | `server/internal/replay/replay_test.go` | three-member replay proof |
| Modify | `server/internal/game/game_test.go` | three-player same-level sim proof if needed |
| Modify | `client/scripts/net_client.gd` | active session list, create listed co-op, join listed co-op, URL parsing tests if possible |
| Modify | `client/scripts/main_menu.gd` | add Multiplayer entry |
| Modify | `client/scripts/main.gd` | host/join routing and panel wiring |
| Modify/Create | `client/scripts/multiplayer_sessions_panel.gd` | active session browser panel |
| Modify | `client/scripts/character_select_panel.gd` | reuse for multiplayer host/join selection if needed |
| Modify | `client/scripts/bot_controller.gd` | menu bot actions for multiplayer panel |
| Modify | `client/scripts/bot_scenario_runner.gd` | menu assertions for listed sessions |
| Modify | `client/tests/test_client_bot.gd` | client scenario validation |
| Modify | `client/tests/test_coop_client.gd` | more than one remote player rendering/unit proof |
| Modify | `make/client.mk` | `play-remote` target and numeric goal handling |
| Modify | `scripts/play.sh` | local multi-client menu launch without pre-created co-op |
| Create | `scripts/play_remote.sh` | remote backend multi-client menu launcher |
| Modify | `tools/play/coop_setup.py` | keep as explicit dev/debug helper or remove from `make play` path |
| Modify | `tools/bot/run.py` | listed co-op and three-peer scenario support |
| Modify | `tools/bot/test_protocol.py` | helper tests |
| Create | `tools/bot/scenarios/27_session_browser_uncapped_coop.json` | protocol bot proof |
| Create/Defer | `tools/bot/scenarios/client/14_multiplayer_menu_session_browser.json` | client menu proof if reliable |

## Task 1 — Store and HTTP Session Contracts

Files:
- Create: `server/migrations/0008_listed_sessions.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 1.1: Add a persisted `sessions.listed BOOLEAN NOT NULL DEFAULT FALSE` column and useful active-list indexes. Keep existing rows private/non-listed by default.

```bash
cd server && go test ./internal/store/... -run TestAccountCharacterSessionFlow -count=1
```

- [x] Step 1.2: Add `Listed` to `store.Session`, add a `SessionSummary` model, update `CreateSession` / `GetSession` to round-trip `listed`, and update comments that still say co-op is host plus one guest.

```bash
cd server && go test ./internal/store/... -run 'Session|Coop' -count=1
```

- [x] Step 1.3: Add `ListActiveListedSessions(ctx)` to `SessionRepo` and implement it with stable ordering: active listed co-op sessions only, newest `updated_at` first, `session_id` tie-break. Include host character display name, member count, connected count, `created_at`, and `updated_at`.

```bash
cd server && go test ./internal/store/... -run 'Active|Listed|Session' -count=1
```

- [x] Step 1.4: Remove the hard `count >= 2` cap from `CreateSessionGuestMember`. Keep duplicate account/character rejection and session active/mode checks. Remove or retire `ErrPartyFull` if it becomes unused.

```bash
cd server && go test ./internal/store/... -run 'Coop|Member|Session' -count=1
```

- [x] Step 1.5: Extend `createSessionRequest` / `createSessionResponse` with `listed`, return it from `sessionResponse`, and set `listed: true` only when requested by menu multiplayer create.

```bash
cd server && go test ./internal/http/... -run 'Create.*Session|Session.*Listed' -count=1
```

- [x] Step 1.6: Add authenticated `GET /v0/sessions/active` returning the spec summary shape and excluding solo, ended, and non-listed sessions. Ensure raw join codes and account ids are never exposed.

```bash
cd server && go test ./internal/http/... -run 'Active|Listed|Session' -count=1
```

- [x] Step 1.7: Update `POST /v0/sessions/{session_id}/join` so listed sessions accept `{ "character_id": "..." }` without `join_code`, while non-listed/private sessions keep the v33 join-code requirement. Preserve cross-account character rejection and private-session non-enumeration.

```bash
cd server && go test ./internal/http/... -run 'Join|Coop|Session' -count=1
```

## Task 2 — N-Member Realtime and Replay Proof

Files:
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `server/internal/http/ws_test.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/replay/replay_test.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 2.1: Audit `session_loop.go` startup, late join, disconnect, `membersByPlayerID`, fanout, and persistence paths for assumptions that only one non-host member exists.

```bash
cd server && go test ./internal/realtime/... ./internal/http/... -run 'Coop|Session' -count=1
```

- [x] Step 2.2: Adjust display-name fallback for non-host members so the party/list remains readable for multiple guests. Prefer character names when readily available from HTTP list summaries; otherwise keep deterministic labels and avoid relying on uniqueness.

```bash
cd server && go test ./internal/http/... ./internal/realtime/... -run 'Coop|Session' -count=1
```

- [x] Step 2.3: Add a three-client WebSocket test: host creates listed co-op, two additional accounts join through listed join, all three connect, receive distinct `local_player_id`, descend/regroup as needed, and see three same-level player entities.

```bash
cd server && go test ./internal/http/... -run 'Three|Coop|WebSocket|Session' -count=1
```

- [x] Step 2.4: Add or update focused `game` tests for three players on one level if current coverage only proves two-player same-level visibility. Assert actor-scoped movement for the third player without exact tuning locks.

```bash
cd server && go test ./internal/game/... -run 'Coop|Three|Player' -count=1
```

- [x] Step 2.5: Add replay coverage with three members and actor-tagged inputs. Verify deterministic member ordering, correct `local_player_id` per member, party length >= 3, and actor-scoped inventory/hotbar/progression persistence.

```bash
cd server && go test ./internal/replay/... -run 'Coop|Three|Replay|Session' -count=1
```

- [x] Step 2.6: Replace tests that expect `party_full` on a third join with tests proving third and fourth joins are accepted, while duplicate account/character still conflict.

```bash
cd server && go test ./internal/store/... ./internal/http/... -run 'PartyFull|Coop|Join|Session' -count=1
```

## Task 3 — Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/27_session_browser_uncapped_coop.json`

- [x] Step 3.1: Add Python bot helpers for `create_listed_coop_session`, `list_active_sessions`, and `join_listed_session` without join code.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 3.2: Generalize the current co-op bot driver beyond host+guest: accept a peer count from scenario data, create/login accounts, ensure characters, connect all peers, and keep per-peer runtime state.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 3.3: Create `27_session_browser_uncapped_coop.json` with at least three accounts. Steps must prove listed create, active list visibility, listed join for two peers, distinct local players, same-level visibility, independent movement, one disconnect, remaining peers still active, reconnect if cheap, and replay verification.

```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=session_browser_uncapped_coop
```

- [x] Step 3.4: Keep existing `23_true_coop_session.json` green as a compatibility proof for the private join-code flow.

```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=true_coop_session
```

## Task 4 — Godot Client Multiplayer Menu

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/main_menu.gd`
- Modify: `client/scripts/main.gd`
- Modify/Create: `client/scripts/multiplayer_sessions_panel.gd`
- Modify: `client/scripts/character_select_panel.gd`
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 4.1: Add `NetClient.list_active_sessions()`, `NetClient.create_listed_coop_session(character_id)`, and `NetClient.join_listed_session(session_id, character_id)`. Preserve existing solo `create_session()` behavior.

```bash
make client-unit
```

- [x] Step 4.2: Harden `NetClient` URL parsing for full `BASE_URL` values used by `play-remote`, including `http://host:port`, `https://host`, optional path stripping if present, and correct `ws://` / `wss://` WebSocket construction.

```bash
make client-unit
```

- [x] Step 4.3: Add a Multiplayer button to the main menu and a simple active-session browser panel with Host Listed Session, Refresh Sessions, Join Selected, and Back.

```bash
make client-unit
```

- [x] Step 4.4: Wire host flow: Multiplayer -> Host Listed Session -> choose/create character -> `create_listed_coop_session` -> `_begin_gameplay_connection(false)`.

```bash
make client-unit
```

- [x] Step 4.5: Wire join flow: Multiplayer -> Refresh Sessions -> choose row -> choose/create character -> `join_listed_session` -> `_begin_gameplay_connection(false)`.

```bash
make client-unit
```

- [x] Step 4.6: Ensure menu/pause/input locks still block gameplay while multiplayer panels are visible. Existing Continue/New Game solo flows must remain unchanged.

```bash
make client-unit
```

- [x] Step 4.7: Extend client co-op unit coverage so one local player plus at least two remote players render/update/remove independently by entity id.

```bash
make client-unit
```

## Task 5 — Client Bot Menu Proof

Files:
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`
- Create/Defer: `tools/bot/scenarios/client/14_multiplayer_menu_session_browser.json`

- [x] Step 5.1: Add bot-controller actions for opening Multiplayer, refreshing sessions, hosting listed session, selecting/joining the first listed session, and asserting multiplayer panel/session-row state.

```bash
make client-unit
```

- [x] Step 5.2: Add `test_client_bot.gd` validation for the new scenario step names and required fields.

```bash
make client-unit
```

- [x] Step 5.3: If reliable with one Godot process, create `14_multiplayer_menu_session_browser.json` to prove menu host/list/join routing against a prepared backend session. If not reliable, document the deferral in this plan during execution and rely on Task 3 protocol proof plus Task 4 unit coverage.

Execution note: deferred the single-process Godot menu scenario for v38. A single authenticated Godot bot process can exercise host or join routing, but cannot reliably prove a second-account listed session browser join without external backend preparation and cross-process coordination. Coverage for v38 is the Task 3 three-peer protocol bot plus Task 4/5 client unit coverage for menu routing, session rows, and multiple remote-player presentation.

```bash
ADDR=:18081 BASE_URL=http://localhost:18081 HEADLESS=1 make bot-client scenario=multiplayer_menu_session_browser
```

- [x] Step 5.4: Re-run existing menu scenario to ensure v24 flow remains intact.

```bash
HEADLESS=1 make bot-client scenario=08_main_menu_flow.json
```

## Task 6 — Launchers and Remote Play

Files:
- Modify: `make/client.mk`
- Modify: `scripts/play.sh`
- Create: `scripts/play_remote.sh`
- Modify: `tools/play/coop_setup.py`
- Modify: `README.md`

- [x] Step 6.1: Update `make/client.mk` to support numeric goals for both `play` and `play-remote`, and add a `play-remote` target that does not depend on `db-up`.

```bash
make help
```

- [x] Step 6.2: Change `scripts/play.sh` so `PLAY_CLIENTS > 1` launches independent menu clients with distinct `ARPG_EMAIL` values, no `ARPG_AUTOSTART`, and no pre-created co-op session. Keep single-client behavior compatible.

```bash
GODOT=/usr/bin/true make play 3
```

- [x] Step 6.3: Add `scripts/play_remote.sh`: validate `PLAY_CLIENTS`, validate `GODOT`, probe `$BASE_URL/readyz`, import assets, then launch N clients with distinct emails and `ARPG_BASE_URL="$BASE_URL"`. Do not build/start server and do not touch local DB.

```bash
GODOT=/usr/bin/true BASE_URL=http://localhost:8080 make play-remote 3
```

- [x] Step 6.4: Keep `tools/play/coop_setup.py` only as an explicit dev/debug helper if still useful; remove it from the normal `make play N` path.

```bash
rg -n "coop_setup|ARPG_AUTOSTART|ARPG_SESSION_ID" scripts make tools/play
```

- [x] Step 6.5: Update `README.md` with `make play`, `make play 3`, and `BASE_URL=<url> make play-remote 3` examples.

```bash
rg -n "play-remote|make play 3|BASE_URL" README.md
```

## Task 7 — Lifecycle Docs and CI

Files:
- Modify: `docs/PROGRESS.md`
- Modify: `docs/specs/v38_spec-session-browser-and-uncapped-coop-menu.md` only if implementation clarifies contract details

- [x] Step 7.1: Update `docs/PROGRESS.md` lifecycle table with v38 once implementation is complete.

- [x] Step 7.2: Add a concise v38 "What this slice proved" entry covering listed sessions, uncapped co-op membership, menu host/join, and `play-remote`.

- [x] Step 7.3: Move any newly deferred scope into the Open gaps table, especially filters/search, Steam invites, load-aware caps, and richer party UI.

```bash
rg -n "v38|session-browser-and-uncapped-coop-menu|play-remote" docs/PROGRESS.md
```

## Final verification

- [x] Store/http focused tests:

```bash
cd server && go test ./internal/store/... ./internal/http/... -run 'Session|Coop|Active|Join' -count=1
```

- [x] Game/realtime/replay focused tests:

```bash
cd server && go test ./internal/game/... ./internal/realtime/... ./internal/replay/... -run 'Coop|Replay|Session|Three' -count=1
```

- [x] Python bot helper tests:

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Protocol bot proof:

```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=session_browser_uncapped_coop
```

- [x] Existing private co-op bot proof:

```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=true_coop_session
```

- [x] Client unit tests:

```bash
make client-unit
```

- [x] Client menu bot proof if implemented: deferred/not applicable per Step 5.3; covered by the Task 3 protocol bot proof plus Task 4/5 unit coverage.

```bash
ADDR=:18081 BASE_URL=http://localhost:18081 HEADLESS=1 make bot-client scenario=multiplayer_menu_session_browser
```

- [x] Client smoke:

```bash
make client-smoke
```

- [x] Launcher smoke with no real Godot launch:

```bash
GODOT=/usr/bin/true make play 3
GODOT=/usr/bin/true BASE_URL=http://localhost:8080 make play-remote 3
```

- [x] Full CI:

```bash
make ci
```

## Deferred scope

- Session filters/search/sorting controls.
- Steam lobby, Steam invites, friend flows, and platform identity.
- Ready checks, chat, emotes, party panel polish, and invite UI.
- Trade, XP sharing, party bonuses, proximity rewards, loot allocation, PvP, and friendly fire.
- Load-aware session capacity limits and multi-process session ownership.
