# v129 Plan - Market Offer Withdrawal

Status: Complete
Goal: Add active-offer cancellation and bidder item recovery.
Architecture: Implement cancellation as a store transaction and expose a narrow authenticated HTTP
route.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/migrations/0024_market_expiration_and_audit.sql` | Add `canceled` offer status and timestamp. |
| Modify | `server/internal/store/models.go` | Carry canceled offer state. |
| Modify | `server/internal/store/repos.go` | Implement `CancelMarketOffer`. |
| Modify | `server/internal/http/market.go` | Add cancel-offer route. |
| Modify | `client/scripts/net_client.gd` | Add client API wrapper. |
| Modify | `server/internal/store/store_test.go` | Store cancellation proof. |
| Modify | `server/internal/http/market_purchase_test.go` | HTTP cancellation proof. |
| Add | `docs/as-built/v129_market-offer-withdrawal.md` | Closeout notes. |

## Tasks

- [x] Add offer cancellation status/timestamp.
- [x] Restore bidder offer items on cancellation.
- [x] Reject foreign cancellation.
- [x] Expose `POST /v0/market/listings/{listing_id}/offers/{offer_id}/cancel`.
- [x] Add focused store and HTTP tests.

## Proof

- [x] `cd server && go test ./internal/store ./internal/http ./internal/replay`

