# v111 Spec: Market Purchase and Delivery

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `market-purchase-and-delivery`

## Purpose

Add direct gold purchase for active player-market listings. Sellers can create a listing with an
optional stash-gold price, and another account can buy a priced active listing. The server atomically
delivers the listed item to the buyer's account stash, transfers stash gold to the seller's account
stash, marks the listing accepted, and refunds any active item offers.

## Non-goals

- No Godot market UI, notifications, accepted-trade audit feed, expiration timers, listing edits,
  taxes/fees, price suggestions, offer gold, or counteroffers.
- No inventory/equipped/hotbar listings; only account-stash listings remain supported.
- No market restrictions for upgraded items; v110 upgraded items remain listable/purchasable.
- No protocol/realtime schema changes.

## Acceptance criteria

- `market_listings` persists non-negative `price_gold`, defaulting to `0` for existing offer-only
  listings.
- `POST /v0/market/listings` accepts optional `price_gold`; list responses include `price_gold`.
- `POST /v0/market/listings/{listing_id}/purchase` buys a priced active listing for a non-seller
  account with enough stash gold.
- Purchase atomically:
  - debits buyer account-stash gold by `price_gold`,
  - credits seller account-stash gold by `price_gold`,
  - delivers the listed item to buyer account stash,
  - marks the listing `accepted`,
  - refunds active item offers on the listing.
- Purchase rejects missing listings, seller self-purchase, unpriced listings, insufficient buyer
  stash gold, and already accepted/canceled listings.
- Store and HTTP tests prove success, gold movement, item delivery, offer refund, self-purchase
  rejection, and insufficient-gold rejection.

## Scope and likely files

- Migration: `server/migrations/0022_market_listing_price_gold.sql`
- Store: `server/internal/store/models.go`, `server/internal/store/interfaces.go`,
  `server/internal/store/repos.go`, `server/internal/store/market_purchase.go`,
  `server/internal/store/store_test.go`
- HTTP: `server/internal/http/market.go`, `server/internal/http/auth_session_test.go`
- Lifecycle docs: `PROGRESS.md`, `docs/as-built/v111_market-purchase-and-delivery.md`

## Test and bot proof

- `cd server && go test ./internal/store -run TestMarket -count=1`
- `cd server && go test ./internal/http -run TestMarket -count=1`
- `make test-go`
- `make maintainability`
- `make ci`

No protocol bot is required because this slice changes authenticated HTTP market ownership transfer,
not realtime gameplay, protocol intents, world generation, combat, or client presentation.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | How are prices set? | Sellers provide optional integer `price_gold`; `0` preserves offer-only listings. |
| Q-2 | What does purchase do to active offers? | Refund them using the existing cancellation/accepted-offer refund path. |
| R-1 | Ownership transfer must be atomic. | Implement in one store transaction with row locks for listing and stash-gold rows. |
| R-2 | Store monolith growth. | Put purchase transaction code in a focused store file rather than growing `repos.go`. |

## ADR alignment

- ADR-0011: advances player-market ownership transfer and stash delivery.
- ADR-0014 D3/D12: makes stash gold and player-market listings part of the long-term economy loop.
