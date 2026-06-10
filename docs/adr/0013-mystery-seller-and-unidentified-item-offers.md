# ADR-0013: Mystery Seller and Unidentified Item Offers

- **Status:** Proposed
- **Date:** 2026-06-09
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, vendors, unidentified-items, item-rarity, gold-sink, economy

---

## Context

The current game has character-owned gold, durable character progression, rolled equipment, vendor
buy/sell flows, item comparison, and generated vendor stock. Future vendor economy should include a
high-risk, high-price gold sink that can produce exciting equipment without exposing the exact item
before purchase.

The same project constraints apply:

- The Go server remains authoritative for stock generation, price calculation, purchase validation,
  item rolling/reveal, ownership transfer, and persistence.
- Godot renders the mystery seller state and sends intents; it does not decide hidden outcomes.
- Hidden offers must be deterministic under the same seed/input stream where they affect replay or
  bot proofs.
- Unknown items must become normal owned item instances after purchase, using the same inventory,
  equipment, appraisal, market, and upgrade constraints as other equipment.

This ADR records the future mystery-seller direction only. It does not define final unique/set item
catalogs, item upgrade mechanics, or player market behavior.

---

## Future Direction

Town should eventually include a mystery seller that offers expensive unidentified equipment. The
seller presents one hidden offer for each eligible equipment item family across weapons, gear, and
jewelry. The buyer sees only the offer category/slot and price before buying. After purchase, the
server reveals the actual rolled item as if the item had just dropped.

Intended behavior:

- The mystery seller has one offer for every eligible equipment family/template in scope: weapons,
  armor/gear slots, and jewelry.
- Each offer is unknown before purchase. The UI may show a silhouette, item family, slot, source
  level window, and price, but not the actual item name, rarity, rolled stats, requirements, or
  effects.
- Buying an offer spends gold first, then creates/reveals the actual item instance server-side and
  delivers it to the character inventory or stash.
- The seller never offers common items. The minimum rarity is magic.
- All non-common rarities are possible when their catalogs exist, including magic, rare, set, and
  unique items.
- Offer source levels are limited to the last five levels of the player's achieved progression
  ceiling: `max(1, achieved_max_level - 4)` through `achieved_max_level`, inclusive.
- Prices are intentionally expensive and should scale by source level, item family, and expected
  rarity/value so the seller functions as a meaningful gold sink rather than a cheap loot shortcut.
- Hidden offer outcomes and prices are data-driven and server-authored so future tuning can adjust
  rarity weights, set/unique eligibility, and family coverage without client gameplay logic.

---

## Open Design Questions

- Whether `achieved_max_level` means character level, deepest dungeon depth, future item level, or a
  derived progression tier.
- Whether the seller refreshes on a timer, on new achieved levels/depths, on town visits, or through
  a paid reroll.
- Whether each equipment family maps to one item template, one slot, one weapon class, or every
  eligible template in the rules catalog.
- Whether ring offers should appear once, twice, or as a generic jewelry roll despite two ring
  equipment slots.
- Whether bought mystery items go directly to inventory, overflow to stash, or require a town stash
  before the feature can ship.
- Whether unrevealed offers are pre-rolled at stock refresh time or rolled only at purchase time.
- How set and unique item eligibility is constrained by level, item family, account state, boss
  progression, or drop-source rules.
- Whether mystery-seller purchases can be refunded, resold immediately, listed on the player market,
  upgraded, or bound on reveal.
- How much information the UI should reveal before purchase so the risk feels fair without turning
  into normal vendor appraisal.

---

## Non-Goals For Current Slices

This ADR does not implement mystery seller routes, protocol schemas, UI, stock persistence, set or
unique item catalogs, stash delivery, rarity weights, pricing formulas, refresh cadence, or purchase
bot scenarios. It records future product direction so future specs can align on intent before
choosing contracts.

---

## Consequences

- Future vendor specs need a hidden-offer contract that separates visible offer metadata from the
  server-owned revealed item payload.
- Itemization needs a non-common rarity floor for this seller and eventually needs set/unique item
  catalogs before the full rarity promise is complete.
- The pricing model must account for the expected value of blind rolls or the seller will either
  trivialize loot progression or feel unusably punishing.
- The feature interacts with stash overflow, market eligibility, item upgrades, and item binding, so
  those ownership rules must be explicit before mystery items become player-facing at scale.
