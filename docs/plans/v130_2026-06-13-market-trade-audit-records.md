# v130 Plan - Market Trade Audit Records

Status: Complete
Goal: Add durable audit rows for market state and ownership transitions.
Architecture: Keep audit inserts inside the same store transactions as the market transition they
describe.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/migrations/0024_market_expiration_and_audit.sql` | Add `market_audit_records`. |
| Modify | `server/internal/store/models.go` | Add audit model. |
| Modify | `server/internal/store/repos.go` | Insert and read audit rows. |
| Modify | `server/internal/store/market_purchase.go` | Audit purchase transitions. |
| Modify | `server/internal/store/store_test.go` | Audit and accept-offer regression proof. |
| Modify | `server/internal/store/test_cleanup_test.go` | Clean audit rows after tests. |
| Modify | `server/internal/http/test_cleanup_test.go` | Clean audit rows after tests. |
| Add | `docs/as-built/v130_market-trade-audit-records.md` | Closeout notes. |

## Tasks

- [x] Create audit table and indexes.
- [x] Insert audit rows inside publish/cancel/offer/accept/reject/purchase/expire transactions.
- [x] Add listing-scoped audit reader for tests/tools.
- [x] Strengthen accept-offer regression for item ownership transfer.
- [x] Extend market test cleanup for audit rows.

## Proof

- [x] `cd server && go test ./internal/store ./internal/http ./internal/replay`

