# arpg-dev

Personal experiment: an isometric action-RPG / looter built almost entirely by
AI agents. See [`docs/adr/0001-technology-stack.md`](docs/adr/0001-technology-stack.md)
for the foundational architecture and [`docs/specs/spec-first-playable-vertical-slice.md`](docs/specs/spec-first-playable-vertical-slice.md)
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
make test-go         # run all Go tests
make validate-shared # validate all shared JSON against schemas
make validate-assets # validate the asset manifest + runtime .glb files
make gen-assets      # regenerate the committed runtime .glb files (deterministic)
make bot             # run the Python protocol bot end-to-end (server must be up)
make bot-visual      # open Godot and visibly autoplay the bot slice
make replay SESSION_ID=<id>   # re-simulate and verify a recorded session
make client-smoke    # Godot headless smoke (skips if Godot not installed)
make ci              # full local CI suite
```

To watch the scripted slice in the real client, run:

```bash
make bot-visual
AUTOPLAY_STEP_DELAY=0.8 make bot-visual  # slower, easier to inspect
```

Typical first run:

```bash
make db-up
make server          # in one terminal
make bot             # in another — plays login -> session -> move -> attack -> pickup -> equip
```

## Status

First playable vertical slice — see the spec and plan under `docs/`.
