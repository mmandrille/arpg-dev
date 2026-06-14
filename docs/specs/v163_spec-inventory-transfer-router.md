# v163 Spec — Inventory transfer router

Date: 2026-06-14
Status: Complete
Codename: inventory-transfer-router

## Purpose

Extract inventory panel transfer and staging routing into a focused client helper so future stash,
market, blacksmith, corpse, unique chest, and equipment work does not keep expanding
`inventory_panel.gd`.

This is a behavior-preserving client maintainability slice. Server authority, protocol messages,
and item rules stay unchanged.

## Non-goals

- No new inventory, stash, market, blacksmith, corpse, shop, hotbar, or equipment behavior.
- No protocol/schema/server changes.
- No UI redesign or new visual controls.
- No change to item eligibility, requirement checks, weapon-set semantics, or server-authored
  mutation outcomes.

## Acceptance criteria

- `InventoryPanel` delegates double-click, shift-click, and drag/drop transfer routing to a focused
  helper file.
- Existing emitted intent names and payloads remain unchanged for shop sell/buy, market staging,
  blacksmith staging/unstaging, stash withdraw/equip, corpse withdraw, unique chest take, equip,
  unequip, consumable use, and hotbar assignment.
- The helper is covered by a focused GDScript unit test or expanded existing unit coverage.
- Existing client panel tests, client bot scenarios, and full CI remain green.
- `inventory_panel.gd` line count decreases and maintainability ratchets pass.

## Likely files

- `client/scripts/inventory_transfer_router.gd`: new focused router.
- `client/scripts/inventory_panel.gd`: delegate routing decisions and emit results.
- `client/tests/test_inventory_transfer_router.gd`: direct route coverage.
- `client/tests/test_stash_panel.gd` / `client/tests/test_shop_panel.gd`: existing integration
  coverage should remain green.
- `PROGRESS.md`, `docs/plans/v163_2026-06-14-inventory-transfer-router.md`, and
  `docs/as-built/v163_inventory-transfer-router.md`.

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=23_account_stash_panel.json`
- `make bot-client scenario=35_market_board_ui.json`
- `make bot-client scenario=39_blacksmith_upgrade_ui.json`
- `make maintainability`
- `make ci`

Visual verification command:

```bash
make bot-visual scenario=23_account_stash_panel.json
```
