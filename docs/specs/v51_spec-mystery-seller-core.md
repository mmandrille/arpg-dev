# Spec: `mystery-seller-core`

Status: Implemented
Date: 2026-06-10
Branch: `main`
Codename: `mystery-seller-core`
Slice: v51 - mystery seller core
Baseline: v50 `account-stash-storage`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared contracts, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - town hub and character progression
- [`../adr/0013-mystery-seller-and-unidentified-item-offers.md`](../adr/0013-mystery-seller-and-unidentified-item-offers.md) - mystery seller product direction
- [`v41_spec-town-vendor-gold-sink.md`](v41_spec-town-vendor-gold-sink.md) - shop interaction and gold sink baseline
- [`v42_spec-vendor-appraisal-and-item-comparison.md`](v42_spec-vendor-appraisal-and-item-comparison.md) - server-authored item row details
- [`v47_spec-shop-stock-lifecycle.md`](v47_spec-shop-stock-lifecycle.md) - durable generated shop stock and buyback lifecycle
- [`v50_spec-account-stash-storage.md`](v50_spec-account-stash-storage.md) - town stash and account storage baseline

## 1. Purpose

The town vendor now has finite generated stock and the account has stash storage, but the economy
still lacks a blind-buy equipment sink. This slice adds the first mystery seller: a town
interactable that shows paid unidentified equipment offers and reveals the actual rolled item only
after the server accepts a purchase.

The mystery seller core should prove:

- Town level `0` contains a `town_mystery_seller` interactable.
- Opening the seller shows hidden equipment offers with only category/slot/source-window/price
  metadata.
- Hidden offers never expose the actual item template, display name, rarity, rolled stats,
  requirements, effects, comparison, or equip preview before purchase.
- Buying a hidden offer spends character gold, consumes the offer, reveals the rolled item, and
  delivers it to the character bag through the normal server-owned inventory path.
- Mystery stock is deterministic, per-character, finite, durable across reconnect and fresh
  sessions, and replay-safe through session-start shop stock snapshots.

This is the smallest end-to-end version of ADR-0013. It uses the existing shop interaction shape
where possible, but the protocol must explicitly distinguish concealed mystery offers from normal
visible generated offers.

The existing `ShopPanel`, item tooltip, inventory state, and bot hooks are already the right
display-only surface. The implementation plan should record this checklist decision and extend the

## 2. Non-goals

- No set or unique item catalogs. v51 mystery offers roll only from currently supported `magic` and
  `rare` item rarities.
- No new item families, affix grammar, special effects, item levels, item upgrade resources, or
  final economy rebalance.
- No paid reroll, timer refresh, daily shop, clock-based refresh, account-wide mystery stock, or
  cross-character shared seller inventory.
- No delivery to stash, overflow stash delivery, market delivery, item binding, refund, or special
  resale rules. Successful purchases go to the current character bag or reject if the bag is full.
- No separate mystery-seller panel if the existing shop panel can render hidden rows cleanly.
- No production NPC art, portrait, dialog, sound, custom icons, or imported UI asset pack.
- No personal loot, shared gold, co-op trading, or remote cross-session push.
- No backward-compatibility promise for stale protocol versions beyond coordinated current-dev
  schema, fixture, bot, and client updates.

## 3. Acceptance Criteria

1. Shared interactable/world rules place `town_mystery_seller` on `dungeon_levels` town level `0`
   with a stable `shop_id` such as `town_mystery_seller`.
2. Shared shop rules define mystery-seller stock declaratively: eligible equipment slots/families,
   source-depth window of the last five achieved depths, minimum rarity `magic`, maximum rarity
   `rare`, positive mystery price multiplier, finite availability, and refresh policy.
3. Opening the mystery seller uses existing `action_intent` range/auto-approach behavior and emits
   an actor-private `shop_opened` or equivalent shop event for the owner only.
