# v51 Plan - Mystery Seller Core

Status: Implemented (`make ci` green)
Goal: Add a town mystery seller with finite concealed equipment offers that reveal only after a server-authorized purchase.
Architecture: Reuse the existing shop interaction, generated stock persistence, session-start stock snapshots, and `shop_buy_intent` path where possible. Add explicit protocol v8 support for concealed `mystery` offer rows so normal generated offers remain fully visible while mystery offers omit item identity until purchase. The Go server owns stock generation, hidden payloads, price validation, purchase mutation, persistence, private fanout, and replay reconstruction; the Godot client renders hidden rows and sends existing shop-buy intents only.
Tech stack: shared JSON rules/protocol v8, Go shop/sim/realtime/replay tests, Python protocol bot, Godot shop panel/client bot, lifecycle docs.

## Baseline and shortcut decision

Baseline is v50 `account-stash-storage` on `main`. Reuse:

- v41-v47 shop open/buy/sell, generated stock, buyback, appraisals, finite availability, and shop stock persistence.
- v42/v43 server-authored item summaries, comparisons, requirement status, and equip previews after item reveal.
- v47 session-start shop stock snapshots for replay-safe historical shop reconstruction.
- v50 town interactable placement pattern and actor-private fanout discipline.
- Existing `ShopPanel` and bot hooks as the client UI surface.
Key implementation decisions:

