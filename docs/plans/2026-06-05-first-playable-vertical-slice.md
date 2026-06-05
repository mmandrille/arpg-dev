# First Playable Vertical Slice - Implementation Plan

Status: Ready for execution

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

- [ ] Step 1.1: Create directory structure matching the spec.
- [ ] Step 1.2: Initialize Go module under `server/`.
- [ ] Step 1.3: Initialize Python tooling for `tools/bot/` and `tools/replay/`.
- [ ] Step 1.4: Initialize minimal Godot `4.6.3-stable` project under `client/`.
- [ ] Step 1.5: Add root-level documented commands for starting Postgres, running the server,
  running tests, running the bot, and running replay verification.
- [ ] Step 1.6: Add exact toolchain pins: `.godot-version`, Go `go.mod` directive, Python
  `pyproject.toml` / `.tool-versions`, and Postgres image version.
- [ ] Step 1.7: Add Docker Compose for local Postgres and make `make db-up` the canonical DB
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

- [ ] Step 2.1: Define WebSocket envelope schema and message payload schemas.
- [ ] Step 2.2: Define `session_snapshot` and `state_delta` schemas with the exact payload shapes
  from the spec.
- [ ] Step 2.3: Define combat, item, monster, and loot table schemas.
- [ ] Step 2.4: Add v0 placeholder data for one player attack, one monster, one loot table, and one
  equippable item.
- [ ] Step 2.5: Add golden fixtures for expected combat, loot, pickup, equip, and final snapshot
  outcomes.
- [ ] Step 2.6: Add schema validation command covering `shared/protocol/`, `shared/rules/`, and
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

- [ ] Step 3.1: Add config loading for address, database URL, auth dev token, `DEBUG_TOKEN`, and
  debug-route settings.
- [ ] Step 3.2: Add structured JSON logger.
- [ ] Step 3.3: Add `/healthz`, `/readyz`, and `/metrics`.
- [ ] Step 3.4: Add request IDs/correlation IDs.
- [ ] Step 3.5: Add basic server tests.

Verification:

```bash
cd server && go test ./...
```

## Task 4: Postgres persistence and migrations

Files:
- Create: `server/internal/store/`
- Create: `server/migrations/`
- Create/Modify: dev DB scripts

- [ ] Step 4.1: Add migrations for accounts, characters, sessions, inventory items, session inputs,
  and session events.
- [ ] Step 4.2: Add repository interfaces and Postgres implementations.
- [ ] Step 4.3: Add integration tests against local Postgres.
- [ ] Step 4.4: Wire `/readyz` to database connectivity.

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

- [ ] Step 5.1: Implement `POST /v0/auth/dev-login`.
- [ ] Step 5.2: Implement bearer-token middleware.
- [ ] Step 5.3: Implement `POST /v0/sessions` create/resume.
- [ ] Step 5.4: Persist account, character, and session records.
- [ ] Step 5.5: Add failure tests for invalid token and invalid session.

Verification:

```bash
cd server && go test ./...
```

## Task 6: Deterministic authoritative simulation

Files:
- Create: `server/internal/game/`
- Modify: `server/internal/store/`
- Modify: `shared/rules/`

- [ ] Step 6.1: Implement fixed-tick solo session state.
- [ ] Step 6.2: Run the authoritative sim at 20 Hz and apply tick-tagged buffered inputs in
  deterministic `(tick, sequence, message_id)` order.
- [ ] Step 6.3: Allocate monotonic per-session unsigned 64-bit entity IDs and encode them as
  decimal strings in JSON.
- [ ] Step 6.4: Load and validate shared v0 rules.
- [ ] Step 6.5: Implement movement, attack, monster death, loot drop, pickup, and equip.
- [ ] Step 6.6: Ensure deterministic seeded RNG and stable entity ordering.
- [ ] Step 6.7: Persist authoritative inputs and events.
- [ ] Step 6.8: Add tests using shared golden fixtures.

Verification:

```bash
cd server && go test ./internal/game/... ./...
```

## Task 7: WebSocket realtime protocol

