# v139 Spec: Market Expiration Read Freshness

Status: Complete
Date: 2026-06-13
Codename: `market-expiration-read-freshness`

## Purpose

Guarantee market expiration side effects have run before market read surfaces return. v128-v130
implemented expiration refunds, seller item restoration, and audit rows, but only
`ListActiveMarketListings` triggered `ExpireMarketListings`. Summary and offer-read paths could
therefore report active stale listings/offers until someone opened the active listing board.

This slice keeps the existing expiration implementation and makes read entrypoints freshness-safe.

## Non-goals

- No market UI changes.
- No new audit browser or API.
- No schema, protocol, replay, gameplay, or pricing changes.
- No market repository file split; v138 keeps that as downstream roadmap scope.

## Acceptance Criteria

- `GetMarketSummary` runs the expiration sweep before counting active published listings and incoming
  bids.
- `ListMarketOffersForSeller` runs the expiration sweep before returning offer rows, so expired
  listings no longer leak active offers and reserved bid items are restored before the read returns.
- Existing active listing reads continue to run the sweep.
- Focused store tests prove summary and seller-offer reads trigger expiration side effects: seller
  item restoration, bidder item refund, active count/bid count cleanup, and `listing_expired` audit.
- `make maintainability` and `make ci` pass.

## Scope and Likely Files

- `server/internal/store/repos.go`
- `server/internal/store/market_expiration_read_test.go`
- `docs/plans/v139_2026-06-13-market-expiration-read-freshness.md`
- `docs/as-built/v139_market-expiration-read-freshness.md`
- `PROGRESS.md`

## Test Proof

- `cd server && go test ./internal/store -run 'TestMarket(ReadSummary|OfferRead)' -count=1`
- `make maintainability`
- `make ci`

No bot scenario is required: this is a store read-side consistency fix with no player-facing route
shape change. Existing HTTP/client market scenarios remain covered by `make ci`.

## Risks

- Expiring before summary/offers adds a write transaction to those reads. The existing list-active
  path already does this; this slice applies the same consistency contract to the other read
  surfaces.
