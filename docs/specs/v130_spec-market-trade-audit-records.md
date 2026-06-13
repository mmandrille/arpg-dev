# v130 Spec - Market Trade Audit Records

Status: Complete
Date: 2026-06-13
Codename: `market-trade-audit-records`

## Purpose

Persist audit records for market ownership transitions so listing, offer, purchase, cancellation,
expiration, and acceptance flows are inspectable in tests and future tools.

## Non-goals

- No player-facing audit UI.
- No admin query endpoint.
- No full ledger reconstruction beyond market action rows.

## Acceptance Criteria

1. Market publish, offer submit, offer accept, offer reject/refund, offer cancel, purchase, listing
   cancel, and listing expiration write audit rows.
2. Audit rows identify listing, offer where relevant, actor, seller/bidder where known, and item ids
   where relevant.
3. Tests can read audit rows by listing id.
4. The accept-offer regression proves the bidder loses offered items and receives the listed item.

