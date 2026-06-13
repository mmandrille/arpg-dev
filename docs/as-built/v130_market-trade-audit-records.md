# v130 As-Built - Market Trade Audit Records

Date: 2026-06-13
Spec: [`docs/specs/v130_spec-market-trade-audit-records.md`](../specs/v130_spec-market-trade-audit-records.md)
Plan: [`docs/plans/v130_2026-06-13-market-trade-audit-records.md`](../plans/v130_2026-06-13-market-trade-audit-records.md)

## What Shipped

- Added `market_audit_records` for market transitions.
- Audit rows are written in the same transaction as ownership/status changes.
- Store tests now assert accepted offers transfer the listed item to the bidder and keep offered
  items out of the bidder stash.

## Proof

- `cd server && go test ./internal/store ./internal/http ./internal/replay`
