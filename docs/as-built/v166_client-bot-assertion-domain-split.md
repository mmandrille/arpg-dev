# v166 As-built — Client Bot Assertion Domain Split

Date: 2026-06-14

## What shipped

- Added `client/scripts/bot_ui_assertion_handlers.gd` for UI/menu-oriented client bot assertions.
- Kept `BotAssertionHandlers.evaluate` as the public dispatcher while delegating handled UI/menu
  assertion types through a handled/result pair.
- Preserved existing client bot assertion step names and scenario DSL behavior.

## Proof

- `make client-unit` passed.
- `client/scripts/bot_assertion_handlers.gd` dropped from 299 to 267 lines; the new helper is 48
  lines and remains below the 600-line target.

## Deferred

- Combat, inventory, shop/stash/market, and world/presentation assertion domains remain in the main
  dispatcher for future focused splits.
