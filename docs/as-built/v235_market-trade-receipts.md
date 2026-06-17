# v235 As-Built - Market Trade Receipts

Date: 2026-06-17

## What shipped

- Added account-scoped market receipt reads over existing immutable `market_audit_records`.
- Added authenticated `GET /v0/market/receipts/mine` with bounded `limit` validation and receipt
  response fields for action, listing, offer, actor, seller, bidder, item, details, and timestamp.
- Added a market-board `Receipts` action that opens a receipt view in the existing market panel.
- Added receipt row rendering and market debug `receipt_rows` for bot assertions.
- Reused the existing market bot view action with a `__receipts` sentinel instead of growing bot
  controller action types.
- Added `52_market_trade_receipts.json`, which cancels an outgoing offer and verifies the
  account-scoped receipt feed shows `offer_canceled`.

## Proof

```bash
cd server && go test ./internal/store -run MarketAudit -count=1
cd server && go test ./internal/http -run MarketReceipt -count=1
cd server && go test ./internal/replay -run '^$'
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=52_market_trade_receipts.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=52_market_trade_receipts.json
```

## Scope limits

- No admin ledger, export, inbox, notification badge, search, pagination, or receipt deletion
  shipped.
- No new audit schema, market semantics, settlement rules, or historical backfill shipped.
- No seller-side rejection UI shipped.
