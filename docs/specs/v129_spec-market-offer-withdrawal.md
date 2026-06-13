# v129 Spec - Market Offer Withdrawal

Status: Complete
Date: 2026-06-13
Codename: `market-offer-withdrawal`

## Purpose

Let the bidder cancel an active market offer and recover the reserved offer items. Offered items
remain removed from usable inventory/stash while the offer is active.

## Non-goals

- No offer browser redesign.
- No independent offer timers.
- No direct inventory delivery; recovered items return to account stash.

## Acceptance Criteria

1. A bidder can cancel only their own active offer.
2. Canceling an offer changes its status to `canceled`.
3. Canceling an offer restores all reserved items to bidder account stash.
4. Foreign users cannot cancel another bidder's offer.
5. HTTP and store tests cover the cancel/refund flow.

