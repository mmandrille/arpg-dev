# v41 Plan - Town Vendor and Gold Sink

Status: Implemented - `make ci` green
Spec: `docs/specs/v41_spec-town-vendor-gold-sink.md`
Baseline: v40 `reachable-dungeon-obstacles` on `main`
Date: 2026-06-09

## Goal

Add a server-authoritative town vendor that gives gold a repeatable sink. The vendor must let players buy red/blue potions and five generated item offers from the deepest dungeon level the character has reached, sell inventory items back for gold, and preserve the character's deepest dungeon depth across sessions.

## Architecture Notes

- Server remains authoritative for shop catalog generation, pricing, affordability checks, inventory capacity checks, gold mutation, and persistence.
- Shop offers are deterministic from character progression and shared rules. The client only renders offers and sends selected offer/item ids.
- Protocol advances to v4 for shop intents and events. Existing protocol examples and validators should move together with the schemas.
- Deepest dungeon depth is durable progression state, updated when a character reaches a deeper negative dungeon floor.
- Co-op behavior follows v38/v39 actor-scoped patterns: only the acting player can buy, sell, and receive private shop responses.
- No external Godot shop/inventory plugin is adopted for this slice. Existing local panels and signals are enough, and a plugin would not reduce server/protocol work.

## File Map

| Area | Files |
| --- | --- |
| Shared rules | `shared/rules/shops.v0.json`, `shared/rules/shops.v0.schema.json`, `shared/rules/interactables.v0.json`, `shared/rules/interactables.v0.schema.json`, `shared/rules/worlds.v0.json`, `shared/rules/worlds.v0.schema.json` |
| Shared goldens | `shared/golden/shop_pricing.json`, `shared/golden/shop_offers.json`, matching schemas if needed |
| Protocol | `shared/protocol/envelope.v4.schema.json`, `shared/protocol/messages.v4.schema.json`, `shared/protocol/state_delta.v4.schema.json`, `shared/protocol/session_snapshot.v4.schema.json`, `shared/protocol/examples/**` |
| Validation | `tools/validate_shared.py`, `client/tests/test_golden.gd` |
| Store | `server/migrations/0010_character_deepest_dungeon_depth.sql`, `server/internal/store/models.go`, `server/internal/store/repos.go`, store tests |
| Game server | `server/internal/game/types.go`, `server/internal/game/rules.go`, `server/internal/game/sim.go`, game tests |
| Input/protocol decode | `server/internal/inputdecode/inputdecode.go`, inputdecode tests |
| HTTP/realtime/replay | `server/internal/http/**`, `server/internal/realtime/**`, `server/internal/replay/**` |
| Bot | `tools/bot/run.py`, `tools/bot/scenarios/29_town_vendor_gold_sink.json` |
| Client | `client/scripts/main.gd`, `client/scripts/net_client.gd`, `client/scripts/inventory_panel.gd`, `client/scripts/shop_panel.gd`, `client/tests/**`, `tools/bot/scenarios/client/15_town_vendor_shop_panel.json` |
| Docs | `PROGRESS.md` during finish |

## Implementation Tasks

### 1. Shared Rules, Goldens, and Protocol v4

- [x] Add `shared/rules/shops.v0.schema.json`.
  - Validate `shop_id`, static offers, generated offer count, source loot table, sell multiplier, price formula config, and deterministic ordering requirements.
  - Allow future rarity multipliers such as `unique` even if current templates only use common/magic/rare.
- [x] Add `shared/rules/shops.v0.json`.
  - `town_vendor` static offers: red potion and blue potion.
  - Generated offers: five equipment drops from the common enemy loot table matching the character's deepest dungeon depth band.
  - Pricing config: slot base, stat weights for template base/rolled stats, rarity multiplier, sell multiplier, and integer rounding.
- [x] Update town rules.
  - Add `town_vendor` with `shop_id` to `shared/rules/interactables.v0.json`.
  - Place one reachable `town_vendor` interactable in the `dungeon_levels` level `0` town world.
- [x] Add shop goldens.
  - `shop_pricing.json`: fixed potion prices plus representative common/magic/rare equipment with low/high rolls.
  - `shop_offers.json`: deterministic offer lists for deepest depths `1`, `2`, and `3+`.
- [x] Create protocol v4 schemas by extending v3.
  - Client intents: `shop_buy_intent`, `shop_sell_intent`.
  - Server events: `shop_opened`, `shop_purchase`, `shop_sale`, with failures represented by stable `intent_rejected` reasons.
  - State deltas: actor-scoped `gold_update`, `inventory_add`, `inventory_remove`, and `character_progression_update.deepest_dungeon_depth`.
- [x] Update protocol examples to v4.
  - Include a successful buy, insufficient-gold failure, successful sell, and invalid/range failure.
- [x] Update `tools/validate_shared.py` so `make validate-shared` validates v4 protocol examples and new shop rules/goldens.
- [x] Update `client/tests/test_golden.gd` to load and validate shop pricing/offers goldens where the client-side golden test coverage expects shared data parity.

