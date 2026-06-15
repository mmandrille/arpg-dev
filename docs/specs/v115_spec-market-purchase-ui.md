# v115 Spec: Market Purchase UI

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `market-purchase-ui`

## Purpose

Make direct gold purchase usable from the Godot market board. A buyer should be able to open the
market board, see a priced listing from another account, press a purchase button, and get clear UI
feedback after the authenticated purchase route accepts.

## Non-goals

- No realtime delivery notification or in-session stash reload after purchase; the v111 route remains
  the authority for persisted delivery.
- No listing cancel/edit, price suggestions, search, sorting, pagination, taxes, expiration, or audit
  feed.
- No protocol changes.

## Acceptance criteria

- `NetClient` exposes `POST /v0/market/listings/{listing_id}/purchase`.
- Browse rows show a purchase button only for non-seller listings with `price_gold > 0`.
- Purchase action refreshes the market listing list and reports a clear status on success/failure.
- Client debug state exposes enough row/status data for bot assertions.
- A client bot scenario prepares a seller listing, funds buyer account-stash gold through gameplay,
  opens the market board, purchases the priced listing, and asserts the purchased row is gone.

## Scope and likely files

- Client: `client/scripts/net_client.gd`, `client/scripts/market_panel.gd`,
  `client/scripts/main.gd`, `client/scripts/bot_controller.gd`,
  `client/scripts/bot_scenario_runner.gd`
- Bot tooling: `scripts/bot_client.sh`, new preflight helper under `tools/bot/`
- Bot scenario: `tools/bot/scenarios/client/36_market_purchase_ui.json`
- Docs: `PROGRESS.md`, `docs/as-built/v115_market-purchase-ui.md`

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=36_market_purchase_ui`
- `make maintainability`
- `make ci`

Manual visual command:

```bash
make bot-visual scenario=36_market_purchase_ui
```
## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | How does the scenario get a foreign listing? | Add a focused seller-listing preflight that uses normal HTTP/WebSocket flows. |
| Q-2 | How does the buyer get stash gold? | Sell compact lab loot through the vendor, then deposit the earned gold into account stash. |
| R-1 | Existing local market rows can pollute count assertions. | Use a unique price and item filter, and assert disappearance of that filtered row. |
| R-2 | In-session stash delivery is not realtime. | Verify purchase acceptance and active-list removal; delivery remains the v111 persisted route contract. |
