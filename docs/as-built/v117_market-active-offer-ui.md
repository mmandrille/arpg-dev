# v117 As-Built: Market Active Offer UI

**Status:** Complete on `main`

## What shipped

- `NetClient` can list and accept offers through the existing market HTTP routes.
- Seller-owned market listing rows now expose `View Offers`.
- The market panel renders active offer rows with bidder hints, status, and offered item names.
- Sellers can accept an active offer from the panel; success refreshes active listings and reports
  `Offer accepted`.
- Client debug state exposes offer rows and the local account id for bot assertions.
- The client bot harness can preflight a seller listing plus a foreign bidder offer.
- Added client bot scenario `38_market_active_offer_ui`.

## Proof

- `make client-unit`
- `make bot-client scenario=38_market_active_offer_ui`
- `make maintainability`
- `make ci`

## Manual Visual

```bash
make bot-visual scenario=38_market_active_offer_ui
```

## Deferred

Offer editing, counteroffers, expiration timers, audit feeds, notifications, pagination, search,
sorting, item comparison, listing cancellation, item locking beyond existing acceptance semantics,
and realtime stash delivery notifications remain deferred.
