# v234 Spec - Buyer Offer Cancel UI

Status: Approved for autoloop
Date: 2026-06-17
Codename: buyer-offer-cancel-ui

## Purpose

Let buyers cancel their own outgoing market offers from the `My Offers` view shipped in v233. The
backend already owns bidder-scoped cancel semantics and returns the offered items to account stash;
this slice exposes that action in the Godot market panel and proves the item refund through a
headless client scenario.

## Non-goals

- No seller-side rejection UI, trade receipts, audit history, notifications, or inbox.
- No confirmation modal, bulk cancel, listing search/sort/pagination, or expiration timer UI.
- No market ownership, pricing, offer matching, or item-transfer rule changes.

## Acceptance Criteria

- Outgoing offer rows in `My Offers` show a `Cancel` action for active offers and disable it once the
  offer is no longer active.
- Clicking `Cancel` calls the existing authenticated
  `POST /v0/market/listings/{listing_id}/offers/{offer_id}/cancel` route through `NetClient`.
- On success, the market panel remains in `My Offers`, shows cancellation status, refreshes outgoing
  offers, and the refunded offered item appears in the account stash state.
- A client bot scenario creates a foreign listing, submits an offer as the client bidder, opens
  `My Offers`, cancels the outgoing offer, verifies the canceled/removed offer state, and verifies
  the offered item is available in stash.

## Scope and Likely Files

- Client UI/action routing: `client/scripts/market_offer_rows.gd`, `client/scripts/market_panel.gd`,
  `client/scripts/main.gd`.
- Bot/scenario: `client/scripts/bot_controller.gd`, `client/scripts/bot_step_catalog.gd`,
  `client/scripts/bot_action_step_validator.gd`, `client/scripts/bot_scenario_runner.gd`,
  `tools/bot/scenarios/client/51_buyer_offer_cancel_ui.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `cd server && go test ./internal/store -run MarketOfferCancel -count=1`
- `cd server && go test ./internal/http -run MarketOfferCancel -count=1`
- `make bot-client scenario=51_buyer_offer_cancel_ui.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. Existing backend and transport helpers already cover cancel semantics.
- Risk: client bot/controller and market panel files are on the maintainability boundary. Keep the
  implementation narrow and offset local growth in touched grandfathered files.
