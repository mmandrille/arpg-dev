# v47 — Shop stock lifecycle

**Proves:** Town-vendor generated stock is finite, per-character, refresh-gated by newly unlocked
non-town waypoints, and paired with session-local buyback without moving shop authority to the client.

- Protocol v6 adds `buyback` offer kind support, source-depth metadata, shop stock change ops, and
  refreshed `offers` / `sell_appraisals` on purchase and sale events.
- Shared shop rules cap generated shop rarity at `rare`, define buyback behavior, and make generated
  source depth roll from the character's achieved dungeon range, with `character level + 1` as the
  lower bound when that is inside the achieved range.
- Generated shop stock is stored per character, frozen into session-start snapshots for replay, and
  consumed only after successful validation; fixed potion offers remain infinite.
- Successful sales create actor-local buyback rows that preserve the sold item payload and can be
  repurchased, then disappear after buyback purchase, leaving town, or ending the session.
- Newly discovered non-town waypoints refresh generated stock exactly once; repeated shop opens or
  duplicate waypoint discoveries do not reroll stock.
- Godot keeps the shop panel open across buy/sell events, applies server-authored refreshed rows,
  exposes buyback/source-depth debug state, and reuses existing item summary/comparison rendering.
- Go lifecycle tests prove finite generated stock consumption; protocol bot scenario
  `33_shop_stock_lifecycle.json` proves source-depth metadata, buyback, waypoint refresh, replay,
  reconnect, and fresh-session behavior; client bot scenario `22_shop_stock_lifecycle.json` proves
  the live Godot panel applies purchase and buyback refreshes while staying open.

**Explicit non-goals:** no fixed potion stock limits, durable buyback across town exits or session
ends, multiple vendors, stash, repair, crafting, gambling, sorting/filtering/search, player trade,
clock-based refresh, production vendor assets, expanded item-level/depth economy bands beyond the
current `1`, `2`, and `3+` loot bands, or unique/set shop offers.
