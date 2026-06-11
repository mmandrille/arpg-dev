# v68 Spec: Market Stash Listing Foundation

## Status

Approved for implementation in the current autoloop using the default split.

## Context

ADR-0011 describes a future player market with offers, expiration, delivery, and atomic ownership
transfer. This slice only establishes the smallest safe foundation: an account can publish one
stash item into an active listing, browse active listings, and cancel its own listing back to stash.

## Goals

- Add persistent active/canceled market listings backed by account-owned stash items.
- Creating a listing removes the item from `account_stash_items` so it has one durable location.
- Listing browsing returns active listings with enough item metadata for API/bot proof.
- Canceling an owned active listing returns the listed item to the seller's account stash.
- Prove account scoping, active listing visibility, and cancel return through store/HTTP tests.

## Non-goals

- No offers, purchases, gold pricing, listing expiration, audit UI, Godot market UI, or delivery
  to another player's stash.
- No listing from inventory/equipment/hotbar; only existing stash rows are listable.
- No item locking model beyond moving the row out of stash while listed.

## Acceptance Criteria

- `POST /v0/market/listings` with a valid `stash_item_id` creates an active listing for the
  authenticated account and removes that item from the account stash.
- `GET /v0/market/listings` returns active listings across accounts.
- `POST /v0/market/listings/{listing_id}/cancel` cancels only the authenticated seller's listing
  and returns the item to that seller's stash.
- Foreign accounts cannot cancel another seller's active listing.
- `make ci` remains green.