4. The opened offer rows use a distinct kind such as `mystery` and include only visible hidden-offer
   fields: `offer_id`, `kind`, category, slot or family label, source-depth min/max, buy price,
   and available state.
5. Hidden offer rows omit actual item identity before purchase: no `item_def_id`,
   `item_template_id`, actual `display_name`, `rarity`, `rolled_stats`, `requirements`,
   `effect_ids`, `summary_lines`, `comparison`, or `equip_preview`.
6. Mystery stock is generated per character and per shop from deterministic server RNG labels that
   include session seed, shop id, character id, and refresh key.
7. Mystery source depth is selected from `[max(1, achieved_depth - 4), achieved_depth]`, where
   achieved depth is the character deepest dungeon depth clamped to at least `1`.
8. Each mystery offer rolls equipment from the selected source-depth loot band and rerolls until
   the item is equipment, belongs to the offer slot/family, and has rarity `magic` or `rare`.
9. The seller exposes at least one hidden offer for each currently supported equipment slot:
   `main_hand`, `off_hand`, `head`, `chest`, `gloves`, `belt`, `boots`, `ring`, and `amulet`.
10. Prices use the mystery seller's configured valuation plus a small mystery multiplier. Final
    tuning against visible generated vendor prices is deferred.
11. Reopening the mystery seller, reconnecting, or starting a fresh session preserves the same
    available hidden rows until a row is purchased or the configured refresh key changes.
12. Buying a mystery offer validates town/range, offer availability, sufficient gold, bag capacity,
    and item roll integrity before mutation.
13. Successful purchase atomically subtracts character gold, consumes the mystery offer, adds the
    revealed item to the current character bag, persists character gold/inventory/stock
    availability, and emits actor-private `gold_update`, `inventory_add`,
    `character_progression_update`, and shop refresh/purchase events.
14. The purchase event includes the revealed item payload after purchase so the client and bot can
    assert the server-owned reveal, rarity, and stats.
15. Failed purchase for insufficient gold, full inventory, missing offer, unavailable offer, or
    invalid shop interaction rejects without consuming stock or mutating inventory/gold.
16. Co-op peers can see the public mystery seller entity but do not receive another account's
    hidden offer payloads, purchases, revealed item data, wallet changes, or stock changes.
17. Replay reconstruction uses session-start mystery stock snapshots plus ordered inputs; replay
    does not read live mutated stock while reconstructing a historical session.
18. `/state`, reconnect, fresh session, and protocol replay all observe the same available mystery
    stock and consumed-row state for the owning character.
19. The Godot shop UI renders mystery rows as intentionally concealed offers, shows price and
    category/slot metadata, hides normal item tooltip details before purchase, and displays the
    revealed inventory item after purchase through existing inventory/shop state updates.
20. The Godot client bot can open the mystery seller, assert concealed rows, buy one affordable
    mystery offer, and assert the row is consumed and inventory/gold update.
21. Existing town vendor, shop stock lifecycle, account stash, gold auto-pickup, and inventory
    scenarios remain green.
22. Protocol examples, shared validation, Go tests, client tests, protocol bot, client bot, replay,
    and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v51_spec-mystery-seller-core.md - this spec
docs/plans/v51_2026-06-10-mystery-seller-core.md - implementation plan
PROGRESS.md - lifecycle update when v51 ships

shared/rules/interactables.v0.json - new town mystery seller interactable
shared/rules/worlds.v0.json - town placement for dungeon_levels level 0
shared/rules/shops.v0.json - mystery seller rules and price multiplier
shared/rules/shops.v0.schema.json - validate mystery stock config
shared/protocol/envelope.v8.schema.json - protocol version bump if hidden offer fields require it
shared/protocol/messages.v8.schema.json - current shop buy intent plus any v8 enum updates
shared/protocol/session_snapshot.v8.schema.json - current snapshot shape if version advances
shared/protocol/state_delta.v8.schema.json - hidden mystery offer fields and reveal event shape
shared/protocol/examples/state_delta.json - mystery seller open and purchase reveal examples
shared/golden/mystery_seller.json - deterministic hidden stock and reveal fixture if useful
shared/golden/mystery_seller.v0.schema.json - fixture schema if added
tools/validate_shared.py - validate protocol v8 and mystery seller rule/golden drift

