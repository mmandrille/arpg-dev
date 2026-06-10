# v51 — Mystery seller core

**Proves:** Town can host a server-owned blind-buy equipment seller without leaking item identity
before purchase.

- Shared rules add `town_mystery_seller` on `dungeon_levels` town level `0`; protocol v8 adds
  concealed `mystery` shop rows and post-purchase revealed item payloads.
- Mystery stock is deterministic, per-character, finite, durable through `character_shop_stock`,
  and replay-safe through session-start shop stock snapshots.
- The seller rolls one hidden row per supported equipment slot from the character progression
  source-depth window, requires `magic` or `rare` rarity, and prices rows through a conservative
  mystery-specific table with final economy tuning deferred.
- Concealed rows expose only safe metadata: offer id, kind, category, slot, source-depth bounds,
  mystery label, availability, and buy price. Actual template, display name, rarity, stats,
  requirements, effects, comparison, and equip preview remain hidden until purchase.
- Successful purchases validate gold, range, capacity, availability, and hidden roll integrity,
  then subtract gold, consume the row, add the revealed item, and emit actor-private purchase,
  gold, inventory, progression, and refreshed-offer payloads.
- Protocol bot scenario `37_mystery_seller_core.json` proves funding through dungeon loot vendor
  sales, hidden-row assertions, reveal-on-purchase, replay, reconnect, `/state`, and fresh-session
  consumed-stock persistence; client bot scenario `24_mystery_seller_core.json` proves the live
  Godot shop panel renders and buys concealed offers.
