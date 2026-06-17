# v242 As-Built - Market Board Search And Sort

Date: 2026-06-17

## What shipped

- Added native market search/sort controls with default, name, price-low, price-high, and status
  modes.
- Added `MarketRowFilters`, a pure helper that filters/sorts listings, offers, and receipts without
  mutating the source arrays.
- Market browse, publish, incoming/outgoing offer, and receipt views now render from filtered
  derived arrays while keeping server-loaded market data unchanged.
- Market debug state now exposes search text, sort mode, sort options, and filtered listing/offer/
  receipt counts.
- Added focused bot helpers for market actions and market row assertions, including
  `set_market_search` and `select_market_sort`.
- Added `59_market_search_sort.json`, which opens a preflight market listing, filters rows, changes
  sort mode, and verifies the market panel state through the real client.
- Reduced `bot_controller.gd` and `bot_scenario_runner.gd` below their previous maintainability
  baselines and lowered `.maintainability/file-size-baseline.tsv` accordingly.

## Proof

```bash
godot --headless --path client --script res://tests/test_market_search_sort.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=59_market_search_sort.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v241-v250 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=59_market_search_sort.json
```

## Scope limits

- No server-side market query, pagination, indexing, persistence, pricing, offer, receipt, or
  ownership behavior changed.
- No item comparison UI shipped; v243 owns market item comparison.
- No external assets, plugins, or imported art shipped.
