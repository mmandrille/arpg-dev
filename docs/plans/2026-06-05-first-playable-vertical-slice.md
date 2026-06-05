# First Playable Vertical Slice - Implementation Plan

Status: Complete (2026-06-05) — all tasks implemented and verified via `make ci`

Goal: prove the ADR-0001 architecture with the smallest playable end-to-end flow.
Architecture: Godot client and Go server as separate apps; server-authoritative game state; Postgres
persistence; JSON-over-WebSocket realtime protocol; Python bot and replay verification.
Tech stack: Go, Godot 4/GDScript, Python, Postgres, JSON Schema.

Related spec: `docs/specs/spec-first-playable-vertical-slice.md`

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/` | Go service module, HTTP routes, WebSocket, sim, persistence, tests |
| Create | `shared/protocol/` | Protocol schemas and examples |
| Create | `shared/golden/` | Cross-language fixtures consumed by Go and GDScript tests |
| Create | `shared/rules/` | Rules schemas and v0 data |
| Create | `tools/bot/` | Python protocol bot integration driver |
| Create | `tools/replay/` | Replay verification CLI wrapper |
| Create | `client/` | Minimal Godot client |
| Create | `docker-compose.yml` | Local Postgres dev runner |
| Create | `.godot-version` | Exact Godot pin: `4.6.3-stable` |
| Create | `pyproject.toml` / `.tool-versions` | Python and toolchain pins |
| Create | `make/` or root scripts | Dev commands for server, DB, bot, tests, CI |
| Modify | `docs/` | As-built notes if implementation differs from the spec |

## Task 1: Repository scaffolding and dev commands

Files:
- Create: `server/`
- Create: `shared/`
- Create: `tools/`
- Create: `client/`
- Create: `docker-compose.yml`
- Create: `.godot-version`
- Create: `pyproject.toml` / `.tool-versions`
- Create/Modify: `make/` or root dev scripts

- [x] Step 1.1: Create directory structure matching the spec.
- [x] Step 1.2: Initialize Go module under `server/`.
- [x] Step 1.3: Initialize Python tooling for `tools/bot/` and `tools/replay/`.
- [x] Step 1.4: Initialize minimal Godot `4.6.3-stable` project under `client/`.
- [x] Step 1.5: Add root-level documented commands for starting Postgres, running the server,
  running tests, running the bot, and running replay verification.
- [x] Step 1.6: Add exact toolchain pins: `.godot-version`, Go `go.mod` directive, Python
  `pyproject.toml` / `.tool-versions`, and Postgres image version.
- [x] Step 1.7: Add Docker Compose for local Postgres and make `make db-up` the canonical DB
  startup path.

Verification:

```bash
find server shared tools client -maxdepth 2 -type d
test "$(cat .godot-version)" = "4.6.3-stable"
make db-up
```

## Task 2: Shared protocol and rules contracts

Files:
- Create: `shared/protocol/*.schema.json`
- Create: `shared/protocol/examples/*.json`
- Create: `shared/rules/*.schema.json`
- Create: `shared/rules/*.json`
- Create: `shared/golden/*.json`

- [x] Step 2.1: Define WebSocket envelope schema and message payload schemas.
- [x] Step 2.2: Define `session_snapshot` and `state_delta` schemas with the exact payload shapes
  from the spec.
- [x] Step 2.3: Define combat, item, monster, and loot table schemas.
- [x] Step 2.4: Add v0 placeholder data for one player attack, one monster, one loot table, and one
  equippable item.
- [x] Step 2.5: Add golden fixtures for expected combat, loot, pickup, equip, and final snapshot
  outcomes.
- [x] Step 2.6: Add schema validation command covering `shared/protocol/`, `shared/rules/`, and
  `shared/golden/`.

Verification:

```bash
make validate-shared
```

## Task 3: Go server baseline

Files:
- Create: `server/cmd/arpg-server/main.go`
- Create: `server/internal/config/`
- Create: `server/internal/http/`
- Create: `server/internal/logging/`
- Create: `server/internal/metrics/`

- [x] Step 3.1: Add config loading for address, database URL, auth dev token, `DEBUG_TOKEN`, and
  debug-route settings.
- [x] Step 3.2: Add structured JSON logger.
- [x] Step 3.3: Add `/healthz`, `/readyz`, and `/metrics`.
- [x] Step 3.4: Add request IDs/correlation IDs.
- [x] Step 3.5: Add basic server tests.

Verification:

```bash
cd server && go test ./...
```

## Task 4: Postgres persistence and migrations

Files:
- Create: `server/internal/store/`
- Create: `server/migrations/`
- Create/Modify: dev DB scripts

- [x] Step 4.1: Add migrations for accounts, characters, sessions, inventory items, session inputs,
  and session events.
- [x] Step 4.2: Add repository interfaces and Postgres implementations.
- [x] Step 4.3: Add integration tests against local Postgres.
- [x] Step 4.4: Wire `/readyz` to database connectivity.

Verification:

```bash
make db-up
cd server && go test ./internal/store/...
```

## Task 5: Auth and session lifecycle

Files:
- Create: `server/internal/auth/`
- Modify: `server/internal/http/`
- Modify: `server/internal/store/`

- [x] Step 5.1: Implement `POST /v0/auth/dev-login`.
- [x] Step 5.2: Implement bearer-token middleware.
- [x] Step 5.3: Implement `POST /v0/sessions` create/resume.
- [x] Step 5.4: Persist account, character, and session records.
- [x] Step 5.5: Add failure tests for invalid token and invalid session.

Verification:

```bash
cd server && go test ./...
```

## Task 6: Deterministic authoritative simulation

Files:
- Create: `server/internal/game/`
- Modify: `server/internal/store/`
- Modify: `shared/rules/`

- [x] Step 6.1: Implement fixed-tick solo session state.
- [x] Step 6.2: Run the authoritative sim at 20 Hz and apply tick-tagged buffered inputs in
  deterministic `(tick, sequence, message_id)` order.
- [x] Step 6.3: Allocate monotonic per-session unsigned 64-bit entity IDs and encode them as
  decimal strings in JSON.
- [x] Step 6.4: Load and validate shared v0 rules.
- [x] Step 6.5: Implement movement, attack, monster death, loot drop, pickup, and equip.
- [x] Step 6.6: Ensure deterministic seeded RNG and stable entity ordering.
- [x] Step 6.7: Persist authoritative inputs and events.
- [x] Step 6.8: Add tests using shared golden fixtures.

Verification:

```bash
cd server && go test ./internal/game/... ./...
```

## Task 7: WebSocket realtime protocol

Files:
- Modify: `server/internal/http/`
- Create: `server/internal/realtime/`
- Modify: `shared/protocol/`

- [x] Step 7.1: Implement authenticated WebSocket upgrade at `/v0/ws`.
- [x] Step 7.2: Validate envelopes and payloads.
- [x] Step 7.3: Emit `session_snapshot` and `state_delta` with payloads matching
  `shared/protocol/` schemas.
- [x] Step 7.4: Emit `intent_accepted`, `intent_rejected`, and `error` messages with authoritative
  server ticks.
- [x] Step 7.5: Add duplicate message handling, late input handling, and structured protocol errors.
- [x] Step 7.6: Add WebSocket integration tests.

Verification:

```bash
cd server && go test ./...
```

## Task 8: Inspection API and replay

Files:
- Create: `server/internal/replay/`
- Modify: `server/internal/http/`
- Create: `tools/replay/`

- [x] Step 8.1: Implement `GET /v0/sessions/{session_id}/state`.
- [x] Step 8.2: Implement `GET /v0/sessions/{session_id}/replay`.
- [x] Step 8.3: Implement replay verification from seed + recorded inputs.
- [x] Step 8.4: Add replay mismatch failure tests.
- [x] Step 8.5: Add CLI wrapper for replay verification.
- [x] Step 8.6: Enforce bearer auth plus `X-Debug-Token` debug authorization on state and replay
  endpoints, with tests for missing/invalid debug authorization.

Verification:

```bash
make replay SESSION_ID=<recorded-session-id>
```

## Task 9: Python protocol bot

Files:
- Create: `tools/bot/`

- [x] Step 9.1: Implement bot login and session creation.
- [x] Step 9.2: Connect over WebSocket using the v0 envelope.
- [x] Step 9.3: Complete move, attack, pickup, and equip.
- [x] Step 9.4: Assert authoritative state through protocol messages and the `/state` API.
- [x] Step 9.5: Assert persistence by resuming or restarting the session path.
- [x] Step 9.6: Add bot command to root dev scripts.

Verification:

```bash
make bot
```

## Task 10: Minimal Godot client

Files:
- Create: `client/project.godot`
- Create: `client/scenes/`
- Create: `client/scripts/`

- [x] Step 10.1: Create placeholder isometric scene with player, monster, loot marker, and inventory
  display.
- [x] Step 10.2: Implement dev login and session creation.
- [x] Step 10.3: Implement WebSocket client and v0 envelope parsing.
- [x] Step 10.4: Send move, attack, pickup, and equip intents from input.
- [x] Step 10.5: Render authoritative snapshots/deltas with placeholder visuals.
- [x] Step 10.6: Add client debug state output for automation.
- [x] Step 10.7: Implement minimal movement prediction and reconcile on authoritative
  `state_delta` / `session_snapshot`.
- [x] Step 10.8: Report reconciliation deltas through client debug output and metrics payloads when
  available.
- [x] Step 10.9: Add GDScript tests that consume at least one fixture from `shared/golden/`.

Verification:

```bash
make client-smoke
```

## Task 11: End-to-end verification

Files:
- Modify: root dev scripts
- Modify: docs as needed

- [x] Step 11.1: Run Postgres and server.
- [x] Step 11.2: Run shared validation.
- [x] Step 11.3: Run Go tests.
- [x] Step 11.4: Run Python bot.
- [x] Step 11.5: Run replay verification for the bot session.
- [x] Step 11.6: Run Godot client smoke test.
- [x] Step 11.7: Update docs with as-built deviations.

Final verification:

```bash
make validate-shared
cd server && go test ./...
make bot
make replay SESSION_ID=<recorded-session-id>
make client-smoke
```

## Task 12: CI validation

Files:
- Create/Modify: CI workflow files
- Modify: root dev scripts

- [x] Step 12.1: Add a single `make ci` command that runs shared validation, Go tests, replay
  fixture verification, Python checks, and available Godot headless tests.
- [x] Step 12.2: Add CI workflow using pinned Go, Python, Postgres, and Godot versions.
- [x] Step 12.3: Ensure CI starts Postgres through the same Docker Compose configuration or an
  equivalent pinned service container.
- [x] Step 12.4: Document any local-only Godot smoke requirement if the CI runner cannot install
  the pinned Godot runtime yet.

Verification:

```bash
make ci
```

## Sequencing notes

- Do not build real art before the bot can complete the slice.
- Do not add multiplayer orchestration before the solo authoritative session is replayable.
- Do not introduce Protobuf before JSON message shapes survive the bot and client flows.
- Do not allow anonymous local-only gameplay; even dev mode must pass through account and session
  concepts.
- Use `/v0/sessions/{session_id}/state` as the canonical live inspection endpoint; do not add a
  parallel `/inspect` route in v0.

## As-Built Notes

Deviations and concrete decisions made during implementation (status: all 12 tasks complete; the
full slice passes via `make ci` — shared validation, Go tests, Python bot, replay verification, and
the Godot 4.6.3 headless smoke).

### Toolchain (verified-latest-at-scaffold)
- **Go**: `go.mod` pins `go 1.24`; the local/CI toolchain is Go 1.25.x. ADR baseline was 1.24.x.
- **Python**: `pyproject.toml` requires `>=3.12`; the local interpreter is 3.14.x. Lower bounds are
  pinned and upper bounds left open so tooling runs on the newer interpreter present at scaffold.
- **Postgres**: `postgres:16.4` via Docker Compose. The Docker daemon is provided by **Colima**
  (not Docker Desktop), and Compose is invoked as `docker-compose` (the v5 standalone binary); the
  `docker compose` plugin is not present. `make db-up` works with either daemon.
- **Godot**: `4.6.3-stable`, installed via Homebrew cask.

### Architecture / contract decisions
- **Inventory persistence is session-scoped**, not character-scoped. `inventory_items` is keyed by
  `(session_id, id)`. Reason: `id` is the protocol `item_instance_id`, allocated by the
  deterministic per-session entity counter (which restarts at 1001 each session), so it is unique
  only within a session. Spec 4.6 lists inventory as character-scoped with no `session_id`; honoring
  that with per-session ids caused cross-session id collisions (a later session's pickup `1004`
  conflicted with an earlier one, and equip mutated the wrong row). "Survives restart" (acceptance
  #8) is proven by reconnecting to the **same** session, which reloads inventory from Postgres into
  a fresh sim. Durable cross-session character inventory is deferred post-v0.
- **`session_inputs.payload` stores the full envelope** (not just the intent payload), because spec
  4.6 has no `type` column and replay needs the message type to re-apply inputs.
- **WebSocket auth accepts `?access_token=` query param** in addition to the `Authorization: Bearer`
  header, because `WebSocketPeer` (Godot) and browsers cannot reliably set handshake headers. The
  Python bot and Go tests use the header; the Godot client uses the query param.
- **Monster death leaves a corpse** (`entity_update` to hp 0), matching the spec `state_delta`
  example; `entity_remove` is used only for loot on pickup.
- **`state_delta.changes`/`events` are always emitted as `[]`** (never `null`); the sim guarantees
  non-nil slices so the wire conforms to the schema and dynamically-typed clients don't choke.
- **Combat has no attack-range gate in v0**: `base_hit_chance` is 1.0 and an attack on a living
  monster always hits. Movement is real (predicted + reconciled) but not a precondition for the
  kill, keeping the bot/client robust.
- **Metrics** use `prometheus/client_golang` for `/metrics` (Prometheus-compatible exposition).
- **Plan step 6.7** (persist authoritative inputs + events) is implemented in the realtime session
  runner (Task 7), where inputs actually arrive, rather than inside the pure sim package.
- **Replay** reconstructs authoritative state from a session's own recorded input stream; it does
  not replay cross-session resumed inventory.
- A small `internal/ids` package (hand-rolled prefixed ULIDs) was added for platform identities;
  these are explicitly outside the deterministic sim (which uses the per-session counter).
