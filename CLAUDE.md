# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project progress (read first)

**Before any new feature or slice:** read [`docs/PROGRESS.md`](docs/PROGRESS.md).

That file is the canonical lifecycle doc: completed slices (v0–v4), current branch/merge status,
what each slice proved, known gaps, deferred backlog, and the agent checklist for starting work.
Update it when a slice ships.

**Current snapshot (2026-06-05):** slices through **v4 `take-a-hit`** are implemented on
`feature/animate-and-react` (`make ci` green); not yet merged to `main`. No v5 spec exists.

## Commands

Everything runs through `make`. Run `make help` for the full list.

```bash
# Infrastructure
make db-up           # start local Postgres (required before server or bot)
make db-reset        # destroy + recreate Postgres (drops all data)

# Running
make server          # run Go server (requires db-up first)
make bot             # run Python protocol bot end-to-end (server must be up)
make bot-visual      # server + interactive Godot window with ARPG_AUTOPLAY=1
make play            # Postgres + server + interactive Godot client

# Testing
make test-go         # all Go tests  (`cd server && go test ./...`)
make client-smoke    # Godot headless smoke tests (requires Godot binary)
make ci              # full local CI suite (shared validation + Go tests + bot + replay)

# Shared contracts
make validate-shared # validate all shared JSON against their schemas
make validate-assets # validate asset manifest + runtime .glb files

# Assets (regenerate deterministic committed files)
make gen-assets      # regenerate runtime .glb files
make gen-anims       # regenerate AnimationLibrary .tres clips (requires Godot)

# Replay
make replay SESSION_ID=<id>   # re-simulate a recorded session and verify output
```

**Single Go test:** `cd server && go test ./internal/game/... -run TestName`

**Single Python test:** `cd <repo-root> && .venv/bin/pytest tools/bot/test_protocol.py::test_name -v`

**Override env vars on the command line:**
```bash
AUTOPLAY_STEP_DELAY=0.8 make bot-visual   # slower playback
GODOT=/path/to/godot make bot-visual
SESSION_ID=abc123 make replay
```

## Architecture

```
client/   Godot 4 (GDScript) — thin client: renders + locally predicts movement/attack feedback
server/   Go authoritative server — owns all game state, loot rolls, inventory, persistence
shared/   Data contracts: JSON schemas, rules-as-data, cross-language golden fixtures
tools/    Python: protocol bot, replay wrapper, shared schema validator, asset pipeline tools
assets/   Source art manifests + asset generation scripts
docs/     ADRs + specs + plans + PROGRESS.md (slice lifecycle — read before new work)
```

### The authoritative boundary (ADR-0001 D2)
The client is a renderer + input layer; **the server owns every outcome that matters** (HP, damage, loot rolls, inventory). Even in solo play the client speaks the full production-shaped protocol over WebSocket. There is no local-only path or client-side shortcut.

### Server internals (`server/internal/`)
- **`game/`** — deterministic authoritative simulation (`Sim`). Given the same seed + ordered inputs it always produces identical output. Enforced: seeded PRNG only (`rng.go`), no `time.Now()` in game logic, stable entity-ID ordering, fixed 20 Hz tick. Rules loaded from `shared/rules/` at startup (`rules.go`).
- **`realtime/`** — WebSocket hub + per-session runner. `Hub.Run()` upgrades the connection, constructs a `Sim`, and enters the session loop.
- **`store/`** — repository interface + Postgres implementation. Sessions, inventory, events all persist here.
- **`auth/`, `http/`, `replay/`** — platform services (auth, REST endpoints, replay command).
- **`cmd/arpg-server`** — main binary (also self-migrates on boot). **`cmd/arpg-replay`** — standalone replay verifier.

### Shared contracts (`shared/`)
- **`protocol/`** — JSON schemas for WebSocket messages (`envelope`, `messages`, `state_delta`, `session_snapshot`). Wire format is JSON over WebSocket in v0; both Go and GDScript consume these.
- **`rules/`** — combat formulas, item definitions, monster definitions, loot tables as JSON data. The Go sim and Godot client both evaluate formulas from this shared catalog — never from language-specific logic.
- **`golden/`** — cross-language fixture files (damage formulas, loot rolls, slice outcomes, retaliation damage). Go tests and Godot `test_golden.gd` both assert against these.

### Client (`client/`)
- `scripts/main.gd` — top-level coordinator: manages `net_client`, entity map (`{node, controller, type}`), routes server events to entity animation controllers, reads `state_delta.events` array for authoritative triggers.
- `scripts/animation_controller.gd` — injected per-entity. Priority state machine: terminal (death) > one-shot (hit/attack) > locomotion (idle/walk). Drives `AnimationPlayer`; never crosses the wire.
- `scripts/net_client.gd` — WebSocket connection, serializes/deserializes protocol envelopes.
- `tests/test_golden.gd` — headless cross-language golden checks run via `make client-smoke`.

### Python tools (`tools/`)
- `bot/run.py` — protocol bot: authenticates, creates a session, sends move/attack/pickup/equip intents over the same WebSocket the real client uses, asserts on authoritative state. Primary agent-playability path.
- `replay/run.py` — replay wrapper (invoked via `make replay`).
- `validate_shared.py` — validates all JSON in `shared/` against their schemas.
- `assets/gen_glb.py` — generates committed runtime `.glb` files deterministically.
- `assets/validate_assets.py` — validates the asset manifest + GLB node presence.

### Animation model (ADR-0007)
Animation is **client-only presentation state** — never on the wire. Player locomotion (`idle/walk/attack`) is driven by client input/prediction signals. Player damage/death and monster reactions (`hit/death`) are driven by authoritative events (`player_damaged`, `player_killed`, `monster_damaged`, `monster_killed`) already present in every `state_delta.events`. No server or protocol change is needed to add new reactions — only a server event emission and a client mapping.

## SDD Process

This project uses Spec-Driven Development. Before touching code for any new feature:

1. Read [`docs/PROGRESS.md`](docs/PROGRESS.md) — baseline slice, open gaps, invariants.
2. Read or write the spec under `docs/specs/spec-<feature>.md`.
3. Write or check the plan under `docs/plans/<date>-<feature>.md`.
4. Consult the relevant ADRs in `docs/adr/` — especially ADR-0001 (foundational) and any feature-specific ones.
5. When the slice completes, update `docs/PROGRESS.md` (lifecycle table + summary + new gaps).

## Key Invariants

- **Determinism in the Go sim is non-negotiable.** No `time.Now()`, `rand.Intn()` without the seeded `RNG`, or map iteration in game logic (`game/` package only). Any violation breaks replay.
- **Shared rules are data, not code.** Formula types live in `shared/rules/`; Go and GDScript each implement the same small evaluator set. Never add executable logic to shared data files.
- **Golden fixtures are cross-language contracts.** Any change to combat formulas or loot logic must update the golden files in `shared/golden/` and both the Go and GDScript golden tests.
- **Protocol JSON schemas are versioned.** Changes to `shared/protocol/` require a schema version bump and must remain backward-compatible or require coordinated client+server update.
- **Asset manifest is the source of truth for asset identity.** `assets/manifests/assets.v0.json` maps `asset_id → runtime_path`. `shared/assets/item_visuals.v0.json` maps `item_def_id → asset_id + mount_socket`. These two files are the only canonical link between gameplay and visuals.
