# v74 As-built — Gameplay Window Chrome

Date: 2026-06-11

## What Shipped

- Migrated inventory, shop, and stash panels to the reusable `DraggableWindow` chrome.
- Preserved existing item drag/drop, buy/sell, reroll, stash search/sort, and panel visibility APIs.
- Moved shop and stash dynamic titles into the reusable titlebar.
- Added close/drag bot helpers and window debug state for the migrated panels.
- Updated focused client tests and smoke coverage for titlebar metadata, close behavior, drag movement, and clamping.

## Proof

- `make client-unit` passed.

## Deferred

- Waypoint, settings, character select, multiplayer, pause, and main menu window migration.
- Persisting custom positions across sessions/restarts.
