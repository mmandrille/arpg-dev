# Spec: `true-coop-session`

Status: Approved for planning (gaps closed 2026-06-08)
Branch: `main`
Slice: v33 - true two-player co-op foundation
Baseline: v32 `test-floor-and-resilient-scenarios`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative Go server, thin Godot client, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - co-op players share one `Sim`
- [`v20_spec-play-session-loop.md`](v20_spec-play-session-loop.md) - default town/dungeon session loop
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) - character-owned inventory, hotbar, waypoints, progression

## 1. Purpose

Add the first real co-op vertical slice: two authenticated clients can join the same active session,
each controls a distinct player entity, both players see each other, and both can move and fight in
the same authoritative world.

The goal is not lobby polish. The goal is to replace the current solo-only realtime assumption with
a deterministic two-player session model:

- A host creates a co-op session.
- A second account joins the session with a join code and selected character.
- The server creates one player entity per session member.
- Each WebSocket connection is bound to exactly one server-derived player entity.
- Client intents never claim their actor; the realtime server derives actor identity from the
  authenticated connection.
- The authoritative session loop runs once per session and broadcasts state to both clients.
- Replays reconstruct the same two-player outcome from the ordered input log.

This slice is intentionally limited to a maximum of two players in one session. It is still true
co-op because both clients control separate player entities inside the same authoritative `Sim`.
Players see and fight alongside each other when they are on the same level. If one player is in the
dungeon and a second player joins later, the second player's character appears in level `0` town and
can descend or use unlocked travel later.

## 2. Non-goals

- No matchmaking service, public session browser, Steam lobby, invites, or friend list.
- No Godot high-level multiplayer, ENet game sync, or peer-authoritative state.
- No more than two players in one session.
- No automatic party travel. Each player's level transitions affect that player only.
- No trade protocol.
- No chat, emotes, ready checks, party UI polish, or lobby scene.
- No loot allocation rules beyond "the player who picks up an item owns it."
- No XP sharing rules. In v33, the player who lands the killing blow receives kill XP.
- No friendly fire.
- No production remote-player art. Reuse the existing humanoid/placeholder presentation.
- No same-account multi-tab/shared-control support. A session member has at most one active client
  connection.
- No persistence migration compatibility just for legacy dev rows; active development may update
  contracts, fixtures, and tests together.

## 3. Files to create or modify

```text
docs/specs/v33_spec-true-coop-session.md             - this slice contract
docs/plans/v33_<YYYY-MM-DD>-true-coop-session.md     - implementation plan
PROGRESS.md                                     - lifecycle update when v33 ships

shared/protocol/messages.v2.schema.json              - realtime envelope + intent contracts with actor derived server-side
shared/protocol/session_snapshot.v2.schema.json      - local_player_id, party, and player entity metadata
shared/protocol/state_delta.v2.schema.json           - player entity/event metadata for multi-player deltas

server/internal/store/*                              - session mode, join code, session members, per-member start snapshots, actor-tagged inputs
server/internal/http/session.go                      - create co-op sessions and join sessions by code
server/internal/http/realtime.go                     - authorize session members, not only session owner
server/internal/http/*_test.go                       - session create/join/auth and two-account WS coverage

server/internal/realtime/hub.go                      - one authoritative session loop per session with multiple subscribers
server/internal/realtime/runner.go                   - split session authority from per-connection read/write handling
server/internal/realtime/protocol.go                 - v2 envelope structs and per-connection player binding

server/internal/game/*.go                            - multiple player entities, actor-scoped inputs, combat/loot/progression ownership
server/internal/game/*_test.go                       - deterministic two-player movement, combat, loot, XP, and actor-only level travel tests
server/internal/replay/*.go                          - replay reconstructs session members and actor-tagged inputs
server/internal/replay/*_test.go                     - two-player replay determinism

client/scripts/net_client.gd                         - accept v2 snapshots/deltas and local_player_id
client/scripts/main.gd                               - distinguish local player from remote players; render remote player entities
client/tests/*                                       - v2 schema/client golden coverage where applicable

tools/bot/run.py                                     - two-account/two-WebSocket co-op scenario support
tools/bot/scenarios/23_true_coop_session.json        - protocol bot proof for two-player co-op
tools/bot/scenarios/client/12_true_coop_session.json - optional Godot client proof if the plan can keep it reliable
```

