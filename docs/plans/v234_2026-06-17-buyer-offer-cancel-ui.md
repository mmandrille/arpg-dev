# v234 Plan - Buyer Offer Cancel UI

Status: Complete
Goal: Let bidders cancel outgoing market offers from the `My Offers` market-board view.
Architecture: Reuse the existing bidder-scoped cancel endpoint and `NetClient` helper, then keep the
client in the `My Offers` read view after canceling.
Tech stack: Godot UI/client bot, existing Go market tests, docs.

## Baseline and Shortcut Decision

Builds on v233 `My Offers` and the existing `CancelMarketOffer` backend route. Asset/plugin
decision: reject external assets/plugins; this is a button/action state change in existing market UI.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/market_offer_rows.gd` | Render outgoing active offers with a cancel action |
| Modify | `client/scripts/market_panel.gd` | Emit/drive cancel action from matching outgoing offers |
| Modify | `client/scripts/main.gd` | Call `NetClient.cancel_market_offer`, refresh `My Offers`, update stash |
| Modify | `client/scripts/bot_*` | Bot click/assert plumbing for canceling outgoing offers |
| Add | `tools/bot/scenarios/client/51_buyer_offer_cancel_ui.json` | Client proof |
| Add | `docs/as-built/v234_buyer-offer-cancel-ui.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/market_panel.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Replace accept-only bot offer plumbing with shared offer-action plumbing to avoid growing
  `bot_controller.gd`.
- [x] Offset any market/main growth locally or extract only if the slice requires it.

Verification:
```bash
make maintainability
```

## Task 1 - Outgoing cancel UI

Files:
- Modify: `client/scripts/market_offer_rows.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/main.gd`

- [x] Show `Cancel` for active outgoing offers and keep `Accept` for seller incoming offers.
- [x] Emit `cancel_offer` with listing/offer IDs from outgoing rows.
- [x] On success, restore returned offer items to local stash, refresh inventory, and reload `My Offers`.

```bash
godot --headless --path client --script res://tests/test_shop_panel.gd
```

## Task 2 - Bot proof

Files:
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_action_step_validator.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Add: `tools/bot/scenarios/client/51_buyer_offer_cancel_ui.json`

- [x] Add `click_market_cancel_offer` scenario support without growing `bot_controller.gd`.
- [x] Prove outgoing offer cancel from `My Offers`.
- [x] Prove the offered item returns to stash.

```bash
make bot-client scenario=51_buyer_offer_cancel_ui.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Add: `docs/as-built/v234_buyer-offer-cancel-ui.md`

- [x] Record focused verification and deferred scope.

```bash
make maintainability
```

## Final Verification

- [x] `cd server && go test ./internal/store -run MarketOfferCancel -count=1`
- [x] `cd server && go test ./internal/http -run MarketOfferCancel -count=1`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=51_buyer_offer_cancel_ui.json HEADLESS=1`
- [x] `make maintainability`
