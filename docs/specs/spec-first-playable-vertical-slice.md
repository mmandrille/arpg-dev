# Spec: `first-playable-vertical-slice`

Status: Ready for implementation
Branch: `feature/first-playable-vertical-slice`
Related: `docs/adr/0001-technology-stack.md`

## 1. Purpose

Build the smallest end-to-end game slice that proves the foundational architecture from ADR-0001:
a Godot client and Go server running as separate apps, with auth/session identity, a remote-capable
WebSocket protocol, authoritative server simulation, Postgres persistence, structured observability,
seeded deterministic replay, and a Python protocol bot that can play and verify the flow.

The slice is intentionally small: one account, one character, one scene, one monster, one loot
drop, one inventory, one equip action. Placeholder visuals are acceptable. The goal is to prove the
architecture before investing in content, polish, or production multiplayer features.

## 2. Non-goals

- Production-grade auth provider integration.
- Matchmaking, parties, PvP, chat, or multiple players in the same session.
- Steam integration.
- Final art, animation polish, character customization, or real asset pipeline output.
- Complex combat, skills, pathfinding, AI behavior trees, or item affix systems.
- Protobuf transport.
- Horizontal scaling or separate realtime/platform deployables.
- Historical/filtered inspection APIs beyond the current live session state.

## 3. Files to create or modify

```text
server/                 - Go service: HTTP routes, WebSocket realtime session, sim, persistence
server/cmd/arpg-server/ - Server entrypoint
server/internal/auth/   - Auth/session baseline
server/internal/config/ - Environment/config loading
server/internal/game/   - Deterministic authoritative simulation
server/internal/http/   - REST routes, WebSocket upgrade, middleware
server/internal/logging/- Structured logging and correlation IDs
server/internal/metrics/- Metrics registration/export
server/internal/realtime/- Authenticated WebSocket session runner and protocol handling
server/internal/replay/ - Session recording and replay verification
server/internal/store/  - Postgres repositories and migrations
server/migrations/      - Postgres schema migrations
shared/protocol/        - JSON message schemas and examples
shared/golden/          - Cross-language fixtures consumed by Go and GDScript tests
shared/rules/           - Combat, item, monster, and loot table schemas/data
tools/bot/              - Python protocol bot and integration assertions
tools/replay/           - Replay CLI wrapper
client/                 - Godot 4 project with minimal playable scene
docs/plans/             - Implementation plan for this spec
docker-compose.yml      - Local Postgres dev runner
.godot-version          - Exact Godot editor/runtime pin
pyproject.toml          - Python tool pinning for bot/replay tooling
```

## 4. Public interfaces and data shapes

### 4.1 HTTP routes