## 4. Data shapes

### 4.1 Session mode and join code

`POST /v0/sessions` keeps solo behavior by default and accepts `mode: "coop"` for co-op sessions.

```json
{
  "mode": "coop",
  "world_id": "dungeon_levels",
  "character_id": "char_..."
}
```

Co-op create response adds a join code:

```json
{
  "session_id": "sess_...",
  "character_id": "char_...",
  "seed": "0123456789abcdef",
  "world_id": "dungeon_levels",
  "mode": "coop",
  "join_code": "join_...",
  "ws_url": "/v0/ws?session_id=sess_..."
}
```

`POST /v0/sessions/{session_id}/join` lets a second authenticated account join with the host's
join code and one of that account's characters. The guest's character appears in level `0` town,
even if the host is already in a dungeon level:

```json
{
  "join_code": "join_...",
  "character_id": "char_..."
}
```

Join response returns the same session id and the guest character:

```json
{
  "session_id": "sess_...",
  "character_id": "char_...",
  "seed": "0123456789abcdef",
  "world_id": "dungeon_levels",
  "mode": "coop",
  "ws_url": "/v0/ws?session_id=sess_..."
}
```

Rules:

- `mode: "solo"` stays the default.
- `mode: "coop"` allows exactly two active members: one host and one guest.
- A session cannot join itself twice with the same account or same character.
- A third member receives `409 party_full`.
- A wrong or missing join code receives `404 session_not_found` to avoid session enumeration.
- Joining an ended session receives `409 session_ended`.

### 4.2 Persistence model

The store must represent a session as a party container rather than a single character-owned run.
Exact table names can be chosen in the plan, but the model must express:

```text
sessions
  id
  account_id              host account; retained for ownership/admin actions
  character_id            host character; retained for solo compatibility and host defaults
  mode                    solo | coop
  join_code_hash          nullable for solo
  seed
  world_id
  status

session_members
  session_id
  account_id
  character_id
  player_entity_id
  role                    host | guest
  status                  active | left
  joined_tick
  left_tick
  connected
  joined_at
  updated_at

session_member_start_snapshots
  session_id
  account_id
  character_id
  progression snapshot
  items snapshot
  hotbar snapshot
  waypoints snapshot

session_inputs
  session_id
  tick
  sequence
  message_id
  correlation_id
  actor_account_id
  actor_character_id
  actor_player_entity_id
  payload                 original client envelope, without trusting actor fields
```

Replay must use the per-member start snapshots and actor-tagged inputs. Live mutable character rows
must not affect replay reconstruction.

### 4.3 Protocol v2 snapshot

Protocol v2 adds local-player identity and party metadata to the snapshot payload. `current_level`
is the recipient/local player's current level; `entities` contains the entities visible on that
level. Snapshot payloads retain the current v1 recipient-owned inventory, equipment, hotbar,
teleporter, progression, and recent-event shapes; the example below is an excerpt focused on the
new co-op fields.

```json
{
  "server_tick": 10,
  "session_id": "sess_...",
  "seed": "0123456789abcdef",
  "current_level": -1,
  "local_player_id": "1001",
  "party": [
    {
      "player_id": "1001",
      "character_id": "char_host",
      "display_name": "Hero",
      "role": "host",
      "connected": true,
      "current_level": -1
    },
    {
      "player_id": "1007",
      "character_id": "char_guest",
      "display_name": "Guest",
      "role": "guest",
      "connected": true,
      "current_level": 0
    }
  ],
  "entities": [
    {
      "id": "1001",
      "type": "player",
      "character_id": "char_host",
      "display_name": "Hero",
      "position": { "x": 4, "y": 10 },
      "hp": 10,
      "max_hp": 10
    },
    {
      "id": "1012",
      "type": "monster",
      "monster_def_id": "dungeon_mob",
      "position": { "x": 8, "y": 10 },
      "hp": 6,
      "max_hp": 6
    }
  ]
}
```

