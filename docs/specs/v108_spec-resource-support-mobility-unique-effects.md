# v108 Spec: Resource Support Mobility Unique Effects

## Goal

Make the remaining resource, support, and mobility unique effects from `shared/rules/unique_effects.v0.json` work in the authoritative server simulation:

- `grave_pact`
- `blood_price`
- `pilgrims_momentum`
- `lantern_of_the_fallen`

## Requirements

- All tuning values must be read from `shared/rules/unique_effects.v0.json`.
- `grave_pact` heals the killing hero for a catalog percent of max HP when the hero was below the configured HP threshold at kill time.
- `blood_price` lets an equipped hero pay missing skill mana with HP while preserving the configured minimum remaining HP.
- `pilgrims_momentum` tracks continuous player movement, charges after the configured duration, and consumes the charge on the next attack for bonus damage and brief knockback.
- `lantern_of_the_fallen` heals the lowest-health nearby connected hero when a nearby enemy dies.
- Effects must reuse existing events where practical (`player_healed`, `skill_effect_started`, `monster_damaged`, entity updates) without protocol schema changes.

## Non-Goals

- No new bespoke client visual assets.
- No changes to corpse looting.
- No new item roll odds or rarity tuning.

## Bot Scenario

Add a protocol bot scenario that proves at least one v108 effect over the live protocol.
