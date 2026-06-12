# v93 As-built: Market Multi-Item Offers

## What shipped

- Added durable market offers with active, accepted, and rejected states.
- Bidders can submit 1 to 10 account-stash items as an offer on another account's active listing.
- Offer creation reserves bidder items by removing them from `account_stash_items` and snapshotting
  them in `market_offer_items`.
- Sellers can list active offers for their own listing and accept one offer.
- Accepting an offer marks the listing accepted, delivers the listed item to the bidder's stash,
  delivers accepted offer items to the seller's stash, and rejects/refunds competing active offers.
- Canceling a listing now also rejects/refunds active offers before returning the listed item to the
  seller's stash.

## Proof

- `cd server && go test ./internal/store -run TestMarket -count=1`
- `cd server && go test ./internal/http -run TestMarket -count=1`
- `make maintainability`
- `make test-go`
- `make ci`

## Deferred

- Direct gold purchase, offer gold, expiration timers, accepted-trade audit feeds, Godot market UI,
  counteroffers, offer edits, and notifications remain deferred.
