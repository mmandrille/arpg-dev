# v107 Spec: Survival Reactive Unique Effects

## Goal

Make the defensive and reactive unique effects from `shared/rules/unique_effects.v0.json` work in the authoritative server simulation:

- `veil_of_the_last_oath`
- `frostglass_ward`
- `mirrorsteel_skin`
- `ashen_reprisal`

## Requirements

- All tuning values must be read from `shared/rules/unique_effects.v0.json`.
- `veil_of_the_last_oath` prevents one lethal incoming hit while off cooldown, applies `last_oath_veil` to the player for 3 seconds, and exposes its 60 second cooldown through the existing skill cooldown state so the client hotbar can render downtime.
- `frostglass_ward` triggers after a large taken hit, slows nearby living monsters with `ice_slow`, grants the player a temporary armor bonus, and respects its cooldown.
- `mirrorsteel_skin` triggers on the first projectile hit taken while off cooldown, reduces the incoming projectile damage, and reflects catalog-scaled damage back to the projectile owner.
- `ashen_reprisal` primes after a block or evade, then the next hero hit within 3 seconds deals bonus fire damage and applies a short burn.
- Triggered effects must emit existing combat/status/cooldown events where practical so protocol bots and the client can observe them without a new schema version.

## Non-Goals

- No new unique item drop tables or loot odds.
- No new bespoke Godot visuals beyond existing `effect_ids`, `skill_effect_started`, and hotbar cooldown presentation.
- No changes to corpse looting.

## Bot Scenario

Add a protocol bot scenario that equips a deterministic survival unique item and observes at least one v107 effect through events/state.

## Godot Plugin Adoption

No client-side UI or asset pipeline is introduced in this slice. Existing status marker and skill cooldown presentation are reused, so the Godot plugin adoption result is **reject** for new plugin use.
