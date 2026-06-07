# arpg-dev

Personal experiment: an isometric action-RPG / looter built almost entirely by
AI agents. See [`docs/adr/0001-technology-stack.md`](docs/adr/0001-technology-stack.md)
for the foundational architecture and [`docs/specs/v1_spec-first-playable-vertical-slice.md`](docs/specs/v1_spec-first-playable-vertical-slice.md)
for the current slice.

Architecture in one line: a **Godot 4 client** and a **Go authoritative server**
as separate apps in one repo, talking JSON over WebSocket, with **Postgres**
persistence, a **deterministic seeded simulation** that is **replayable**, and a
**Python protocol bot** that plays and verifies the game the same way the client
does.

```
client/   Godot 4 (GDScript) thin client — renders + predicts, never authoritative
server/   Go realtime game server + platform services (auth, sessions, persistence)
shared/   data contracts: protocol schemas, rules-as-data, cross-language golden fixtures
tools/    Python protocol bot + replay verification wrapper + shared schema validator
docs/     ADRs, specs, plans
```

## Toolchain (pinned)

| Component | Pin | Source of truth |
|-----------|-----|-----------------|
| Godot | `4.6.3-stable` | `.godot-version` |
| Go | `1.24` | `server/go.mod` |
| Python | `3.12` (lower bound) | `pyproject.toml` / `.tool-versions` |
| Postgres | `16.4` | `docker-compose.yml` |

## Dev commands

Everything runs through the `Makefile`. Run `make help` for the full list.

```bash
make db-up           # start local Postgres (canonical DB startup path)
make server          # run the Go server against local Postgres
make test            # unit tests: shared validation, Go, Python, client unit
make test-go         # run all Go tests
make test-py         # run Python unit tests (tools/)
make validate-shared # validate all shared JSON against schemas
make validate-assets # validate the asset manifest + runtime .glb files
make gen-assets      # regenerate the committed runtime .glb files (deterministic)
make bot             # run the Python protocol bot end-to-end (server must be up)
make bot-visual      # scripts/bot_visual.sh: server + Godot autoplay (see below)
make replay SESSION_ID=<id>   # re-simulate and verify a recorded session
make client-smoke    # Godot headless smoke (skips if Godot not installed)
make ci              # full local CI suite
```

To watch the scripted slice in the real client (move → attack → pickup → equip,
including player hit reactions), run `make bot-visual`. It builds the server,
waits for `/readyz`, then opens Godot with `ARPG_AUTOPLAY=1` via
`scripts/bot_visual.sh`. Close the Godot window to stop the server.

```bash
make bot-visual
AUTOPLAY_STEP_DELAY=0.8 make bot-visual  # slower, easier to inspect
GODOT=/path/to/godot make bot-visual       # override Godot binary
```

To play manually, run `make play`. It starts Postgres + the local server and opens
the Godot main menu; create or continue a named character to start a fresh
`dungeon_levels` session. Direct session startup is a dev override only:
`ARPG_AUTOSTART=1 ARPG_WORLD_ID=dungeon_levels make play`.

Headless cross-language golden checks (including `retaliation_damage.json`) run in
`make client-smoke` via `client/tests/test_golden.gd`.

Typical first run:

```bash
make bot-visual
```

## Status

First playable vertical slice — see the spec and plan under `docs/`.
