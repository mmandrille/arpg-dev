# v115 Plan — Market Purchase UI

Status: Complete
Goal: Prove a Godot client can buy another account's priced market listing.
Architecture: Keep ownership transfer server-owned through the existing v111 authenticated HTTP
purchase route. The client only exposes the action, status, and active-list refresh.
Tech stack: Godot client, bot preflight helper, client bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v111 purchase/delivery backend and v114 market board UI.
## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/net_client.gd` | Add market purchase HTTP helper. |
| Modify | `client/scripts/market_panel.gd` | Add purchase button, bot purchase helper, debug status rows. |
| Modify | `client/scripts/main.gd` | Route purchase action and refresh market data. |
| Modify | `client/scripts/bot_controller.gd` | Add market purchase bot action dispatch. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add market status/listing absence assertions. |
| Modify | `scripts/bot_client.sh` | Add seller-listing preflight type. |
| Create | `tools/bot/client_market_preflight.py` | Prepare seller account listing through normal flows. |
| Create | `tools/bot/scenarios/client/36_market_purchase_ui.json` | Client proof for buy + refresh. |
| Modify | `PROGRESS.md` | Mark v115 complete during finish. |
| Create | `docs/as-built/v115_market-purchase-ui.md` | Record shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep market UI behavior in `market_panel.gd`.
- [x] Existing bot registry edits stayed within the v114 grandfathered baseline allowance.
- [x] Run `make maintainability` before final CI.

## Task 1 — Purchase action UI

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Add `purchase_market_listing(listing_id)`.
- [x] Step 1.2: Show a purchase button for non-seller priced listings.
- [x] Step 1.3: Route purchase success/failure through status text and active-list refresh.
```bash
make client-unit
```

## Task 2 — Seller preflight and bot proof

Files:
- Modify: `scripts/bot_client.sh`
- Create: `tools/bot/client_market_preflight.py`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/36_market_purchase_ui.json`

- [x] Step 2.1: Add a `market_listing` client-bot preflight that creates a seller listing at a unique price.
- [x] Step 2.2: Add bot action support for clicking a market purchase row.
- [x] Step 2.3: Add wait/assert support for filtered listing absence through existing market row filters.
- [x] Step 2.4: Add a scenario that funds buyer stash gold, purchases the listing, and asserts refresh.
```bash
make bot-client scenario=36_market_purchase_ui
```

## Task 3 — Lifecycle docs and CI

Files:
- Modify: `docs/plans/v115_2026-06-13-market-purchase-ui.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v115_market-purchase-ui.md`

- [x] Step 3.1: Mark plan tasks complete as they pass.
- [x] Step 3.2: Update `PROGRESS.md` latest slice, next slice, lifecycle row, and recently closed note.
- [x] Step 3.3: Add the v115 as-built note.
```bash
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=36_market_purchase_ui`
- [x] `make maintainability`
- [x] `make ci`
