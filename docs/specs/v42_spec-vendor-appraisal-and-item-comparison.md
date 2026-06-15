# Spec: `vendor-appraisal-and-item-comparison`

Status: Draft
Date: 2026-06-09
Branch: `main`
Slice: v42 - vendor appraisals and item comparison
Baseline: v41 `town-vendor-gold-sink`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - town hub and NPC economy direction
- [`v41_spec-town-vendor-gold-sink.md`](v41_spec-town-vendor-gold-sink.md) - first town vendor, deterministic offers, buy/sell, deepest dungeon depth

## 1. Purpose

v41 gives gold a use, but the current vendor panel is not decision-ready: offer rows show little
more than a price and a Buy button, and sell rows do not explain what item is being sold or why its
value is what it is. This slice adds server-authored item appraisals and client-side comparison
presentation so the player can understand vendor offers before spending gold.

The server remains authoritative for all pricing. The shop open event includes enough item summary
and appraisal data for the client to render buy prices, sell prices, item slots, rarity, stats,
requirements, and simple comparison deltas against the actor's currently equipped item in the same
slot. The client only displays this data and continues to send the same buy/sell intents from v41.

The thin proof:

- Generated equipment offers visibly show name, slot, rarity, buy price, and stat lines.
- Fixed potion offers visibly show item name, item kind, effect summary, and buy price.
- Sell rows visibly show item name, slot/kind, rarity, stat/effect summary, and server sell price.
- Equipment offers compare against currently equipped gear in the same slot with +/- stat deltas.
- Co-op/private shop behavior stays actor-scoped.

## 2. Non-goals

- No stash, buyback tab, crafting, repair, gambling, trade, or multiple vendor types.
- No full item search, sorting, filtering, loot filter, or bulk buy/sell.
- No procedural name generation, affix grammar, unique/set catalog, or item economy rebalance.
- No derived-total preview that predicts all character stats after equip; v42 compares direct item
  stats only.
- No new active skill, mana spender, skill bar behavior, or combat tuning.
- No production vendor art, portrait, audio, or town art pass.
  the default is to extend the in-repo `ShopPanel` because v41 already owns the panel and server
  protocol adapter.

## 3. Acceptance Criteria

1. `shop_opened` offers include display-ready item metadata for both fixed consumables and generated
   equipment: name, kind/category, slot where relevant, rarity where relevant, price, stats/effects,
   requirements, and comparison data for equipment.
2. The server computes sell appraisals for the actor's unequipped sellable inventory items and sends
   them with `shop_opened`.
3. Sell appraisals include item instance id, display name, kind/category, slot where relevant, rarity
   where relevant, sell price, stats/effects, requirements, and comparison data where applicable.
4. Equipped items are not sellable in the appraisal list and still reject `shop_sell_intent` with
   `item_equipped`.
5. The Godot shop panel renders more than bare prices: each buy and sell row includes item identity,
   price, slot/kind, and at least one stat/effect/requirement/comparison line when data exists.
6. Comparison lines use the currently equipped item in the corresponding slot and show signed direct
   stat deltas for generated equipment offers and sellable equipment rows.
7. The panel remains usable on the existing 560x520 shop surface without text overlap; long item
   names clip or wrap inside the row rather than covering buttons.
8. Buying and selling still use the v41 `shop_buy_intent` / `shop_sell_intent` payloads and server
   validation remains authoritative for gold, capacity, range, and ownership.
9. Shared validation, Go tests, client unit tests, protocol bot, client bot, and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v42_spec-vendor-appraisal-and-item-comparison.md - this spec
docs/plans/v42_2026-06-09-vendor-appraisal-and-item-comparison.md - implementation plan
PROGRESS.md - lifecycle update when v42 ships

shared/protocol/state_delta.v4.schema.json - add appraisal/comparison payloads to shop_opened
shared/protocol/examples/state_delta.json - include detailed shop_opened appraisals
shared/golden/shop_appraisals.json - deterministic appraisal/comparison fixture
shared/golden/shop_appraisals.v0.schema.json - appraisal fixture schema
tools/validate_shared.py - validate appraisal golden if needed

