# v175 Plan — Elite Objective HUD

Status: Ready for implementation
Goal: Add a compact current-floor HUD tracker for generated elite-objective reward chest state.
Architecture: The server remains unchanged. The client derives one tracker state from current entities: `elite_objective` chest presence/open state plus alive `monster_pack_leader` count. A focused tracker script owns rendering and debug state; `main.gd` only feeds derived dictionaries and exposes bot debug state.
Tech stack: Godot client UI/tests, client bot scenario JSON, lifecycle docs.

## Baseline and Shortcut Decision

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/elite_objective_tracker.gd` | Render tracker state and expose debug data |
| Modify | `client/scripts/main.gd` | Create tracker, derive state, update on snapshots/deltas, expose debug |
| Add | `client/tests/test_elite_objective_tracker.gd` | Unit coverage for tracker states |
| Modify | `scripts/client_smoke.sh` | Add tracker unit gate |
| Add | `client/scripts/bot_elite_objective_assertions.gd` | Bot matcher for tracker state |
| Modify | `client/scripts/bot_step_catalog.gd` | Register tracker assertion/wait steps |
| Modify | `client/scripts/bot_wait_handlers.gd` | Wait on tracker state |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert tracker state |
| Add | `tools/bot/scenarios/client/44_elite_objective_hud.json` | Client bot proof |
| Add | `docs/as-built/v175_elite-objective-hud.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: tracker rendering and bot matching live in focused new files; `main.gd` receives minimal state derivation and stays under its lowered baseline.

Verification:
```bash
make maintainability
```

## Task 1 — Tracker UI

Files:
- Add: `client/scripts/elite_objective_tracker.gd`
- Add: `client/tests/test_elite_objective_tracker.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 1.1: Implement tracker state rendering for hidden, active, claim, and complete.
- [x] Step 1.2: Expose `set_state` and `get_debug_state`.
- [x] Step 1.3: Add a focused client unit gate.
```bash
make client-unit
```

## Task 2 — Main Client Wiring

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Instantiate tracker in gameplay UI.
- [x] Step 2.2: Derive state from elite objective chest and alive elite leader records on snapshot/delta.
- [x] Step 2.3: Include tracker state in bot debug output.
```bash
make client-unit
make maintainability
```

## Task 3 — Bot Scenario

Files:
- Add: `client/scripts/bot_elite_objective_assertions.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_wait_handlers.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/44_elite_objective_hud.json`

- [x] Step 3.1: Add wait/assert support for elite objective tracker debug state.
- [x] Step 3.2: Create a pinned client scenario that descends to the elite objective floor and asserts active leader count.
```bash
make bot-client scenario=44_elite_objective_hud.json
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v175_elite-objective-hud.md`
- Modify: `docs/plans/v175_2026-06-14-elite-objective-hud.md`
- Modify: `docs/specs/v175_spec-elite-objective-hud.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark plan tasks complete and write as-built notes.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle and next-slice pointer.
```bash
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=44_elite_objective_hud.json`
- [x] `make ci`
