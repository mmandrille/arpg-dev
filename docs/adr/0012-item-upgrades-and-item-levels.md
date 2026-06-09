# ADR-0012: Item Upgrades and Item Levels

- **Status:** Proposed
- **Date:** 2026-06-09
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, item-progression, crafting, affixes, dungeon-resources, economy

---

## Context

The current game has rolled equipment, item requirements, durable inventory/equipment, depth-scaled
loot, and a first pass at vendor economy. Future item progression should build on those systems
without moving mutation authority into the client.

The same project constraints apply:

- The Go server remains authoritative for resource consumption, success rolls, item mutation, and
  persistence.
- Shared rules remain data; success rates, costs, and upgrade recipes should be data-driven where
  possible.
- Upgrade results must be deterministic for replay/test proofs when driven by the same seed/input
  stream.
- Item ownership and market eligibility must remain clear after upgrades.

This ADR records the future item-upgrade direction only. It does not define mercenaries or player
market mechanics.

---

## Future Direction

Players should be able to use resources looted from advanced dungeon content to upgrade equipment.
Items gain their own level/progression, and upgrades can either add a new roll or improve a random
existing roll with a chance of success.

Intended behavior:

- Upgrade resources drop from advanced dungeon content and are persisted as owned items/resources.
- Upgrade attempts target one equipment item.
- Each attempt has a server-authored success rate.
- On success, the item either gains an extra roll or improves one random existing roll, depending on
  the upgrade recipe/rules.
- Items have their own level/progression so repeated upgrades are trackable and can gate future
  recipes, costs, or success rates.
- Upgrade results must be deterministic for replay/test proofs when driven by the same seed/input
  stream.

---

## Open Design Questions

- Whether failed attempts consume resources only, reduce durability, brick an item, or do nothing
  beyond cost consumption.
- Whether the player chooses "add roll" vs "improve roll" or the recipe determines it.
- Whether item level is a single integer, per-affix level, item XP, or recipe-tier history.
- How success chances scale with dungeon depth, item rarity, item level, and resource tier.
- Whether upgraded items become account-bound, tradeable, or market-restricted.
- Whether upgrade resources are inventory items, stash materials, wallet-like currencies, or a mix.
- Whether adding a roll can exceed normal rolled-item affix limits or uses a separate upgrade slot.

---

## Non-Goals For Current Slices

This ADR does not implement upgrade routes, schemas, UI, resource drops, success formulas,
persistence changes, or item mutation mechanics. It records future product direction so future specs
can align on intent before choosing contracts.

---

## Consequences

- Upgrade specs will need a durable item-level/progression schema and a deterministic mutation
  contract for rolled items.
- Upgrade resources will likely need advanced dungeon treasure classes, stash/material handling, and
  UI that explains success chance and possible outcomes.
- Upgrading creates high-value item mutations, so it needs atomic server persistence, auditability,
  and bot/golden coverage before being player-facing.
