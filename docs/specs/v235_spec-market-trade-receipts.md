# v235 Spec - Market Trade Receipts

Status: Approved for autoloop
Date: 2026-06-17
Codename: market-trade-receipts

## Purpose

Make the market's existing audit trail visible to players as account-scoped trade receipts. Players
should be able to open the market board, view recent market actions that involved their account, and
confirm completed or canceled trade outcomes without inspecting logs or tests.

## Non-goals

- No admin ledger, export, inbox, notification badge, search, pagination, or receipt deletion.
- No new audit table, market semantics, settlement rules, or historical backfill.
- No seller-side rejection UI; this slice only displays receipts for actions already recorded.

## Acceptance Criteria

- Authenticated `GET /v0/market/receipts/mine` returns recent market audit records where the
  current account is the actor, seller, or bidder, newest-first.
- Receipt records include action, listing id, offer id when present, item id, stash item id,
  seller/bidder/actor ids, details, and timestamp.
- The Godot market panel exposes a player-facing `Receipts` view from the market board.
- Receipt rows are readable, include action/status and item context, and are available in market
  debug state for client-bot assertions.
- A client bot scenario creates an outgoing offer as the client, cancels it, opens `Receipts`, and
  asserts the canceled receipt row is visible.

## Scope and Likely Files

- Store/HTTP: `server/internal/store/market_repo.go`, `server/internal/store/interfaces.go`,
  `server/internal/http/market.go`.
- Client: `client/scripts/net_client.gd`, `client/scripts/market_panel.gd`,
  `client/scripts/market_offer_rows.gd`, `client/scripts/main.gd`.
- Bot/scenario: `client/scripts/bot_scenario_runner.gd`, `client/scripts/bot_step_catalog.gd`,
  `tools/bot/scenarios/client/52_market_trade_receipts.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `cd server && go test ./internal/store -run MarketAudit -count=1`
- `cd server && go test ./internal/http -run MarketReceipt -count=1`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=52_market_trade_receipts.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. v130 already records the audit rows; this slice exposes and renders them.
- Risk: market client files are on line-count boundaries. Prefer helper-row rendering and sentinel
  reuse over new controller actions.
