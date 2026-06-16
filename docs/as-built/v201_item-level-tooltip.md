# v201 As-built: Item Level Tooltip

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Updated the shared item tooltip footer to prefer explicit authoritative `item_level` metadata as
  `Item level N`.
- Preserved the existing requirement-level footer fallback for items without explicit item-level
  metadata.
- Kept character level requirements visible in the tooltip requirements block when an item also has
  an item-level footer.
- Added focused client unit assertions for vendor and inventory tooltip paths.

## Verification

- `make client-unit`
- `make maintainability`
- `make ci`

## Deferred

- Item-level-driven stat scaling, upgrade odds, and loot-band rebalance remain future itemization
  slices.