server/internal/game/types.go - appraisal/comparison protocol views
server/internal/game/shop.go - item summaries, sell appraisals, comparison deltas
server/internal/game/shop_test.go - appraisal and comparison coverage
server/internal/realtime/session_loop_test.go - recipient-scoped shop event shape if needed

client/scripts/shop_panel.gd - richer row rendering and debug state
client/tests/test_shop_panel.gd - offer/sell row summary and comparison tests
client/scripts/bot_controller.gd, client/scripts/bot_scenario_runner.gd - assertions if client bot needs new debug fields

tools/bot/run.py - assertions for appraisals/comparison payloads
tools/bot/scenarios/30_vendor_appraisal_quotes.json - protocol proof
tools/bot/scenarios/client/16_vendor_item_comparison.json - client UI proof
```

Protocol note: v42 may extend `state_delta.v4.schema.json` in place if the project treats v4 as the
current coordinated protocol version from v41. If the plan finds current validators require a new
major protocol file, use v5 consistently across envelope/messages/snapshot/delta examples. The
default is no new client intent shape.

## 5. Data Shape Draft

`shop_opened` gains optional `sell_appraisals`:

```json
{
  "event_type": "shop_opened",
  "entity_id": "1013",
  "shop_id": "town_vendor",
  "offers": [
    {
      "offer_id": "generated:depth3:000",
      "kind": "generated",
      "item_def_id": "cave_bow",
      "item_template_id": "cave_bow",
      "display_name": "Magic Cave Bow",
      "rarity": "magic",
      "slot": "main_hand",
      "rolled_stats": { "damage_min": 3, "damage_max": 7 },
      "requirements": { "level": 1 },
      "buy_price": 120,
      "comparison": {
        "slot": "main_hand",
        "equipped_item_instance_id": "1004",
        "deltas": [
          { "stat": "damage_min", "offered": 3, "equipped": 2, "delta": 1 },
          { "stat": "damage_max", "offered": 7, "equipped": 5, "delta": 2 }
        ]
      }
    }
  ],
  "sell_appraisals": [
    {
      "item_instance_id": "1008",
      "item_def_id": "cave_gloves",
      "item_template_id": "cave_gloves",
      "display_name": "Rare Cave Gloves",
      "rarity": "rare",
      "slot": "gloves",
      "rolled_stats": { "armor": 4 },
      "sell_price": 38
    }
  ]
}
```

Fixed consumables may use the existing `item_def_id` plus optional display summary fields rather
than inventing a fake slot. Direct item stats compare only matching stat keys, and missing equipped
items compare against zero with no `equipped_item_instance_id`.

## 6. Test And Bot Proof

- `make validate-shared` validates protocol examples and appraisal fixture shape.
- Go shop tests prove sell appraisals, fixed item summaries, generated item summaries, comparison
  deltas, and equipped-item exclusion.
- `make bot scenario=vendor_appraisal_quotes` proves protocol-visible offer/appraisal metadata and
  that buy/sell still mutate gold/inventory as v41 did.
- `make client-unit` covers `ShopPanel` row summaries and comparison debug state.
- `HEADLESS=1 make bot-client scenario=16_vendor_item_comparison.json` proves the real client opens
  the shop and exposes visible item detail rows.
- `make ci` is the final gate.

## 7. Open Questions And Risks

| # | Question / risk | Default |
|---|-----------------|---------|
| Q-1 | Protocol version bump vs extending v4 current files. | Extend v4 unless validators require v5. |
| Q-2 | Direct stat deltas vs derived character stat preview. | Direct stat deltas only. |
| R-1 | Long generated item names and comparison text can overlap buttons. | Use wrapped/clipped labels in fixed-height rows and expose client tests/debug state. |
| R-2 | Comparison must not duplicate authority in the client. | Server emits comparison/appraisal data; client renders it only. |
