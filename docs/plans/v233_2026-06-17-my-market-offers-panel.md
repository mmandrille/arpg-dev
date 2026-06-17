# v233 Plan - My Market Offers Panel

Status: Complete
Goal: Let bidders view their own outgoing market offers from the client market board.
Architecture: Add a narrow server read model for bidder-owned offers, then render it through the
existing market panel. No ownership mutation happens in this slice.
Tech stack: Go store/HTTP, Godot UI/client bot, Python preflight, docs.

## Baseline and Shortcut Decision

Builds on v93/v117 market offers and v231 market listing cancel UI. Asset/plugin decision: reject
external assets/plugins; offer rows reuse existing item-icon and tooltip rendering.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/store/market_repo.go` | Query bidder-owned offers |
| Modify | `server/internal/store/interfaces.go` | Expose store method |
| Modify | `server/internal/http/market.go` | Add authenticated mine endpoint |
| Modify | `client/scripts/net_client.gd` | Add client HTTP helper |
| Modify | `client/scripts/market_panel.gd` | Render outgoing offer rows |
| Modify | `client/scripts/main.gd` | Wire market action |
| Modify | `client/scripts/bot_*` | Bot click/assert plumbing |
| Modify | `tools/bot/client_market_preflight.py` | Support bidder-as-client preflight |
| Add | `tools/bot/scenarios/client/50_my_market_offers_panel.json` | Client proof |
| Add | `docs/as-built/v233_my-market-offers-panel.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/market_panel.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep changes narrow and offset local growth in grandfathered files.

Verification:
```bash
make maintainability
```

## Task 1 - Server read endpoint

Files:
- Modify: `server/internal/store/market_repo.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/http/market.go`

- [x] Add `ListMarketOffersForBidder`.
- [x] Add `GET /v0/market/offers/mine`.
- [x] Cover store/HTTP behavior with existing market tests where practical.

```bash
cd server && go test ./internal/store -run MarketOffer -count=1
cd server && go test ./internal/http -run MarketOffer -count=1
```

## Task 2 - Client market panel

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_shop_panel.gd`

- [x] Add a `My Offers` load action and outgoing-offer render/debug state.
- [x] Wire the action through `NetClient`.
- [x] Keep the existing market/shop panel regression green.

```bash
godot --headless --path client --script res://tests/test_shop_panel.gd
```

## Task 3 - Bot proof

Files:
- Modify: `client/scripts/bot_facade.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `tools/bot/client_market_preflight.py`
- Add: `tools/bot/scenarios/client/50_my_market_offers_panel.json`

- [x] Add click/assert support for loading outgoing offers.
- [x] Add a bidder-as-client market preflight.
- [x] Prove the panel in a headless client scenario.

```bash
make bot-client scenario=50_my_market_offers_panel.json HEADLESS=1
```

## Task 4 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v233_my-market-offers-panel.md`

- [x] Record focused verification and deferred scope.

```bash
make maintainability
```

## Final Verification

- [x] `cd server && go test ./internal/store -run MarketOffer -count=1`
- [x] `cd server && go test ./internal/http -run 'MarketOffer|DeleteCharacter' -count=1`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=50_my_market_offers_panel.json HEADLESS=1`
- [x] `make maintainability`
