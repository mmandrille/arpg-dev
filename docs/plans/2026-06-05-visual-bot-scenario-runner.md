# Visual Bot Scenario Runner - Implementation Plan

Goal: make every bot scenario discoverable, recordable, replay-verifiable, and visually playable through `make bot-visual`.
Architecture: scenario JSON catalog drives Python protocol runs; server exposes a debug replay timeline; Godot plays a manifest of replay timelines.
Tech stack: Python bot tooling, Go replay/http, Godot GDScript, Make/bash.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/bot/scenarios/vertical_slice.json` | First declarative scenario |
| Modify | `tools/bot/run.py` | Discover and execute scenario catalog, write manifest |
| Modify | `tools/bot/test_protocol.py` | Catalog and helper tests |
| Modify | `server/internal/replay/replay.go` | Build protocol-shaped timeline envelopes |
| Modify | `server/internal/http/inspect.go` | Add debug-gated timeline route |
| Modify | `server/internal/http/replay_test.go` | Cover timeline route |
| Modify | `client/scripts/main.gd` | Visual replay playlist mode |
| Modify | `scripts/bot_visual.sh` | Record scenarios before launching Godot |
| Modify | `make/agents.mk` | Update target description |
| Modify | `docs/PROGRESS.md` | Record v6 when complete |

## Task 1: Scenario Catalog

Files:
- Create: `tools/bot/scenarios/vertical_slice.json`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Add JSON scenario loader.
- [x] Replace single hardcoded bot flow with generic supported actions.
- [x] Keep existing assertions as named assertions.
- [x] Add `--scenario`, `--list-scenarios`, and `--write-manifest`.
- [x] Run Python tests.

```bash
make tools
.venv/bin/python -m pytest -q tools
```

## Task 2: Replay Timeline Endpoint

Files:
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/http/inspect.go`
- Modify: `server/internal/http/replay_test.go`

- [x] Add timeline reconstruction from stored inputs.
- [x] Return protocol-shaped `session_snapshot` and `state_delta` envelopes.
- [x] Gate route with auth and debug token.
- [x] Add endpoint tests.

```bash
cd server && go test ./internal/replay ./internal/http
```

## Task 3: Visual Playlist

Files:
- Modify: `client/scripts/main.gd`
- Modify: `scripts/bot_visual.sh`
- Modify: `make/agents.mk`

- [x] Add Godot replay-manifest mode.
- [x] Fetch timeline for each manifest session.
- [x] Apply envelopes through existing handlers at autoplay cadence.
- [x] Update `bot-visual` orchestration.
- [x] Exit Godot normally after the visual replay playlist completes.

```bash
make bot-visual
```

## Task 4: Documentation and Final Verification

Files:
- Modify: `docs/PROGRESS.md`

- [x] Update lifecycle table and progress summary.
- [x] Run focused tests.
- [x] Run `make ci` if time/environment allows.

```bash
make ci
```
