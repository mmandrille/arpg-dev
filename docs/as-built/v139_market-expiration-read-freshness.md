# v139 As-built: Market Expiration Read Freshness

Date: 2026-06-13
Spec: [`docs/specs/v139_spec-market-expiration-read-freshness.md`](../specs/v139_spec-market-expiration-read-freshness.md)
Plan: [`docs/plans/v139_2026-06-13-market-expiration-read-freshness.md`](../plans/v139_2026-06-13-market-expiration-read-freshness.md)

## What shipped

- `GetMarketSummary` now runs the existing market expiration sweep before counting active published
  listings and incoming bids.
- `ListMarketOffersForSeller` now runs the same sweep before returning seller offer rows, so stale
  listings cannot leak active offers after their expiration time.
- Added focused store coverage in `server/internal/store/market_expiration_read_test.go` proving both
  read paths restore the seller item, refund bidder items, clear active rows/counts, and preserve the
  `listing_expired` audit trail.

## Proof

```bash
cd server && go test ./internal/store -run 'TestMarket(ReadSummary|OfferRead)' -count=1
make maintainability
make ci
```

No bot scenario was added because the slice changes store-side read freshness only; HTTP/client
contracts and response shapes remain unchanged.

## Deferred

- Extracting the market repository block from `server/internal/store/repos.go`.
- New player-facing market expiration UI or audit browser.
