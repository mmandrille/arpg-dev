# Spec: `session-browser-and-uncapped-coop-menu`

Status: Draft
Branch: `main`
Slice: v38 - multiplayer menu session browser, listed co-op sessions, uncapped party size, and remote play launcher
Baseline: v37 `combat-control-and-boss-ai-fixes`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, thin client, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - co-op players share one `Sim`
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - multiplayer menu/lobby shortcut review
- [`v24_spec-main-menu-and-character-start.md`](v24_spec-main-menu-and-character-start.md) - main menu and character start shell
- [`v33_spec-true-coop-session.md`](v33_spec-true-coop-session.md) - authoritative co-op session foundation
- [`v37_spec-combat-control-and-boss-ai-fixes.md`](v37_spec-combat-control-and-boss-ai-fixes.md) - current gameplay baseline

## 1. Purpose

Co-op is now authoritative, but it is not player-facing from the menu. `make play 3` currently
pre-creates one co-op session through a helper script and launches every Godot client with
`ARPG_AUTOSTART=1` and the same `ARPG_SESSION_ID`. That proves the server path, but it bypasses the
menu and prevents each client from choosing whether to host or join.

This slice adds a thin multiplayer menu loop:

- A player can create a listed co-op session from the main menu.
- A player can see active listed sessions from the backend and join one.
- Joining a listed session does not require manually copying a join code.
- Co-op sessions no longer have an artificial application-level player cap.
- `make play 3` launches three independent clients at the menu against the local backend.
- `make play-remote 3 BASE_URL=<url>` launches three independent menu clients against an already
  running backend without starting local Postgres or a local server.

The goal is not production matchmaking. The goal is to make the existing authoritative co-op model
usable from the normal startup shell, while keeping the Go server as the only owner of sessions,
membership, player entities, persistence, combat, loot, replay, and WebSocket fanout.

## 2. Non-goals

- No Steam lobby, Steam invites, friend list, account discovery, or platform identity integration.
- No session filters, search, sorting UI beyond a stable simple default. Filters/search are future
  work.
- No chat, emotes, ready checks, lobby staging room, party role selection, or party UI polish.
- No trade protocol, XP sharing, party bonuses, proximity reward rules, loot allocation rules, PvP,
  or friendly fire.
- No performance cap or fixed maximum player count in this slice. If real load problems appear,
  load-aware limits and capacity planning are future work.
- No distributed session ownership, multi-process realtime routing, shard assignment, or split
  deployables.
- No production menu art/audio.
- No Godot high-level multiplayer, ENet sync, or peer-authoritative game state.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v38_spec-session-browser-and-uncapped-coop-menu.md     - this slice contract
docs/plans/v38_<YYYY-MM-DD>-session-browser-and-uncapped-coop-menu.md - implementation plan
PROGRESS.md                                                   - lifecycle update when v38 ships
README.md                                                          - manual play and remote play command docs

