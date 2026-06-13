# v128 As-Built - Market Listing Expiration

Date: 2026-06-13
Spec: [`docs/specs/v128_spec-market-listing-expiration.md`](../specs/v128_spec-market-listing-expiration.md)
Plan: [`docs/plans/v128_2026-06-13-market-listing-expiration.md`](../plans/v128_2026-06-13-market-listing-expiration.md)

## What Shipped

- Listings now receive a server-authored 24-hour `expires_at`.
- Active market listing reads first expire stale rows.
- Expiration restores the listed item to seller stash and rejects/refunds active offers.

## Proof

- `cd server && go test ./internal/store ./internal/http ./internal/replay`