Recipient-owned fields are scoped to the receiving client:

- `local_player_id`
- `inventory`
- `equipped`
- `hotbar_capacity`
- `hotbar`
- `character_progression`
- `discovered_teleporters`

World fields are shared:

- `server_tick`
- `session_id`
- `seed`
- `current_level` for each party member
- `party`
- `entities`
- `recent_events`

### 4.4 Protocol v2 intents

Client intent payloads do not include `player_id`, `account_id`, or `character_id`.

The realtime server derives the actor from the authenticated WebSocket's `session_member` binding
and sets `game.Input.ActorPlayerID` before buffering and persisting the input.

If a client sends an actor field anyway, schema validation rejects it with `intent_rejected:
invalid_payload`.

### 4.5 Protocol v2 deltas and events

`state_delta` remains level-scoped and now allows player metadata on player entity changes:

```json
{
  "server_tick": 24,
  "level": -1,
  "changes": [
    {
      "op": "entity_update",
      "entity": {
        "id": "1007",
        "type": "player",
        "character_id": "char_guest",
        "display_name": "Guest",
        "position": { "x": 8, "y": 10 },
        "hp": 10,
        "max_hp": 10
      }
    }
  ],
  "events": [
    {
      "event_type": "monster_damaged",
      "source_entity_id": "1007",
      "target_entity_id": "1012",
      "damage": 3,
      "outcome": "hit"
    }
  ]
}
```

Events that affect character-owned state must identify the actor through `source_entity_id` or
`entity_id`:

- `monster_damaged`: `source_entity_id` is the attacking player.
- `monster_killed`: `source_entity_id` is the killing player.
- `loot_picked_up`: `entity_id` is the player who picked up the item.
- `experience_gained`: `entity_id` is the player receiving XP.
- `character_leveled`: `entity_id` is the player whose character leveled.
- `player_damaged` / `player_killed`: `target_entity_id` is the damaged/killed player.

## 5. Architecture and flow

```text
Host account logs in
  -> POST /v0/sessions { mode: "coop", character_id }
  -> server creates session, host session_member, host start snapshot, join_code
  -> host opens /v0/ws?session_id=...
  -> realtime hub starts one authoritative session loop for session_id
  -> host connection binds to host player_entity_id
  -> host receives snapshot with local_player_id = host player

Guest account logs in
  -> POST /v0/sessions/{session_id}/join { join_code, character_id }
  -> server validates join code, active session, member capacity, character ownership
  -> server creates guest session_member and guest start snapshot
  -> server creates the guest player entity in level 0 town
  -> guest opens /v0/ws?session_id=...
  -> realtime hub attaches guest connection to existing session loop
  -> guest connection binds to guest player_entity_id
  -> guest receives snapshot with local_player_id = guest player
  -> host receives party metadata showing the guest joined; host sees the guest entity only when both are on the same level

Both clients play
  -> each client sends ordinary v2 intents
  -> realtime server derives actor from connection
  -> one session loop applies both actors' inputs in deterministic tick/sequence order
  -> session loop broadcasts level-scoped deltas to clients whose local player is on that level
  -> each client applies own inventory/progression changes only from recipient-scoped payload fields
  -> replay reconstructs host + guest start snapshots and actor-tagged inputs
```

## 6. Game rules for v33 co-op

### 6.1 Player entities

- A co-op session has one player entity per active session member.
- Host player spawns at the world/session spawn point.
- Guest player always spawns in level `0` town using the town player spawn or the nearest legal
  placement next to it. If no legal town spawn position exists, joining fails with
  `409 no_spawn_position`.
- Players are non-solid to each other in v33 to avoid pathfinding deadlocks and body-block griefing.
- Player IDs are deterministic within replay: host is created first; guest creation order follows
  persisted session-member join order.