server/migrations/*.sql                                            - listed session field/indexes if schema changes
server/internal/store/models.go                                    - listed/session summary models and guest role semantics
server/internal/store/interfaces.go                                - list active sessions and uncapped member create contract
server/internal/store/repos.go                                     - remove two-player cap, list active listed sessions
server/internal/store/store_test.go                                - uncapped membership and active-list coverage

server/internal/http/session.go                                    - listed create, active session list, listed join behavior
server/internal/http/auth_session_test.go                          - list/join/visibility/account isolation tests
server/internal/http/ws_test.go                                    - three-client WebSocket/session visibility coverage

server/internal/realtime/session_loop.go                           - audit N-member startup, late join, fanout, disconnect
server/internal/realtime/hub.go                                    - connected-member guard remains per account/character
server/internal/replay/replay.go                                   - audit N-member reconstruction and connectivity
server/internal/replay/replay_test.go                              - three-member replay coverage

server/internal/game/sim.go                                        - audit party view, same-level visibility, monster targeting
server/internal/game/game_test.go                                  - three-player same-level movement/combat proof if needed

client/scripts/net_client.gd                                       - list active sessions, create listed co-op, join listed session
client/scripts/main_menu.gd                                        - multiplayer entry points
client/scripts/main.gd                                             - multiplayer menu routing and connection start
client/scripts/character_select_panel.gd                           - character choice before host/join if reused
client/scripts/multiplayer_sessions_panel.gd                       - active session browser if separated
client/scripts/bot_controller.gd                                   - multiplayer menu bot helpers
client/scripts/bot_scenario_runner.gd                              - multiplayer menu assertions
client/tests/test_client_bot.gd                                    - client scenario validation
client/tests/test_coop_client.gd                                   - N remote-player rendering/unit coverage

make/client.mk                                                     - add play-remote target and preserve make play N
scripts/play.sh                                                    - local server path launches N menu clients, no pre-created co-op
scripts/play_remote.sh                                             - remote backend path, no local db/server
tools/play/coop_setup.py                                           - delete or keep only for explicit dev/debug use

tools/bot/run.py                                                   - three-account listed co-op protocol scenario support
tools/bot/test_protocol.py                                         - helper coverage
tools/bot/scenarios/27_session_browser_uncapped_coop.json          - protocol proof
tools/bot/scenarios/client/14_multiplayer_menu_session_browser.json - client menu proof if reliable
```

## 4. Data shapes

### 4.1 Listed session create

`POST /v0/sessions` keeps existing solo and co-op behavior. For menu-created multiplayer sessions,
the client sends `mode: "coop"` and `listed: true`.

```json
{
  "mode": "coop",
  "listed": true,
  "world_id": "dungeon_levels",
  "character_id": "char_..."
}
```

Response keeps the current create response shape and may still include a `join_code` for private
copy/paste or dev flows:

```json
{
  "session_id": "sess_...",
  "character_id": "char_...",
  "seed": "0123456789abcdef",
  "world_id": "dungeon_levels",
  "mode": "coop",
  "listed": true,
  "join_code": "join_...",
  "ws_url": "/v0/ws?session_id=sess_..."
}
```

Rules:

- `mode: "solo"` stays default for existing automation and menu Continue/New Game.
- Menu multiplayer create sends `mode: "coop"`, `world_id: "dungeon_levels"`, selected
  `character_id`, and `listed: true`.
- Listed co-op sessions appear in the active session list while `status: "active"` and at least
  one active member is currently connected.
- When the last connected member leaves a listed co-op session, the backend ends that session so it
  disappears from discovery and cannot be joined again.
- Existing non-listed join-code behavior remains available.
- The server stores only a hash of any join code, as in v33.

### 4.2 Active listed sessions

New authenticated endpoint:

```text
GET /v0/sessions/active
```

Response:

```json
{
  "sessions": [
    {
      "session_id": "sess_...",
      "world_id": "dungeon_levels",
      "mode": "coop",
      "listed": true,
      "host_character_id": "char_host",
      "host_display_name": "Mara",
      "member_count": 3,
      "connected_count": 2,
      "created_at": "2026-06-08T12:00:00Z",
      "updated_at": "2026-06-08T12:03:00Z"
    }
  ]
}
```

Rules:

- Only active listed co-op sessions with `connected_count > 0` are returned.
- Ended sessions are not returned.
- Empty listed co-op sessions are not returned, even if legacy rows still have `status: "active"`.
- Solo sessions are not returned.
- Search and filters are deferred; default order should be stable and useful, preferably newest
  `updated_at` first with `session_id` as a tie-breaker.
- The list must not expose raw join codes or account ids.
- `member_count` counts active session members, whether connected or temporarily disconnected.
- `connected_count` counts currently connected members.

### 4.3 Join listed session

Joining from the session browser should not require the player to manually copy a join code. The
implementation can either add a dedicated route or extend the existing join route with listed
semantics. Preferred smallest contract:

```text
POST /v0/sessions/{session_id}/join
```

Request for a listed session:

```json
{
  "character_id": "char_..."
}
```

Request for a private/non-listed session remains the existing join-code shape:

```json
{
  "join_code": "join_...",
  "character_id": "char_..."
}
```

Response:

```json
{
  "session_id": "sess_...",
  "character_id": "char_...",
  "seed": "0123456789abcdef",
  "world_id": "dungeon_levels",
  "mode": "coop",
  "listed": true,
  "ws_url": "/v0/ws?session_id=sess_..."
}
```

Rules:

- If `listed: true`, `join_code` is not required.
- If `listed: false`, the existing join-code validation remains required.
- `character_id` must belong to the authenticated account.
- A same account or same character cannot join the same session twice as separate members.
- A second simultaneous WebSocket for the same session member remains rejected with
  `member_already_connected`.
- There is no fixed player-count cap.
- Joining an ended session returns `409 session_ended`.
- Joining a non-existent, non-listed, or unauthorized session must avoid leaking private session
  existence beyond what the active list already exposes.

### 4.4 Uncapped session membership

v33 intentionally capped co-op at two players. v38 removes that cap.

Persistence model stays one row per member:

```text
session_members
  session_id
  account_id
  character_id
  player_entity_id
  role
  status
  connected
  current_level
  joined_tick
  left_tick
  joined_at
  updated_at
```

Rules:

- Exactly one member has role `host`.
- Every non-host member may keep role `guest` for now; no unique guest role is required.
- Member ordering for replay and party display remains deterministic:
  host first, then joined tick, then account id, then character id.
- Player entity IDs remain deterministic from session member order and late-join order.
- There is no schema or application constant that rejects member count above two.

### 4.5 BASE_URL and WebSocket URL handling

`BASE_URL` remains the launcher/setup name.

Rules:

- Local play defaults `BASE_URL` to `http://localhost:8080`.
- Godot continues receiving `ARPG_BASE_URL`.
- HTTP requests use `BASE_URL` / `ARPG_BASE_URL` for auth, character, session list, create, join,
  and end-session calls.
- WebSocket URLs are derived from the same full base URL and the server-provided `ws_url`.
- `http://` maps to `ws://`; `https://` maps to `wss://`.
- Host and port parsing must support full backend URLs, including explicit ports.

## 5. Architecture and flow

### 5.1 Local `make play 3`

```text
make play 3
  -> db-up
  -> build/start local Go server
  -> wait for BASE_URL/readyz
  -> launch 3 Godot clients
  -> each client logs in with a distinct dev email
  -> each client opens the main menu
  -> any client can create a listed co-op session
  -> other clients can refresh active sessions and join from the list
  -> each client opens its own WebSocket after host/join succeeds
```

Important change: `make play N` must not pre-create a co-op session or force
`ARPG_AUTOSTART=1` for multi-client launches. It launches menu clients.

### 5.2 Remote `make play-remote 3`

```text
BASE_URL=https://example-backend make play-remote 3
  -> do not run db-up
  -> do not build/start local Go server
  -> check BASE_URL/readyz if available
  -> launch 3 Godot clients with ARPG_BASE_URL=BASE_URL
  -> each client opens the main menu
```

Rules:

- `play-remote` is a convenience launcher for an already running backend.
- `BASE_URL` still defaults to `http://localhost:8080`, but remote usage should document overriding
  it explicitly.
- `play-remote` must not depend on local Postgres, local migrations, or `scripts/play.sh` server
  startup.

### 5.3 Menu host flow

```text
Main menu
  -> Multiplayer
  -> Host Listed Session
  -> choose or create character
  -> POST /v0/sessions { mode: "coop", listed: true, world_id: "dungeon_levels", character_id }
  -> connect WebSocket
  -> gameplay starts in town
```

### 5.4 Menu join flow

```text
Main menu
  -> Multiplayer
  -> Browse Active Sessions
  -> GET /v0/sessions/active
  -> choose a session
  -> choose or create character
  -> POST /v0/sessions/{session_id}/join { character_id }
  -> connect WebSocket
  -> joined player starts in town
```

### 5.5 Existing solo flow

Continue and New Game from v24 remain solo/fresh-session flows unless the player explicitly enters
the multiplayer menu. Existing bot/smoke/dev automation paths remain explicit through environment
variables.

## 6. Game and server rules

### 6.1 No fixed party cap

The server must remove the v33 two-player cap from the join path. The implementation may still be
limited by actual process resources, database resources, local machine capacity, and any future
deployment constraints. It must not reject the third, fourth, or later member because of a hardcoded
game rule.

Any existing tests or comments that assert a third member receives `party_full` must be updated to
the v38 contract.

### 6.2 N-player realtime behavior

All v33 co-op rules generalize to N members:

- Each connected member controls exactly one player entity.
- The server derives actor identity from the authenticated connection.
- Level visibility remains recipient-scoped.
- Same-level players see each other.
- Different-level players remain invisible in entity lists but appear in party metadata.
- Player/player collision remains non-solid.
- Disconnect removes only that member's player entity and keeps the session alive while at least
  one member remains connected.
- Reconnect restores that member's actor binding and respawns them in town.
- The session ends only when existing end-session rules decide it has ended; one member cannot end
  the session for everyone while others remain connected.

### 6.3 N-player replay

Replay must reconstruct every persisted session member, not only host plus one guest.

Rules:

- Host start snapshot loads first.
- Non-host start snapshots load in deterministic member order.
- Actor-tagged inputs replay against the correct player entity for every member.
- Current member connectivity is applied for all members.
- Replay and `/state` remain deterministic for three or more members.

## 7. Client behavior

### 7.1 Main menu

The main menu should expose multiplayer without replacing existing solo controls. Suggested first
shape:

```text
Continue
New Game
Multiplayer
Settings
Exit
```

The multiplayer panel should support:

- Host Listed Session.
- Refresh Sessions.
- Join selected listed session.
- Back.

The UI can be plain and functional. It should use existing in-repo Control nodes and styling. No
plugin adoption is required for v38 unless the plan finds a low-cost display-only menu helper.

### 7.2 Character selection

Host and join both need a character. The plan can choose one of these small implementations:

1. Reuse the current character picker panel before host/join.
2. Add a compact picker inside the multiplayer panel.

Default: reuse the current character picker to avoid duplicating character creation/list behavior.

### 7.3 Session list rendering

Each session row should show stable, useful summary text:

- Host display name.
- Member count and connected count.
- World id or simple label.
- Created/updated recency if cheap.

No search box, filter panel, pagination, avatars, portraits, or rich party panel in v38.

### 7.4 Remote players

The existing v33/v34 remote player presentation must work for more than one remote player:

- Local player remains `PlayerAnchor`.
- Every visible remote player is rendered as a remote player node.
- Camera and local HP UI stay bound to `local_player_id`.
- Remote player movement, hit, death, remove, and reconnect behavior apply by entity id.

## 8. Plugin and shortcut decision

The required project research document was checked for this spec. Decision for v38:

| Candidate | Decision | Reason |
|-----------|----------|--------|
| Godot high-level multiplayer / ENet / `@rpc` | Reject | Conflicts with the authoritative Go sim, WebSocket protocol, and replay. |
| Steam lobby templates / GodotSteam | Reject for v38 | Useful later for Steam invites, but listed backend sessions are the thin vertical slice now. |
| Godot multiplayer/lobby plugin | Reject | No plugin should own session membership, game state sync, combat, inventory, or persistence. |
| Existing Godot Control menu code | Reuse | Current UI is enough for a functional browser without dependency risk. |

The v38 plan must record the adoption checklist result. If a plugin or demo is considered during
planning, document adopt/borrow/reject there.

## 9. Acceptance criteria

1. Interactive `make play` still starts local Postgres, starts the local server, and opens one
   Godot client at the main menu.
2. `make play 3` starts local Postgres, starts one local server, and opens three Godot clients at
   the main menu without pre-creating a session or setting `ARPG_AUTOSTART=1`.
3. `make play-remote` opens one Godot client against `BASE_URL` without starting local Postgres or
   the local server.
4. `make play-remote 3 BASE_URL=<url>` opens three Godot clients against the configured backend URL
   without local server startup.
5. `BASE_URL` remains the public launcher variable; Godot receives the same backend as
   `ARPG_BASE_URL`.
6. WebSocket connection construction works from both `http://host:port` and `https://host` base
   URLs by selecting `ws://` or `wss://` correctly.
7. `POST /v0/sessions` can create a listed co-op session for an owned character.
8. `GET /v0/sessions/active` returns active listed co-op sessions and excludes solo, ended, and
   non-listed sessions.
9. Active session list rows include session id, world id, host display name, member count, connected
   count, and timestamps.
10. The main menu exposes a multiplayer flow to host a listed session.
11. The main menu exposes a multiplayer flow to refresh and join an active listed session.
12. Joining a listed session from the menu does not require manually entering a join code.
13. Joining a non-listed/private session still requires the existing join code path.
14. The server rejects cross-account character use for listed create and join.
15. The same account or same character cannot join the same session twice as separate members.
16. A third authenticated account can join the same listed co-op session.
17. A fourth or later authenticated account is not rejected by an application-level party cap.
18. Three connected clients receive distinct `local_player_id` values for the same session.
19. Three players on the same level appear in each same-level recipient snapshot as three player
   entities.
20. Movement from any one of the three clients updates only that client's actor player.
21. Disconnecting one of three clients removes only that player's visible entity and leaves the
   other two clients connected to the same session loop.
22. Replay reconstructs a three-member co-op session deterministically from persisted members,
   start snapshots, and actor-tagged inputs.
23. Existing solo session create, main-menu Continue/New Game, protocol bot, replay, client smoke,
   and visual replay paths remain green.
24. `make ci` passes.

## 10. Testing plan

1. Store/http focused tests:

```bash
cd server && go test ./internal/store/... ./internal/http/... -run 'Session|Coop|Active|Join' -count=1
```

Coverage:

- listed create/list,
- non-listed exclusion,
- ended exclusion,
- cross-account character rejection,
- third and fourth member join,
- duplicate account/character rejection.

2. Game/realtime/replay focused tests:

```bash
cd server && go test ./internal/game/... ./internal/realtime/... ./internal/replay/... -run 'Coop|Replay|Session' -count=1
```

Coverage:

- N-player same-level visibility,
- actor-scoped movement,
- disconnect/reconnect,
- three-member replay.

3. Protocol bot proof:

```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=session_browser_uncapped_coop
```

The scenario must use at least three authenticated accounts and selected/default characters. It
should create one listed co-op session, list it, join two additional members through listed join,
connect three WebSockets, assert distinct local players, assert same-level visibility, move each
actor independently, disconnect one member, and verify replay.

4. Client unit/bot proof:

```bash
make client-unit
ADDR=:18081 BASE_URL=http://localhost:18081 HEADLESS=1 make bot-client scenario=multiplayer_menu_session_browser
```

If multi-process Godot menu proof is too brittle, the plan may substitute focused client unit tests
for list rendering and menu routing plus the protocol bot's three-client proof, but must explain
the deferral.

5. Launcher manual checks:

```bash
make play 3
BASE_URL=http://localhost:8080 make play-remote 3
```

Manual `play-remote` assumes a backend is already running. Verify the launcher does not start a
local server or local Postgres.

6. Final gate:

```bash
make ci
```

## 11. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Menu-created co-op sessions are listed by default. | The current goal is easy local/dev discovery; filters/search come later. |
| 2 | Keep `BASE_URL` as the launcher/setup variable. | It already exists and is clear enough; avoid unnecessary renaming. |
| 3 | Add `make play-remote`. | Remote play should not boot local infrastructure. |
| 4 | Remove the hard two-player cap. | The project should not bake an artificial cap before real performance data exists. |
| 5 | Keep join-code support for private/non-listed sessions. | Existing v33 behavior remains useful and does not conflict with listed sessions. |
| 6 | Listed session join does not require copy/paste join code. | A browser row already exposes the session intentionally; manual code entry is unnecessary friction. |
| 7 | Non-host members can keep role `guest`. | Role is not a party slot count; richer roles are future UI/game design. |
| 8 | Use existing Control UI, not a lobby plugin. | The slice is backend/session flow, not production lobby polish. |

## 12. Open questions

| # | Question | Default if unanswered |
|---|----------|----------------------|
| Q-1 | Should listed sessions be visible across all dev accounts on the backend? | Yes; active listed sessions are backend-visible to authenticated users. |
| Q-2 | Should the active list include current dungeon level or deepest party level? | No for v38; member/connected counts are enough. |
| Q-3 | Should a menu-created listed session still return a join code? | Yes, keep current response compatibility unless implementation finds a reason to omit it. |
| Q-4 | Should `play-remote` probe `/readyz` before launching clients? | Yes, warn/fail clearly if unreachable. |
