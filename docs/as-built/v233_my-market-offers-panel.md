# v233 As-Built - My Market Offers Panel

Date: 2026-06-17

## What shipped

- Added authenticated `GET /v0/market/offers/mine` so bidders can load their own outgoing market
  offers with offer items and listed-item metadata.
- Added a Godot market-board `My Offers` action that renders outgoing offers in the existing offers
  tab, including status text, listing item context, offered item icons, and bot-visible debug rows.
- Extracted market offer row construction to `client/scripts/market_offer_rows.gd` to keep
  `market_panel.gd` inside the file-size ratchet while sharing incoming/outgoing row rendering.
- Extended the client bot market preflight so a scenario can create the offer as the client account.
- Added `50_my_market_offers_panel.json`, which opens the market board, loads `My Offers`, and
  asserts the outgoing offer row for a foreign listing.
- Repaired character cleanup for session-start resource wallet rows so market client-bot preflights
  can clean accounts without HTTP 500 cleanup noise.

## Proof

```bash
cd server && go test ./internal/store -run 'TestDeleteCharacterRemovesProgressionAndSessions|MarketOffer' -count=1
cd server && go test ./internal/http -run 'MarketOffer|DeleteCharacter' -count=1
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=50_my_market_offers_panel.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=50_my_market_offers_panel.json
```

## Scope limits

- No buyer-side cancel button shipped; v234 owns canceling outgoing offers from this view.
- No trade receipts, audit history, inbox, notifications, pagination, sorting, or expiration timer
  UI shipped.
- No market ownership or item-transfer semantics changed.
