# v241 Spec - Market Expiration Countdown

Status: Complete
Date: 2026-06-17
Codename: market-expiration-countdown

## Purpose

Show player-facing expiration timing on market listing rows so buyers and sellers can tell that
published items are time-limited before they purchase, cancel, or make offers. The slice should
reuse the existing server-owned `market_listings.expires_at` value and keep expiration enforcement
unchanged.

## Non-goals

- No changes to market expiration duration, sweep policy, ownership transfer, refunds, pricing, or
  audit semantics.
- No realtime countdown ticks, notification inbox, badge, pagination, search, or sorting.
- No new art assets, plugins, or external UI dependencies.
- No seller offer rejection or market history expansion.

## Acceptance criteria

- Market listing HTTP responses include `expires_at` for active listings and nested offer listing
  context.
- Godot market rows render a stable expiration line derived from `expires_at`, with a safe fallback
  when the field is missing or malformed.
- Market debug listing rows expose `expires_at`, `expiration_label`, and whether the expiration line
  is visible for client bot assertions.
- A client bot scenario opens the market board and verifies a listing row exposes expiration text.
- Focused server/client tests cover response serialization and expiration label/debug behavior.

## Scope and likely files

- Server HTTP: `server/internal/http/market.go`, focused market HTTP tests if needed.
- Client: `client/scripts/market_panel.gd`, a focused helper under `client/scripts/`,
  `client/tests/test_shop_panel.gd`.
- Bot: `client/scripts/bot_scenario_runner.gd`, `client/scripts/bot_step_catalog.gd`,
  `tools/bot/scenarios/client/58_market_expiration_countdown.json`.
- Docs: `docs/plans/v241_2026-06-17-market-expiration-countdown.md`,
  `docs/as-built/v241_market-expiration-countdown.md`, `PROGRESS.md`,
  `docs/progress/slice-lifecycle.md`.

## Test and bot proof

- `cd server && go test ./internal/http -run Market -count=1`
- `godot --headless --path client --script res://tests/test_market_listing_rows.gd`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=58_market_expiration_countdown.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=58_market_expiration_countdown.json
```

## Client asset/plugin decision

Reject external assets/plugins. This is text-only market UI polish that borrows the existing market
panel row style and item icon rendering.

## Open questions and risks

- Exact wall-clock remaining time can drift in tests; bot proof should assert the presence and
shape of the expiration label rather than an exact duration.
- `market_panel.gd` is over 600 lines and already at its allowance, so the implementation must
  extract helper code or otherwise keep the touched file from growing.
