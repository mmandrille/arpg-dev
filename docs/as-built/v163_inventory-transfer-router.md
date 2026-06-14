# v163 As-built — Inventory transfer router

Date: 2026-06-14
Status: Complete

## What shipped

- Added `client/scripts/inventory_transfer_router.gd` as a focused, inert routing helper for
  inventory double-click, shift-click, and drag/drop decisions.
- `InventoryPanel` now delegates transfer routing and only applies returned decisions by emitting
  existing intent signals or unstaging blacksmith items.
- Preserved existing route payloads for shop sell/buy, market stage, blacksmith stage/unstage,
  stash withdraw/equip, corpse withdraw, unique chest take, equip, unequip, consumable use, and
  hotbar assignment.
- Added `client/tests/test_inventory_transfer_router.gd` and wired it into `scripts/client_smoke.sh`.
- Updated CODEMAP so stash, shop/vendor, market, and town-service work points at the router.
- Reduced `client/scripts/inventory_panel.gd` from 1583 to 1534 lines.

## Verification

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=23_account_stash_panel.json`
- `make bot-client scenario=35_market_board_ui.json`
- `make bot-client scenario=39_blacksmith_upgrade_ui.json`
- `make ci`

Visual verification command:

```bash
make bot-visual scenario=23_account_stash_panel.json
```
