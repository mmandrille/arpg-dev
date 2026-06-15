# v176 Plan — Elite Objective Minimap Pin

Status: Ready for implementation
Goal: Add a compact minimap-style HUD pin for the current floor's closed elite-objective reward chest.
Architecture: The server remains unchanged. A focused state helper derives normalized minimap coordinates from existing client entity records and local player position. A focused Control script owns drawing and debug state; `main.gd` only creates, updates, and exposes the widget.
Tech stack: Godot client UI/tests, client bot scenario JSON, lifecycle docs.

## Baseline and Shortcut Decision

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/elite_objective_minimap.gd` | Render minimap widget and expose debug data |
| Add | `client/scripts/elite_objective_minimap_state.gd` | Derive minimap state from entity records and player position |
| Modify | `client/scripts/main.gd` | Create minimap, update it on snapshots/deltas, expose debug |
| Add | `client/tests/test_elite_objective_minimap.gd` | Unit coverage for minimap states and coordinate clamping |
| Modify | `scripts/client_smoke.sh` | Add minimap unit gate |
| Add | `client/scripts/bot_elite_objective_minimap_assertions.gd` | Bot matcher for minimap debug state |
| Modify | `client/scripts/bot_step_catalog.gd` | Register minimap assertion/wait steps |
| Modify | `client/scripts/bot_wait_handlers.gd` | Wait on minimap state |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert minimap state |
| Add | `tools/bot/scenarios/client/45_elite_objective_minimap_pin.json` | Client bot proof |
| Add | `docs/as-built/v176_elite-objective-minimap-pin.md` | As-built summary |
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
- [ ] Defer extraction with rationale: not needed; minimap rendering, state derivation, and bot matching live in focused new files.

Verification:
```bash
make maintainability
```

## Task 1 — Minimap UI and State

Files:
- Add: `client/scripts/elite_objective_minimap.gd`
- Add: `client/scripts/elite_objective_minimap_state.gd`
- Add: `client/tests/test_elite_objective_minimap.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 1.1: Implement state derivation for hidden, active, and complete states from entity records.
- [x] Step 1.2: Implement minimap drawing and `get_debug_state`.
- [x] Step 1.3: Add focused client unit coverage and smoke gate.
```bash
make client-unit
```

## Task 2 — Main Client Wiring

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Instantiate the minimap in gameplay UI.
- [x] Step 2.2: Update minimap state after snapshots and deltas.
- [x] Step 2.3: Include minimap state in bot debug output while keeping `main.gd` within the maintainability allowance.
```bash
make client-unit
make maintainability
```

## Task 3 — Bot Scenario

Files:
- Add: `client/scripts/bot_elite_objective_minimap_assertions.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_wait_handlers.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/45_elite_objective_minimap_pin.json`

- [x] Step 3.1: Add wait/assert support for elite objective minimap debug state.
- [x] Step 3.2: Create a pinned client scenario that descends to the elite objective floor and asserts the active pin.
```bash
make bot-client scenario=45_elite_objective_minimap_pin.json
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v176_elite-objective-minimap-pin.md`
- Modify: `docs/plans/v176_2026-06-15-elite-objective-minimap-pin.md`
- Modify: `docs/specs/v176_spec-elite-objective-minimap-pin.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark plan tasks complete and write as-built notes.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle and next-slice pointer.
```bash
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=45_elite_objective_minimap_pin.json`
- [x] `make ci`
