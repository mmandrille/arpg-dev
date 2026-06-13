# v111 Plan — Market Purchase and Delivery

Status: Ready for implementation
Goal: Add direct stash-gold purchase and stash delivery for priced market listings.
Architecture: Keep listing purchase as an authenticated HTTP/store feature. Persist `price_gold` on
market listings, keep `0` as offer-only, and perform purchase in one store transaction that locks the
listing, stash gold rows, and active offers. No Godot or realtime protocol changes.
Tech stack: Postgres migration, Go store, Go HTTP, lifecycle docs.

## Baseline and shortcut decision

Builds on v68 market listing foundation, v93 multi-item offers, and v110 repeat item upgrades. No
client UI/art work is in scope, so the Godot plugin adoption checklist is not required.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/migrations/0022_market_listing_price_gold.sql` | Add non-negative listing price. |
| Modify | `server/internal/store/models.go` | Carry `PriceGold` on market listing model. |
| Modify | `server/internal/store/interfaces.go` | Add purchase method to repository interface. |
| Modify | `server/internal/store/repos.go` | Include `price_gold` in listing create/list/cancel/scan paths. |
| Create | `server/internal/store/market_purchase.go` | Implement atomic purchase transaction. |
| Modify | `server/internal/store/store_test.go` | Prove purchase/delivery/refund/rejection behavior. |
| Modify | `server/internal/http/market.go` | Accept/list price and expose purchase route. |
| Modify | `server/internal/http/auth_session_test.go` | Prove purchase route behavior. |
| Modify | `server/internal/replay/replay_test.go` | Keep fake repo interface complete. |
| Modify | `PROGRESS.md` | Mark v111 complete during finish. |
| Create | `docs/as-built/v111_market-purchase-and-delivery.md` | Record shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/store/repos.go`
- [x] `server/internal/store/store_test.go`
- [x] `server/internal/http/auth_session_test.go`

Decision:
- [x] Extract focused helper/module/test file as part of this slice: purchase transaction lives in
  `server/internal/store/market_purchase.go`.
- [x] Defer larger test-file extraction with rationale: existing market tests already live in the
  large store/http test files; keep additions compact and revisit if ratchet fails.

Verification:
```bash
make maintainability
```

## Task 1 — Listing price persistence

Files:
- Create: `server/migrations/0022_market_listing_price_gold.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/repos.go`

- [x] Step 1.1: Add `price_gold INTEGER NOT NULL DEFAULT 0` with a non-negative check.
- [x] Step 1.2: Add `PriceGold` to market listing model and scan/create/list/cancel paths.
```bash
cd server && go test ./internal/store -run TestMarketListing -count=1
```

## Task 2 — Atomic purchase transaction

Files:
- Modify: `server/internal/store/interfaces.go`
- Create: `server/internal/store/market_purchase.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 2.1: Implement `PurchaseMarketListing(ctx, buyerAccountID, listingID)` in one
  transaction.
- [x] Step 2.2: Reject seller self-purchase, unpriced listings, missing/inactive listings, and
  insufficient buyer stash gold.
- [x] Step 2.3: Debit buyer stash gold, credit seller stash gold, deliver item to buyer stash, mark
  listing accepted, and refund active offers.
- [x] Step 2.4: Add store tests for success, offer refund, self-purchase, and insufficient gold.
```bash
cd server && go test ./internal/store -run TestMarket -count=1
```

## Task 3 — HTTP purchase route

Files:
- Modify: `server/internal/http/market.go`
- Modify: `server/internal/http/auth_session_test.go`
- Modify: `server/internal/replay/replay_test.go`

- [x] Step 3.1: Add optional `price_gold` to listing create request and listing response.
- [x] Step 3.2: Add `POST /v0/market/listings/{listing_id}/purchase`.
- [x] Step 3.3: Update route tests for purchase success and rejection paths.
```bash
cd server && go test ./internal/http -run TestMarket -count=1
```

## Task 4 — Lifecycle docs and CI

Files:
- Modify: `docs/plans/v111_2026-06-13-market-purchase-and-delivery.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v111_market-purchase-and-delivery.md`

- [x] Step 4.1: Mark plan tasks complete as they pass.
- [x] Step 4.2: Update `PROGRESS.md` latest slice, next slice, lifecycle row, and recently closed note.
- [x] Step 4.3: Add the v111 as-built note.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `cd server && go test ./internal/store -run TestMarket -count=1`
- [x] `cd server && go test ./internal/http -run TestMarket -count=1`
- [x] `make test-go`
- [x] `make ci`
