# v121 As-Built - Inventory market blacksmith flow

Date: 2026-06-13
Status: Complete

## What Shipped

- Added inventory-origin market listing support while preserving existing stash-origin listing
  requests.
- Added inventory-origin multi-item market offers, reserving character inventory items into account
  stash before creating the offer.
- Added inventory-origin blacksmith upgrades through `POST /v0/account-stash/items/upgrade`.
- Made blacksmith upgrade payment spend character gold first, then account stash gold for any
  remainder.
- Updated Godot market and blacksmith panels so town services can stage character inventory items
  directly, with inventory opening alongside the service panel.
- Added focused HTTP and Godot unit coverage for inventory-origin listing, offer, and upgrade flows.

## Proof

- `cd server && go test ./internal/http`
- `make client-unit`
- `make maintainability`
- `make bot-client scenario=35_market_board_ui`
- `make bot-client scenario=39_blacksmith_upgrade_ui`

## Scope Limits

- Market ownership, offer acceptance, purchase, listing cancellation, and upgrade math stayed on the
  existing authoritative paths.
- No upgrade recipes, success chance, bricking, material costs, or equipped-item upgrades.
