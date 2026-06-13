# v115 As-Built: Market Purchase UI

**Status:** Complete on `main`

## What shipped

- `NetClient` can call `POST /v0/market/listings/{listing_id}/purchase`.
- Market browse rows now show a `Buy` button for non-seller listings with `price_gold > 0`.
- Purchase success/failure routes through the existing market panel status text and refreshes active
  listings.
- Added a `market_listing` client-bot preflight that creates a seller listing through normal
  HTTP/WebSocket flows.
- Added client bot scenario `36_market_purchase_ui`.

## Proof

- `make client-unit`
- `make bot-client scenario=36_market_purchase_ui`
- `make maintainability`
- `make ci`

## Deferred

Realtime delivery notifications, in-session stash reload after purchase, purchase history, listing
cancel/edit controls, search, sorting, pagination, taxes, expiration, and audit feeds remain
deferred.