All HTTP responses are JSON unless noted.

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/healthz` | none | Liveness check. |
| `GET` | `/readyz` | none | Readiness check, including database connectivity. |
| `GET` | `/metrics` | deployment-gated | Prometheus-compatible service metrics. |
| `POST` | `/v0/auth/dev-login` | none | Development login using email/dev token. |
| `POST` | `/v0/sessions` | bearer token | Create or resume a solo game session. |
| `GET` | `/v0/sessions/{session_id}/state` | bearer token + debug authorization | Current authoritative state for agents. |
| `GET` | `/v0/sessions/{session_id}/replay` | bearer token + debug authorization | Replay metadata and latest verification result. |
| `GET` | `/v0/ws` | bearer token | WebSocket upgrade for realtime game protocol. |

The canonical live inspection route for v0 is `GET /v0/sessions/{session_id}/state`, matching
ADR-0001 D8.4's minimal `GET /sessions/{id}/state` form within the versioned API namespace.
`/replay` is debug metadata only; it is not a general historical query API.

Debug authorization in v0 requires both a valid account bearer token and `X-Debug-Token` matching
the server `DEBUG_TOKEN` setting. Local dev may default `DEBUG_TOKEN` to `local-debug-token`.
Remote deployments must set a non-default token and must not expose debug routes without this
header.

### 4.2 Auth shapes

`POST /v0/auth/dev-login`

```json
{
  "email": "dev@example.test",
  "dev_token": "local-dev-token"
}
```

Response:

```json
{
  "account_id": "acct_01H00000000000000000000000",
  "access_token": "opaque-dev-token",
  "expires_at": "2026-06-05T12:00:00Z"
}
```

The implementation may use a local dev-token strategy, but the server and client must treat the
result as real account identity. Anonymous local-only play is not part of this architecture.

### 4.3 Session shapes

`POST /v0/sessions`

```json
{
  "mode": "solo",
  "resume_session_id": null
}
```

Response:

```json
{
  "session_id": "sess_01H00000000000000000000000",
  "character_id": "char_01H00000000000000000000000",
  "seed": "hex-encoded-server-seed",
  "ws_url": "/v0/ws?session_id=sess_01H00000000000000000000000"
}
```

### 4.4 WebSocket envelope

All realtime messages use this JSON envelope in v0:

```json
{
  "type": "move_intent",
  "message_id": "msg_01H00000000000000000000000",
  "session_id": "sess_01H00000000000000000000000",
  "tick": 42,
  "correlation_id": "corr_01H00000000000000000000000",
  "payload": {}
}
```

The authoritative server tick rate is **20 Hz** (50 ms per tick). Client-to-server intents stamp
`tick` with the intended authoritative tick. The server buffers valid intents and applies them in
deterministic order within that tick: `(tick, sequence, message_id)`. `duration_ticks` is measured
in authoritative ticks, so `duration_ticks: 5` means 250 ms of requested movement. Server-to-client
messages stamp `tick` with the authoritative tick that produced the payload. Late or duplicate
inputs are acknowledged or rejected with structured protocol messages, never silently ignored.

Required client-to-server message types:

| Type | Payload |
|------|---------|
| `client_ready` | `{ "client_version": "dev", "last_seen_tick": 0 }` |
| `move_intent` | `{ "direction": { "x": 1, "y": 0 }, "duration_ticks": 5 }` |
| `attack_intent` | `{ "target_id": "1002" }` |
| `pick_up_intent` | `{ "entity_id": "1003" }` |
| `equip_intent` | `{ "item_instance_id": "1004", "slot": "weapon" }` |

Required server-to-client message types:

| Type | Payload |
|------|---------|
| `session_snapshot` | Full authoritative state needed to render the scene. |
| `state_delta` | Ordered authoritative changes for a tick. |
| `intent_accepted` | `{ "accepted_message_id": "...", "server_tick": 42 }` |
| `intent_rejected` | `{ "rejected_message_id": "...", "reason": "invalid_target" }` |
| `error` | `{ "code": "bad_message", "message": "human-readable summary" }` |

Minimal `session_snapshot` payload:

```json
{
  "server_tick": 42,
  "session_id": "sess_01H00000000000000000000000",
  "seed": "hex-encoded-server-seed",
  "entities": [
    {
      "id": "1001",
      "type": "player",
      "position": { "x": 10, "y": 5 },
      "hp": 10,
      "max_hp": 10
    },
    {
      "id": "1002",
      "type": "monster",
      "monster_def_id": "training_dummy",
      "position": { "x": 12, "y": 5 },
      "hp": 3,
      "max_hp": 3
    }
  ],
  "inventory": [],
  "equipped": { "weapon": null },
  "recent_events": []
}
```

Minimal `state_delta` payload:

```json
{
  "server_tick": 43,
  "changes": [
    {
      "op": "entity_update",
      "entity": {
        "id": "1002",
        "type": "monster",
        "position": { "x": 12, "y": 5 },
        "hp": 0,
        "max_hp": 3
      }
    },
    {
      "op": "entity_spawn",
      "entity": {
        "id": "1003",
        "type": "loot",
        "item_def_id": "rusty_sword",
        "position": { "x": 12, "y": 5 }
      }
    },
    {
      "op": "inventory_add",
      "item": {
        "item_instance_id": "1004",
        "item_def_id": "rusty_sword",
        "slot": "weapon",
        "equipped": false
      }
    }
  ],
  "events": [
    {
      "event_type": "monster_killed",
      "entity_id": "1002",
      "correlation_id": "corr_01H00000000000000000000000"
    }
  ]
}
```

`changes` are applied in array order. Unknown `op` values are protocol errors in tests.
Entity IDs are decimal strings in JSON to preserve unsigned 64-bit values without precision loss.

### 4.5 Shared rules data

Minimum shared data files:

```text
shared/rules/combat.v0.schema.json
shared/rules/items.v0.schema.json
shared/rules/loot_tables.v0.schema.json
shared/rules/monsters.v0.schema.json
shared/rules/combat.v0.json
shared/rules/items.v0.json
shared/rules/loot_tables.v0.json
shared/rules/monsters.v0.json
```

Protocol schemas live under `shared/protocol/`. At minimum they cover the envelope, all listed
client/server payloads, `session_snapshot`, and `state_delta`. Golden fixtures live under
`shared/golden/` and must be consumed by both Go and GDScript tests.

Rules must be declarative. For this slice, combat may be deterministic and simple:

```json
{
  "base_hit_chance": 1.0,
  "player_damage": { "min": 2, "max": 4 },
  "monster_hp": 3
}
```

### 4.6 Persistence model

Minimum persisted entities:

| Entity | Required fields |
|--------|-----------------|
| `accounts` | `id`, `email`, `created_at` |
| `characters` | `id`, `account_id`, `name`, `created_at` |
| `sessions` | `id`, `account_id`, `character_id`, `seed`, `status`, `created_at`, `updated_at` |
| `inventory_items` | `id`, `account_id`, `character_id`, `item_def_id`, `slot`, `equipped`, `created_at` |
| `session_events` | `id`, `session_id`, `tick`, `sequence`, `event_type`, `correlation_id`, `payload`, `created_at` |
| `session_inputs` | `id`, `session_id`, `tick`, `sequence`, `message_id`, `correlation_id`, `payload`, `created_at` |

### 4.7 Deterministic simulation identifiers

Authoritative game entities use monotonic, server-assigned unsigned 64-bit IDs from a per-session
counter. The counter advances only from deterministic sim events in stable spawn order. Authoritative
iteration sorts by entity ID and never depends on Go map order, pointer addresses, wall-clock time,
or database insertion order. Replay must reproduce the same entity IDs from the same seed and input
stream.

## 5. Architecture and flow

```text
Godot client
  -> POST /v0/auth/dev-login
  -> POST /v0/sessions
  -> WebSocket /v0/ws
  -> sends move/attack/pickup/equip intents
  -> predicts movement locally and reconciles to authoritative snapshots/deltas