server/internal/game/rules.go - shop rule parsing/validation for mystery stock
server/internal/game/types.go - hidden offer/reveal protocol fields
server/internal/game/shop.go - mystery stock generation, hidden offer views, pricing
server/internal/game/sim.go - purchase validation/mutation and actor-private events
server/internal/game/shop_test.go - deterministic stock, hidden fields, reveal, rejection tests
server/internal/realtime/* - private fanout and stock persistence if existing hooks need extension
server/internal/replay/* - replay with session-start mystery stock
server/internal/store/* - likely reuse character_shop_stock; extend only if hidden metadata needs more columns

client/scripts/shop_panel.gd - render concealed mystery rows and debug state
client/scripts/main.gd - apply mystery purchase/open events if the shape changes
client/scripts/bot_controller.gd - bot action/debug hooks if existing shop hooks are insufficient
client/scripts/bot_scenario_runner.gd - client-bot assertions for concealed rows
client/tests/test_shop_panel.gd - concealed row rendering and buy hook coverage
client/tests/test_client_bot.gd - scenario validation for mystery seller assertions

tools/bot/run.py - protocol helpers/assertions for mystery offers and reveal
tools/bot/test_protocol.py - helper tests for hidden row validation
tools/bot/scenarios/37_mystery_seller_core.json - protocol proof
tools/bot/scenarios/client/24_mystery_seller_core.json - Godot client proof
```

Protocol note: v51 likely needs protocol v8 because current shop rows are visible generated offers
with item identity fields. The plan may reuse `shop_buy_intent` for purchase, but the schemas need
to permit concealed `mystery` offer rows and a post-purchase reveal payload without making hidden
fields optional for normal offers in an ambiguous way.

Persistence note: the existing `character_shop_stock` and `session_start_shop_stock` rows can store
the hidden rolled payload if the implementation adds a mystery kind/offer id convention and keeps
the actual payload server-side until purchase. A migration is only necessary if current columns
cannot express the required hidden metadata.

## 5. Test And Bot Proof

- Shared validation proves the new interactable, shop rules, protocol v8, examples, and any golden
  fixtures are valid.
- Go tests cover shop rule validation, deterministic mystery stock, non-common rarity floor,
  hidden pre-purchase offer rows, purchase reveal, insufficient gold rejection, full-bag rejection,
  consumed-row persistence, and co-op privacy.
- Protocol bot scenario `37_mystery_seller_core.json` opens the seller, asserts hidden offers do
  not leak item identity, buys an affordable mystery offer, asserts gold spend and revealed
  inventory item, verifies reconnect, `/state`, replay, and fresh-session consumed stock.
- Client bot scenario `24_mystery_seller_core.json` opens the live Godot panel, asserts concealed
  row debug state, buys one offer, and asserts the row/inventory/gold UI update.
- Existing vendor and stash scenarios stay green because v51 extends the shop surface used by both.

## 6. Open Questions And Risks

- The exact offer-family list is plan-level detail. The expected minimum is one row per supported
  equipment slot; weapon subfamilies beyond `main_hand` can be deferred.
- Current item templates only define `common`, `magic`, and `rare` item rarities. This slice should
  enforce `magic`/`rare` and leave set/unique promises to future itemization work.
- If existing shop stock persistence cannot distinguish hidden mystery rows from visible generated
  rows cleanly, the plan should add a small migration rather than overloading fields with fragile
  string parsing.
- If the current character usually lacks enough gold for the mystery offer in short bot runs, keep
  the seller-specific pricing table conservative and leave final price tuning to a later economy
  pass.
