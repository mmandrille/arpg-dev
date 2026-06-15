# Spec: `town-vendor-gold-sink`

Status: Draft
Branch: `main`
Slice: v41 - town vendor, deterministic shop offers, and spendable gold
Baseline: v40 `reachable-dungeon-obstacles`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - town hub, character progression, deepest level reached
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) - durable character inventory, equipment, and waypoints
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item payloads
- [`v25_spec-treasure-classes-and-guarded-chests.md`](v25_spec-treasure-classes-and-guarded-chests.md) - treasure class rolls
- [`v29_spec-dungeon-equipment-drop-expansion.md`](v29_spec-dungeon-equipment-drop-expansion.md) - depth-banded generated equipment drops
- [`v39_spec-ui-currency-and-mana-polish.md`](v39_spec-ui-currency-and-mana-polish.md) - durable gold and mana UI

## 1. Purpose

Gold is now durable character state and drops from dungeon play, but it has no player-facing use.
This slice adds the first town gold sink: a single vendor in level `0` that sells consumables plus
five deterministic rolled equipment offers tied to the character's deepest achieved dungeon depth.

The server remains authoritative for all economy decisions. The client opens a display-only shop
panel, renders the offer list received from the server, and sends buy/sell intents. The Go sim
validates price, range, inventory capacity, item ownership, equipped state, and gold balance before
mutating character-owned inventory and wallet state.

This slice also persists `deepest_dungeon_depth` on the character. That field is used by the vendor
catalog and closes part of ADR-0008's character progression baseline without introducing quests,
stash, or a broader NPC system.

The thin vertical proof:

- Town has one reachable vendor interactable.
- Actioning the vendor opens a server-generated shop offer list.
- The fixed catalog sells `red_potion` and `blue_potion`.
- The generated catalog contains five rolled equipment offers from the common dungeon mob loot
  table for the character's deepest achieved dungeon depth.
- Prices are deterministic and derived from rarity plus rolled/base stats.
- Buying spends gold and adds an item to the acting character's inventory.
- Selling an unequipped sellable item removes it and credits gold.
- Reconnect, replay, fresh sessions, and co-op actor scoping preserve the economy result.

## 2. Non-goals

- No stash, crafting, repair, gambling, vendors by type, trade, or player-to-player economy.
- No full NPC dialog protocol, quest system, reputation, or vendor affinity.
- No random daily vendor rotation, server clock dependency, or durable vendor inventory rows.
- No generated shop offers from champion/rare/unique monster rarity. v41 uses common enemy drops
  only, at the character's deepest achieved depth.
- No item comparison UI, sorting/filtering, search, buyback, bulk buy/sell, shift-click shortcuts,
  or final vendor UX.
- No production vendor art, portrait, audio, voice, shop SFX, or town art pass.
- No broader item-level/depth economy rebalance.
- No Magic Find, affix grammar, unique/set catalog, or special-effect execution.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v41_spec-town-vendor-gold-sink.md - this slice contract
docs/plans/v41_<YYYY-MM-DD>-town-vendor-gold-sink.md - implementation plan
PROGRESS.md - lifecycle update when v41 ships

server/migrations/0010_character_deepest_dungeon_depth.sql - durable depth field and session-start snapshot field
server/internal/store/models.go - deepest depth on character progression and session-start progression
server/internal/store/interfaces.go - progression repo surface if needed
server/internal/store/repos.go - load/upsert/session-start snapshot persistence for deepest depth
server/internal/store/store_test.go - migration and persistence coverage

