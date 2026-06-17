# v234 As-Built - Buyer Offer Cancel UI

Date: 2026-06-17

## What shipped

- Changed outgoing `My Offers` rows to show an active `Cancel` action while keeping seller incoming
  rows on `Accept`.
- Routed outgoing offer cancellation through the existing `NetClient.cancel_market_offer` helper and
  bidder-scoped backend route.
- On cancel success, restored the returned offered items into local account stash state, refreshed
  inventory UI, and reloaded `My Offers` with an "Offer canceled" status.
- Replaced accept-only market bot offer plumbing with shared accept/cancel offer-action plumbing so
  `bot_controller.gd` stayed inside the maintainability ratchet.
- Added `51_buyer_offer_cancel_ui.json`, which opens `My Offers`, cancels the active outgoing offer,
  verifies the row becomes canceled, opens stash, and verifies the offered item was refunded.

## Proof

```bash
cd server && go test ./internal/store -run MarketOfferCancel -count=1
cd server && go test ./internal/http -run MarketOfferCancel -count=1
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=51_buyer_offer_cancel_ui.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=51_buyer_offer_cancel_ui.json
```

## Scope limits

- No seller-side rejection UI, trade receipts, notifications, inbox, or audit-history panel shipped.
- No confirmation modal, bulk cancel, search/sort/pagination, or expiration timer UI shipped.
- No market ownership, pricing, matching, or item-transfer semantics changed.
