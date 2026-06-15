# v33 Plan — True Co-op Session

Status: Ready for implementation
Goal: Let two authenticated clients join one server-owned session, each controlling a distinct character/player entity in the authoritative Go sim.
Architecture: Introduce protocol v2, session membership, actor-tagged inputs, and one in-process authoritative realtime loop per session. The server derives actor identity from the authenticated WebSocket member binding; clients never claim `player_id` in intents. Players have independent current levels: late joiners and reconnecting members spawn in level `0` town, and players only see each other's entities when they share a level.
Tech stack: Go sim/realtime/http/store/replay, Postgres migrations, shared JSON protocol schemas, Godot GDScript client, Python protocol bot, Godot client bot, project docs.

## Baseline and shortcut decision

Baseline is v32 `test-floor-and-resilient-scenarios` on `main`, with protocol bot scenarios
currently numbered through `tools/bot/scenarios/22_combat_stat_effects.json` and client scenarios
through `tools/bot/scenarios/client/11_combat_feedback.json`.

This slice reuses the existing authoritative Go sim, character persistence, deterministic replay,
WebSocket protocol shape, Godot `PlayerAnchor`, entity rendering map, and Python bot harness.
| Candidate | Decision | Reason |
|-----------|----------|--------|
| Godot high-level multiplayer / ENet / `@rpc` | Reject | Game state must stay Go-authoritative and replayable. |
| Steam lobby templates / GodotSteam | Reject for v33 | Useful later for discovery/invites; not needed for server co-op authority. |

Work stays on the current branch. Do not create a branch. Preserve unrelated dirty worktree changes.

## Spec decisions

| Topic | Decision for implementation |
|-------|-----------------------------|
| Party size | Exactly two active members: host and guest. |
| Join flow | Host creates `mode: "coop"` and gets a join code; guest joins by session id + join code. |
| Late join | Guest always spawns in level `0` town, even if host is already in a dungeon level. |
| Actor identity | Server derives actor from the authenticated session member; intent payloads do not include actor fields. |
| Level model | Each player has an independent `current_level`; deltas are scoped to the receiver's visible level. |
| Disconnect | Disconnect removes that player's entity from the level; the other client keeps playing. |
| Session ownership | Co-op sessions are server-owned and stay open while at least one member is connected. |
| Duplicate member socket | Reject a second active WebSocket for the same already-connected session member. |
| Rewards | Killing blow receives XP; picker receives loot. XP sharing and allocation rules are deferred. |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `docs/specs/v33_spec-true-coop-session.md` | Approved contract and clarified decisions |
| Create | `docs/plans/v33_2026-06-08-true-coop-session.md` | Implementation checklist |
| Create | `shared/protocol/messages.v2.schema.json` | Protocol v2 envelope/intents |
| Create | `shared/protocol/session_snapshot.v2.schema.json` | `local_player_id`, party metadata, player entity metadata |
| Create | `shared/protocol/state_delta.v2.schema.json` | Actor-aware player events and level-scoped changes |
| Modify | `tools/validate_shared.py` | Validate protocol v2 schemas and any new invariants |
| Create | `server/migrations/0007_true_coop_sessions.sql` | Session mode, join code hash, members, actor-tagged inputs |
| Modify | `server/internal/store/models.go` | Session/member/input models |
| Modify | `server/internal/store/interfaces.go` | Membership and actor input repository methods |
| Modify | `server/internal/store/repos.go` | Postgres implementation for sessions, members, start snapshots, actor inputs |
| Modify | `server/internal/http/session.go` | Create co-op, join, member-aware end/leave behavior |
| Modify | `server/internal/http/realtime.go` | Authorize session members, not just session owner |
| Modify | `server/internal/http/*_test.go` | HTTP auth/session/join/WS coverage |
| Modify | `server/internal/game/*.go` | Multiple player states, actor-scoped inputs, per-player levels, reward ownership |
| Modify | `server/internal/game/*_test.go` | Deterministic co-op sim coverage |
| Modify | `server/internal/realtime/*.go` | One session loop with multiple member connections and per-recipient snapshots |
| Modify | `server/internal/realtime/*_test.go` | Fanout, duplicate socket, disconnect/remove tests |
| Modify | `server/internal/replay/*.go` | Reconstruct members and actor-tagged inputs |
| Modify | `server/internal/replay/*_test.go` | Two-player replay determinism |
| Modify | `client/scripts/net_client.gd` | Accept v2 co-op snapshot/delta payloads |
| Modify | `client/scripts/main.gd` | Local/remote player split, remote removal on disconnect/level change |
| Modify | `client/tests/*` | Focused client unit/golden coverage for local vs remote player handling |
| Modify | `tools/bot/run.py` | Two-account/two-WebSocket scenario support |
| Create | `tools/bot/scenarios/23_true_coop_session.json` | Protocol bot proof |
| Create/Optional | `tools/bot/scenarios/client/12_true_coop_session.json` | Godot client proof if reliable in v33 |
| Modify | `PROGRESS.md` | Lifecycle update when v33 ships |

