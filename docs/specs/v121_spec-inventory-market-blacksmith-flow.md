# v121 Spec: Inventory Market Blacksmith Flow

Status: Approved
Date: 2026-06-13
Codename: `inventory-market-blacksmith-flow`

## Purpose

Make the market board and blacksmith operate from the character inventory instead of requiring the
player to pre-place items in account stash. Opening either town service also opens the inventory.
Players can drag or double-click inventory items into the relevant service action area, then confirm
the market publish/offer or blacksmith upgrade action.

## Non-goals

- No change to market ownership, offer acceptance, purchase, listing cancellation, or upgrade math.
- No new account-stash browsing requirement for market publish, market offers, or blacksmith upgrade.
- No upgrade recipes, success chance, bricking, material costs, or equipped-item upgrades.

## Contract changes

- `POST /v0/market/listings` keeps `stash_item_id` support and also accepts
  `item_instance_id` + `character_id`. When inventory fields are used, the server atomically moves
  the character item into account stash, then creates the listing from that reserved stash item.
- `POST /v0/market/listings/{listing_id}/offers` keeps `stash_item_ids` support and also accepts
  `item_instance_ids` + `character_id`. When inventory fields are used, the server atomically moves
  those character items into account stash, then creates the offer from the reserved stash ids.
- `POST /v0/account-stash/items/{stash_item_id}/upgrade` remains valid. New
  `POST /v0/account-stash/items/upgrade` accepts `item_instance_id` + `character_id`, reserves the
  character item in account stash, upgrades it, and returns the upgraded stash item. Upgrade payment
  spends character inventory gold first and account stash gold only for any remaining cost.

## Acceptance criteria

- Market panel opens with the inventory panel visible.
- Publish tab has a clear drop/stage area and publish price control; double-clicking or dragging an
  inventory item stages it, and pressing Publish creates a listing from the inventory item.
- Offer tab has a 2-row by 5-column staging grid; double-clicking or dragging inventory items fills
  the grid up to 10 items, and pressing Offer submits all staged items.
- Blacksmith panel opens with the inventory panel visible, has a central item block, lists the
  next upgrade cost and stat improvements for the staged item, and upgrades by button press.
- Blacksmith affordability uses character gold plus stash gold; successful payment drains character
  gold before stash gold.
- Existing stash-based market and upgrade tests remain valid.
- Server tests cover inventory-origin listing, offer, and upgrade requests.

## Testing plan

- `cd server && go test ./internal/http`
- `make client-unit`
- `make maintainability`

Manual visual proof command after implementation:

```bash
make bot-visual scenario=35_market_board_ui
make bot-visual scenario=39_blacksmith_upgrade_ui
```