# v242 Plan - Market Board Search And Sort

Status: Complete
Goal: Add client-owned market board search and sort controls with bot proof.
Architecture: Market data continues loading through the existing HTTP endpoints. The Godot panel
keeps the authoritative arrays unchanged and derives filtered/sorted display arrays through helper
classes. A small bot-market-action helper absorbs new filter/sort bot actions without growing
`bot_controller.gd`.
Tech stack: Godot UI/helpers, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v241 market row helper extraction and `expires_at` debug rows. Asset/plugin decision:
reject external UI assets/plugins; borrow existing market row styling and local Godot controls.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/market_panel.gd` | Add search/sort controls and use filtered display lists. |
| Create | `client/scripts/market_filter_controls.gd` | Own market search input and sort selector UI. |
| Create | `client/scripts/market_row_filters.gd` | Pure filter/sort logic for listings, offers, and receipts. |
| Create | `client/scripts/bot_market_actions.gd` | Focused market bot action dispatcher. |
| Modify | `client/scripts/bot_controller.gd` | Delegate market actions to the focused helper. |
| Modify | `client/scripts/bot_action_step_validator.gd` | Validate market search/sort action steps. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register market search/sort action steps and filter expectations. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match market panel filter/sort state. |
| Create | `client/tests/test_market_search_sort.gd` | Unit coverage for filter/sort helpers and panel debug state. |
| Create | `tools/bot/scenarios/client/59_market_search_sort.json` | Client-bot proof for search/sort. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Slice lifecycle close-out. |
| Create | `docs/as-built/v242_market-board-search-and-sort.md` | As-built proof summary. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/market_panel.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline or avoid growth through extraction?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 - Filter helpers and panel controls

Files:
- Modify: `client/scripts/market_panel.gd`
- Create: `client/scripts/market_filter_controls.gd`
- Create: `client/scripts/market_row_filters.gd`
- Create: `client/tests/test_market_search_sort.gd`

- [x] Step 1.1: Add reusable search/sort controls to the market panel.
- [x] Step 1.2: Derive filtered/sorted listing, offer, and receipt arrays without mutating source arrays.
- [x] Step 1.3: Expose filter query, sort mode, options, and filtered counts in debug state.
- [x] Step 1.4: Add focused unit coverage for listing/offer/receipt search and sort behavior.

```bash
godot --headless --path client --script res://tests/test_market_search_sort.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
```

## Task 2 - Client bot actions and proof

Files:
- Create: `client/scripts/bot_market_actions.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_action_step_validator.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/59_market_search_sort.json`

- [x] Step 2.1: Extract market bot action dispatch and add `set_market_search` / `select_market_sort`.
- [x] Step 2.2: Add market panel filter/sort state matching for client bot wait/assert steps.
- [x] Step 2.3: Add a focused client bot scenario that filters a market listing and switches sort mode.

```bash
make bot-client scenario=59_market_search_sort.json HEADLESS=1
```

## Task 3 - Lifecycle docs and close-out

Files:
- Modify: `docs/specs/v242_spec-market-board-search-and-sort.md`
- Modify: `docs/plans/v242_2026-06-17-market-board-search-and-sort.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Create: `docs/as-built/v242_market-board-search-and-sort.md`

- [x] Step 3.1: Mark spec/plan complete and add lifecycle/as-built documentation.
- [x] Step 3.2: Update `PROGRESS.md` current status to v242 with final batch CI pending.

```bash
make maintainability
```

## Final verification

- [x] `godot --headless --path client --script res://tests/test_market_search_sort.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=59_market_search_sort.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` deferred to `$autoloop` after the selected queue completes.
