# v75 As-built — Persistent Window Layout

Date: 2026-06-11

## What Shipped

- Extended `DraggableWindow` with optional local layout persistence in `user://window_layout.cfg`.
- Added stable layout keys for character stats, skills, inventory, shop, and stash.
- Saved clamped positions after bot/user drags and restored saved positions when panels are built.
- Disabled normal layout persistence under `CLIENT_UNIT_ONLY=1` so unit tests do not pollute player/developer layout files.
- Added focused helper coverage for save/load behavior using a test override path.

## Proof

- `make client-unit` passed.

## Deferred

- Reset-layout UI.
- Server/account-synced UI layout.
- Draggable chrome migration for waypoint, settings, character select, multiplayer, pause, and main menu windows.
