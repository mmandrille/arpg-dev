# Spec: `shop-stock-lifecycle`

Status: Draft
Date: 2026-06-09
Branch: `main`
Codename: `shop-stock-lifecycle`
Slice: v47 - per-character shop stock lifecycle
Baseline: v46 `client-join-game-proof`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - town hub, character persistence, waypoints
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - client UI shortcut checklist
- [`v41_spec-town-vendor-gold-sink.md`](v41_spec-town-vendor-gold-sink.md) - first town vendor, fixed/generated offers, buy/sell, pricing
- [`v42_spec-vendor-appraisal-and-item-comparison.md`](v42_spec-vendor-appraisal-and-item-comparison.md) - server-authored shop rows and comparisons
- [`v43_spec-equipment-requirements-and-preview.md`](v43_spec-equipment-requirements-and-preview.md) - shop requirement status and equip previews

## 1. Purpose

The town vendor currently derives the same five generated equipment offers from a deterministic
formula whenever the catalog is requested. That avoids rerolling through repeated opens, but the
shop still behaves like a stateless quote sheet: buying a generated item does not remove the offer,
and selling an item destroys it rather than making it available for buyback.

This slice gives the vendor a real, server-owned stock lifecycle while keeping the existing thin
client boundary:

- Fixed potion offers remain infinite.
- The five generated equipment offers become durable per-character stock rows.
- Buying generated stock consumes that row, so it disappears from the shop.
- Selling an unequipped sellable item creates a temporary buyback row for that character.
- Buying a buyback row returns the same item payload and removes the row.
- Temporary buyback stock clears when the character leaves town or starts a new session.
- Generated stock refreshes only when the character unlocks a new non-town teleporter/waypoint.
- Generated shop stock never rolls above `rare` rarity.

The source depth for each generated offer is rolled before the item roll. For each of the five
generated offers, the server first selects a source dungeon depth from the character's achieved
depth range, then rolls the equipment item from the common dungeon-mob loot table for that source
depth. We choose `character_level + 1` as the lower bound because it matches the desired example:
a level 24 character with dungeon depth 50 rolls source depths `25..50`. If the character level is
greater than or equal to the deepest achieved dungeon depth, the level floor is ignored and the
source depth rolls from any achieved depth.

Client shortcut decision: reject external shop/inventory plugin adoption for this slice. The
existing `ShopPanel`, inventory panel, item tooltip, drag/drop, and client bot hooks already cover
the required UI surface. The plan should still record the adoption checklist result.

## 2. Non-goals

- No new vendor types, vendor personalities, NPC dialog, stash, repair, crafting, gambling, search,
  sorting, filters, bulk operations, player trade, or marketplace.
- No production vendor art, portraits, audio, SFX, town art, or imported UI asset pack.
- No clock-based daily/hourly refresh, server-time dependency, or check-count-based refresh.
- No generated stock from champion/rare/unique monster loot-depth offsets; shop stock uses common
  dungeon-mob treasure classes only.
- No richer item-level economy rebalance or new depth bands. Current `1`, `2`, and `3+` loot bands
  may make high source depths resolve to the same loot table until a future economy slice expands
  them.
- No unique item or set item catalog. v47 only caps shop-generated rarity so future rarity catalogs
  cannot leak above `rare`.
- No durable buyback. Sold-item buyback is intentionally session-local and town-local.
- No preservation of old generated stock compatibility beyond coordinated contract, fixture, and
  test updates for the current development state.

## 3. Acceptance Criteria

1. Shared shop rules describe generated stock count, max generated rarity `rare`, source-depth
   policy, refresh trigger, buyback pricing, and fixed infinite offers.
2. Generated stock is stored per character and per shop. Reopening the shop, reconnecting, or
   starting a fresh session does not produce new generated stock unless the refresh trigger changed.
3. Initial generated stock is created lazily when the character first opens the shop, using a
   deterministic, character-scoped RNG label that does not depend on how many times the shop is
   opened.
4. Unlocking a new non-town teleporter/waypoint refreshes the character's five generated stock rows
   exactly once for that newly unlocked waypoint state.
