# v243 Spec - Market Item Comparison

Status: Complete
Date: 2026-06-17
Codename: market-item-comparison

## Purpose

Make market listing rows decision-ready by showing direct stat comparisons against the player's
currently equipped item for the same slot.

## Non-goals

- No server-side market payload/query changes, pagination, pricing, ownership, or comparison
  authority changes.
- No derived full-character stat preview; compare direct item stats only.
- No listing edit, offer recommendation, auto-equip, or purchase confirmation flow.
- No external assets/plugins.

## Acceptance criteria

- Market listing rows expose at least one comparison line when the listed item has direct stats,
  including the no-equipped-item case where equipped stats are zero.
- Market item tooltips include the same comparison entries using the shared tooltip comparison
  section.
- Comparison uses shared item/template data and current client inventory/equipped state without
  mutating listing or inventory arrays.
- Market debug listing rows expose comparison count and text for bot/unit proof.
- Client unit coverage proves comparison delta generation and market panel debug output.
- Client bot proof opens a real preflight market listing and asserts a comparison line is available.

## Scope and likely files

- Client UI/helpers: `client/scripts/market_panel.gd`, new focused helper under `client/scripts/`.
- Client wiring: `client/scripts/main.gd` to pass equipped state to the market panel.
- Client bot assertions: `client/scripts/bot_market_row_assertions.gd`,
  `client/scripts/bot_step_catalog.gd`, and a new scenario.
- Tests/docs: focused Godot unit test, `PROGRESS.md`, lifecycle row, and as-built.

## Test and bot proof

- `godot --headless --path client --script res://tests/test_market_item_comparison.gd`
- `godot --headless --path client --script res://tests/test_market_search_sort.gd`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=60_market_item_comparison.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=60_market_item_comparison.json
```

## Client asset/plugin decision

Reject external assets/plugins. The slice reuses existing item stat labels, market row labels, and
the shared tooltip comparison section already used by shop and inventory.

## Risks

- `market_panel.gd` remains grandfathered, so comparison rendering must mostly live in a helper.
- Market listings are loaded from account-stash snapshots, so client-side comparison must tolerate
  sparse rows and fall back to rule/template stats.
