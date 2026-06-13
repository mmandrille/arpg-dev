# v114 As-Built: Market Board UI

**Status:** Complete on `main`

## What shipped

- Market listing creation from the Godot client can now send a chosen `price_gold`.
- The market panel publish tab has a simple gold price control with a deterministic 25 gold default.
- Browse rows display each listing price before the listing id and seller hint.
- Market debug state exposes listing rows, stash rows, and publish price for client bot assertions.
- Added client bot scenario `35_market_board_ui`.
- Updated the grandfathered file-size baseline for `client/scripts/bot_controller.gd` and
  `client/scripts/bot_scenario_runner.gd`; the slice had to register new market bot actions and
  wait/assert validation in the existing bot registries.

## Proof

- `make client-unit`
- `make bot-client scenario=35_market_board_ui`
- `make maintainability`
- `make ci`

## Deferred

Direct purchase UI, delivery feedback, listing cancel/edit controls, search, sorting, pagination,
taxes, expiration, and offer-management polish remain deferred.