### 6.2 Movement and actions

- `move_intent` and `move_to_intent` move only the actor player.
- `action_intent` uses the actor player's position, reach, attack stats, and equipment.
- Targeting another player is rejected with `intent_rejected: invalid_target`.
- Friendly fire is disabled.
- Dead players cannot move, attack, pick up, equip, allocate stats, or use items.

### 6.3 Combat and monsters

- Monsters can target any living connected player on the monster's current level.
- Initial v33 targeting rule: pick the nearest living player within aggro range, with deterministic
  tie-break by entity id.
- Monster retaliation damages the targeted player, not a global `sim.playerID`.
- Combat events must identify source and target player/monster IDs.
- Existing combat formulas remain unchanged except for actor/target selection.

### 6.4 Loot and inventory

- Ground loot is shared world state until picked up.
- The player who successfully picks up a loot entity receives the item in their own character
  inventory.
- If both players attempt to pick up the same item in the same tick, deterministic input sequence
  decides the winner; the loser receives `intent_rejected: target_missing` or equivalent current
  rejection after the item has been removed.
- Equip, unequip, drop, use, assign-hotbar, use-hotbar, and allocate-stat intents affect only the
  actor player's character state.
- Dropped items appear as shared world loot.

### 6.5 XP and progression

- In v33, the player who lands the killing blow receives XP.
- Character level-up and stat allocation are per actor character.
- XP sharing, party bonus rules, and proximity checks are deferred.

### 6.6 Level transitions

- Each player has their own current level in v33.
- `descend_intent`, `ascend_intent`, and `teleport_intent` move only the actor player.
- When a player changes levels, the server emits level-scoped remove/spawn deltas so clients on the
  origin level stop seeing that player and clients on the destination level start seeing that player.
- Players on different levels remain in the same session but do not receive each other's world
  entity deltas.
- Shared party metadata still reports each member's `current_level` and `connected` state.

### 6.7 Disconnect and reconnect

- Disconnecting one client marks that member disconnected but keeps the session loop alive while at
  least one member remains connected.
- A disconnected player's entity is removed from its level. It cannot be damaged or targeted while
  disconnected.
- Reconnecting the same account/character to the same active session reuses the existing
  `session_member` and respawns that player's entity in level `0` town.
- A second simultaneous connection for an already-connected session member is rejected with
  `member_already_connected`; this is not the same as two different members playing co-op.
- No client can end a co-op session for the other member. The server keeps the session open while at
  least one member is connected and owns final shutdown when no clients remain connected.

## 7. Client behavior

The Godot client must stop treating every `type: "player"` entity as the local `PlayerAnchor`.

Required behavior:

- Use `local_player_id` from the snapshot to bind the existing `PlayerAnchor` to the local player.
- Render other `type: "player"` entities as remote player nodes under `entities_root`.
- Reconcile only the local player's predicted position.
- Do not predict remote players; apply authoritative positions from deltas.
- Camera follows only the local player.
- Local health bar displays only the local player's HP.
- Damage numbers and event animations resolve by entity id, so remote player hits/deaths can be
  shown without being mistaken for local player events.
- When a remote player disconnects or changes level, remove that remote player node from the local
  visible entity set.
- Existing bot/client automation remains able to run a solo session unchanged.

No new multiplayer lobby UI is required. Bot and manual testing can use REST responses, join code,
and environment variables.
## 9. Acceptance criteria

1. Host can create `mode: "coop"` session and receives a non-enumerable join code.
2. A second authenticated account can join the active co-op session with the join code and a
   character it owns.
3. A third join attempt is rejected with `party_full`.
4. Wrong account/character ownership and wrong join code do not leak session existence.
5. Host and guest can both open WebSockets to the same `session_id`.
6. Both clients receive snapshots for the same session with different `local_player_id` values.
7. A guest who joins after session creation appears in level `0` town.
8. Host movement updates only the host player; guest movement updates only the guest player.
9. When both players are on the same level, both snapshots include two player entities and both
   clients can see each other.
