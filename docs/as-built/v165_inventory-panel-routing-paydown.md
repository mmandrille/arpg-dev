# v165 As-built — Inventory Panel Routing Paydown

Date: 2026-06-14

## What shipped

- Reused `InventoryTransferRouter` as the single source for equipment slot-kind parsing in
  inventory drag/drop and slot rendering paths.
- Removed the duplicate slot-kind helper from `inventory_panel.gd`, keeping the panel focused on UI
  state lookup and signal emission.
- Added router unit coverage for equipment slot recognition and parsing.

## Proof

- Focused router test: `godot --headless --path client --script res://tests/test_inventory_transfer_router.gd`
  passed with 42 assertions.
- `inventory_panel.gd` dropped from 1534 to 1528 lines, and the maintainability baseline was lowered
  from 1559 to 1528.

## Deferred

- Further inventory panel reductions remain available, especially tooltip rendering and paper-doll
  presentation helpers.