5. Unlocking a waypoint that was already discovered does not refresh stock.
6. Each generated stock row rolls `source_depth` first, then rolls an equipment item from the common
   dungeon-mob treasure class for that source depth.
7. Source-depth bounds are:
   - if `character_level + 1 <= deepest_dungeon_depth`, roll uniformly from
     `[character_level + 1, deepest_dungeon_depth]`;
   - otherwise roll uniformly from `[min_shop_depth, max(1, deepest_dungeon_depth)]`.
8. Generated stock skips no-drop rolls, currency, consumables, quest items, non-equipment entries,
   and any future item rarity above `rare`.
9. The generated catalog contains exactly five available equipment offers unless shared rules are
   invalid; exhausting configured roll attempts is a test/validation failure, not a silent partial
   catalog.
10. Fixed red and blue potion offers remain buyable indefinitely and are not consumed by purchases.
11. Buying generated stock validates range, town state, offer availability, gold, inventory
    capacity, and requirements metadata before mutating; success removes the generated offer,
    subtracts gold, adds the item, persists character inventory/gold/shop stock, and emits an
    actor-scoped shop mutation event.
12. Failed generated-stock buys do not consume the offer or mutate inventory/gold.
13. Selling an unequipped sellable item removes it from character inventory, clears hotbar
    references if needed, adds sell gold, and creates a temporary buyback offer for the same
    character in the current session while the character remains in town.
14. Buyback offers expose the same server-authored display fields as generated offers/appraisals:
    item identity, rarity, stats, requirements, equip preview, comparison, and buy price.
15. Buying a buyback offer validates gold and capacity, removes the buyback row, subtracts gold,
    and returns the original item instance/payload to the character inventory.
16. Buyback rows clear when the character leaves town level `0` through stairs, waypoint travel, or
    any other level transition away from town.
17. Buyback rows are absent at fresh session start and are not stored in character persistence or
    session-start snapshots.
18. Shop offers and sell appraisals remain actor-scoped in co-op. One character's generated stock,
    buyback rows, purchases, and sales do not change another character's shop view.
19. The Godot shop panel removes consumed generated/buyback rows after successful purchase, shows
    temporary buyback rows after sale, keeps fixed potion offers visible, and keeps inventory/sell
    rows synchronized without requiring the player to close and reopen the panel.
20. Protocol examples, shared validation, Go tests, client tests, protocol bot, client bot, replay,
    and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v47_spec-shop-stock-lifecycle.md - this spec
docs/plans/v47_2026-06-09-shop-stock-lifecycle.md - implementation plan
PROGRESS.md - lifecycle update when v47 ships

shared/rules/shops.v0.json - stock lifecycle rules, max rarity, buyback pricing
shared/rules/shops.v0.schema.json - validate new shop-stock fields
shared/protocol/envelope.v6.schema.json - protocol version bump if needed
shared/protocol/messages.v6.schema.json - allow buyback offer ids if intent schema requires it
shared/protocol/session_snapshot.v6.schema.json - current protocol shape if stock metadata becomes snapshot-visible
shared/protocol/state_delta.v6.schema.json - `buyback` offer kind and shop mutation refresh payloads
shared/protocol/examples/state_delta.json - shop open, generated buy, sell-to-buyback, buyback purchase
shared/golden/shop_stock_lifecycle.json - deterministic stock/depth/rarity/buyback fixture
shared/golden/shop_stock_lifecycle.v0.schema.json - fixture schema
tools/validate_shared.py - validate stock lifecycle rules and golden drift

