# v117 Spec: Market Active Offer UI

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `market-active-offer-ui`

## Purpose

Let sellers inspect and accept active item offers from the Godot market board. This builds on the
existing v93 HTTP/store offer routes and the v114-v115 market panel without changing market
ownership semantics.

## Non-goals

- No offer editing, counteroffers, expiration timers, audit feed, notifications, pagination, search,
  sorting, or item comparison UI.
- No new market protocol/WebSocket messages.
- No listing cancel UI.
- No multi-item offer composition UI beyond the existing one-stash-item Make Offer path.
- No realtime stash delivery notification after acceptance; persisted delivery remains the v93
  store/HTTP contract.

## Acceptance criteria

- `NetClient` exposes authenticated `GET /v0/market/listings/{listing_id}/offers` and
  `POST /v0/market/listings/{listing_id}/offers/{offer_id}/accept`.
- Seller-owned listing rows show an action to inspect active offers.
- The market panel renders offer rows with offer id, bidder account hint, status, and offered item
  names/counts.
- Sellers can accept one active offer from the panel; the listing list refreshes and a clear status
  message reports success/failure.
- Non-seller listing rows keep the existing Make Offer and Buy behavior.
- Client debug state exposes offer rows and status for bot assertions.
- A client bot scenario prepares a seller listing plus a foreign item offer, opens the market board
  as the seller, inspects offers, accepts one, and verifies the active listing row is gone.

## Scope and likely files

- Client: `client/scripts/net_client.gd`, `client/scripts/market_panel.gd`,
  `client/scripts/main.gd`, `client/scripts/bot_controller.gd`,
  `client/scripts/bot_scenario_runner.gd`
- Bot tooling: `scripts/bot_client.sh`, `tools/bot/client_market_preflight.py`
- Bot scenario: `tools/bot/scenarios/client/38_market_active_offer_ui.json`
- Docs: `docs/plans/v117_2026-06-13-market-active-offer-ui.md`,
  `docs/as-built/v117_market-active-offer-ui.md`, `PROGRESS.md`

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=38_market_active_offer_ui`
- `make maintainability`
- `make ci`

Manual visual command:

```bash
make bot-visual scenario=38_market_active_offer_ui
```

## Plugin / shortcut note

Reject external UI plugins/assets. The existing market panel, draggable window chrome, and client
bot preflight harness cover the slice.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Should this add offer expiration or audit records? | No. Keep this to seller inspection and acceptance of already-active offers. |
| Q-2 | How does the scenario create a foreign offer? | Extend the existing market preflight helper to optionally create a buyer account, stash item, and offer against the seller listing. |
| R-1 | Existing local market rows can pollute assertions. | Use a unique item/price and assert disappearance of that filtered listing after acceptance. |
| R-2 | Offer acceptance can imply stash delivery UI. | Defer realtime delivery UI; assert HTTP acceptance and active-list removal only. |

## ADR alignment

- ADR-0011: advances player-facing market offer acceptance while preserving server-owned atomic
  transfer.
- ADR-0014 D12: keeps market work moving toward the long-term endgame economy without adding new
  resource complexity.

