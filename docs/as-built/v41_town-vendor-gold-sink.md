# v41 — Town vendor gold sink

**Proves:** Town can host an authoritative vendor that spends and returns durable character gold,
while generated stock keys off the player's deepest reached dungeon depth.

- Protocol v4 adds shop buy/sell intents, shop events, private actor-scoped offer payloads, and
  `deepest_dungeon_depth` in character progression snapshots/deltas.
- Shared shop rules define the `town_vendor` with fixed red/blue potion offers plus five
  deterministic generated offers sourced from common dungeon-mob treasure classes.
- The vendor exists in town and opens through normal `action_intent`; opening the shop is
  display-only and does not mutate inventory or gold.
- Buying validates range, town state, offer id, gold, and capacity; success subtracts gold,
  adds the purchased item, persists progression/inventory, and emits `shop_purchase`.
- Selling validates owned unequipped inventory items; success removes the item, clears hotbar refs
  if needed, adds sell gold, persists, and emits `shop_sale`.
- Deepest dungeon depth persists on the character when a player reaches a deeper negative floor,
  then survives fresh sessions and replay verification.
- Godot adds a shop panel for fixed/generated offers, buy buttons, sell rows, gold/status text, and
  client-bot debug/actions while reusing the existing protocol send path.
- Protocol bot scenario `29_town_vendor_gold_sink.json` proves depth tracking, shop open, fixed and
  generated offers, buy/sell gold deltas, inventory changes, reconnect, replay, and fresh-session
  persistence.
- Client bot scenario `15_town_vendor_shop_panel.json` proves the rendered panel, offer counts,
  sell row, fixed-potion purchase, and visible inventory update in headless Godot.

**Explicit non-goals:** no vendor stock refresh timers, multiple vendors, buyback tab, item
comparison UI, stash, player trade, crafting, or final economy balance.
