# v93 Spec: Market Multi-Item Offers

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-12
Codename: `market-multi-item-offers`

## Purpose

Build on v68 market listings so other accounts can submit item offers for a listed stash item.
The seller can inspect all active offers on their listing, accept exactly one offer, or cancel the
listing. Acceptance atomically moves the listed item to the bidder's account stash and the offered
items to the seller's account stash.

## Non-goals

- No direct gold purchase, offer gold, item pricing, expiration timers, audit UI, or Godot market UI.
- No listing from inventory/equipment/hotbar; listed items still come from account stash.
- No offers larger than 10 items and no empty offers.
- No counteroffers, offer edits, offer expiration, accepted-trade audit feed, or notifications.
- Self-purchase/self-offer is rejected.

## Acceptance criteria

- A non-seller account can submit an offer with 1 to 10 owned stash item ids for an active listing.
- Submitting an offer removes every offered item from the bidder's account stash and records one
  active offer with item snapshots.
- The seller can list active offers for their own listing and sees each offer's offered item rows.
- Foreign accounts cannot view a seller's offer list.
- The seller can accept one active offer; the listing becomes `accepted`, that offer becomes
  `accepted`, competing offers become `rejected`, the listed item appears in the bidder's stash, and
  offered items appear in the seller's stash.
- Canceling a listing with active offers returns the listed item to the seller and returns all
  offered items to their bidders.
- Sellers cannot offer on their own listings.
- Store and HTTP tests prove ownership transfer, foreign rejection, and cancel return.

## Scope and likely files

- `server/migrations/0017_market_offers.sql` - market offer and offer item persistence.
- `server/internal/store/models.go` - offer model structs and statuses.
- `server/internal/store/interfaces.go` - repository methods.
- `server/internal/store/repos.go` - transactional offer create/list/accept/cancel behavior.
- `server/internal/store/store_test.go` - atomic ownership tests.
- `server/internal/http/market.go` - HTTP offer routes and response payloads.
- `server/internal/http/auth_session_test.go` - authenticated route tests.
- `docs/plans/v93_2026-06-12-market-multi-item-offers.md` - implementation plan.
- `docs/as-built/v93_market-multi-item-offers.md` and `PROGRESS.md` - close-out docs.

## Test and bot proof

- `cd server && go test ./internal/store ./internal/http`
- `make test-go`
- `make ci`

No protocol bot or Godot bot is required because this slice is authenticated HTTP/store economy
infrastructure and has no gameplay WebSocket or client UI surface.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Direct gold purchase or item offers? | Owner selected item offers: bidders submit 1-10 items, seller accepts one or cancels. |
| Q-2 | Can sellers offer on their own listings? | No; reject self-offers. |
| R-1 | Offer acceptance can duplicate or lose items if not transactional. | Implement offer create, accept, and cancel inside store transactions with row locks. |
| R-2 | Rejected competing offers must not strand bidder items. | Acceptance rejects competing active offers and restores their items to original bidder stashes. |

## ADR alignment

- ADR-0011: advances player market item-for-item offers and stash delivery while deferring
  expiration, audit records, and richer offer UI.
- ADR-0014 D12: feeds the player-market endgame direction without adding direct gold purchase yet.
