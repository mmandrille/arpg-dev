# v235 Plan - Market Trade Receipts

Status: Complete
Goal: Show account-scoped market trade receipts from the market board.
Architecture: Add a narrow account-filtered read endpoint over existing market audit rows, then
render those rows in the existing market panel.
Tech stack: Go store/HTTP, Godot UI/client bot, docs.

## Baseline and Shortcut Decision

Builds on v130 market audit records and v233-v234 market panel actions. Asset/plugin decision:
reject external assets/plugins; receipts are text rows in the existing market UI.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/internal/store/market_receipts.go` | Query recent audit records for an account |
| Modify | `server/internal/store/interfaces.go` | Expose receipt query |
| Add | `server/internal/http/market_receipts.go` | Add authenticated receipts endpoint |
| Modify | `server/internal/http/market.go` | Register receipt route |
| Modify | `client/scripts/net_client.gd` | Add receipt HTTP helper |
| Modify | `client/scripts/market_panel.gd` | Show receipt view and debug rows |
| Modify | `client/scripts/market_offer_rows.gd` | Render receipt rows/buttons |
| Modify | `client/scripts/main.gd` | Wire list receipts action |
| Add | `client/scripts/bot_market_receipt_assertions.gd` | Receipt row matching |
| Modify | `client/scripts/bot_wait_handlers.gd`, `client/scripts/bot_step_catalog.gd` | Receipt assertions |
| Add | `tools/bot/scenarios/client/52_market_trade_receipts.json` | Client proof |
| Add | `docs/as-built/v235_market-trade-receipts.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/market_panel.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `server/internal/store/market_repo.go`

Decision:
- [x] Reuse `click_market_view_offers` sentinel plumbing for receipts instead of adding a bot action.
- [x] Put row rendering in `market_offer_rows.gd` and offset any main/market-panel growth locally.

Verification:
```bash
make maintainability
```

## Task 1 - Receipt endpoint

Files:
- Add: `server/internal/store/market_receipts.go`
- Modify: `server/internal/store/interfaces.go`
- Add: `server/internal/http/market_receipts.go`
- Modify: `server/internal/http/market.go`

- [x] Add `ListMarketAuditRecordsForAccount`.
- [x] Add `GET /v0/market/receipts/mine`.
- [x] Cover account filtering and HTTP auth behavior with focused tests.

```bash
cd server && go test ./internal/store -run MarketAudit -count=1
cd server && go test ./internal/http -run MarketReceipt -count=1
```

## Task 2 - Client receipt view

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/market_offer_rows.gd`
- Modify: `client/scripts/main.gd`

- [x] Add `Receipts` button from market browse view.
- [x] Render receipt rows in the offers tab surface with readable action/item details.
- [x] Expose receipt rows in market debug state.

```bash
godot --headless --path client --script res://tests/test_shop_panel.gd
```

## Task 3 - Bot proof

Files:
- Add: `client/scripts/bot_market_receipt_assertions.gd`
- Modify: `client/scripts/bot_wait_handlers.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Add: `tools/bot/scenarios/client/52_market_trade_receipts.json`

- [x] Add receipt row wait/assert matching.
- [x] Prove an outgoing offer cancel produces a visible `offer_canceled` receipt.

```bash
make bot-client scenario=52_market_trade_receipts.json HEADLESS=1
```

## Task 4 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Add: `docs/as-built/v235_market-trade-receipts.md`

- [x] Record focused verification and deferred scope.

```bash
make maintainability
```

## Final Verification

- [x] `cd server && go test ./internal/store -run MarketAudit -count=1`
- [x] `cd server && go test ./internal/http -run MarketReceipt -count=1`
- [x] `cd server && go test ./internal/replay -run '^$'`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=52_market_trade_receipts.json HEADLESS=1`
- [x] `make maintainability`