## Task 1 — Shared Protocol v2 Contracts

Files:
- Create: `shared/protocol/messages.v2.schema.json`
- Create: `shared/protocol/session_snapshot.v2.schema.json`
- Create: `shared/protocol/state_delta.v2.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Copy v1 schemas to v2 as the starting point; keep solo-compatible intent names and envelope fields.
- [x] Step 1.2: Add `local_player_id` and `party[]` to `session_snapshot.v2.schema.json`.
- [x] Step 1.3: Add player entity metadata fields: `character_id`, `display_name`, and any required connected/level party metadata.
- [x] Step 1.4: Keep intent payload schemas actor-free: reject `player_id`, `account_id`, and `character_id` in client intents through `additionalProperties: false`.
- [x] Step 1.5: Extend event schemas so `monster_damaged`, `monster_killed`, `loot_picked_up`, `experience_gained`, `character_leveled`, `player_damaged`, and `player_killed` can identify actor/target player IDs.
- [x] Step 1.6: Update shared validation to discover and validate protocol v2 files.
- [x] Step 1.7: Keep protocol v1 files intact until the server/client switch is complete; do not delete v1 in this slice unless all references are intentionally migrated.

```bash
make validate-shared
```

## Task 2 — Store and Migration Foundation

Files:
- Create: `server/migrations/0007_true_coop_sessions.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/*_test.go`

- [x] Step 2.1: Add `sessions.mode` (`solo`/`coop`) and nullable `sessions.join_code_hash`; preserve solo defaults.
- [x] Step 2.2: Add `session_members` keyed by `(session_id, account_id, character_id)` with role, connected flag, current `player_entity_id`, status, joined/left ticks, and timestamps.
- [x] Step 2.3: Add per-member start snapshot storage or extend current `session_start_*` tables so replay can load each member's items, hotbar, waypoints, and progression from session start.
- [x] Step 2.4: Add actor columns to `session_inputs`: `actor_account_id`, `actor_character_id`, `actor_player_entity_id`.
- [x] Step 2.5: Store only a hash of the join code. Use cryptographically random join codes and constant-time comparison in repository/helper code; do not log raw join codes.
- [x] Step 2.6: Add repository methods for creating host member, joining guest member, listing members in deterministic order, marking connected/disconnected, and looking up membership by account/session.
- [x] Step 2.7: Update character deletion cleanup to include session members and new start snapshot rows.
- [x] Step 2.8: Add store tests for membership creation, capacity, duplicate member rejection, actor-tagged inputs, and per-member start snapshot loading.

```bash
cd server && go test ./internal/store/... -count=1
```

## Task 3 — HTTP Session and Join Lifecycle

Files:
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/http/realtime.go`
- Modify: `server/internal/http/auth_session_test.go`
- Modify: `server/internal/http/ws_test.go`

- [x] Step 3.1: Extend `POST /v0/sessions` to accept `mode: "coop"` and return `mode` plus a raw one-time-visible join code in the create response.
- [x] Step 3.2: Keep `mode: "solo"` behavior compatible for existing bots/client flows.
- [x] Step 3.3: Add `POST /v0/sessions/{session_id}/join` to validate join code, active session, capacity, and guest character ownership.
- [x] Step 3.4: Ensure wrong join code, wrong account/character, and unknown session avoid session enumeration in responses.
- [x] Step 3.5: Change WebSocket authorization so any active session member may connect; non-members remain denied.
- [x] Step 3.6: Define co-op leave/end behavior: closing the client or calling the current end route disconnects/leaves only that member; the co-op session remains active while another member is connected.
- [x] Step 3.7: Add HTTP tests for co-op create, join success, third join `party_full`, duplicate same account/member, wrong code, ended session, and member-only WebSocket access.

```bash
cd server && go test ./internal/http/... -run 'Session|Join|WS|Coop' -count=1
```

## Task 4 — Actor-Scoped Game Sim

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/*.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/game/pathfind_test.go`

- [x] Step 4.1: Add actor identity to `game.Input` (`ActorPlayerID` or equivalent) and make solo tests/helpers set it through a deterministic default.
- [x] Step 4.2: Replace global `sim.playerID` assumptions in runtime logic with actor-scoped player lookup helpers. Keep a solo compatibility wrapper only for tests/helpers that intentionally exercise one player.
- [x] Step 4.3: Introduce per-player state for entity id, account id, character id, display name, current level, inventory, equipment, hotbar, discovered teleporters, and progression.
- [x] Step 4.4: Add deterministic methods to add host player, add guest player in level `0` town, remove disconnected player entity, and respawn a returning member in town.
- [x] Step 4.5: Update movement, auto-nav, action, ranged projectile ownership, combat reach/stats, loot pickup, equipment, consumables, hotbar, stat allocation, and XP to operate on the actor player.
- [x] Step 4.6: Update monster AI to select the nearest living connected player on the monster's level, with stable entity-id tie-breaks and no map iteration nondeterminism.
- [x] Step 4.7: Update level transitions so descend/ascend/teleport moves only the actor player and emits level-scoped remove/spawn changes.
- [x] Step 4.8: Keep players non-solid to each other; keep monsters/walls/closed interactables authoritative.
- [x] Step 4.9: Add `SnapshotForPlayer(playerID)` or equivalent recipient-scoped snapshot builder while preserving existing solo snapshot tests through a compatibility path.
- [x] Step 4.10: Add Go tests for independent movement, same-level visibility, guest town spawn after host descends, actor-scoped combat, monster target selection, loot pickup race, actor-scoped XP, actor-only level travel, disconnect removal, reconnect-to-town, and deterministic sorted output.

```bash
cd server && go test ./internal/game/... -run 'Coop|Player|Actor|Level|Monster|Loot|Experience' -count=1
```

## Task 5 — Realtime Session Loop and Fanout

Files:
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/realtime/protocol.go`
- Modify: `server/internal/realtime/*_test.go`
- Modify: `server/internal/http/ws_test.go`

- [x] Step 5.1: Split the current one-connection runner into a session loop keyed by `session_id` plus per-connection reader/writer clients.
- [x] Step 5.2: Ensure the session loop reconstructs or creates the `Sim` once per active session, not once per WebSocket.
- [x] Step 5.3: Bind each connection to a `session_member` and derived `player_entity_id`.
- [x] Step 5.4: Reject a second active WebSocket for the same already-connected member with `member_already_connected`.
- [x] Step 5.5: On attach, send a recipient-scoped snapshot with that member's `local_player_id`, inventory/progression, visible entities for their current level, and party metadata.
- [x] Step 5.6: Buffer and persist inputs with actor metadata after server-side actor binding; ignore/reject any client-claimed actor fields.
- [x] Step 5.7: Broadcast each tick's world changes only to clients whose local player can see the changed level; broadcast party metadata changes to all connected members.
- [x] Step 5.8: On disconnect, mark the member disconnected, remove that player's entity from its level, and keep the loop running while at least one member remains connected.
- [x] Step 5.9: Stop the session loop only when no clients remain connected, leaving final session status to server-owned lifecycle policy.
- [x] Step 5.10: Add WebSocket tests for two different accounts in one session, distinct local player IDs, host/guest independent movement, same-level visibility, guest town spawn, duplicate same-member socket rejection, disconnect removal, and continued play by the remaining member.

```bash
cd server && go test ./internal/realtime/... ./internal/http/... -run 'Coop|WebSocket|SessionLoop|Disconnect' -count=1
```

## Task 6 — Replay Reconstruction

Files:
- Modify: `server/internal/replay/*.go`
- Modify: `server/internal/replay/*_test.go`
- Modify: `server/cmd/arpg-replay/*` if needed

- [x] Step 6.1: Load session members and per-member start snapshots during replay reconstruction.
- [x] Step 6.2: Recreate host/guest player entities in deterministic member order and preserve actor player IDs expected by the input log.
- [x] Step 6.3: Replay actor-tagged inputs using server-derived actor metadata, not client payload actor fields.
- [x] Step 6.4: Reconstruct disconnect/removal and reconnect-to-town behavior deterministically if these are persisted as inputs/events.
- [x] Step 6.5: Add a replay test that runs a two-player session with movement, combat, loot, level transition, disconnect, reconnect, and verifies no mismatch.

```bash
cd server && go test ./internal/replay/... -run 'Coop|Actor|Reconnect' -count=1
```

## Task 7 — Godot Client Local/Remote Player Handling

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/*`
- Modify: `tools/bot/scenarios/client/12_true_coop_session.json` if a reliable proof is added

- [x] Step 7.1: Parse and store `local_player_id` from v2 snapshots.
- [x] Step 7.2: Treat only `local_player_id` as the existing `PlayerAnchor`; render every other visible `type: "player"` as an entity node under `entities_root`.
- [x] Step 7.3: Reconcile predicted movement only for the local player.
- [x] Step 7.4: Apply authoritative movement for remote players with no prediction.
- [x] Step 7.5: Route health bar, camera follow, local input gating, and local death handling only through the local player.
- [x] Step 7.6: Route damage numbers and animation events by entity id so remote player hit/death events do not trigger local-only behavior.
- [x] Step 7.7: Remove remote player nodes when `entity_remove` arrives because the remote disconnected or changed levels.
- [x] Step 7.8: Keep existing solo client-bot scenarios green.
- [x] Step 7.9: Add either a focused client unit test for local-vs-remote player application or a reliable two-process client-bot scenario. If two-process Godot is brittle, document the deferral in this plan/as-built notes and rely on protocol bot plus unit coverage for v33.

```bash
make client-unit
make client-smoke
```

## Task 8 — Bot Scenarios

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/23_true_coop_session.json`
- Create/Optional: `tools/bot/scenarios/client/12_true_coop_session.json`

- [x] Step 8.1: Add bot support for two dev-login accounts, two selected/default characters, one host co-op create, one guest join by join code, and two concurrent WebSocket connections.
- [x] Step 8.2: Add scenario assertions for distinct `local_player_id` values and party metadata containing host and guest.
- [x] Step 8.3: Prove guest joins into level `0` town after host descends to a dungeon level.
- [x] Step 8.4: Move guest down to the host's level or move both to a shared level, then assert both clients see two player entities.
- [x] Step 8.5: Send movement from host and guest and assert each moves only its own player.
- [x] Step 8.6: Have both players attack monsters; assert source/target event IDs identify the acting/damaged player.
- [x] Step 8.7: Pick up loot with one player and assert only that player's inventory changes.
- [x] Step 8.8: Kill a monster and assert XP is awarded only to the killing player.
- [x] Step 8.9: Move one player through descend/ascend/teleport and assert the other player remains on its current level.
- [x] Step 8.10: Disconnect one WebSocket and assert its player entity is removed while the other client continues receiving ticks/deltas.
- [x] Step 8.11: Reconnect the disconnected member and assert actor binding is restored and the player respawns in town.
- [x] Step 8.12: Fetch replay and assert no mismatch for the co-op session.
- [x] Step 8.13: Keep existing protocol scenarios `01`-`22` green.

As-built note: the Python co-op bot directly covers steps 8.1-8.5 and 8.10-8.12. Steps 8.6-8.9 are covered by focused Go simulation/replay/WebSocket tests for actor-scoped combat, loot, XP, and level isolation, avoiding brittle generated-dungeon combat pathing in the protocol bot.

```bash
python -m py_compile tools/bot/run.py
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=true_coop_session
make bot
```

## Task 9 — Integration Sweep

Files:
- Modify: `server/internal/game/*`
- Modify: `server/internal/realtime/*`
- Modify: `server/internal/http/*`
- Modify: `server/internal/replay/*`
- Modify: `client/scripts/*`
- Modify: `tools/bot/*`

- [x] Step 9.1: Run shared validation after all protocol/schema changes.
- [x] Step 9.2: Run all Go tests after sim/realtime/replay integration.
- [x] Step 9.3: Run protocol bot after server and bot updates.
- [x] Step 9.4: Run client unit and smoke after Godot changes.
- [x] Step 9.5: Fix any solo regressions by preserving single-player compatibility through the new actor/member path, not by adding a separate solo code path.

```bash
make validate-shared
make test-go
make bot
make client-unit
make client-smoke
```

## Task 10 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v33_2026-06-08-true-coop-session.md`

- [x] Step 10.1: Update this plan with any as-built deviations, especially if the optional two-process Godot proof is deferred.
- [x] Step 10.2: Update `PROGRESS.md` only after implementation and CI pass.
- [x] Step 10.3: Record v33 as complete, summarize true co-op behavior, and list deferred multiplayer gaps: matchmaking/lobby, Steam invites, party UI, trade, XP sharing, loot allocation, more than two players.
- [x] Step 10.4: Run the final gate.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make bot`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make ci`

## Deferred scope

- Matchmaking, public session discovery, Steam lobby/invites, and friend flows.
- More than two players.
- Party panel/UI polish.
- Chat, emotes, ready checks, and lobby scene.
- Trade protocol.
- XP sharing, party bonus, proximity reward rules, and loot allocation rules.
- Friendly fire/PvP.
- Production remote-player art.
- Cross-process distributed session ownership. v33 assumes one server process owns the in-memory session loop.