Verification:

```sh
make validate-shared
```

### 2. Persist Deepest Dungeon Depth

- [x] Add migration `0010_character_deepest_dungeon_depth.sql`.
  - Add `deepest_dungeon_depth INTEGER NOT NULL DEFAULT 0` to `character_progression`.
  - Add the same field to `session_start_character_progression`.
  - Add non-negative constraints consistent with existing gold constraints.
- [x] Extend store models.
  - `CharacterProgression.DeepestDungeonDepth`.
  - `SessionStartCharacterProgression.DeepestDungeonDepth`.
- [x] Update store repository queries.
  - Select/insert/upsert deepest depth in progression methods.
  - Include deepest depth in session-start snapshot creation and load paths.
- [x] Add store tests for default `0`, persistence after update, and replay/session-start snapshot retention.
- [x] Update HTTP/session snapshot mapping so the field reaches game session construction and replay.

Verification:

```sh
cd server && go test ./internal/store/...
cd server && go test ./internal/http/...
```

### 3. Load Shop Rules and Implement Pricing/Offer Generation

- [x] Add game rule structs and validation for `shops.v0.json`.
  - Validate referenced item ids and loot tables exist.
  - Validate generated offer count is positive and stable.
  - Reject invalid pricing multipliers or negative prices.
- [x] Implement a deterministic shop pricing helper.
  - Formula from the spec:
    - `base_score = slot_base[item.slot] + sum(template.base_stats[stat] * stat_weights[stat])`
    - `roll_score = sum(item.rolled_stats[stat] * stat_weights[stat])`
    - `raw_buy = (base_score + roll_score) * rarity_multipliers[item.rarity]`
    - `buy_price = ceil_to_multiple(max(1, raw_buy), round_buy_to)`
    - `sell_price = max(1, floor(buy_price * sell_multiplier))`
  - Fixed consumables use their configured fixed price.
- [x] Implement generated offer creation.
  - Use deepest depth band `max(1, deepest_dungeon_depth)`.
  - Roll from common enemy loot table for that band.
  - Filter to equipment items only.
  - Continue deterministic rolls until five offers are produced.
  - Include stable offer ids, item payload, buy price, and source metadata.
- [x] Add Go tests against shop goldens.
  - Pricing golden parity.
  - Offer golden parity for depths `1`, `2`, and `3+`.
  - Validation failures for missing references and bad formula config.

Verification:

```sh
cd server && go test ./internal/game/... -run 'TestShop|TestRulesLoad'
make validate-shared
```

### 4. Add Shop Intents and Simulation Behavior

- [x] Extend `server/internal/inputdecode/inputdecode.go`.
  - Decode and validate `shop_buy_intent`.
  - Decode and validate `shop_sell_intent`.
  - Mark both as client intents.
  - Add tests for valid payloads, malformed payloads, and stored input replay decoding.
- [x] Extend `game.Input`.
  - Add shop buy/sell fields.
  - Include shop intents in dead-player rejection logic.
- [x] Add town vendor interactable behavior.
  - Existing `action_intent` against `town_vendor` opens the shop instead of generic interaction only.
  - Server emits actor-scoped `shop_opened` with current static and generated offers.
  - Opening the shop does not mutate gold or inventory.
- [x] Add buy handling.
  - Require the actor to be alive, in town, and within vendor action range.
  - Resolve offer id from current deterministic catalog.
  - Reject unknown offer, insufficient gold, full inventory, or non-actionable state.
  - On success, subtract gold, add item, persist through existing inventory/progression changes, and emit success event.
  - On failure, emit actor-scoped failure event with a stable reason.
- [x] Add sell handling.
  - Require alive actor, town vendor range, and owned inventory item id.
  - Reject equipped items unless the spec is explicitly updated later; v41 should sell only unequipped inventory entries.
  - Remove item, add sell price, persist inventory and gold, and emit success event.
  - Emit stable failure reasons for invalid item, equipped item, and range/actionability failures.
- [x] Update deepest dungeon depth during dungeon travel.
  - When a player reaches a negative dungeon floor deeper than their stored value, update progression.
  - Append `character_progression_update` for the actor.
  - Persist the updated field alongside gold and XP progression.
- [x] Add game tests.
  - Opening vendor returns static and generated offers.
  - Buying potion succeeds and mutates gold/inventory.
  - Buying generated item succeeds at sufficient gold.
  - Insufficient gold and full inventory fail without mutation.
  - Selling unequipped item removes it and adds gold.
  - Equipped item sell is rejected.
  - Deepest depth advances only when a deeper floor is reached.
  - Co-op shop events and mutations affect only the acting player.

Verification:

```sh
cd server && go test ./internal/inputdecode ./internal/game/...
```

### 5. Wire Realtime, HTTP, Replay, and Persistence

- [x] Update realtime input routing to pass shop intents through unchanged from protocol decode into simulation.
- [x] Ensure shop events are delivered only to the actor when they contain private catalog or failure data.
- [x] Persist buy/sell results through existing inventory and progression save paths.
  - Gold writes should remain atomic with inventory mutation at the session result boundary used by existing inventory/gold flows.