10. Host and guest can each damage and kill monsters using their own position, reach, equipment, and
    stats.
11. Monster aggro/retaliation can damage either connected player on the same level and events
    identify the damaged player.
12. Loot pickup transfers the item to the picking player's character inventory only.
13. Equip/use/drop/hotbar/stat intents affect only the actor player's character state.
14. Killing-blow XP is awarded to the killing player's character only.
15. Descend/ascend/teleport moves only the actor player and emits deterministic level-scoped
    remove/spawn deltas/events.
16. Disconnecting one client removes that player's entity, leaves the other connected, and keeps the
    authoritative session loop running.
17. A co-op session remains server-owned and open while at least one member is connected.
18. Reconnecting a disconnected member restores the actor binding and respawns that player's entity
    in level `0` town.
19. Replay reconstructs a two-player session deterministically from per-member start snapshots and
    actor-tagged inputs.
20. The Godot client renders the local player and one remote player without confusing local
    prediction, camera follow, or health UI.
21. Existing solo session, protocol bot, client bot, replay, shared validation, and smoke coverage
    remain green.
22. `make ci` passes.

## 10. Testing plan

1. Shared/protocol validation:

```bash
make validate-shared
```

2. Focused Go tests while refactoring sim/realtime:

```bash
cd server && go test ./internal/game/... ./internal/realtime/... ./internal/http/... ./internal/replay/... -count=1
```

3. Protocol bot proof:

```bash
ADDR=:18080 BASE_URL=http://localhost:18080 make bot scenario=true_coop_session
```

The bot scenario must use two accounts, two characters, and two WebSocket connections. It should
assert distinct `local_player_id` values, independent movement, shared monster combat, actor-scoped
loot pickup, actor-scoped XP, actor-scoped level transitions, disconnect removal, reconnect-to-town,
and replay success.

4. Godot client proof:

```bash
ADDR=:18081 BASE_URL=http://localhost:18081 HEADLESS=1 make bot-client scenario=true_coop_session
```

If a two-process Godot proof is too brittle for v33, the plan may substitute a focused client unit
test plus protocol bot coverage, but must explain the deferral.

5. Final gate:

```bash
make ci
```

## 11. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | v33 implements true two-player co-op, not spectator-only fanout. | The requested slice is an initial multiplayer step with two controllable players. |
| 2 | Maximum party size is two. | Keeps the first co-op contract testable while surfacing the correct architecture. |
| 3 | Guest joins by join code, not by open session id alone. | Avoids accidental session enumeration and keeps matchmaking out of scope. |
| 4 | Intents do not include actor identity. | Actor identity must be authenticated and server-derived, not client-claimed. |
| 5 | Protocol v2 is introduced for local player identity and party metadata. | Existing v1 schemas are single-local-player shaped and strict. |
| 6 | Each player has their own `current_level` in v33. | A late joiner appears in town even if another player is already in the dungeon, matching ADR-0008 D6. |
| 7 | Level travel moves only the actor player. | Players can regroup manually and clients receive level-scoped visibility. |
| 8 | Players are non-solid to each other. | Avoids pathfinding/body-block complexity in the first co-op slice. |
| 9 | Killing blow receives XP; picker receives loot. | Minimal deterministic ownership rules with no party reward design yet. |
| 10 | Steam/Godot multiplayer shortcuts are rejected for game state. | ADR-0001 authority and replay remain the governing architecture. |
| 11 | Disconnect removes the player's entity from the level. | A disconnected client should not leave a targetable body behind. |
| 12 | A co-op session remains open while at least one member is connected. | Sessions are server-owned; no client can terminate another member's active session. |
| 13 | A simultaneous second WebSocket for the same already-connected member is rejected. | Co-op is two different session members, not shared control by one member. |
| 14 | Rejoining after disconnect respawns the member in town. | Keeps reconnect deterministic and avoids resurrecting stale combat positions. |

## 12. Open questions

None.
