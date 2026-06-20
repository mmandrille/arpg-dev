# v313 As-Built - Vendor And Unique Chest Stability

Status: Complete
Date: 2026-06-20

## Summary

- Reused `InventoryRenderGuard` in `ShopPanel` so identical town vendor and mystery seller refreshes keep existing offer slot controls instead of rebuilding hover targets.
- Extended the render guard state key with shop identity, title, offers, and sell appraisals.
- Added a focused shop tooltip stability test proving stable vendor and mystery seller slot identity across identical refreshes.
- Added focused coverage proving vendor and mystery seller tooltips use the shared item tooltip panel and recursively ignore mouse input.
- Updated the main inventory refresh path so account-stash data is only pushed into the stash panel while it is actually showing account stash, preserving open unique chest and corpse contents.
- Expanded stash panel coverage to prove a fake account stash with `3` uniques and `4` sets cannot overwrite an open unique chest with more rows.
- Added server-side unique test chest backfill so stale persisted chest state receives missing current catalog entries.
- Added a server regression test for persisted unique chest catalog gaps.

## Verification

```bash
cd server && go test ./internal/game -run 'TestUniqueTestChest(OpensContainerAndTakesSelectedItem|BackfillsPersistedCatalogGaps|RepeatActivationReopensRemainingItems)' -count=1
godot --headless --path client --script res://tests/test_shop_tooltip_stability.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
godot --headless --path client --script res://tests/test_stash_panel.gd
cd server && go test ./internal/game -run UniqueTestChest -count=1
make client-unit
make maintainability
```

All checks passed on 2026-06-20.
