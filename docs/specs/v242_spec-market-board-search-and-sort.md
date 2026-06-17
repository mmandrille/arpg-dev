# v242 Spec - Market Board Search And Sort

Status: Complete
Date: 2026-06-17
Codename: market-board-search-and-sort

## Purpose

Add display-only search and sort controls to the Godot market board so players can scan active
listings, their own listings, offers, outgoing offers, and receipts without changing server market
queries or ownership rules.

## Non-goals

- No server-side query parameters, pagination, indexing, persistence, or market semantics changes.
- No offer rejection, listing edit, notifications, or receipt export.
- No item comparison UI; v243 owns market item comparison.
- No external UI plugins or assets.

## Acceptance criteria

- The market panel shows a search input and sort selector using existing UI styling.
- Search filters visible market listings by item name/id, seller, price, expiration text, status,
  offer items, and receipt action/item where applicable.
- Sort supports at least default/source order, item/name, price low/high, and status/action modes;
  unavailable sort keys degrade without hiding rows.
- Filtering and sorting are client-only and do not mutate `listings`, `active_offers`, or
  `market_receipts`.
- Market debug state exposes filter text, sort mode, option labels, and filtered row counts.
- Client unit coverage proves filter/sort behavior. Client bot proof sets market filter/sort
  controls through the real panel and asserts filtered market rows.

## Scope and likely files

- Client UI: `client/scripts/market_panel.gd`, focused helpers under `client/scripts/`.
- Client bot: `client/scripts/bot_controller.gd`, `client/scripts/bot_action_step_validator.gd`,
  `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_scenario_runner.gd`,
  `tools/bot/scenarios/client/59_market_search_sort.json`.
- Tests: new focused Godot unit test for market filter controls/helpers.
- Docs: spec, plan, as-built, `PROGRESS.md`, and lifecycle row.

## Test and bot proof

- `godot --headless --path client --script res://tests/test_market_search_sort.gd`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=59_market_search_sort.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=59_market_search_sort.json
```

## Client asset/plugin decision

Reject external assets/plugins. This is native Godot Control UI that borrows the existing market
panel row styling and the same text-input/sort-control pattern already used in stash/session panels.

## Open questions and risks

- `bot_controller.gd` is already at its file-size allowance; adding actions requires a focused
  market bot-action extraction in the same slice.
- Search/sort must remain presentation-only. Any server-side filtering belongs to a later market
  pagination/query slice.
