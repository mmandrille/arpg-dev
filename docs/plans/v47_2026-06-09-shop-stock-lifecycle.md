# v47 Plan — Shop Stock Lifecycle

Status: Ready for implementation
Goal: Make town-vendor generated stock finite, per-character, refresh-gated by new waypoint unlocks, and support session-local buyback.
Architecture: The Go sim remains authoritative for stock generation, buy/sell validation, inventory mutation, gold mutation, and buyback lifecycle. Durable generated stock is loaded into the sim from store/session-start snapshots and updated through explicit TickResult changes so replay stays deterministic. Buyback rows are sim/session-local only and are cleared when the actor leaves town; they are never persisted. The Godot client renders server-authored shop rows and sends existing shop intents.
Tech stack: shared JSON rules/schemas/goldens, Go `server/internal/game` + `store` + `realtime` + `replay`, Godot GDScript shop panel, Python protocol bot and client bot.

## Baseline and shortcut decision

Baseline is v46 `client-join-game-proof`, with v41 `town-vendor-gold-sink`, v42 `vendor-appraisal-and-item-comparison`, v43 `equipment-requirements-and-preview`, v44 `skill-points-and-magic-bolt`, and v45/v46 menu/session flows already complete.

Key implementation decisions:

- Add protocol v6 because v5 only allows `fixed` and `generated` shop offer kinds and offer-id prefixes.
- Add durable generated stock tables and session-start stock snapshot tables in migration `0013`.
- Represent generated stock inside the sim as per-player/per-character shop state loaded at session start or lazily created on first open.
- Emit complete server-authored shop rows after successful buy/sell mutations so an open Godot panel can update without reopening.
- Update `AddCharacterWaypoint` to report whether a waypoint was newly inserted, allowing stock refresh once per newly unlocked non-town waypoint.
- Keep fixed potion offers stateless and infinite.
- Keep buyback rows runtime-only: no durable store rows and no session-start snapshot rows.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/plans/v47_2026-06-09-shop-stock-lifecycle.md` | This implementation plan. |
| Modify | `docs/specs/v47_spec-shop-stock-lifecycle.md` | Only if execution discovers approved spec corrections. |
| Modify | `PROGRESS.md` | Lifecycle update when v47 ships. |
| Modify | `shared/rules/shops.v0.json` | Add stock lifecycle, max rarity, source-depth, and buyback rules. |
| Modify | `shared/rules/shops.v0.schema.json` | Validate new shop rule fields. |
| Add | `shared/protocol/envelope.v6.schema.json` | Protocol version bump. |
| Add | `shared/protocol/messages.v6.schema.json` | Allow `buyback:` offer ids for `shop_buy_intent`. |
| Add | `shared/protocol/session_snapshot.v6.schema.json` | Current snapshot schema version if validator requires v6 parity. |
| Add | `shared/protocol/state_delta.v6.schema.json` | Add `buyback` offer kind, stock metadata, and refreshed shop rows on mutations. |
| Modify | `shared/protocol/examples/state_delta.json` | Include finite generated buy, sell-to-buyback, and buyback purchase examples. |
| Add | `shared/golden/shop_stock_lifecycle.json` | Pin source-depth, rarity cap, finite stock, and buyback fixture. |
| Add | `shared/golden/shop_stock_lifecycle.v0.schema.json` | Validate stock lifecycle golden. |
| Modify | `tools/validate_shared.py` | Validate new rules, schemas, examples, and golden drift. |
| Add | `server/migrations/0013_character_shop_stock.sql` | Durable generated stock and session-start stock snapshot tables. |
| Modify | `server/internal/store/models.go` | Shop stock persistence models and session-start snapshot fields. |
| Modify | `server/internal/store/interfaces.go` | Shop stock repo methods and waypoint insert result. |
| Modify | `server/internal/store/repos.go` | Persist, replace, consume, load, and snapshot generated stock. |
| Modify | `server/internal/store/store_test.go` | Store/migration/session-start stock coverage. |
| Modify | `server/internal/game/rules.go` | Parse and validate stock lifecycle rules. |
| Modify | `server/internal/game/types.go` | Offer kind/source-depth fields, shop stock change ops, persisted stock views. |
| Modify | `server/internal/game/shop.go` | Generated stock state, source-depth roll, rarity cap, buyback row creation. |
| Modify | `server/internal/game/sim.go` | Open/buy/sell lifecycle, waypoint refresh hooks, town-exit buyback cleanup. |
| Modify | `server/internal/game/shop_test.go` | Unit/golden coverage for stock lifecycle. |
| Modify | `server/internal/game/game_test.go` | Transition/waypoint cleanup and refresh integration if needed. |
| Modify | `server/internal/realtime/hub.go` | Load durable stock for session creation and attach flows. |
| Modify | `server/internal/realtime/session_loop.go` | Persist shop stock changes and updated waypoint insert result. |
| Modify | `server/internal/realtime/session_loop_test.go` | Actor-scoped shop row refresh and persistence coverage. |
| Modify | `server/internal/replay/replay.go` | Reconstruct stock from session-start snapshots and ordered inputs. |
| Modify | `server/internal/replay/replay_test.go` | Replay stock lifecycle coverage and fake repo surface. |
| Modify | `server/internal/http/*_test.go` | `/state` or fresh-session parity if shop stock is exposed. |
| Modify | `client/scripts/shop_panel.gd` | Render/update finite generated and buyback rows. |
| Modify | `client/scripts/main.gd` | Apply shop row refresh payloads on purchase/sale events. |
| Modify | `client/tests/test_shop_panel.gd` | Panel row removal, buyback count, debug-state tests. |
| Modify | `client/tests/test_client_bot.gd` | Validate new client bot shop assertions. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add buyback/stock assertion support if needed. |
| Modify | `tools/bot/run.py` | Add stock lifecycle assertions and buyback helpers. |
| Add | `tools/bot/scenarios/33_shop_stock_lifecycle.json` | Protocol bot proof. |
| Add | `tools/bot/scenarios/client/22_shop_stock_lifecycle.json` | Godot client proof. |

## Task 1 — Shared Contracts And Rules

Files:
- Modify: `shared/rules/shops.v0.json`
- Modify: `shared/rules/shops.v0.schema.json`
- Add: `shared/protocol/envelope.v6.schema.json`
- Add: `shared/protocol/messages.v6.schema.json`
- Add: `shared/protocol/session_snapshot.v6.schema.json`
- Add: `shared/protocol/state_delta.v6.schema.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Add: `shared/golden/shop_stock_lifecycle.json`
- Add: `shared/golden/shop_stock_lifecycle.v0.schema.json`
- Modify: `tools/validate_shared.py`
- Modify: `client/tests/test_golden.gd`

- [x] Step 1.1: Extend `shops.v0.json` with generated-stock lifecycle fields: `source_depth_policy`, `max_rarity`, `refresh_on`, larger `max_roll_attempts`, and `buyback` settings with multiplier and clear-on-leave-town behavior.
```bash
make validate-shared
```

- [x] Step 1.2: Extend `shops.v0.schema.json` to validate lifecycle fields, max rarity membership, positive buyback multiplier, and fixed-offer rules.
```bash
make validate-shared
```

- [x] Step 1.3: Copy v5 protocol schemas to v6 and update shop offer definitions to allow `kind: "buyback"`, `buyback:` offer ids, `source_depth` or equivalent depth metadata, and complete refreshed `offers` / `sell_appraisals` payloads on `shop_purchase` and `shop_sale`.
```bash
make validate-shared
```

- [x] Step 1.4: Update protocol examples with `shop_opened`, generated purchase consuming a generated offer, sale creating buyback, and buyback purchase consuming buyback.
```bash
make validate-shared
```

- [x] Step 1.5: Add `shop_stock_lifecycle` golden fixture with source-depth cases: level 24/depth 50 -> `25..50`, level 60/depth 50 -> `1..50`, level 1/depth 0 -> `1..1`, plus max rarity `rare` and representative buyback pricing.
```bash
make validate-shared
```

- [x] Step 1.6: Update `tools/validate_shared.py` and `client/tests/test_golden.gd` to validate the new golden and keep existing `shop_pricing` / `shop_offers` checks aligned with finite stock.
```bash
make validate-shared
make client-unit
```

## Task 2 — Store Persistence And Session-Start Stock Snapshots

Files:
- Add: `server/migrations/0013_character_shop_stock.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 2.1: Add durable `character_shop_stock` table with account, character, shop, refresh key, offer id/index, source depth, item template id, rolled payload, buy price, available/consumed state, and timestamps.
```bash
cd server && go test ./internal/store/... -run TestMigrations -count=1
```

- [x] Step 2.2: Add `session_start_shop_stock` table so replay freezes generated stock that existed when a session started.
```bash
cd server && go test ./internal/store/... -run 'SessionStart|ShopStock' -count=1
```

- [x] Step 2.3: Add store models such as `CharacterShopStockItem` and include `ShopStock` on `SessionStartSnapshot`.
```bash
cd server && go test ./internal/store/... -run 'ShopStock|SessionStart' -count=1
```

- [x] Step 2.4: Add repository methods to list durable stock, replace stock for a refresh key, consume/restore stock availability, and load stock from session-start snapshots.
```bash
cd server && go test ./internal/store/... -run ShopStock -count=1
```

- [x] Step 2.5: Change `AddCharacterWaypoint` to return whether a row was newly inserted, then update store tests for first insert vs duplicate insert.
```bash
cd server && go test ./internal/store/... -run 'Waypoint|ShopStock' -count=1
```

- [x] Step 2.6: Extend `CreateSessionStartSnapshot`, `LoadSessionStartSnapshotForMember`, and `LoadSessionStartSnapshots` to write/read shop stock while preserving existing item, waypoint, hotbar, and progression behavior.
```bash
cd server && go test ./internal/store/... -run 'SessionStart|Coop|ShopStock' -count=1
```

## Task 3 — Go Rules, Sim State, And Shop Stock Generation

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/shop_test.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Extend `ShopGeneratedOffers`, `ShopDef`, and validation for source-depth policy, max rarity, refresh trigger, buyback settings, and pricing support for buyback rows.
```bash
cd server && go test ./internal/game/... -run 'TestShopRules|TestLoadRules|TestInvalid' -count=1
```

- [x] Step 3.2: Add sim-owned shop stock state keyed by player/character and shop, plus load helpers for persisted/session-start generated stock.
```bash
cd server && go test ./internal/game/... -run 'TestShop.*Stock|TestSnapshot' -count=1
```

- [x] Step 3.3: Replace stateless `generatedShopOffers` use with stock-backed catalog assembly: infinite fixed offers, available generated stock rows, and temporary buyback rows.
```bash
cd server && go test ./internal/game/... -run 'TestShopGenerated|TestShopOpen' -count=1
```

- [x] Step 3.4: Implement deterministic source-depth selection before item roll, using `character_level + 1` when it is within deepest achieved depth, else any achieved depth from `min_depth`.
```bash
cd server && go test ./internal/game/... -run 'TestShopStockSourceDepth|TestShopStockGolden' -count=1
```

- [x] Step 3.5: Enforce the max rarity cap so generated shop stock skips future rarities above `rare` while common/magic/rare remain valid.
```bash
cd server && go test ./internal/game/... -run 'TestShopStockRarityCap|TestShopStockGolden' -count=1
```

- [x] Step 3.6: Add generated-stock lifecycle change ops in `TickResult` for lazy stock creation, refresh replacement, generated offer consumption, and buyback row refresh payloads.
```bash
cd server && go test ./internal/game/... -run 'TestShop.*Stock|TestShop.*Buy' -count=1
```

- [x] Step 3.7: Update `handleShopBuy` so fixed offers remain infinite, generated offers consume only after all validation passes, and buyback offers return the same item payload then disappear.
```bash
cd server && go test ./internal/game/... -run 'TestShopBuy.*|TestShopStock.*' -count=1
```

- [x] Step 3.8: Update `handleShopSell` so successful sells create temporary buyback rows with full server-authored offer metadata, while equipped/unsellable rejects do not mutate state.
```bash
cd server && go test ./internal/game/... -run 'TestShopSell.*|TestShopBuyback.*' -count=1
```

- [x] Step 3.9: Clear actor buyback rows on any transition away from town level `0`, including stairs and waypoint travel.
```bash
cd server && go test ./internal/game/... -run 'TestShopBuybackClears|TestTeleport|TestDungeonLevel' -count=1
```

- [x] Step 3.10: Refresh durable generated stock exactly once when a newly discovered non-town waypoint is unlocked, and do not refresh on duplicate discovery or town waypoint.
```bash
cd server && go test ./internal/game/... -run 'TestShopStockRefresh|TestTeleporterDiscovery' -count=1
```

## Task 4 — Realtime Persistence And Replay Integration

Files:
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `server/internal/realtime/session_loop_test.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/replay/replay_test.go`
- Modify: `server/internal/http/*_test.go`

- [x] Step 4.1: Load durable generated shop stock when creating session-start snapshots and when attaching members, then pass it into the sim using new load helpers.
```bash
cd server && go test ./internal/realtime/... ./internal/replay/... -run 'SessionStart|ShopStock' -count=1
```

- [x] Step 4.2: Persist sim-emitted stock replacement/consume changes to `character_shop_stock`, including refresh-key replacement after new waypoint unlock.
```bash
cd server && go test ./internal/realtime/... -run 'Shop|Waypoint|Persist' -count=1
```

- [x] Step 4.3: Update persistence handling for the new `AddCharacterWaypoint` boolean result and keep existing waypoint/progression persistence green.
```bash
cd server && go test ./internal/realtime/... ./internal/store/... -run 'Waypoint|Teleporter|Progression' -count=1
```

- [x] Step 4.4: Keep shop events actor-scoped and include refreshed `offers` / `sell_appraisals` only for the acting character after purchase/sale.
```bash
cd server && go test ./internal/realtime/... -run 'ShopDeltasAreActorScoped|Shop' -count=1
```

- [x] Step 4.5: Update replay reconstruction to load session-start shop stock for solo and co-op sessions, while letting in-session lazy stock creation replay deterministically from ordered inputs.
```bash
cd server && go test ./internal/replay/... -run 'ShopStock|Reconstruct|Coop' -count=1
```

- [x] Step 4.6: Add replay coverage for generated buy disappearance, sell-to-buyback, buyback purchase, buyback cleanup on leaving town, and no buyback in fresh session start.
```bash
cd server && go test ./internal/replay/... -run 'ShopStockLifecycle|Verify' -count=1
```

- [x] Step 4.7: Update `/state` or HTTP tests if shop stock is surfaced in inspection output; otherwise document that stock remains shop-event scoped.
```bash
cd server && go test ./internal/http/... -run 'State|Shop' -count=1
```

## Task 5 — Go Golden And Regression Tests

Files:
- Modify: `server/internal/game/shop_test.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `shared/golden/shop_stock_lifecycle.json`

- [x] Step 5.1: Add golden-backed tests for stock source-depth selection, stable offer ids/order, rarity cap, and deterministic rolled payloads.
```bash
cd server && go test ./internal/game/... -run TestShopStockLifecycleGolden -count=1
```

- [x] Step 5.2: Add tests proving fixed potion buys remain infinite and generated equipment buys consume only the selected generated offer.
```bash
cd server && go test ./internal/game/... -run 'TestShopFixedOffersInfinite|TestShopGeneratedStockConsumed' -count=1
```

- [x] Step 5.3: Add tests for failed generated and buyback purchases: insufficient gold, inventory full, unknown/consumed offer, invalid shop, and out of range.
```bash
cd server && go test ./internal/game/... -run 'TestShopBuyFailure|TestShopBuybackFailure' -count=1
```

- [x] Step 5.4: Add tests for sell-to-buyback, buyback purchase returning the same payload, equipped-item rejection, unsellable-item rejection, and hotbar cleanup.
```bash
cd server && go test ./internal/game/... -run 'TestShopSell|TestShopBuyback|TestHotbar' -count=1
```

- [x] Step 5.5: Add co-op tests proving each character's stock, buyback rows, and refreshed shop events are private to that actor.
```bash
cd server && go test ./internal/game/... ./internal/realtime/... -run 'TestShop.*Coop|TestShopDeltasAreActorScoped' -count=1
```

## Task 6 — Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Add: `tools/bot/scenarios/33_shop_stock_lifecycle.json`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 6.1: Update bot state ingestion so any shop event carrying `offers` / `sell_appraisals` refreshes the cached shop rows, not only `shop_opened`.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 6.2: Add bot filters/assertions for `offer_kind: "buyback"`, source-depth ranges, exact/absent offer ids, generated offer count after purchase, and buyback row count.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 6.3: Add a helper or assertion for opening the shop after reconnect/fresh session and confirming durable generated consumption persists while buyback does not.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 6.4: Create `33_shop_stock_lifecycle.json`: open shop, assert five generated rows and source-depth metadata, sell item to buyback, process a purchase refresh, sell the purchased item, rebuy one buyback row, leave town, return, and confirm buyback is gone. Generated offer disappearance is covered by focused Go lifecycle tests so the protocol scenario does not depend on character-specific generated-stock affordability.
```bash
make bot scenario=shop_stock_lifecycle
```

- [x] Step 6.5: Extend the scenario to unlock a new non-town teleporter/waypoint and assert generated stock refreshes exactly once.
```bash
make bot scenario=shop_stock_lifecycle
```

- [x] Step 6.6: Keep existing vendor and requirements scenarios green under finite generated stock.
```bash
make bot scenario=town_vendor_gold_sink
make bot scenario=vendor_appraisal_quotes
make bot scenario=equipment_requirements_and_preview
```

## Task 7 — Godot Shop Panel Updates

Files:
- Modify: `client/scripts/shop_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_shop_panel.gd`

- [x] Step 7.1: Add `buyback_offer_count` and refreshed offer-row debug data to `ShopPanel.get_debug_state()`.
```bash
make client-unit
```

- [x] Step 7.2: Add `ShopPanel` method to apply server-authored refreshed `offers` and `sell_appraisals` while preserving visible panel state and current gold/inventory state.
```bash
make client-unit
```

- [x] Step 7.3: Update `main.gd` shop purchase/sale handling to apply refreshed rows from the server when present, then show status text.
```bash
make client-unit
```

- [x] Step 7.4: Render buyback rows using existing item tooltip/summary/comparison UI and distinguish them through kind/debug data without introducing new art.
```bash
make client-unit
```

- [x] Step 7.5: Add focused `test_shop_panel.gd` coverage for generated row removal, buyback row insertion/removal, fixed offer persistence, sell-row sync, and no text/control overlap regressions in debug rows.
```bash
make client-unit
```

## Task 8 — Client Bot Scenario

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`
- Add: `tools/bot/scenarios/client/22_shop_stock_lifecycle.json`

- [x] Step 8.1: Add client bot assertions for generated offer count, buyback offer count, fixed offer count, and refreshed shop status after purchase/sale.
```bash
make client-unit
```

- [x] Step 8.2: Create `22_shop_stock_lifecycle.json` to open the real shop panel, sell a rolled item into buyback, process a purchase refresh, sell the purchased item into buyback, rebuy one buyback row, and assert generated/fixed rows remain visible. Generated offer disappearance is covered by focused Go lifecycle tests and `test_shop_panel.gd`, because the Godot client scenario cannot assume a generated row is affordable for every character-specific stock roll.
```bash
HEADLESS=1 make bot-client scenario=22_shop_stock_lifecycle
```

- [x] Step 8.3: Extend the client scenario to prove fixed potion offer stays visible after generated/buyback mutations and panel remains open.
```bash
HEADLESS=1 make bot-client scenario=22_shop_stock_lifecycle
```

- [x] Step 8.4: Run existing shop-related client bot scenarios to catch regressions.
```bash
HEADLESS=1 make bot-client scenario=15_town_vendor_shop_panel
HEADLESS=1 make bot-client scenario=16_vendor_item_comparison
HEADLESS=1 make bot-client scenario=17_equipment_requirements_and_preview
```

## Task 9 — Full Regression, Docs, And CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v47_spec-shop-stock-lifecycle.md` only if execution finds approved spec corrections
- Modify: `docs/plans/v47_2026-06-09-shop-stock-lifecycle.md` if execution discovers approved deviations

- [x] Step 9.1: Run shared, Go, client, protocol bot, and client bot focused suites before lifecycle docs.
```bash
make validate-shared
make test-go
make client-unit
make client-smoke
make bot scenario=shop_stock_lifecycle
HEADLESS=1 make bot-client scenario=22_shop_stock_lifecycle
```

- [x] Step 9.2: Update `PROGRESS.md` lifecycle table and "What each slice proved" for v47, including explicit non-goals and the current loot-band caveat.
```bash
rg -n "v47|shop-stock-lifecycle|Latest completed slice|town vendor|buyback" PROGRESS.md
```

- [x] Step 9.3: Run full local CI.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot scenario=shop_stock_lifecycle`
- [x] `HEADLESS=1 make bot-client scenario=22_shop_stock_lifecycle`
- [x] `make bot`
- [x] `make bot-client`
- [x] `make ci`

## Deferred scope

- Fixed potion stock limits.
- Durable buyback across town exits or session ends.
- Multiple vendor types, stash, repair, crafting, gambling, sorting, filters, search, bulk operations, or player trade.
- Clock-based daily/hourly refresh.
- Expanded item-level/depth economy bands beyond the current `1`, `2`, and `3+` loot bands.
- Unique/set item catalogs and unique/set shop offers.