Files:
- Modify: `server/internal/http/`
- Create: `server/internal/realtime/`
- Modify: `shared/protocol/`

- [ ] Step 7.1: Implement authenticated WebSocket upgrade at `/v0/ws`.
- [ ] Step 7.2: Validate envelopes and payloads.
- [ ] Step 7.3: Emit `session_snapshot` and `state_delta` with payloads matching
  `shared/protocol/` schemas.
- [ ] Step 7.4: Emit `intent_accepted`, `intent_rejected`, and `error` messages with authoritative
  server ticks.
- [ ] Step 7.5: Add duplicate message handling, late input handling, and structured protocol errors.
- [ ] Step 7.6: Add WebSocket integration tests.

Verification:

```bash
cd server && go test ./...
```

## Task 8: Inspection API and replay

Files:
- Create: `server/internal/replay/`
- Modify: `server/internal/http/`
- Create: `tools/replay/`

- [ ] Step 8.1: Implement `GET /v0/sessions/{session_id}/state`.
- [ ] Step 8.2: Implement `GET /v0/sessions/{session_id}/replay`.
- [ ] Step 8.3: Implement replay verification from seed + recorded inputs.
- [ ] Step 8.4: Add replay mismatch failure tests.
- [ ] Step 8.5: Add CLI wrapper for replay verification.
- [ ] Step 8.6: Enforce bearer auth plus `X-Debug-Token` debug authorization on state and replay
  endpoints, with tests for missing/invalid debug authorization.

Verification:

```bash
make replay SESSION_ID=<recorded-session-id>
```

## Task 9: Python protocol bot

Files:
- Create: `tools/bot/`

- [ ] Step 9.1: Implement bot login and session creation.
- [ ] Step 9.2: Connect over WebSocket using the v0 envelope.
- [ ] Step 9.3: Complete move, attack, pickup, and equip.
- [ ] Step 9.4: Assert authoritative state through protocol messages and the `/state` API.
- [ ] Step 9.5: Assert persistence by resuming or restarting the session path.
- [ ] Step 9.6: Add bot command to root dev scripts.

Verification:

```bash
make bot
```

## Task 10: Minimal Godot client

Files:
- Create: `client/project.godot`
- Create: `client/scenes/`
- Create: `client/scripts/`

- [ ] Step 10.1: Create placeholder isometric scene with player, monster, loot marker, and inventory
  display.
- [ ] Step 10.2: Implement dev login and session creation.
- [ ] Step 10.3: Implement WebSocket client and v0 envelope parsing.
- [ ] Step 10.4: Send move, attack, pickup, and equip intents from input.
- [ ] Step 10.5: Render authoritative snapshots/deltas with placeholder visuals.
- [ ] Step 10.6: Add client debug state output for automation.
- [ ] Step 10.7: Implement minimal movement prediction and reconcile on authoritative
  `state_delta` / `session_snapshot`.
- [ ] Step 10.8: Report reconciliation deltas through client debug output and metrics payloads when
  available.
- [ ] Step 10.9: Add GDScript tests that consume at least one fixture from `shared/golden/`.

Verification:

```bash
make client-smoke
```

## Task 11: End-to-end verification

Files:
- Modify: root dev scripts
- Modify: docs as needed

- [ ] Step 11.1: Run Postgres and server.
- [ ] Step 11.2: Run shared validation.
- [ ] Step 11.3: Run Go tests.
- [ ] Step 11.4: Run Python bot.
- [ ] Step 11.5: Run replay verification for the bot session.
- [ ] Step 11.6: Run Godot client smoke test.
- [ ] Step 11.7: Update docs with as-built deviations.

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

- [ ] Step 12.1: Add a single `make ci` command that runs shared validation, Go tests, replay
  fixture verification, Python checks, and available Godot headless tests.
- [ ] Step 12.2: Add CI workflow using pinned Go, Python, Postgres, and Godot versions.
- [ ] Step 12.3: Ensure CI starts Postgres through the same Docker Compose configuration or an
  equivalent pinned service container.
- [ ] Step 12.4: Document any local-only Godot smoke requirement if the CI runner cannot install
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
