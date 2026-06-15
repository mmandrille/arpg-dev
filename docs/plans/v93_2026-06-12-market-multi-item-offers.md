# v93 Plan — Market Multi-Item Offers

Status: Ready for implementation
Goal: Let bidders submit 1-10 stash items as offers on a market listing and let the seller accept one or cancel the listing.
Architecture: The store remains the ownership authority. Offer creation, acceptance, and cancellation run inside transactions that lock the listing, offer rows, and relevant stash items before moving rows between market tables and account stash. HTTP routes expose the store contracts without adding Godot UI or gameplay protocol.
Tech stack: Go store, SQL migrations, Go HTTP, lifecycle docs.

## Baseline and shortcut decision

Builds on v68 `market-stash-listing-foundation`, where active listings already move one seller stash
UI, camera, inventory presentation, or art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/migrations/0017_market_offers.sql` | Offer and offered-item tables |
| Modify | `server/internal/store/models.go` | Offer models/status constants |
| Modify | `server/internal/store/interfaces.go` | Repository offer methods |
| Modify | `server/internal/store/repos.go` | Transactional offer/list/accept/cancel behavior |
| Modify | `server/internal/store/store_test.go` | Ownership and rejection tests |
| Modify | `server/internal/http/market.go` | Offer routes and JSON payloads |
| Modify | `server/internal/http/auth_session_test.go` | Authenticated offer route proof |
| Modify | `docs/plans/v93_2026-06-12-market-multi-item-offers.md` | Execution checklist |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Create | `docs/as-built/v93_market-multi-item-offers.md` | As-built summary |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/store/repos.go`
- [x] `server/internal/store/store_test.go`
- [x] `server/internal/http/auth_session_test.go`

Decision:
- [x] Defer extraction with rationale: this slice extends existing market/store tests and repository methods in place so the transactional ownership proof stays close to v68 code; no new source/test file is introduced over 600 lines.
- [x] Documented maintenance exception: `make maintainability` also surfaced pre-existing baseline drift
  in untouched client/game files. The baseline was updated for the current repo state, including the
  v93 market growth, and a follow-up structural split remains preferred before further market/test
  expansion.

Verification:
```bash
make maintainability
```

## Task 1 — Persistence Contract

Files:
- Create: `server/migrations/0017_market_offers.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`

- [x] Step 1.1: Add `market_offers` with `active`, `accepted`, and `rejected` statuses.
- [x] Step 1.2: Add `market_offer_items` snapshot rows preserving bidder account, stash item id, source character, item def, and rolled stats.
- [x] Step 1.3: Add store models and repository method signatures for create/list/accept offer flows.

```bash
cd server && go test ./internal/store -run TestMarket
```

## Task 2 — Store Transactions

Files:
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 2.1: Implement offer creation for 1-10 bidder stash items, rejecting sellers and invalid counts.
- [x] Step 2.2: Implement seller-only offer listing.
- [x] Step 2.3: Implement seller acceptance that delivers the listed item to the bidder, delivers accepted offer items to the seller, rejects/refunds competing offers, and marks the listing accepted.
- [x] Step 2.4: Extend listing cancellation so active offers are rejected/refunded before the listed item returns to the seller.
- [x] Step 2.5: Add store tests for create/list/accept, self-offer rejection, foreign offer-list rejection, and cancel refund.

```bash
cd server && go test ./internal/store -run 'TestMarket'
```

## Task 3 — HTTP Routes

Files:
- Modify: `server/internal/http/market.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 3.1: Add `POST /v0/market/listings/{listing_id}/offers`.
- [x] Step 3.2: Add `GET /v0/market/listings/{listing_id}/offers`.
- [x] Step 3.3: Add `POST /v0/market/listings/{listing_id}/offers/{offer_id}/accept`.
- [x] Step 3.4: Add route tests for submit/list/accept plus rejection cases.

```bash
cd server && go test ./internal/http -run 'TestMarket'
```

## Task 4 — Lifecycle Docs And CI

Files:
- Modify: `docs/plans/v93_2026-06-12-market-multi-item-offers.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v93_market-multi-item-offers.md`

- [x] Step 4.1: Mark plan tasks complete.
- [x] Step 4.2: Add the v93 lifecycle row and current status updates.
- [x] Step 4.3: Add the v93 as-built summary and deferred scope.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `cd server && go test ./internal/store -run 'TestMarket'`
- [x] `cd server && go test ./internal/http -run 'TestMarket'`
- [x] `make test-go`
- [x] `make ci`

## Deferred scope

- Direct gold purchase, offer gold, expiration timers, accepted-trade audit feeds, Godot market UI,
  counteroffers, offer edits, and notifications remain deferred.
