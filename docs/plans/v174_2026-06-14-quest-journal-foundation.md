# v174 Plan — Quest Journal Foundation

Status: Ready for implementation
Goal: Add a `J`-toggle client quest journal that reflects current-floor reward chest objective state.
Architecture: The server remains unchanged; the client derives journal rows from already-authoritative entity metadata (`quest_reward`, `interactable_def_id`, and `state`). A focused `quest_journal_panel.gd` owns layout and debug state, while `main.gd` only creates the panel, toggles it, and feeds it derived objective dictionaries. Bot checks read the same debug state.
Tech stack: Godot client UI/tests, client bot scenario JSON, lifecycle docs.

## Baseline and Shortcut Decision

Builds on v155 random quest reward floors and v173 quest reward chest metadata. Godot plugin adoption checklist: reject external plugins because this is a small draggable-window style panel using existing in-repo UI patterns; no general quest framework or asset dependency is needed.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/quest_journal_panel.gd` | Render/toggle quest objective state and expose debug data |
| Modify | `client/scripts/main.gd` | Create panel, bind `J`, derive current objective state, expose bot debug |
| Add | `client/tests/test_quest_journal_panel.gd` | Unit coverage for empty, active, and complete states |
| Modify | `scripts/client_smoke.sh` | Add quest journal panel unit gate |
| Modify | `client/scripts/bot_scenario_runner.gd` | Validate and assert quest journal debug state |
| Add | `tools/bot/scenarios/client/43_quest_journal_foundation.json` | Client bot proof on pinned reward floor |
| Add | `docs/as-built/v174_quest-journal-foundation.md` | As-built summary |
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
- [x] Defer extraction with rationale: `quest_journal_panel.gd` owns UI behavior, `bot_quest_journal_assertions.gd` owns bot matching, and `main.gd` only keeps minimal wiring with its baseline lowered after redundant blank-line reductions.

Verification:
```bash
make maintainability
```

## Task 1 — Journal Panel

Files:
- Add: `client/scripts/quest_journal_panel.gd`
- Add: `client/tests/test_quest_journal_panel.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 1.1: Implement a small draggable-window-compatible quest journal panel with empty, active, and complete objective states.
- [x] Step 1.2: Expose `set_objectives`, `toggle`, `hide_display`, and `get_debug_state`.
- [x] Step 1.3: Add a focused unit test gate for panel state and debug output.
```bash
make client-unit
```

## Task 2 — Main Client Wiring

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Instantiate the journal panel in gameplay UI and include it in window raising/closing behavior.
- [x] Step 2.2: Add `J` key handling and derive objectives from `quest_reward` chest records.
- [x] Step 2.3: Include quest journal visibility/state in bot debug output.
```bash
make client-unit
make maintainability
```

## Task 3 — Bot Scenario

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Add: `tools/bot/scenarios/client/43_quest_journal_foundation.json`

- [x] Step 3.1: Add `assert_quest_journal` / wait support against bot debug state.
- [x] Step 3.2: Create a pinned client scenario that descends to `v155_bot_quest_0015`, opens the journal, and asserts the active reward objective.
```bash
make bot-client scenario=43_quest_journal_foundation.json
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v174_quest-journal-foundation.md`
- Modify: `docs/plans/v174_2026-06-14-quest-journal-foundation.md`
- Modify: `docs/specs/v174_spec-quest-journal-foundation.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark plan tasks complete and write as-built notes.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle and next-slice pointer.
```bash
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=43_quest_journal_foundation.json`
- [x] `make ci`
