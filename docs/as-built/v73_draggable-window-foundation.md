# v73 As-built — Draggable Window Foundation

Date: 2026-06-11

## What Shipped

- Added `DraggableWindow`, a reusable Godot client chrome helper with darker titlebar, right-aligned `X`, titlebar-only drag, viewport clamping, and debug state.
- Migrated the character stats and skills panels onto the helper while preserving their existing toggle/show/hide APIs and default positions.
- Added panel bot helpers for close and drag proof.
- Added focused GDScript unit coverage for titlebar metadata, close behavior, drag movement, and clamp behavior.

## Proof

- `make client-unit` passed.

## Deferred

- Inventory, shop, stash, waypoint, character-select, settings, and multiplayer panel migration.
- Persisting custom panel positions across sessions/restarts.
