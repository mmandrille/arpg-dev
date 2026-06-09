# ADR-0011: Player Market and Multi-Item Trade Offers

- **Status:** Proposed
- **Date:** 2026-06-09
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, player-market, trading, stash, item-ownership, economy

---

## Context

The current game has character-owned items, rolled equipment, durable gold, and a town vendor gold
sink. A future player market needs stronger ownership guarantees than the current vendor flow
because items can be offered by multiple real players asynchronously.

The same project constraints apply:

- The Go server remains authoritative for listings, offers, ownership transfer, expiration, and
  persistence.
- Godot renders market state and sends intents; it does not decide trade validity.
- Item ownership transitions must be atomic and auditable.
- Expired, canceled, rejected, and accepted offers must release or transfer reserved items
  predictably.

This ADR records the future player market direction only. It does not define mercenaries or item
upgrade mechanics.

---

## Future Direction

Players should be able to publish an item to a real player market. Other players can submit offers
containing one or more items. The seller has 24 hours to accept one offer; otherwise the listing is
delisted/expired.

Intended behavior:

- A listing publishes one owned item from a player inventory or stash.
- Other players can offer multiple items in exchange for the listed item.
- Offered items are reserved/locked while the offer is active so they cannot be spent, traded, sold,
  upgraded, equipped if that would break ownership guarantees, or deleted out from under the offer.
- The seller can accept one offer within 24 hours.
- On acceptance, ownership transfers atomically: the listed item goes to the buyer, the offered
  items go to the seller.
- Received items are delivered to the receiving player's town stash.
- If the listing expires or is canceled, reserved items unlock and the listing is delisted.

---

## Open Design Questions

- Whether listings support gold, resources, or only item-for-items at first.
- Whether offers expire independently or only with the listing.
- How to present item comparison and equipment requirements in market UI.
- Whether accepted trades create an audit/event record visible to both players.
- Whether the seller's listed item is removed from playable inventory immediately or only locked.
- Whether upgraded, bound, equipped, hotbar-assigned, or mercenary-exported items can be listed.
- Whether stash delivery requires a town visit, a notification inbox, or immediate stash insertion.

---

## Non-Goals For Current Slices

This ADR does not implement market routes, schemas, UI, persistence tables, listing expiration jobs,
or item-locking mechanics. It records future product direction so future specs can align on intent
before choosing contracts.

---

## Consequences

- A town stash is likely required before player-facing market delivery can be complete.
- Market implementation needs an item-lock/reservation model that works across inventory, stash,
  equipment, vendor sale, upgrading, and any future trade system.
- Trading creates high-value ownership transitions, so it needs atomic server persistence,
  auditability, expiration handling, and bot coverage before being player-facing.