shared/rules/shops.v0.json - vendor catalog, generated-offer source, and pricing constants
shared/rules/shops.v0.schema.json - shop rule schema
shared/rules/interactables.v0.json - town vendor interactable definition
shared/rules/interactables.v0.schema.json - shop-capable interactable fields
shared/rules/worlds.v0.json - add vendor to `dungeon_levels` town level
shared/rules/worlds.v0.schema.json - preset validation if shop interactables need stricter fields
shared/protocol/envelope.v4.schema.json - protocol version bump if needed by validator convention
shared/protocol/messages.v4.schema.json - `shop_buy_intent` and `shop_sell_intent`
shared/protocol/session_snapshot.v4.schema.json - `deepest_dungeon_depth` in character progression
shared/protocol/state_delta.v4.schema.json - shop events and progression shape
shared/protocol/examples/session_snapshot.json - deepest depth and vendor example
shared/protocol/examples/state_delta.json - shop open/buy/sell examples
shared/golden/shop_pricing.json - deterministic price formula fixture
shared/golden/shop_pricing.v0.schema.json - pricing fixture schema
shared/golden/shop_offers.json - deterministic generated offer fixture
shared/golden/shop_offers.v0.schema.json - offer fixture schema
tools/validate_shared.py - shop rule validation and pricing/offer golden drift checks

