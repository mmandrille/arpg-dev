# v65 Plan - Stash Search And Sorting

Status: Complete
Goal: Add display-only search and sorting to the existing Godot account stash panel.
Architecture: Keep the server stash contract unchanged. The client derives a filtered/sorted view
from `stash_items`, renders only that view, and continues to send withdraw intents by
`stash_item_id`.
Tech stack: Godot `StashPanel`, client bot runner/controller, client unit tests, client bot scenario.

## Baseline and shortcut decision

Baseline is v64 `mystery-seller-paid-reroll` on `main`.

Godot plugin shortcut decision: **borrow pattern only / reject adoption**. The adoption checklist in
`docs/researchs/godot-plugins-and-shortcuts.md` was reviewed. Inventory UI plugins such as GLoot or
Godot-Inventory remain useful references for filtering controls, but this slice only adds a small
display adapter to the in-repo server-backed stash panel. Adopting a plugin would add unnecessary
state authority risk.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/specs/v65_spec-stash-search-and-sorting.md` | Slice spec |
| Add | `docs/plans/v65_2026-06-11-stash-search-and-sorting.md` | This plan |
| Modify | `client/scripts/stash_panel.gd` | Search/sort controls, filtered rendering, debug state, bot hooks |
| Modify | `client/scripts/main.gd` | Bot wrappers for stash search/sort |
| Modify | `client/scripts/bot_controller.gd` | Bot actions for stash search/sort |
| Modify | `client/scripts/bot_scenario_runner.gd` | Assertions/actions for filtered stash rows |
| Modify | `client/tests/test_stash_panel.gd` | Unit coverage for search/sort/filtered withdraw |
| Modify | `client/tests/test_client_bot.gd` | Scenario validation/assertion coverage |
| Add | `tools/bot/scenarios/client/30_stash_search_and_sorting.json` | Godot client proof |
| Modify | `PROGRESS.md` | Lifecycle update |
| Add | `docs/as-built/v65_stash-search-and-sorting.md` | As-built summary |

## Task 1 - Stash Panel View Model

- [x] Step 1.1: Add search text and sort mode state plus a filtered/sorted stash row helper.
```bash
godot --headless --path client --script res://tests/test_stash_panel.gd
```

- [x] Step 1.2: Render filtered rows while preserving capacity/empty slots and withdraw-by-id.
```bash
godot --headless --path client --script res://tests/test_stash_panel.gd
```

## Task 2 - UI Controls And Bot Hooks

- [x] Step 2.1: Add a search field and sort option control to the stash panel.
```bash
make client-unit
```

- [x] Step 2.2: Add bot-callable methods and client bot actions/assertions for search/sort.
```bash
make client-unit
```

## Task 3 - Scenario Proof

- [x] Step 3.1: Add a client bot scenario that deposits several items, filters to one item class,
  changes sort mode, and withdraws a filtered item.
```bash
make bot-client scenario=30_stash_search_and_sorting
```

## Task 4 - Lifecycle And CI

- [x] Step 4.1: Update plan checkboxes, `PROGRESS.md`, and as-built docs.
```bash
rg -n "v65|stash-search-and-sorting|Latest completed slice|stash search" PROGRESS.md docs/as-built docs/plans/v65_2026-06-11-stash-search-and-sorting.md
```

- [x] Step 4.2: Run final verification.
```bash
make client-unit
make bot-client scenario=30_stash_search_and_sorting
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=30_stash_search_and_sorting`
- [x] `make ci`

## Deferred scope

Server-side sorting/filtering, stash tabs, advanced item queries, stack grouping, keyboard shortcut
polish, and direct item actions from stash remain deferred.
