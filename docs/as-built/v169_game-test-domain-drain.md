# v169 As-built — Game Test Domain Drain

Date: 2026-06-14

## What shipped

- Added `server/internal/game/gold_auto_pickup_test.go` for gold auto-pickup tests.
- Removed the moved tests and gold-only helper from `game_test.go`.
- Left shared helpers used by other test domains in `game_test.go`.

## Proof

- Focused Go test passed:
  `cd server && go test ./internal/game -run 'Test(GoldAutoPickup|NonGoldLootDoesNotAutoPickup|ManualGoldPickupStillWorksInRange)'`.
- `game_test.go` dropped from 9129 to 8905 lines, and its maintainability baseline was lowered from
  9116 to 8905.

## Deferred

- Additional `game_test.go` domains, such as teleporter/world transitions, boss floors, hotbar, and
  co-op visibility/scaling, remain candidates for later focused test files.
