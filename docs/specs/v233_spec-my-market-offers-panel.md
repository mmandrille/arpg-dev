# v233 Spec - My Market Offers Panel

Status: Approved for autoloop
Date: 2026-06-17
Codename: my-market-offers-panel

## Purpose

Give bidders a player-facing way to review their own active market offers from the Godot market
board. The server already owns offer reservation, acceptance, rejection, and cancel semantics; this
slice exposes bidder-owned offers through a narrow authenticated read route and renders them in the
existing market UI.

## Non-goals

- No buyer-side cancel button; v234 owns canceling outgoing offers from this view.
- No trade receipt/audit history; v235 owns receipts.
- No listing search, sorting, pagination, expiration timers, or notification inbox.
- No new market ownership rules or item transfer semantics.

## Acceptance Criteria

- Authenticated `GET /v0/market/offers/mine` returns offers where the current account is the bidder,
  ordered newest-first, including offer status, item snapshots, and minimal listed-item metadata.
- The Godot market panel exposes a `My Offers` view/tab/action that shows outgoing offer rows without
  requiring the bidder to own the listing.
- Outgoing offer rows include status, listing item, offered item icons, and are available in market
  debug state for headless client-bot assertions.
- A client bot scenario creates a foreign listing, submits an offer as the client bidder, opens the
  market board, loads `My Offers`, and asserts the outgoing offer row is visible.

## Scope and Likely Files

- Store/HTTP: `server/internal/store/market_repo.go`, `server/internal/store/interfaces.go`,
  `server/internal/http/market.go`.
- Client: `client/scripts/net_client.gd`, `client/scripts/market_panel.gd`, narrow wiring in
  `client/scripts/main.gd` and bot facade/controller files.
- Bot/scenario: `tools/bot/client_market_preflight.py`,
  `tools/bot/scenarios/client/50_my_market_offers_panel.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `cd server && go test ./internal/store ./internal/http -run MarketOffer -count=1`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=50_my_market_offers_panel.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. The view is read-only and uses existing offer/listing ownership rules.
- Risk: market client files are over the line-count target. Keep behavior focused and offset touched
  grandfathered-file growth.
