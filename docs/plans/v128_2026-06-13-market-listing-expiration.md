# v128 Plan - Market Listing Expiration

Status: Complete
Goal: Add server-owned market listing expiration and refund behavior.
Architecture: Keep expiration in the store transaction layer. Listing reads opportunistically expire
stale rows before returning active listings.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/migrations/0024_market_expiration_and_audit.sql` | Listing expiration columns/status. |
| Modify | `server/internal/store/models.go` | Carry expiration timestamps/status. |
| Modify | `server/internal/store/repos.go` | Expire listings and refund offers atomically. |
| Modify | `server/internal/store/store_test.go` | Expiration/refund regression. |
| Add | `docs/as-built/v128_market-listing-expiration.md` | Closeout notes. |

## Tasks

- [x] Add `expires_at` and `expired_at` to listings.
- [x] Add `expired` listing status.
- [x] Implement `ExpireMarketListings`.
- [x] Filter/expire stale listings before active board reads.
- [x] Verify seller and bidder stash refunds in store tests.

## Proof

- [x] `cd server && go test ./internal/store ./internal/http ./internal/replay`

