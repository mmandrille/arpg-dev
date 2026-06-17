# v243 As-Built - Market Item Comparison

Date: 2026-06-17

## What shipped

- Added `MarketItemComparison`, a focused Godot helper that derives direct stat deltas from a market
  listing, shared item rules/templates, current inventory, and the equipped map.
- Market listing rows now show comparison lines such as `+3 Armor vs equipped` alongside existing
  stat and expiration lines.
- Market item tooltips pass comparison entries into the shared tooltip comparison section, so
  hover inspection matches row/debug state.
- Market debug listing rows now expose `comparison_count`, `comparison_lines`, and
  `comparison_visible`.
- Main market panel refreshes pass the current equipped map into `MarketPanel` without changing
  market HTTP payloads.
- Client bot market row assertions now support `comparison_at_least` and `comparison_contains`.
- Added `60_market_item_comparison.json`, which opens a preflight market listing and proves
  comparison text through the real client.

## Proof

```bash
godot --headless --path client --script res://tests/test_market_item_comparison.gd
godot --headless --path client --script res://tests/test_market_search_sort.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=60_market_item_comparison.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v241-v250 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=60_market_item_comparison.json
```

## Scope limits

- No server-side market comparison payload, query, pagination, price, ownership, listing, offer, or
  receipt behavior changed.
- No derived full-character stat preview or purchase/equip recommendation shipped.
- No external assets, plugins, or imported art shipped.
