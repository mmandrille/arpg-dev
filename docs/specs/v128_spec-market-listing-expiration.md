# v128 Spec - Market Listing Expiration

Status: Complete
Date: 2026-06-13
Codename: `market-listing-expiration`

## Purpose

Make market listings expire after their full 24-hour period. When an active listing expires, the
seller receives the listed item back in account stash, active offers are rejected, and bidders get
their reserved offer items back.

## Non-goals

- No background scheduler beyond opportunistic expiration on market listing reads and an explicit
  store method.
- No client notification inbox.
- No independent offer expiration.

## Acceptance Criteria

1. Active listings have an `expires_at` timestamp 24 hours after publication.
2. Expired listings no longer appear on the active market board.
3. Expiration atomically restores the listed item to the seller stash.
4. Expiration atomically refunds active offer items to bidders.
5. Store tests cover the expiration/refund flow without real-time sleeps.