- [x] Include deepest depth in session start snapshots.
- [x] Update replay tests so shop intents replay deterministically.
  - Replay should reproduce shop offer ids, prices, inventory deltas, gold deltas, and progression depth.
- [x] Update HTTP tests for session snapshot schema/version expectations.

Verification:

```sh
cd server && go test ./internal/realtime/... ./internal/replay/... ./internal/http/...
```

### 6. Add Mandatory Server Bot Scenario

- [x] Extend `tools/bot/run.py` state tracking.
  - Track latest `shop_opened` offers by shop id and offer id.
  - Track `deepest_dungeon_depth` from progression updates/session state.
  - Reuse existing gold and inventory tracking.
- [x] Add bot actions/assertions.
  - `open_shop`.
  - `assert_shop_offer_count`.
  - `buy_shop_offer`.
  - `sell_inventory_item`.
  - `assert_gold_changed`.
  - `assert_deepest_dungeon_depth`.
  - `assert_shop_event`.
- [x] Add `tools/bot/scenarios/29_town_vendor_gold_sink.json`.
  - Enter the dungeon and reach at least depth 3.
  - Return to town.
  - Open the vendor and assert two static potion offers plus five generated offers.
  - Buy a fixed potion.
  - Buy one generated offer after collecting enough gold in the pinned route.
  - Sell one unequipped inventory item.
  - Assert gold decreases on buy, increases on sell, inventory changes, and `deepest_dungeon_depth >= 3`.
- [x] Keep scenario deterministic with a pinned seed and explicit movement/action steps.

Verification:

```sh
make bot scenario=town_vendor_gold_sink
make bot
```

### 7. Add Godot Shop UI and Client Bot Proof

- [x] Add `client/scripts/shop_panel.gd`.
  - Render shop title, player gold, static offers, generated offers, buy buttons, sellable inventory rows, and failure/status text.
  - Disable buy buttons when the client knows gold is insufficient, but rely on server rejection as authority.
  - Keep panel layout consistent with existing inventory/equipment panels.
- [x] Wire `main.gd`.
  - Open shop panel on `shop_opened`.
  - Update gold and inventory display after buy/sell deltas.
  - Close or refresh panel when the player leaves town/vendor range if existing UI patterns support it.
- [x] Wire client protocol sends through the existing `NetClient` path.
  - Send `shop_buy_intent` with selected shop entity id and `offer_id`.
  - Send `shop_sell_intent` with selected inventory item id.
- [x] Update inventory panel interactions as needed so sellable item ids can be selected without disrupting equip/unequip behavior.
- [x] Add or update client unit tests.
  - Shop panel renders offer count and prices.
  - Buy/sell signals send correct protocol payloads.
  - Failure text renders without overlapping compact layouts.
- [x] Add `tools/bot/scenarios/client/15_town_vendor_shop_panel.json`.
  - Launch client, approach vendor, open shop, verify panel visibility and offer rows, click buy, click sell, and assert visible gold/inventory update.

Verification:

```sh
make client-unit
make client-smoke
HEADLESS=1 SCENARIO=15_town_vendor_shop_panel ./scripts/bot_client_local.sh
```

### 8. Full Validation and Documentation Closeout

- [x] Run focused tests after each area, then full CI.
- [x] Update `PROGRESS.md` during finish with:
  - v41 completed summary.
  - New protocol version v4.
  - Shop/vendor behavior.
  - Deepest dungeon depth persistence.
  - Bot scenario names.
- [x] Check for accidental compatibility shims or unused v3 protocol paths. This project does not need backward compatibility unless a test or tool still depends on old fixtures.
- [x] Verify no branch was created and unrelated dirty client changes were not reverted.

Final verification:

```sh
make validate-shared
cd server && go test ./...
make client-unit
make client-smoke
make bot scenario=town_vendor_gold_sink
HEADLESS=1 SCENARIO=15_town_vendor_shop_panel ./scripts/bot_client_local.sh
make ci
```

## Acceptance Checklist

- [x] Vendor exists in town and opens through normal interact action.
- [x] `shop_opened` contains red potion, blue potion, and five generated offers.
- [x] Generated offers reflect the character's deepest achieved dungeon depth.
- [x] Buy succeeds for affordable offers and persists item/gold.
- [x] Buy fails cleanly for insufficient gold, full inventory, bad offer, or invalid state.
- [x] Sell succeeds for unequipped inventory items and persists item removal/gold.
- [x] Sell fails cleanly for missing or equipped items.
- [x] Deepest dungeon depth persists across sessions and replays.
- [x] Co-op shop mutations are actor-scoped.
- [x] Shared protocol/rules/goldens validate.
- [x] Server bot and client bot prove the complete flow.

## Deferred

- Vendor stock refresh timers or daily rotations.
- Multiple vendor NPCs or specialized item categories.
- Buyback tab.
- Item comparison UI.
- Economy balancing beyond the simple v41 formula.
- External inventory/shop plugins.
