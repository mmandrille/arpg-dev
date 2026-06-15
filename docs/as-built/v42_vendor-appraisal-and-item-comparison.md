# v42 — Vendor appraisal and item comparison

**Proves:** The vendor can show decision-ready, server-authored item summaries, sell appraisals,
and direct item-stat comparisons without moving economy authority into the Godot client.

- Protocol v4 `shop_opened` payloads now include display-ready offer metadata: slot/category,
  summary lines, requirements/effects, and direct comparison deltas where equipment stats apply.
- Shared `shop_appraisals` golden fixtures pin one fixed consumable, a generated equipment offer,
  an unequipped sell appraisal, and equipped-item exclusion.
- The Go sim computes stable item summary lines, generated-offer comparison data, and actor-scoped
  sell appraisals for unequipped sellable inventory items.
- Equipped items stay excluded from sell appraisals and still reject `shop_sell_intent` with the
  existing equipped-item guard.
- Godot renders richer buy/sell rows with item identity, price, kind/slot, stat/effect summaries,
  and comparison lines while continuing to send unchanged v41 buy/sell intents.
- Protocol bot scenario `30_vendor_appraisal_quotes.json` proves offer details, sell appraisals,
  comparison deltas, buy/sell mutations, `/state`, reconnect, and replay.
- Client bot scenario `16_vendor_item_comparison.json` proves the real shop panel exposes fixed
  offer summaries, generated offer comparisons, sell appraisal rows, buy, and sell in headless
  Godot.

**Explicit non-goals:** no derived character-stat preview after hypothetical equip, buyback,
stash, repair, crafting, search, sorting, filters, bulk operations, external shop/inventory UI