server/internal/game/rules.go - parse/validate shop rules and pricing constants
server/internal/game/types.go - shop offer/event/progression protocol views
server/internal/game/sim.go - deepest depth update, vendor open, buy, and sell handlers
server/internal/game/game_test.go - unit tests for pricing, offers, buy/sell, depth, replay shape
server/internal/realtime/* - persist actor-scoped gold/inventory/progression changes as needed
server/internal/replay/* - replay reconstruction with deepest depth and shop transactions
server/internal/http/*_test.go - session start/fresh-session persistence if store path is touched

client/scripts/net_client.gd - send shop buy/sell intents
client/scripts/main.gd - vendor action event handling, panel lifecycle, debug state
client/scripts/shop_panel.gd - display-only shop UI and sell controls
client/tests/* - focused shop panel/debug tests if helper extraction is useful

tools/bot/run.py - shop intent helpers and assertions
tools/bot/scenarios/29_town_vendor_gold_sink.json - protocol bot proof
tools/bot/scenarios/client/15_town_vendor_shop_panel.json - client shop-panel proof if reliable
```

Protocol note: v41 adds new client intents and new event payload fields, so the spec expects a
coordinated protocol v4 update. The implementation plan may decide whether old v0-v3 schema files
remain as historical fixtures or whether examples point only to v4.

## 4. Data shapes

### 4.1 Character progression: deepest dungeon depth

Persist a non-negative integer `deepest_dungeon_depth` with character progression:

```json
{
  "character_progression": {
    "level": 1,
    "experience": 0,
    "unspent_stat_points": 0,
    "gold": 125,
    "deepest_dungeon_depth": 3,
    "base_stats": { "str": 5, "dex": 5, "vit": 5, "magic": 5 },
    "derived_stats": {}
  }
}
```

Rules:

- `0` means the character has not achieved a dungeon floor yet.
- Entering dungeon level `-N` sets `deepest_dungeon_depth = max(current, N)` for the acting
  character.
- Town level `0` never decreases the value.
- This value is part of live snapshots, deltas, `/state`, session-start snapshots, replay
  reconstruction, and fresh-session persistence.
- Existing characters migrate with `deepest_dungeon_depth = 0`.

### 4.2 Shop rules

New file: `shared/rules/shops.v0.json`.

Preferred shape:

```json
{
  "version": 0,
  "shops": {
    "town_vendor": {
      "name": "Town Vendor",
      "fixed_offers": [
        { "offer_id": "red_potion", "item_def_id": "red_potion", "buy_price": 20 },
        { "offer_id": "blue_potion", "item_def_id": "blue_potion", "buy_price": 20 }
      ],
      "generated_offers": {
        "offer_count": 5,
        "source": "common_dungeon_mob",
        "min_depth": 1,
        "max_roll_attempts": 64
      },
      "pricing": {
        "sell_multiplier": 0.25,
        "round_buy_to": 5,
        "rarity_multipliers": {
          "common": 1.0,
          "magic": 1.6,
          "rare": 2.4,
          "unique": 4.0
        },
        "slot_base": {
          "main_hand": 30,
          "off_hand": 25,
          "head": 20,
          "chest": 35,
          "gloves": 18,
          "belt": 22,
          "boots": 18,
          "ring": 24,
          "amulet": 28
        },
        "stat_weights": {
          "damage_min": 10,
          "damage_max": 12,
          "armor": 7,
          "block_percent": 4,
          "max_hp": 5,
          "hotbar_slots": 20,
          "inventory_rows": 35
        }
      }
    }
  }
}
```

Validation must reject:

- duplicate offer ids,
- unknown `item_def_id` references,
- fixed shop offers for `gold` or quest items,
- non-positive prices,
- generated offer counts below `1`,
- missing rarity multipliers for every current item-template rarity,
- missing stat weights for current template `base_stats` / `rollable_stats`,
- non-positive or inconsistent pricing constants.

The plan may move formula constants under a top-level `pricing_profiles` section if reuse is useful,
but the v41 catalog needs only one shop.

### 4.3 Shop-capable interactable

Add a shop-capable vendor definition:

```json
{
  "interactables": {
    "town_vendor": {
      "name": "Town Vendor",
      "initial_state": "ready",
      "shop_id": "town_vendor"
    }
  }
}
```

`town_vendor` is not a blocker and has no transition. It uses existing `action_intent` reach and
auto-approach behavior. Activating it does not mutate world state; it returns a server-authored
offer list for the acting player.

### 4.4 Shop offer view

The server sends offers in a `shop_opened` event. Fixed and generated offers use one shape:

```json
{
  "event_type": "shop_opened",
  "entity_id": "1012",
  "shop_id": "town_vendor",
  "offers": [
    {
      "offer_id": "fixed:red_potion",
      "kind": "fixed",
      "item_def_id": "red_potion",
      "display_name": "Red Potion",
      "buy_price": 20
    },
    {
      "offer_id": "generated:depth3:000",
      "kind": "generated",
      "item_def_id": "cave_blade",
      "item_template_id": "cave_blade",
      "display_name": "Magic Cave Blade",
      "rarity": "magic",
      "rolled_stats": { "damage_max": 3, "max_hp": 2 },
      "requirements": { "level": 1 },
      "effect_ids": [],
      "buy_price": 95
    }
  ]
}
```

Notes:

- `offer_id` is stable for a given `shop_id`, character id, session seed, deepest depth, and offer
  index.
- Generated offers are not durable inventory items until purchased.
- The server may regenerate offers from deterministic inputs for buy validation instead of storing
  them in mutable sim state.
- The client must not locally reroll or reprice offers.

### 4.5 Shop buy intent

New message type: `shop_buy_intent`.

```json
{
  "type": "shop_buy_intent",
  "payload": {
    "shop_entity_id": "1012",
    "offer_id": "generated:depth3:000"
  }
}
```

Validation:

- actor player exists and is alive,
- shop entity exists on the actor's current level,
- shop entity references a known `shop_id`,
- actor is in action range or auto-approach resolves before dispatch,
- offer id exists in the current deterministic catalog,
- actor has enough gold,
- actor has inventory capacity for the purchased item,
- fixed offer is purchasable and generated offer has valid rolled payload.

Success emits:

- `inventory_add`,
- `gold_update`,
- `character_progression_update`,
- `shop_purchase` event with `shop_id`, `offer_id`, `item_instance_id`, `price`, and `total_gold`,
- `intent_accepted`.

Rejections use clear reasons such as `invalid_shop`, `invalid_offer`, `out_of_range`,
`insufficient_gold`, or `inventory_full`.

### 4.6 Shop sell intent

New message type: `shop_sell_intent`.

```json
{
  "type": "shop_sell_intent",
  "payload": {
    "shop_entity_id": "1012",
    "item_instance_id": "1044"
  }
}
```

Validation:

- actor player exists and is alive,
- shop entity exists and is in range,
- item instance belongs to the acting character,
- item is in inventory, not equipped,
- item is sellable,
- item is not `gold`, not quest-category, and not otherwise marked unsellable by rules.

Success emits:

- `inventory_remove`,
- `gold_update`,
- `character_progression_update`,
- `shop_sale` event with `shop_id`, `item_instance_id`, `price`, and `total_gold`,
- `intent_accepted`.

Selling an equipped item must reject with `item_equipped`. Selling another player's item in co-op
must reject with `item_not_found` or an equivalent ownership-safe reason and must not leak the
other player's inventory details.

### 4.7 Pricing formula

v41 intentionally uses a simple first-pass price formula. It must be deterministic, data-driven
through `shops.v0.json`, and covered by golden fixtures.

For generated rolled equipment:

```text
base_score = slot_base[item.slot]
           + sum(template.base_stats[stat] * stat_weights[stat])

roll_score = sum(item.rolled_stats[stat] * stat_weights[stat])

raw_buy = (base_score + roll_score) * rarity_multipliers[item.rarity]

buy_price = ceil_to_multiple(max(1, raw_buy), round_buy_to)
sell_price = max(1, floor(buy_price * sell_multiplier))
```

For fixed consumables:

```text
buy_price = fixed_offers[].buy_price
sell_price = max(1, floor(buy_price * sell_multiplier))
```

Implementation notes:

- `item.slot` for template `ring` remains `ring` for pricing, even though equipment placement has
  `ring_left` / `ring_right`.
- `effect_ids` contribute `0` in v41 because special-effect execution is still deferred.
- Requirements do not change price in v41.
- Currency and quest items are not sellable.
- The plan may choose integer math with multiplier basis points to avoid floating drift. Golden
  fixtures own the exact rounding result.

### 4.8 Generated shop offers

Generated offers are derived from the common dungeon mob loot table for the acting character's
deepest achieved depth:

```text
depth = max(shops.town_vendor.generated_offers.min_depth, character.deepest_dungeon_depth)
loot_table = dungeon_generation.loot_band_for_depth(depth).monster_loot_table
treasure_class = loot_tables[loot_table].treasure_class_id
```

Offer generation rules:

- Use common dungeon mob loot only. Do not apply monster rarity loot-depth offsets.
- Use a labeled deterministic RNG stream, for example
  `session_seed|shop|town_vendor|character_id|depth|offers`.
- Generate exactly five sellable equipment offers.
- Skip no-drop rolls, currency drops, consumables, quest items, and non-equipment entries.
- Continue rolling until `offer_count` is reached or `max_roll_attempts` is exhausted.
- Exhausting attempts is a rules/configuration error, not a silent partial catalog.
- Each generated offer carries a rolled item payload equivalent to a loot-spawned rolled item, but
  no durable item instance id exists until purchase.

This gives the vendor depth-aware equipment without creating durable shop inventory rows.

## 5. Architecture and flow

### 5.1 Opening the vendor

```text
player clicks town_vendor
  -> client sends action_intent(target_id)
  -> server validates current-level shop interactable and reach/auto-approach
  -> server derives offer catalog for acting character
  -> server emits shop_opened event with offers
  -> client opens shop_panel.gd from the event payload
```

The event is display data. No gold or inventory mutates when the shop opens.

### 5.2 Buying

```text
shop_panel buy button
  -> client sends shop_buy_intent(shop_entity_id, offer_id)
  -> server regenerates/looks up offer
  -> server validates gold and capacity
  -> server allocates item_instance_id
  -> server deducts gold
  -> server adds item to actor inventory
  -> server emits inventory_add, gold_update, character_progression_update, shop_purchase
  -> client refreshes inventory/gold and leaves the shop panel open
```

Generated offers can be bought repeatedly in v41. Each purchase creates a new item instance with
the same rolled payload and price. Limited stock, refresh costs, and buyback are deferred.

### 5.3 Selling

```text
shop_panel sell action from inventory row
  -> client sends shop_sell_intent(shop_entity_id, item_instance_id)
  -> server validates ownership, unequipped state, sellable category, and range
  -> server computes sell price from the same pricing profile
  -> server removes item from actor inventory
  -> server credits gold
  -> server emits inventory_remove, gold_update, character_progression_update, shop_sale
  -> client refreshes inventory/gold and removes the sold row
```

The first client UI can expose selling through a simple selected-inventory list inside the shop
panel. Drag-to-sell, item comparison, and buyback are non-goals.

### 5.4 Deepest depth update

Successful level transitions update the acting character's `deepest_dungeon_depth`:

```text
descend/teleport/ascend places actor on level -N
  -> if N > current deepest_dungeon_depth
       progression.deepest_dungeon_depth = N
       emit character_progression_update
       persist with actor character
```

This update must be actor-scoped in co-op. A guest reaching level `-3` must not update the host's
progression unless the host also reaches that level.

### 5.5 Co-op actor scoping

Vendor transactions are character-owned, not session-owned.

- Inventory, gold, hotbar, and progression changes from buy/sell are delivered only to the acting
  player, following the current recipient-scoped filtering model.
- World-visible entity changes are not expected for shop transactions.
- Events may be visible to all clients only if they reveal no private inventory details. The safer
  v41 default is to deliver `shop_opened`, `shop_purchase`, and `shop_sale` only to the actor.
- Persistence must use the acting member's account/character, not the session host's.

## 6. Client presentation
### 6.2 Shop panel expectations

The panel should:

- open only from a `shop_opened` event,
- show current gold,
- show fixed offers and five generated offers with name, rarity, key stats, and buy price,
- disable or reject buy affordances when gold is insufficient after authoritative rejection,
- show sellable unequipped inventory rows with sell price,
- omit equipped, quest, and currency rows from the sell list or render them disabled,
- close on Escape / panel toggle / return to menu, following existing panel exclusivity behavior,
- expose debug state for client bot assertions.

The client may keep the visual design minimal and consistent with the current inventory/stats
panels. Production vendor art and final UX polish are deferred.

## 7. Determinism and replay

- Shop offer generation must not use wall-clock time, map iteration order, or client-local state.
- Buy/sell handlers must allocate item ids through the existing deterministic entity/item id stream.
- Price calculations must be stable across Go tests, shared validation, replay, and Godot display.
- Replay reconstruction must reproduce `shop_opened`, `shop_purchase`, `shop_sale`, inventory,
  gold, and deepest depth changes from the same seed and ordered inputs.
- Session-start snapshots must freeze the character's progression, including deepest depth, for
  replay. Later live character changes must not mutate historical replay start state.

## 8. Acceptance criteria

1. `make validate-shared` validates `shops.v0.json`, shop references, fixed offer prices,
   generated offer settings, pricing constants, and shop interactable references.
2. Existing characters migrate with `deepest_dungeon_depth = 0`, and fresh characters start at `0`.
3. Entering dungeon level `-N` persists `deepest_dungeon_depth = max(previous, N)` for the acting
   character and emits an actor-scoped `character_progression_update`.
4. `dungeon_levels` town contains one reachable `town_vendor` interactable.
5. `action_intent` on the vendor emits `shop_opened` with exactly two fixed offers and five
   generated rolled equipment offers.
6. Generated shop offers use the common dungeon mob loot table for `max(1, deepest_dungeon_depth)`
   and do not apply champion/rare/unique rarity offsets.
7. Generated offer rolls are deterministic for a pinned seed, character id, and depth.
8. `shared/golden/shop_offers.json` pins at least depth `1` and depth `3+` generated catalog cases.
9. `shared/golden/shop_pricing.json` pins fixed consumable prices, common/magic/rare rolled
   equipment prices, sell prices, and rounding behavior.
10. `shop_buy_intent` succeeds when the actor has enough gold and inventory capacity, then emits
    `inventory_add`, `gold_update`, `character_progression_update`, and `shop_purchase`.
11. `shop_buy_intent` rejects `insufficient_gold`, `inventory_full`, invalid offer ids, invalid
    shops, and out-of-range shop access without mutating inventory or gold.
12. `shop_sell_intent` succeeds for actor-owned unequipped sellable items, then emits
    `inventory_remove`, `gold_update`, `character_progression_update`, and `shop_sale`.
13. `shop_sell_intent` rejects equipped items, quest items, currency items, unknown item ids, and
    another co-op member's item without leaking private inventory details.
14. Co-op tests prove host and guest vendor transactions mutate only the acting character's wallet
    and inventory.
15. Reconnect snapshots, `/state`, replay reconstruction, and fresh-session creation preserve
    bought items, sold item removal, gold totals, and deepest depth.
16. Godot opens a display-only shop panel from `shop_opened`, sends buy/sell intents, and updates
    from authoritative deltas rather than local mutation.
17. Protocol bot scenario `29_town_vendor_gold_sink.json` proves deepest-depth offers, buy, sell,
    persistence, reconnect, `/state`, and replay.
18. Client bot scenario `15_town_vendor_shop_panel.json` proves the vendor panel opens and displays
    fixed offers, generated offers, gold, and sell rows, unless the plan documents a reliability
    reason to defer the client-bot proof.
19. `make ci` is green before v41 is marked complete.

## 9. Bot proof

Add `tools/bot/scenarios/29_town_vendor_gold_sink.json`.

Expected high-level flow:

```text
create character/session
descend to enough dungeon depth to set deepest_dungeon_depth >= 3
collect enough gold for at least one fixed and one generated offer
return to town through existing stairs/teleporter path
action town_vendor
assert shop_opened fixed offers red_potion + blue_potion
assert five generated equipment offers from common mob depth band
buy red_potion
assert gold decreased and inventory_add red_potion
buy one generated equipment offer if gold allows; otherwise use a pinned setup that guarantees enough gold
assert rolled inventory item matches selected offer payload
sell an unequipped sellable item
assert inventory_remove and gold increased by formula sell price
reconnect and assert wallet/inventory/deepest depth
verify /state
verify replay
start fresh session for same character and assert wallet/inventory/deepest depth persisted
```

The scenario should prefer semantic assertions over brittle exact item ids, except where an id is
the direct subject of the buy/sell operation.

## 10. Testing plan

1. `make validate-shared`
   - validates shop schema, pricing constants, shop/interactable refs, and golden drift.
2. `go test ./internal/game/...`
   - unit coverage for deepest-depth update, generated offers, price formula, buy/sell handlers,
     actor scoping, and replay determinism.
3. `go test ./internal/store/...`
   - migration/session-start snapshot coverage for `deepest_dungeon_depth`.
4. `go test ./internal/http/...`
   - fresh-session persistence and `/state` parity if HTTP snapshots expose the new field.
5. `make client-unit`
   - shop panel data/render helpers if added.
6. `make bot scenario=town_vendor_gold_sink`
   - protocol proof.
7. `make bot-client scenario=15_town_vendor_shop_panel.json`
   - client shop-panel proof if reliable.
8. `make ci`
   - final gate.

## 11. Open questions

| # | Question | Status |
|---|----------|--------|
| Q-1 | Buy only, or buy and sell in v41? | Resolved: buy and sell. |
| Q-2 | Initial fixed catalog? | Resolved: `red_potion`, `blue_potion`, plus five generated common-mob rolled equipment offers. |
| Q-3 | How should prices be calculated? | Resolved: first-pass formula from rarity, base stats, rolled stats, slot base, and fixed consumable prices. |
| Q-5 | What counts as "max dungeon level achieved"? | Resolved: durable `deepest_dungeon_depth`, updated when the actor enters a deeper negative level. |

## 12. Deferred follow-ups

- Vendor stock rotation, limited stock, buyback, and refresh costs.
- Separate vendor types such as blacksmith, alchemist, gambler, or stash keeper.
- Repair, crafting, enchanting, sockets, salvage, or gold sinks beyond buy/sell.
- Item comparison, filtering, sorting, keyboard shortcuts, and richer shop UX.
- Production NPC/town art and audio.
- Final item-level/depth economy and price balancing.
- Magic Find, unique/set catalogs, affixes, and special-effect pricing.
