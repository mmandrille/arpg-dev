# v111 As-Built: Market Purchase and Delivery

**Status:** Complete on `main`

## What shipped

- Added `price_gold` to `market_listings`, defaulting existing listings to offer-only price `0`.
- Listing create/list responses now carry optional `price_gold`.
- Added `POST /v0/market/listings/{listing_id}/purchase`.
- A purchase atomically debits buyer account-stash gold, credits seller account-stash gold, delivers
  the listed item to the buyer stash, marks the listing `accepted`, and refunds active item offers.
- Purchase rejects seller self-purchase, unpriced listings, inactive listings, and insufficient
  buyer stash gold.
- The purchase transaction lives in `server/internal/store/market_purchase.go` to avoid growing
  `repos.go`.
- New store and HTTP purchase tests live in focused market purchase test files to preserve the
  maintainability ratchet.

## Proof

- `cd server && go test ./internal/store -run TestMarket -count=1`
- `cd server && go test ./internal/http -run TestMarket -count=1`
- `make maintainability`
- `make test-go`
- `make ci`

## Deferred

Godot market UI, notifications, accepted-trade audit feeds, expiration timers, listing edits,
taxes/fees, price suggestions, offer gold, counteroffers, and market restrictions for upgraded
items remain deferred.
