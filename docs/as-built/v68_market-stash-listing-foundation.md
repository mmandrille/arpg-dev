# v68 As-built: Market Stash Listing Foundation

## What shipped

- Added durable `market_listings` rows with `active` and `canceled` status.
- Creating a listing moves one account stash item out of `account_stash_items` into an active
  listing, preserving single-location item ownership.
- Canceling an owned active listing marks it canceled and restores the same stash item id to the
  seller's account stash.
- Added authenticated HTTP routes for listing active market rows, creating a listing from stash,
  and canceling an owned listing.

## Proof

- `(cd server && go test ./internal/store ./internal/http)`
- `make ci`
