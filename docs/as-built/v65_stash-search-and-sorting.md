# v65 As-Built - Stash Search And Sorting

Status: Complete
Date: 2026-06-11
Branch: `main`
Spec: [`../specs/v65_spec-stash-search-and-sorting.md`](../specs/v65_spec-stash-search-and-sorting.md)
Plan: [`../plans/v65_2026-06-11-stash-search-and-sorting.md`](../plans/v65_2026-06-11-stash-search-and-sorting.md)

## What Shipped

- Added a local search field to the Godot account stash panel.
- Added stash sort modes for acquired/default order, name, rarity, and slot.
- Kept stash filtering/sorting display-only; withdraw intents still use server-authored
  `stash_item_id`.
- Extended stash debug state with search text, sort mode, filtered item count, and filtered rows.
- Added client bot actions `set_stash_search` and `select_stash_sort`, plus `assert_stash_filter`.
- Added client scenario `30_stash_search_and_sorting.json` proving deposit, search, sort, filtered
  withdraw, and restored unfiltered row count.

## Verification

- `godot --headless --path client --script res://tests/test_stash_panel.gd`
- `godot --headless --path client --script res://tests/test_client_bot.gd`
- `make bot-client scenario=30_stash_search_and_sorting`
- `make client-unit`
- `make ci`

Full `make ci` passed with 9 phases.

## Notes

- No shared protocol, Go server, persistence, or replay contract changed in this slice.
- External inventory plugin adoption was rejected after the required shortcut checklist because the
  existing server-backed `StashPanel` only needed a small local view adapter.
