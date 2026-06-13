# v114 Spec: Market Board UI

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `market-board-ui`

## Purpose

Make the town market board a proven player-facing browsing surface for priced listings. The Godot
market panel should open from the town board, publish a stash item with a simple gold price, refresh
active listings, and show listing price in the browse row.

## Non-goals

- No listing purchase button or delivery flow; v115 owns purchase UI.
- No listing edit, cancel UI, offer management polish, taxes, expiration, audit feed, search, sort,
  or pagination.
- No protocol changes; market remains authenticated HTTP plus existing market service event.
- No external UI plugin adoption.

## Acceptance criteria

- Market publish actions can send `price_gold` through existing authenticated HTTP listing create.
- The market panel exposes a simple publish price control with a deterministic default.
- Browse listing rows display `price_gold`.
- Client debug state exposes listing row ids/prices for bot assertions.
- A client bot scenario opens the town market board, publishes a stash item at a price, refreshes
  the browse tab, and asserts the priced listing is visible.

## Scope and likely files

- Client: `client/scripts/net_client.gd`, `client/scripts/market_panel.gd`,
  `client/scripts/main.gd`, `client/scripts/bot_scenario_runner.gd`, `client/scripts/bot_controller.gd`
- Bot: new `tools/bot/scenarios/client/35_market_board_ui.json`
- Tests/docs: client bot, `PROGRESS.md`, `docs/as-built/v114_market-board-ui.md`

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=35_market_board_ui`
- `make maintainability`
- `make ci`

Manual visual command:

```bash
make bot-visual scenario=35_market_board_ui
```

## Plugin / shortcut note

Reject external UI plugins/assets. The existing in-repo draggable panel and market board model are
already sufficient for this narrow UI proof.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | What default listing price should the UI use? | Use a small deterministic default of 25 gold for the first proof. |
| R-1 | Publishing requires a stash item. | Bot scenario collects loot and deposits it into stash before opening market. |
| R-2 | Existing market panel is partially implemented. | Keep v114 to browse/publish proof and defer purchase to v115. |
