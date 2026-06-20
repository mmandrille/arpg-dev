# v313 Spec - Vendor And Unique Chest Stability

Status: Complete
Date: 2026-06-20
Codename: vendor-and-unique-chest-stability

## Purpose

Fix remaining item-tooltip hover instability in shop-style panels and ensure old persisted unique test chest state receives newly available unique/set catalog items.

## Non-goals

- No vendor, mystery seller, or unique chest visual redesign.
- No new external UI plugins or assets; adopt/reject decision: reject external UI plugins and reuse the v312 render guard and tooltip mouse guard.
- No item balance or catalog tuning.

## Acceptance Criteria

- Town vendor and mystery seller panels do not rebuild unchanged offer slots on identical refreshes.
- Vendor and mystery seller custom item tooltips recursively ignore mouse input.
- Unique test chest state created before catalog growth is backfilled with missing current catalog entries.
- Generic inventory/account-stash refreshes do not overwrite an open unique chest's container item list.
- Existing unique chest repeat-open behavior still shows the same current state and does not add items to inventory.

## Testing Plan

```bash
cd server && go test ./internal/game -run 'TestUniqueTestChest(OpensContainerAndTakesSelectedItem|BackfillsPersistedCatalogGaps|RepeatActivationReopensRemainingItems)' -count=1
godot --headless --path client --script res://tests/test_shop_tooltip_stability.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
godot --headless --path client --script res://tests/test_stash_panel.gd
make client-unit
make maintainability
```