- Add `town_mystery_seller` as a separate interactable with `shop_id: "town_mystery_seller"` so vendor and mystery seller stock lifecycles stay independent.
- Add an optional `mystery_offers` rule block to shops. Existing `town_vendor` remains visible generated stock; `town_mystery_seller` uses mystery rows.
- Reuse `character_shop_stock` and `session_start_shop_stock` if possible. The hidden rolled payload can stay in `rolled_payload`; the client only receives a concealed view until purchase.
- Use offer ids prefixed with `mystery:` for hidden rows and `kind: "mystery"` in protocol views.
- Generate one available hidden row for each current equipment slot: `main_hand`, `off_hand`, `head`, `chest`, `gloves`, `belt`, `boots`, `ring`, and `amulet`.
- Enforce `magic` or `rare` rarity for v51. Set/unique catalogs remain deferred.
- Price mystery rows with the seller-specific pricing table plus a slight mystery premium (`1.01` multiplier, 1-gold rounding) so the core proof stays deterministic while final economy tuning against visible vendor prices remains deferred.
- Use existing `shop_buy_intent`; add reveal data to the successful purchase event rather than adding a new buy intent.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/specs/v51_spec-mystery-seller-core.md` | Slice spec |
| Add | `docs/plans/v51_2026-06-10-mystery-seller-core.md` | This implementation plan |
| Modify | `PROGRESS.md` | Lifecycle update when v51 ships |
| Modify | `shared/rules/interactables.v0.json` | Add `town_mystery_seller` |
| Modify | `shared/rules/worlds.v0.json` | Place mystery seller on town level `0` |
| Modify | `shared/rules/shops.v0.json` | Add `town_mystery_seller` and mystery offer rules |
| Modify | `shared/rules/shops.v0.schema.json` | Validate optional mystery offer block |
| Add | `shared/protocol/envelope.v8.schema.json` | Protocol version bump for concealed shop rows |
| Add | `shared/protocol/messages.v8.schema.json` | Current message schema with v8 protocol id |
| Add | `shared/protocol/session_snapshot.v8.schema.json` | Current snapshot schema with v8 protocol id |
| Add | `shared/protocol/state_delta.v8.schema.json` | Mystery offer fields and reveal event shape |
| Modify | `shared/protocol/examples/state_delta.json` | Mystery seller open/purchase examples |
| Add | `shared/protocol/examples/mystery_shop_buy_intent.json` | Existing shop buy intent example for mystery row |
| Add if useful | `shared/golden/mystery_seller.json` | Deterministic hidden stock/reveal fixture |
| Add if useful | `shared/golden/mystery_seller.v0.schema.json` | Fixture schema |
| Modify | `tools/validate_shared.py` | Validate v8 and mystery shop rules/goldens |
| Modify | `server/internal/game/rules.go` | Parse/validate mystery offer rules |
| Modify | `server/internal/game/types.go` | Hidden offer/reveal protocol fields |
| Modify | `server/internal/game/shop.go` | Mystery stock generation, hidden views, pricing, reveal item creation |
| Modify | `server/internal/game/sim.go` | Mystery purchase validation/mutation/events |
| Modify | `server/internal/game/shop_test.go` | Stock, hidden fields, reveal, rejection, persistence-oriented sim tests |
| Modify if needed | `server/internal/store/*` | Only if current shop stock rows cannot represent hidden rows |
| Modify if needed | `server/internal/realtime/*` | Private fanout/persistence extension if existing shop hooks need changes |
| Modify if needed | `server/internal/replay/*` | Replay extension if stock loading needs mystery-specific handling |
| Modify | `client/scripts/shop_panel.gd` | Render concealed rows and debug state |
| Modify | `client/scripts/main.gd` | Route changed reveal/open payloads if needed |
| Modify | `client/scripts/bot_controller.gd` | Bot callable mystery buy/open helpers if existing shop hooks are insufficient |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client-bot mystery assertions |
| Modify | `client/tests/test_shop_panel.gd` | Concealed row rendering/debug tests |
| Modify | `client/tests/test_client_bot.gd` | Scenario validation for mystery actions/assertions |
| Modify | `tools/bot/run.py` | Protocol mystery helpers/assertions |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for concealed row/reveal assertions |
| Add | `tools/bot/scenarios/37_mystery_seller_core.json` | Protocol proof |
| Add | `tools/bot/scenarios/client/24_mystery_seller_core.json` | Godot client proof |

## Task 1 - Shared Rules And Protocol V8

Files:
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/rules/shops.v0.json`
- Modify: `shared/rules/shops.v0.schema.json`
- Add: `shared/protocol/envelope.v8.schema.json`
- Add: `shared/protocol/messages.v8.schema.json`
- Add: `shared/protocol/session_snapshot.v8.schema.json`
- Add: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Add: `shared/protocol/examples/mystery_shop_buy_intent.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `town_mystery_seller` to `interactables.v0.json` with a stable `shop_id: "town_mystery_seller"` and ready initial state.
```bash
make validate-shared
```

- [x] Step 1.2: Place `town_mystery_seller` on `dungeon_levels` town level `0` in `worlds.v0.json`, away from the vendor and stash.
```bash
make validate-shared
```

- [x] Step 1.3: Extend `shops.v0.schema.json` with an optional `mystery_offers` block that validates enabled state, eligible slots, source depth window size, min/max rarity, price multiplier, refresh policy, and max roll attempts.
```bash
make validate-shared
```

- [x] Step 1.4: Add `town_mystery_seller` to `shops.v0.json` with no fixed stock, no buyback, one hidden row per current equipment slot, last-five-depth source window, `magic` minimum rarity, `rare` maximum rarity, and positive mystery price multiplier.
```bash
make validate-shared
```

- [x] Step 1.5: Copy protocol v7 schemas to v8 and update current protocol validation to require v8 files.
```bash
make validate-shared
```

- [x] Step 1.6: Extend v8 shop offer schema to allow `kind: "mystery"` rows with visible concealed fields and without pre-purchase item identity fields.
```bash
make validate-shared
```

- [x] Step 1.7: Extend v8 shop purchase/reveal event schema so successful mystery purchase carries the revealed `ItemView` payload after mutation.
```bash
make validate-shared
```

- [x] Step 1.8: Add protocol examples for mystery seller open, concealed offer rows, mystery shop buy intent, and post-purchase reveal.
```bash
make validate-shared
```

## Task 2 - Server Mystery Stock Rules And Hidden Views

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/shop_test.go`

- [x] Step 2.1: Parse `MysteryOffers` in `ShopDef` and validate eligible slots, rarity floor/cap, positive price multiplier, and sufficient roll attempts.
```bash
cd server && go test ./internal/game/... -run 'TestShopRules|TestMystery' -count=1
```

- [x] Step 2.2: Add hidden-offer fields to `ShopOfferView`, such as `concealed`, `mystery_label`, `source_depth_min`, and `source_depth_max`, while preserving visible generated/fixed/buyback rows.
```bash
cd server && go test ./internal/game/... -run 'Test.*Shop.*Protocol|TestMystery' -count=1
```

- [x] Step 2.3: Generate mystery stock rows deterministically through existing shop stock state, using offer ids prefixed by `mystery:` and one available row per configured slot.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Stock|TestShopStock' -count=1
```

- [x] Step 2.4: Select source depth from `[max(1, deepest_dungeon_depth - 4), max(1, deepest_dungeon_depth)]`, then roll eligible equipment until slot and `magic`/`rare` rarity constraints pass.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Depth|TestMystery.*Rarity' -count=1
```

- [x] Step 2.5: Convert mystery stock rows into concealed `ShopOfferView` values that do not expose item template, actual display name, rarity, stats, requirements, effects, comparison, or equip preview.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Hidden|TestShopOpen' -count=1
```

- [x] Step 2.6: Add deterministic stock/reveal tests, and add/update a golden fixture if useful for stable expected hidden rows and revealed rarity.
```bash
cd server && go test ./internal/game/... -run 'TestMystery|TestShopStockLifecycleGolden' -count=1
```

## Task 3 - Purchase Mutation, Persistence, Private Fanout, And Replay

Files:
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/shop_test.go`
- Modify if needed: `server/internal/realtime/*`
- Modify if needed: `server/internal/replay/*`
- Modify if needed: `server/internal/store/*`

- [x] Step 3.1: Update `findShopOffer`/purchase flow so mystery rows validate against hidden stock payloads and reveal from the persisted rolled payload, not from the concealed protocol row.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Purchase|TestShopOpenBuyAndSell' -count=1
```

- [x] Step 3.2: Implement successful mystery purchase: validate town/range, availability, gold, capacity, and roll payload; subtract gold; consume stock; add revealed item; emit ack, `inventory_add`, `gold_update`, `character_progression_update`, stock availability, and reveal purchase event.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Purchase|TestShopStockFinite' -count=1
```

- [x] Step 3.3: Implement rejection paths for insufficient gold, full inventory, missing offer, unavailable offer, and invalid interaction without consuming stock or mutating inventory/gold.
```bash
cd server && go test ./internal/game/... -run 'TestMystery.*Reject|TestShop.*Reject' -count=1
```

- [x] Step 3.4: Ensure existing realtime persistence handles mystery stock availability and revealed inventory/gold changes, or extend it with focused tests if the existing generated-stock hooks miss mystery rows.
```bash
cd server && go test ./internal/realtime/... ./internal/store/... -run 'Test.*Shop|Test.*Mystery' -count=1
```

- [x] Step 3.5: Ensure replay loads session-start mystery stock snapshots and replays purchase reveal deterministically without reading live stock.
```bash
cd server && go test ./internal/replay/... ./internal/game/... -run 'Test.*Replay|TestMystery' -count=1
```

- [x] Step 3.6: Add or update co-op/private routing tests so another account sees the public mystery seller entity but not hidden rows, reveal payloads, wallet changes, or stock changes.
```bash
cd server && go test ./internal/game/... ./internal/realtime/... -run 'Test.*Mystery.*Coop|Test.*Mystery.*Private|TestShopStockAndBuybackArePerCharacterInCoop' -count=1
```

## Task 4 - Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Add: `tools/bot/scenarios/37_mystery_seller_core.json`

- [x] Step 4.1: Add protocol bot helpers/assertions for concealed mystery rows, hidden-field absence, mystery purchase, revealed item payload, consumed offer count, and gold spend.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 4.2: Add `37_mystery_seller_core.json` using a deterministic seed that can fund one mystery purchase through vendor sale while preserving the seller-specific mystery premium.
```bash
make bot scenario=37_mystery_seller_core.json
```

- [x] Step 4.3: Drive the scenario to acquire enough gold, open the mystery seller, assert one row per configured slot and no hidden identity leaks, buy one affordable row, and assert revealed item rarity is `magic` or `rare`.
```bash
make bot scenario=37_mystery_seller_core.json
```

- [x] Step 4.4: Verify reconnect, `/state`, replay, and fresh-session consumed-stock persistence for the scenario.
```bash
make bot scenario=37_mystery_seller_core.json
```

- [x] Step 4.5: Keep shop stock, vendor, stash, and gold auto-pickup protocol regressions green.
```bash
make bot scenario=33_shop_stock_lifecycle.json
make bot scenario=30_vendor_appraisal_quotes.json
make bot scenario=36_account_stash_storage.json
make bot scenario=35_gold_autopickup_shared_loot.json
```

## Task 5 - Godot Shop UI And Client Bot

Files:
- Modify: `client/scripts/shop_panel.gd`
- Modify if needed: `client/scripts/main.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_shop_panel.gd`
- Modify: `client/tests/test_client_bot.gd`
- Add: `tools/bot/scenarios/client/24_mystery_seller_core.json`

- [x] Step 5.1: Render `mystery` shop rows as concealed offers with category/slot/source-window/price, and keep normal visible generated/fixed/buyback rows unchanged.
```bash
make client-unit
```

- [x] Step 5.2: Ensure concealed rows do not show item tooltip details, actual rarity, rolled stats, comparison, requirements, or equip preview before purchase.
```bash
make client-unit
```

- [x] Step 5.3: Update shop panel debug state and bot hooks to count mystery rows, assert hidden fields, buy a mystery offer, and observe row consumption/inventory/gold updates.
```bash
make client-unit
```

- [x] Step 5.4: Add `24_mystery_seller_core.json` proving the Godot client opens the mystery seller, renders concealed rows, buys one offer, and updates inventory/gold/offer rows.
```bash
make bot-client scenario=24_mystery_seller_core.json HEADLESS=1
```

- [x] Step 5.5: Keep existing shop, stash, inventory, and gold client-bot regressions green.
```bash
make bot-client HEADLESS=1
```

## Task 6 - Lifecycle Docs And CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v51_2026-06-10-mystery-seller-core.md`

- [x] Step 6.1: Add v51 to the slice lifecycle table when implementation finishes.
```bash
make ci
```

- [x] Step 6.2: Add a concise v51 summary under "What each slice proved", including hidden offers, reveal-on-purchase, non-common rarity floor, deterministic stock, and bot/client proof.
```bash
make ci
```

- [x] Step 6.3: Add `37_mystery_seller_core.json` and `24_mystery_seller_core.json` to the scripted scenario catalog.
```bash
make ci
```

- [x] Step 6.4: Record deferred mystery-seller scope in Open gaps: set/unique catalogs, paid reroll, timer refresh, stash overflow delivery, price tuning, final family coverage, production art/audio.
```bash
make ci
```

- [x] Step 6.5: Keep this plan's checkboxes accurate during execution and update status to implemented only after CI is green.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `cd server && go test ./internal/realtime/...`
- [x] `cd server && go test ./internal/replay/...`
- [x] `cd server && go test ./internal/store/...`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make bot scenario=37_mystery_seller_core.json`
- [x] `make bot scenario=33_shop_stock_lifecycle.json`
- [x] `make bot scenario=30_vendor_appraisal_quotes.json`
- [x] `make bot scenario=36_account_stash_storage.json`
- [x] `make bot scenario=35_gold_autopickup_shared_loot.json`
- [x] `make bot-client scenario=24_mystery_seller_core.json HEADLESS=1`
- [x] `make bot`
- [x] `make bot-client HEADLESS=1`
- [x] `make ci`

## Deferred scope

- No set/unique item catalogs or mystery offers above current `rare` rarity.
- No paid reroll, clock/timer refresh, daily seller, or account-wide seller stock.
- No stash overflow delivery, market delivery, refunds, binding, or special resale rules.
- No new item families, affix grammar, item levels, upgrade resources, or final item economy bands.
- No personal loot, co-op trade, shared gold, or remote cross-session push.
