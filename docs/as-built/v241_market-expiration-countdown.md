# v241 As-Built - Market Expiration Countdown

Date: 2026-06-17

## What shipped

- Exposed `expires_at` on market listing HTTP responses, including nested listing context in offer
  rows.
- Added `MarketListingRows`, a focused Godot helper for market listing titles, detail text, short
  labels, expiration labels, and debug row construction.
- Rendered a display-only expiration line on market listing rows. Valid timestamps show an
  `Expires in ...` label, expired timestamps show `Expired`, and malformed/missing values degrade
  to a safe compact fallback or no line.
- Added market debug row fields: `expires_at`, `expiration_label`, and `expiration_visible`.
- Extended client bot market row matching with `expiration_visible` and `expiration_contains`.
- Added `58_market_expiration_countdown.json`, which opens a preflight market listing and verifies
  expiration text in the row debug state.
- Shrunk `client/scripts/market_panel.gd` by extracting helper logic, keeping the touched
  grandfathered file within the maintainability ratchet.

## Proof

```bash
cd server && go test ./internal/http -run Market -count=1
godot --headless --path client --script res://tests/test_market_listing_rows.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=58_market_expiration_countdown.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v241-v250 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=58_market_expiration_countdown.json
```

## Scope limits

- No market expiration duration, sweep policy, ownership transfer, refund, pricing, audit,
  notification, pagination, search, sorting, or realtime refresh behavior changed.
- No external assets, plugins, or imported art shipped.
