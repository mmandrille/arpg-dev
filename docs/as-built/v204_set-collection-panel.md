# v204 As-built: Set Collection Panel

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `SetCollectionPanel`, a focused Godot UI component that loads enabled set packages from
  `shared/rules/set_items.v0.json`.
- The panel summarizes all enabled sets with owned/equipped counts, per-piece states
  (`missing`, `owned`, `equipped`), and active/inactive set bonus rows.
- Integrated a compact `Sets` button into the inventory window and exposed set collection debug
  state through `InventoryPanel.get_debug_state()`.
- Kept the slice presentation-only: no server, store, protocol, or shared-rule contract changes.
- Reused existing unique chest tabs in bot mode so the client proof can take a set piece from the
  debug chest and assert inventory collection progress.

## Verification

- `make maintainability`
- `make client-unit`
- `SCENARIO=set_collection_panel HEADLESS=1 ./scripts/bot_client_local.sh`
- `make ci`

## Deferred

- Structured server set metadata, account-wide collection persistence, set rewards, collection
  achievements, new set art, and collection window layout persistence remain future work.