Go server
  -> validates auth/session
  -> runs 20 Hz fixed-tick authoritative sim
  -> persists inventory/session/events to Postgres
  -> emits structured logs, metrics, snapshots, and deltas

Python bot
  -> uses the same auth/session/WebSocket path
  -> completes the slice
  -> asserts authoritative state through protocol and state API

Replay tool
  -> loads seed + recorded inputs
  -> re-simulates
  -> compares derived events/state against recorded authoritative output
```

## 6. Observability and replay requirements

- Logs are structured JSON and include at least `severity`, `component`, `session_id`, `tick`,
  `message_id`, and `correlation_id` when available.
- Metrics include process health, active sessions, WebSocket connections, tick duration, message
  latency, rejected intents, persistence errors, replay verification failures, and reconciliation
  deltas reported by the client when available.
- The state API exposes structured state only. Agents must not need pixel scraping to verify the
  authoritative game state.
- Remote deployments must protect state and replay endpoints with bearer auth plus debug
  authorization. Local dev may additionally bind them to localhost.
- Replay verification must fail if the same seed and input stream produce different authoritative
  events or final state.

## 7. Acceptance criteria

1. `server` starts locally and exposes `/healthz`, `/readyz`, `/metrics`, auth, session creation,
   state, replay metadata, and WebSocket routes.
2. Server uses Postgres for account, character, session, inventory, input, and event persistence.
3. Client and bot both authenticate before creating or resuming a session.
4. Client and bot both use the same WebSocket message envelope.
5. Server owns hit/miss, damage, death, loot drops, pickup, equip, HP, and inventory state.
6. Shared rule files validate against JSON schemas.
7. Go tests and GDScript tests consume at least one common golden fixture from `shared/`.
8. The Python bot completes: login → create session → move → attack monster → pick up loot →
   equip item → assert persisted inventory after restart or session resume.
9. Replay CLI verifies the recorded bot session from seed + inputs.
10. Structured logs and metrics are emitted during the bot run.
11. Godot client can complete the same minimal flow with placeholder visuals, including pickup and
   equip, while reporting reconciliation deltas.
12. Invalid auth, invalid session, malformed WebSocket message, invalid target, and duplicate
   pickup all produce structured errors without crashing the server.

## 8. Resolved questions and deferred decisions

| # | Question | Status |
|---|----------|--------|
| 1 | Which Postgres dev runner is preferred: Docker Compose, local Postgres, or another tool? | Resolved: Docker Compose |
| 2 | Which production auth provider should replace dev-token auth later? | Deferred to ADR-0005 |
| 3 | Which metrics stack should local dashboards use? | Resolved: Prometheus-format `/metrics`; dashboards deferred |
| 4 | Should replay logs be stored only in Postgres for v0, or also exported as JSONL files? | Resolved: Postgres only for v0 |
| 5 | What exact Godot version should be pinned for the first client project? | Resolved: `4.6.3-stable` |

## 9. Testing plan

1. Server unit tests for auth, session creation, sim rules, persistence repositories, and replay.
2. Server integration tests against local Postgres.
3. Shared schema validation for all files under `shared/rules/`, `shared/protocol/`, and
   `shared/golden/`.
4. Python bot integration test against a running local server.
5. Replay CLI test using a recorded bot fixture.
6. Godot headless smoke test that loads the project, connects to the server, completes move,
   attack, pickup, and equip, and verifies client state through the debug API.
7. CI runs shared validation, Go tests, replay fixture verification, Python bot smoke, and Godot
   headless smoke where the runtime is available.
