# v139 Plan - Market Expiration Read Freshness

Status: Complete
Goal: Ensure market read entrypoints sweep expired listings before returning summary or offer data.
Architecture: Reuse the existing `ExpireMarketListings` transaction and call it at the top of the
remaining read surfaces that expose active market state. Keep the change in the store layer so HTTP
routes, bots, and client panels keep their existing response shapes.
Tech stack: Go store integration tests, lifecycle docs.

## Baseline and shortcut decision

Builds on v138 `codemap-and-reduction-ratchet` and the v130 review finding that market summary/detail
reads should not depend on the active-list route to run expiration side effects. No client UI,
required.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/store/repos.go` | Run expiration before market summary and seller-offer reads |
| Create | `server/internal/store/market_expiration_read_test.go` | Add focused regression tests for read-triggered expiration side effects |
| Create | `docs/as-built/v139_market-expiration-read-freshness.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle closeout |
| Modify | `docs/specs/v139_spec-market-expiration-read-freshness.md` | Status closeout |
| Modify | `docs/plans/v139_2026-06-13-market-expiration-read-freshness.md` | Checkbox closeout |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/store/repos.go`
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)? No;
  `repos.go` grows only by the two read freshness calls, and the full market repository extraction
  remains deferred to keep this correctness fix narrow.

Decision:
- [x] Defer extraction with rationale: v138 explicitly scoped `repos.go` market extraction as a
  downstream slice. This change is a small consistency call plus focused tests; moving the full
  market repository block now would expand the blast radius.

Verification:
```bash
make maintainability
```

## Task 1 - Store freshness calls

Files:
- Modify: `server/internal/store/repos.go`

- [x] Step 1.1: Run `ExpireMarketListings` before `GetMarketSummary` counts active listings/bids.
- [x] Step 1.2: Run `ExpireMarketListings` before `ListMarketOffersForSeller` resolves and lists offers.

```bash
cd server && go test ./internal/store -run 'TestMarket(ReadSummary|OfferRead)' -count=1
```

## Task 2 - Store regression tests

Files:
- Create: `server/internal/store/market_expiration_read_test.go`

- [x] Step 2.1: Add a summary-read test that forces a listing into the past, calls
  `GetMarketSummary`, and proves seller/bidder items are restored plus counts are zero.
- [x] Step 2.2: Add an offer-read test that forces a listing into the past, calls
  `ListMarketOffersForSeller`, and proves active offers are no longer returned plus audit records
  include `listing_expired`.

```bash
cd server && go test ./internal/store -run 'TestMarket(ReadSummary|OfferRead)' -count=1
```

## Task 3 - Lifecycle docs and CI

Files:
- Create: `docs/as-built/v139_market-expiration-read-freshness.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Add v139 lifecycle references and as-built note.
- [x] Step 3.2: Mark spec and plan complete.

```bash
make ci
```

## Final verification

- [x] `cd server && go test ./internal/store -run 'TestMarket(ReadSummary|OfferRead)' -count=1`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Extracting the market repository block from `repos.go`.
- New market audit browser or player-facing expiration UI.
