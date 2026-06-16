# v207 Plan - Mercenary Roster UI

Status: Complete
Goal: Add a Godot mercenary roster panel that opens from the v206 board event and shows the fixed guard offer plus the current hired mercenary.
Architecture: Keep server behavior unchanged. The client listens to `mercenary_board_opened` and `mercenary_hired`, owns display state in a focused panel script, and exposes debug state for unit/client-bot proof.
Tech stack: Godot GDScript UI, existing client bot runner/catalog, docs.

## Baseline and shortcut decision

Builds on v206. Asset/plugin decision: borrow existing in-repo draggable panel and companion-bar patterns; reject external assets/plugins and new bitmap/vector art.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/mercenary_panel.gd` | Mercenary offer/roster UI and debug state. |
| Add | `client/scripts/mercenary_panel_bridge.gd` | Focused event-to-panel glue so `main.gd` stays under the ratchet. |
| Add | `client/scripts/bot_debug_progression_setup.gd` | Client-bot-only debug gold seeding for deterministic hire proof. |
| Add | `client/scripts/bot_mercenary_panel_assertions.gd` | Focused mercenary panel and companion-bar assertion matching. |
| Modify | `client/scripts/main.gd` | Wire panel construction, board/hire events, gold refresh, companion refresh, and debug state. |
| Modify | `client/scripts/net_client.gd` | Add a debug progression PUT helper for client-bot setup. |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate new wait/assert mercenary panel steps. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add step detail only; matching lives in the focused assertion helper. |
| Modify | `client/scripts/bot_wait_handlers.gd` | Route mercenary panel wait steps. |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Route mercenary panel assert steps. |
| Modify | `scripts/bot_client.sh` | Pass scenario debug gold to the Godot client bot. |
| Modify | `scripts/client_smoke.sh` | Include the mercenary panel unit in `make client-unit`. |
| Add | `client/tests/test_mercenary_panel.gd` | Focused panel unit coverage. |
| Add | `tools/bot/scenarios/client/47_mercenary_roster_ui.json` | Headless client proof. |
| Modify | lifecycle docs | Spec/plan/as-built/progress/scenario catalog updates. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `client/scripts/bot_step_catalog.gd` stayed below the threshold.

Decision:
- [x] Keep UI state/rendering in new `mercenary_panel.gd`.
- [x] Split panel glue and bot matching into focused helpers to keep `main.gd` and the bot runner under their allowances.

Verification:
```bash
make maintainability
```

## Task 1 - Panel UI

Files:
- Add: `client/scripts/mercenary_panel.gd`
- Add: `client/tests/test_mercenary_panel.gd`

- [x] Step 1.1: Build a compact draggable `Mercenaries` panel with fixed dimensions, offer text, status, and roster rows.
- [x] Step 1.2: Implement `show_board`, `set_gold`, `apply_hired_event`, `set_companions`, and `get_debug_state`.
- [x] Step 1.3: Add unit coverage for offer price/gold/affordability and hired companion roster state.
```bash
make client-unit
```

## Task 2 - Client Wiring

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Preload/create the panel alongside existing town service panels.
- [x] Step 2.2: Handle `mercenary_board_opened` and `mercenary_hired` events.
- [x] Step 2.3: Refresh panel gold/affordability on inventory/gold state changes.
- [x] Step 2.4: Feed owned companion state to the panel after `_sync_companion_bar`.
- [x] Step 2.5: Expose `mercenary_panel` and `mercenary_panel_visible` in bot debug state.
```bash
make client-unit
```

## Task 3 - Bot Assertions and Scenario

Files:
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Add: `tools/bot/scenarios/client/47_mercenary_roster_ui.json`
- Modify: `docs/progress/scenario-catalog.md`

- [x] Step 3.1: Add `wait_mercenary_panel`, `assert_mercenary_panel`, and visible-state assertion support.
- [x] Step 3.2: Add a client scenario that clicks `town_mercenary_board`, waits for panel/hire, and asserts the companion bar and roster.
- [x] Step 3.3: Keep the v206 protocol hiring board scenario green.
```bash
make bot-client scenario=mercenary_roster_ui
make bot scenario=mercenary_hiring_board
```

## Task 4 - Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v207_spec-mercenary-roster-ui.md`
- Modify: `docs/plans/v207_2026-06-15-mercenary-roster-ui.md`
- Add: `docs/as-built/v207_mercenary-roster-ui.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 4.1: Mark spec/plan complete and add as-built proof.
- [x] Step 4.2: Update progress/lifecycle docs after verification.
```bash
make ci
```

## Final Verification

- [x] `make client-unit`
- [x] `make bot-client scenario=mercenary_roster_ui`
- [x] `make bot scenario=mercenary_hiring_board`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Separate hire confirmation, player mercenary listings, player-set pricing, death/loss rules, multiple active hires, snapshot refresh, equipment/loot/XP/potion behavior, production art, and companion stance commands remain deferred.