server/migrations/*_character_shop_stock.sql - durable per-character generated stock rows
server/internal/store/models.go - shop stock persistence models
server/internal/store/interfaces.go - shop stock repository surface
server/internal/store/repos.go - load/upsert/consume/refresh generated shop stock
server/internal/store/store_test.go - persistence, session-start stock snapshot, migration coverage

server/internal/game/rules.go - parse/validate stock lifecycle rules
server/internal/game/types.go - shop offer kind/source-depth/stock protocol views
server/internal/game/shop.go - stock generation, depth roll, rarity cap, buyback offer views
server/internal/game/sim.go - buy/sell stock lifecycle, town-exit buyback cleanup, waypoint refresh hook
server/internal/game/shop_test.go - golden, finite stock, buyback, refresh, co-op scoping
server/internal/realtime/* - load/persist stock changes and actor-scoped shop mutation events
server/internal/replay/* - replay reconstruction from session-start stock and ordered inputs
server/internal/http/*_test.go - `/state`/fresh-session parity if stock becomes inspectable

client/scripts/shop_panel.gd - render/update finite stock and buyback rows
client/scripts/main.gd - consume server shop mutation payloads while panel is open
client/tests/test_shop_panel.gd - stock row removal/buyback row/debug-state tests
client/scripts/bot_controller.gd, client/scripts/bot_scenario_runner.gd - debug/assertion hooks if needed

tools/bot/run.py - stock lifecycle assertions and buyback helpers
tools/bot/scenarios/33_shop_stock_lifecycle.json - protocol proof
tools/bot/scenarios/client/22_shop_stock_lifecycle.json - client UI proof
```

Protocol note: current v5 shop schemas only allow `fixed` and `generated` offer kinds and
`fixed:`/`generated:` offer ids. v47 likely needs protocol v6 to add `buyback` offer rows and to
send complete updated shop rows after buy/sell mutations. The plan may choose whether mutation
events carry `offers`/`sell_appraisals` directly or use a dedicated actor-scoped stock refresh
event, but the client must receive server-authored complete rows after every successful mutation.

## 5. Data And Behavior Draft

### 5.1 Shop rules

`shops.v0.json` should remain declarative. Suggested generated-offer additions:

```json
{
  "generated_offers": {
    "offer_count": 5,
    "source": "common_dungeon_mob",
    "min_depth": 1,
    "source_depth_policy": "character_level_plus_one_to_deepest_else_any_achieved",
    "max_rarity": "rare",
    "refresh_on": "new_non_town_waypoint",
    "max_roll_attempts": 128
  },
  "buyback": {
    "enabled": true,
    "scope": "session_town_visit",
    "buy_price_multiplier": 1.0,
    "clear_on_leave_town": true
  }
}
```

The exact field names are plan-level detail, but the rules must validate:

- generated offer count is positive,
- max rarity exists in item-template rarity rules and is not above `rare`,
- buyback multiplier is positive,
- max roll attempts can satisfy the configured offer count,
- fixed offers still point to sellable non-currency item defs.

### 5.2 Generated stock persistence

Generated stock is durable character state. Logical row fields:

```text
account_id
character_id
shop_id
refresh_key
offer_index
offer_id
source_depth
item_template_id
rolled_payload
buy_price
available
```

`refresh_key` represents the character's discovered non-town teleporter/waypoint state. It can be
a monotonic counter, highest unlocked teleporter depth if waypoint unlocks stay monotonic, or a
stable signature of the sorted non-town waypoint levels. The required behavior is "new waypoint
level refreshes once; repeated discovery does not."

Session-start snapshots must freeze the persistent generated stock state for replay, just as they
freeze inventory, progression, hotbar, and waypoints today. Later live character stock changes must
not mutate historical replay start state.

### 5.3 Source-depth roll

For each generated offer slot:

```text
max_depth = max(1, character.deepest_dungeon_depth)
floor = character.level + 1

if floor <= max_depth:
  min_depth = floor
else:
  min_depth = shops[shop_id].generated_offers.min_depth

source_depth = uniform_int(min_depth, max_depth)
loot_table = dungeon_generation.loot_band_for_depth(source_depth).monster_loot_table
treasure_class = loot_tables[loot_table].treasure_class_id
roll equipment from treasure_class
reject item rarities above max_rarity
```

Examples:

- Character level `24`, deepest achieved depth `50`: source depth rolls from `25..50`.
- Character level `60`, deepest achieved depth `50`: source depth rolls from `1..50`.
- Character level `1`, deepest achieved depth `0`: source depth rolls from `1..1`.

The chosen source depth is exposed in shop offer payloads as `depth` or `source_depth` so tests and
debug UI can prove the rule.

### 5.4 Opening the shop

```text
player actions town_vendor
  -> server resolves actor character and shop
  -> server loads or lazily creates generated stock for that character/shop/refresh_key
  -> server combines infinite fixed offers + available generated stock + temporary buyback rows
  -> server computes sell appraisals for unequipped sellable inventory
  -> server emits actor-scoped shop_opened with complete rows
```

Opening the shop never mutates gold or inventory. It may lazily create missing generated stock or
refresh stale generated stock if the character unlocked a new non-town waypoint since the last
stock generation.

### 5.5 Buying

Fixed offers:

- remain available after purchase,
- allocate a normal item instance,
- use existing fixed price and capacity checks.

Generated offers:

- must exist and be available for the acting character,
- are consumed only after all validation passes,
- persist unavailable/consumed state so reopening or starting a fresh session cannot buy the same
  generated row again.

Buyback offers:

- must exist in the actor's temporary town-visit buyback set,
- return the original item instance id and rolled payload,
- are removed only after all validation passes.

All successful buys send updated gold, progression, inventory, and complete shop rows for the
currently open panel.

### 5.6 Selling and buyback

Selling keeps the v41/v42 validation path:

- the item must belong to the acting character,
- equipped items reject with `item_equipped`,
- unsellable items reject without mutation,
- sell price is server-authored.

On success, the item is removed from character inventory and added to the actor's temporary buyback
stock. The buyback price defaults to the normal shop buy price for that exact item payload, adjusted
by `buyback.buy_price_multiplier`. With multiplier `1.0`, selling a generated item for 25% of its
buy value and rebuying at full buy value cannot generate gold.

Buyback rows clear when the actor leaves level `0`. They are not written to durable character stock
or session-start snapshots and are empty on fresh session start.

## 6. Test And Bot Proof

Expected verification:

1. `make validate-shared` validates shop stock rules, protocol examples, and stock lifecycle
   goldens.
2. Go unit tests cover source-depth bounds, deterministic generated stock, rarity cap, finite
   generated purchase, fixed infinite purchase, failed-buy no mutation, sell-to-buyback, buyback
   purchase, town-exit cleanup, fresh-session no buyback, and co-op actor scoping.
3. Store tests cover durable generated stock creation, refresh-key replacement, consumed rows,
   session-start stock snapshots, and no persistence for buyback rows.
4. Replay tests prove generated stock, purchases, sales, buyback purchase, and town-exit cleanup
   reconstruct from the same session-start stock and ordered inputs.
5. Protocol bot scenario `33_shop_stock_lifecycle.json` proves:
   - open shop and observe five generated offers,
   - buy one generated offer and observe it disappears,
   - reopen/reconnect and confirm it remains gone,
   - sell an item and observe a buyback row,
   - rebuy it and confirm the row disappears and inventory regains the item,
   - sell another item, leave town, return, and confirm buyback is gone,
   - unlock a new non-town teleporter and confirm generated stock refreshes once.
6. Client bot scenario `22_shop_stock_lifecycle.json` proves the Godot panel updates visible offer
   counts, generated/buyback row counts, fixed potion visibility, inventory rows, and status text
   without requiring a panel reopen.
7. `make client-unit`, `make client-smoke`, `make bot`, `make bot-client`, and `make ci` pass.

## 7. Open Questions And Risks

No product questions block planning. User decisions already resolved:

- Source-depth lower bound uses `character_level + 1`.
- Fixed potions are infinite.
- Shop stock is per character.
- Buyback does not survive leaving town or session end.

Risks for planning:

- **Protocol shape:** v47 almost certainly needs a coordinated v6 schema update for `buyback`
  offers and server-authored panel refresh after mutations.
- **Replay and persistence:** durable generated stock must be captured in session-start snapshots;
  temporary buyback must not be captured.
- **Waypoint refresh trigger:** current store `AddCharacterWaypoint` may need to report whether a
  waypoint was newly inserted so stock refresh happens once per new non-town waypoint.
- **Current loot bands:** depths above `3` currently resolve to the `3+` band, so the new source
  depth algorithm is correct but not very visible until a later item-depth economy expansion.
