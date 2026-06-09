# ADR-0010: Mercenaries From Player Characters

- **Status:** Proposed
- **Date:** 2026-06-09
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, mercenaries, companions, async-player-characters, combat-AI

---

## Context

The current game has character-scoped inventory/equipment, rolled items, durable progression, and
co-op session membership. Future companion systems should preserve the same design constraints:

- The Go server remains authoritative for combat, ownership, AI, death, rewards, and persistence.
- Godot renders server-owned state and sends intents; it does not decide companion behavior.
- Replay and bot scenarios need deterministic proofs for combat and death paths.
- A hired character derived from another player needs clear rules so the source player's real
  character is not accidentally damaged or deleted.

This ADR records the future mercenary direction only. It does not define market trading or item
upgrade mechanics.

---

## Future Direction

Players should be able to pay to contract a mercenary that fights at their side. A mercenary is
derived from another player's character, including that character's items and stats at the time it
is made available for hire.

Intended behavior:

- A hired mercenary follows the hiring player through gameplay spaces where companions are allowed.
- Mercenaries use server-owned AI, acquire aggro against enemies, and can attack monsters.
- Mercenaries can be damaged and killed through normal combat authority.
- When a hired mercenary dies, the hire is lost; the player must contract another mercenary.
- The source player's real character should not be deleted or damaged when their exported mercenary
  instance dies. The hired unit is a derived/leased combat copy unless a future design explicitly
  opts into higher-stakes lending.

---

## Open Design Questions

- Whether players list their characters for hire manually or the system auto-publishes eligible
  characters.
- Whether mercenary price is fixed, player-set, depth-scaled, or based on character power.
- Whether mercenaries can loot, consume potions, trigger aggro sharing, or receive XP/rewards.
- How many mercenaries can follow a player or party at once.
- Whether mercenary gear snapshots refresh automatically or only when republished.
- Whether mercenary death creates a cooldown, insurance cost, or reputation effect for the hiring
  player or source character.

---

## Non-Goals For Current Slices

This ADR does not implement companion routes, schemas, persistence tables, AI, UI, hiring prices, or
combat behavior. It records future product direction so future specs can align on intent before
choosing contracts.

---

## Consequences

- Future mercenary specs need clear AI, follow, aggro, and death semantics that do not corrupt the
  source player's actual character.
- Mercenary snapshots may require an explicit character-export or character-snapshot persistence
  model.
- Mercenary combat needs bot/replay coverage because it affects server-owned aggro, damage, death,
  and rewards.
