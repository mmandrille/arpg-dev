# v42 Plan - Vendor Appraisal and Item Comparison

Status: Implemented - `make ci` green
Spec: `docs/specs/v42_spec-vendor-appraisal-and-item-comparison.md`
Baseline: v41 `town-vendor-gold-sink` on `main`
Date: 2026-06-09

## Goal

Make the town vendor decision-ready by showing server-authored item summaries, sell appraisals, and
direct stat comparisons instead of bare price/button rows.

## Architecture

The server remains authoritative for pricing and comparison data. `shop_opened` is extended with
display-ready offer summaries and sell appraisals for the acting player's unequipped sellable bag
items. Buy/sell intents stay unchanged from v41. The Godot client renders the server-provided data
and does not locally reprice, reroll, or validate economy outcomes.

## Tech stack

Shared JSON protocol/examples, Go authoritative sim/tests, Python protocol bot, Godot GDScript
panel/unit tests/client bot, and lifecycle docs.

## Baseline And Shortcut Decision

This builds directly on v41's `ShopPanel`, `shop_opened`, deterministic shop pricing, and
`shop_buy_intent` / `shop_sell_intent`. Godot plugin checklist result: **reject external plugin**
for this slice. GLoot/Godot-Inventory would add adapter and vendoring work while v41 already has a
small in-repo panel backed by the correct authoritative protocol.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/specs/v42_spec-vendor-appraisal-and-item-comparison.md` | Slice contract |
| Create | `docs/plans/v42_2026-06-09-vendor-appraisal-and-item-comparison.md` | Implementation plan |
| Modify | `shared/protocol/state_delta.v4.schema.json` | Appraisal/comparison schema |
| Modify | `shared/protocol/examples/state_delta.json` | Detailed `shop_opened` example |
| Create | `shared/golden/shop_appraisals.json` | Deterministic appraisal fixture |
| Create | `shared/golden/shop_appraisals.v0.schema.json` | Fixture schema |
| Modify | `tools/validate_shared.py` | Validate appraisal fixture |
| Modify | `server/internal/game/types.go` | Shop item summary/appraisal/comparison views |
| Modify | `server/internal/game/shop.go` | Server-authored summaries, appraisals, comparison deltas |
| Modify | `server/internal/game/sim.go` | Include appraisals on `shop_opened` |
| Modify | `server/internal/game/shop_test.go` | Go appraisal/comparison tests |
| Modify | `server/internal/realtime/session_loop_test.go` | Shop event shape if needed |
| Modify | `tools/bot/run.py` | Appraisal/comparison assertions |
| Create | `tools/bot/scenarios/30_vendor_appraisal_quotes.json` | Protocol proof |
| Modify | `client/scripts/shop_panel.gd` | Rich row rendering/debug state |
| Modify | `client/tests/test_shop_panel.gd` | Panel detail/comparison tests |
| Modify | `client/scripts/bot_controller.gd` / `client/scripts/bot_scenario_runner.gd` | Client bot debug/assertion if needed |
| Create | `tools/bot/scenarios/client/16_vendor_item_comparison.json` | Client UI proof |
| Modify | `PROGRESS.md` | Lifecycle closeout |

## Task 1 - Shared Protocol And Golden Fixture

Files:
- Create: `shared/golden/shop_appraisals.json`
- Create: `shared/golden/shop_appraisals.v0.schema.json`
- Modify: `shared/protocol/state_delta.v4.schema.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `shop_item_comparison`, `shop_item_comparison_delta`, and `shop_sell_appraisal`
  definitions to `state_delta.v4.schema.json`.
- [x] Step 1.2: Extend `shop_offer` with optional `slot`, `category`, `summary_lines`, and
  `comparison`.
- [x] Step 1.3: Extend `shop_opened` events with optional `sell_appraisals`.
- [x] Step 1.4: Add a `shop_appraisals` golden fixture covering one fixed consumable, one generated
  offer compared against equipped gear, one unequipped sell appraisal, and equipped-item exclusion.
- [x] Step 1.5: Validate the new golden fixture in `tools/validate_shared.py`.

```bash
make validate-shared
```

## Task 2 - Server Appraisals And Comparison Views

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/shop_test.go`
- Modify: `server/internal/realtime/session_loop_test.go`

- [x] Step 2.1: Add protocol view structs for comparison deltas, comparison groups, and sell
  appraisals.
- [x] Step 2.2: Add helper functions that summarize item category/slot/effects/stats into stable
  display lines.
- [x] Step 2.3: Add server-side direct stat comparison against currently equipped item in the same
  slot.
- [x] Step 2.4: Include comparison data on generated equipment offers.
- [x] Step 2.5: Include sell appraisals for unequipped sellable inventory items on `shop_opened`.
- [x] Step 2.6: Add Go tests for offer summaries, sell appraisals, equipped exclusion, and comparison
  deltas.

```bash
cd server && go test ./internal/game/... -run 'TestShop'
cd server && go test ./internal/realtime/...
```

## Task 3 - Protocol Bot Proof

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/30_vendor_appraisal_quotes.json`

- [x] Step 3.1: Track `sell_appraisals` from `shop_opened` events.
- [x] Step 3.2: Add assertions for offer detail fields, sell appraisal count, sell price, and
  comparison deltas.
- [x] Step 3.3: Add a scenario that opens the vendor after obtaining/equipping gear, verifies offer
  detail/comparison data, verifies sell appraisals, then buys and sells as v41 did.

```bash
make bot scenario=vendor_appraisal_quotes
make bot
```

## Task 4 - Godot Shop Panel Rendering

Files:
- Modify: `client/scripts/shop_panel.gd`
- Modify: `client/tests/test_shop_panel.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/16_vendor_item_comparison.json`

- [x] Step 4.1: Replace compact bare rows with stable detail rows that show item name, price,
  slot/kind, stats/effects, and comparison lines.
- [x] Step 4.2: Render `sell_appraisals` when present, falling back to local inventory rows only for
  old/debug data.
- [x] Step 4.3: Expose debug state for visible buy/sell detail lines and comparison row count.
- [x] Step 4.4: Update `test_shop_panel.gd` for summaries, sell prices, disabled buy state, and
  comparison debug data.
- [x] Step 4.5: Add or update client bot assertions and a scenario for the real vendor panel.

```bash
make client-unit
HEADLESS=1 make bot-client scenario=16_vendor_item_comparison.json
```

## Task 5 - Lifecycle Docs And Final Verification

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v42_2026-06-09-vendor-appraisal-and-item-comparison.md`

- [x] Step 5.1: Add v42 lifecycle row, numbering note, summary, bot scenario list, and deferred
  gaps to `PROGRESS.md`.
- [x] Step 5.2: Mark plan checkboxes complete as tasks pass.
- [x] Step 5.3: Run final verification.

```bash
make validate-shared
cd server && go test ./...
make client-unit
make bot scenario=vendor_appraisal_quotes
HEADLESS=1 make bot-client scenario=16_vendor_item_comparison.json
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./...`
- [x] `make client-unit`
- [x] `make bot scenario=vendor_appraisal_quotes`
- [x] `HEADLESS=1 make bot-client scenario=16_vendor_item_comparison.json`
- [x] `make ci`

## Deferred

- Derived character-stat previews after hypothetical equip.
- Buyback, stash, repair, crafting, search, sorting, filters, and bulk operations.
- External inventory/shop UI plugin adoption.
- Item/economy rebalance beyond v41 pricing.
