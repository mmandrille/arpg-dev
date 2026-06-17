# v243 Plan - Market Item Comparison

Status: Complete
Goal: Show direct market listing comparisons against current equipped gear.
Architecture: Keep market listings server-loaded as-is. The Godot client derives comparison deltas
from listing stats, shared item templates, current inventory, and equipped state through a focused
helper, then renders those lines in rows/tooltips/debug state.
Tech stack: Godot UI/helpers, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v242 market filter helpers and existing shop/inventory comparison presentation.
Asset/plugin decision: reject external assets/plugins; reuse local tooltip and label conventions.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/market_item_comparison.gd` | Compute direct stat deltas, labels, tooltip entries, and debug enrichment. |
| Modify | `client/scripts/market_panel.gd` | Render comparison lines and pass comparison entries to tooltips. |
| Modify | `client/scripts/main.gd` | Pass current equipped map into market panel refreshes. |
| Modify | `client/scripts/bot_market_row_assertions.gd` | Match market comparison debug expectations. |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate market comparison expectations. |
| Create | `client/tests/test_market_item_comparison.gd` | Unit coverage for helper and panel debug state. |
| Create | `tools/bot/scenarios/client/60_market_item_comparison.json` | Client-bot proof for a market comparison row. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Slice lifecycle close-out. |
| Create | `docs/as-built/v243_market-item-comparison.md` | As-built proof summary. |

## Maintenance ratchet

Hotspot / over-limit files touched:
- [x] `client/scripts/market_panel.gd`
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline/allowance?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 - Comparison helper and market rendering

Files:
- Create: `client/scripts/market_item_comparison.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/main.gd`
- Create: `client/tests/test_market_item_comparison.gd`

- [x] Step 1.1: Add a pure helper for item stat extraction, equipped-slot selection, and deltas.
- [x] Step 1.2: Render comparison lines in market listing rows and shared item tooltips.
- [x] Step 1.3: Enrich debug listing rows with comparison count/text.
- [x] Step 1.4: Add focused unit coverage.

## Task 2 - Bot proof

Files:
- Modify: `client/scripts/bot_market_row_assertions.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Create: `tools/bot/scenarios/client/60_market_item_comparison.json`

- [x] Step 2.1: Add comparison expectations for market listing rows.
- [x] Step 2.2: Add a focused client bot scenario that asserts comparison text on a preflight listing.

## Task 3 - Lifecycle docs and close-out

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Create: `docs/as-built/v243_market-item-comparison.md`

- [x] Step 3.1: Mark docs complete and record verification.
- [x] Step 3.2: Update current status to v243 with final batch CI pending.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_market_item_comparison.gd`
- [x] `godot --headless --path client --script res://tests/test_market_search_sort.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=60_market_item_comparison.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` deferred to `$autoloop` after the selected queue completes.
