# v64 Plan - Mystery Seller Paid Reroll

Status: Complete
Goal: Let players spend gold to replace current mystery seller hidden stock.
Architecture: Add a server-owned `shop_reroll_intent` for mystery sellers only. Reroll cost lives in
shared shop rules. The sim mutates gold and generated mystery stock atomically, emits stock replace
and shop reroll events, and persistence reuses existing `shop_stock_replace` handling. The client
adds a display-only reroll control to the existing shop panel.
Tech stack: shared JSON rules/protocol, Go sim/shop tests, Python bot, Godot shop panel/client bot,
lifecycle docs.

## Baseline and shortcut decision

Baseline is v63 `runtime-sim-error-construction` on `main`, with v51 mystery seller stock already
persisted through generated shop stock rows.

Godot plugin shortcut decision: **reject external plugin adoption**. The adoption checklist in
`docs/researchs/godot-plugins-and-shortcuts.md` was reviewed before this autoloop. This slice only
adds one button/control to the existing server-backed `ShopPanel`; external shop/inventory plugins
would add authority risk and unnecessary UI surface.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/specs/v64_spec-mystery-seller-paid-reroll.md` | Slice spec |
| Add | `docs/plans/v64_2026-06-11-mystery-seller-paid-reroll.md` | This plan |
| Modify | `shared/rules/shops.v0.json` | Add mystery reroll cost |
| Modify | `shared/rules/shops.v0.schema.json` | Validate reroll cost |
| Modify | `shared/protocol/messages.v8.schema.json` | Add `shop_reroll_intent` |
| Add | `shared/protocol/examples/mystery_shop_reroll_intent.json` | Protocol example |
| Modify | `server/internal/game/types.go` | Input/event/shop offer reroll fields |
| Modify | `server/internal/game/rules.go` | Parse/validate reroll cost |
| Modify | `server/internal/game/shop.go` | Paid reroll stock replacement helpers |
| Modify | `server/internal/game/handlers.go` | Register and implement reroll intent |
| Modify | `server/internal/game/shop_test.go` | Reroll success/rejection/persistence-oriented tests |
| Modify | `client/scripts/shop_panel.gd` | Mystery reroll control/debug state |
| Modify | `client/scripts/bot_controller.gd` | Bot action wrapper |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client-bot reroll assertions/actions |
| Modify | `client/scripts/main.gd` | Shop reroll status/refresh handling |
| Modify | `client/tests/test_shop_panel.gd` | Reroll button/intent coverage |
| Modify | `client/tests/test_client_bot.gd` | Scenario validation |
| Modify | `tools/bot/run.py` | Protocol reroll action/assertions |
| Add | `tools/bot/scenarios/39_mystery_seller_paid_reroll.json` | Protocol proof |
| Add | `tools/bot/scenarios/client/29_mystery_seller_paid_reroll.json` | Godot client proof |
| Modify | `PROGRESS.md` | Lifecycle update |
| Add | `docs/as-built/v64_mystery-seller-paid-reroll.md` | As-built summary |

## Task 1 - Shared Rules And Protocol

- [x] Step 1.1: Add `reroll_cost: 50` to `town_mystery_seller.mystery_offers` and schema validation.
```bash
make validate-shared
```

- [x] Step 1.2: Add `shop_reroll_intent` to protocol v8 messages plus a mystery reroll example.
```bash
make validate-shared
```

## Task 2 - Server Reroll Mutation

- [x] Step 2.1: Add input payload and event fields for shop reroll cost/refresh metadata.
```bash
cd server && go test ./internal/game/... -run TestMystery -count=1
```

- [x] Step 2.2: Implement mystery-only paid reroll with deterministic refresh-key suffix, gold spend,
  stock replacement, and `shop_reroll` event.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Reroll|TestMystery' -count=1
```

- [x] Step 2.3: Cover insufficient gold, non-mystery shop rejection, and different stock after reroll.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Reroll|TestShopOpenBuyAndSell' -count=1
```

## Task 3 - Protocol Bot

- [x] Step 3.1: Add protocol bot action/assertion support for `reroll_shop`.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 3.2: Add protocol scenario `39_mystery_seller_paid_reroll.json`.
```bash
make bot scenario=mystery_seller_paid_reroll
```

## Task 4 - Godot Shop Panel

- [x] Step 4.1: Add mystery reroll button/debug state to `ShopPanel` and emit `shop_reroll_intent`.
```bash
make client-unit
```

- [x] Step 4.2: Add client bot action/assertions and scenario `29_mystery_seller_paid_reroll.json`.
```bash
make bot-client scenario=29_mystery_seller_paid_reroll
```

## Task 5 - Lifecycle Docs And CI

- [x] Step 5.1: Update plan checkboxes, `PROGRESS.md`, and as-built docs.
```bash
rg -n "v64|mystery-seller-paid-reroll|Latest completed slice|reroll" PROGRESS.md docs/as-built docs/plans/v64_2026-06-11-mystery-seller-paid-reroll.md
```

- [x] Step 5.2: Run final verification.
```bash
make validate-shared
make test-go
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make ci`

## Deferred scope

- Timer/daily refreshes, account-wide stock, final economy tuning, paid rerolls for normal vendors,
  and stash overflow delivery remain deferred.
