# v129 As-Built - Market Offer Withdrawal

Date: 2026-06-13
Spec: [`docs/specs/v129_spec-market-offer-withdrawal.md`](../specs/v129_spec-market-offer-withdrawal.md)
Plan: [`docs/plans/v129_2026-06-13-market-offer-withdrawal.md`](../plans/v129_2026-06-13-market-offer-withdrawal.md)

## What Shipped

- Added bidder-owned offer cancellation.
- Canceling an offer restores reserved items to bidder account stash.
- Added HTTP and store regression coverage.

## Proof

- `cd server && go test ./internal/store ./internal/http ./internal/replay`

