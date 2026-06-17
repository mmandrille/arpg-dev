# v231 As-Built - Market Cancel Offer UI

Date: 2026-06-16

## What shipped

- Added a Godot `NetClient.cancel_market_listing` helper for the existing authenticated
  `POST /v0/market/listings/{listing_id}/cancel` route.
- Added a seller-owned market row `Cancel` button next to `View Offers`.
- Routed cancel actions through `main.gd`, showing success/failure status, restoring the returned
  listed item into local stash state, and refreshing market rows.
- Added bot action plumbing and scenario validation for `click_market_cancel_listing`.
- Added `48_market_cancel_listing_ui.json`, which preflights a seller listing with an active bidder
  offer, cancels it from the market panel, verifies the seller row disappears, and verifies the
  listed item returns to stash.
- Kept touched grandfathered client files within the maintainability ratchet by offsetting
  `main.gd` and `bot_controller.gd` growth locally.

## Proof

```bash
make client-unit
make bot-client scenario=48_market_cancel_listing_ui.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
passed after the selected v226-v232 feature queue completed.

Manual visual proof, if desired:

```bash
make bot-visual scenario=48_market_cancel_listing_ui.json
```

## Scope limits

- No backend cancel semantics, listing edit, confirmation modal, audit UI, search, sorting,
  pagination, taxes, expiration timers, or notification inbox shipped.
- No buyer-side offer cancel UI shipped.
- No new realtime stash delivery channel shipped beyond the returned cancel response and local
  panel refresh.
