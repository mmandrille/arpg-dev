# v105 Spec: Unique Burn Effect Live

## Goal

Make the first unique effect mechanically live: an equipped item with `everburning_wound`
causes successful hero damage to apply a burning damage-over-time status to the target and shows
a clear burning cue on the Godot client.

## Player-facing behavior

- Any equipped unique item carrying `effect_ids: ["everburning_wound"]` makes all successful hero
  damage apply burn to the damaged monster.
- Burn ticks once per second for 10 seconds.
- Each tick deals 10% of the original hit damage, rounded deterministically, with a minimum of 1
  only when the original hit dealt positive damage.
- Reapplying burn replaces the target's active `everburning_wound` burn with a new duration and
  tick value based on the new original hit.
- Burn damage uses damage type `fire`.
- The target visibly appears burning while the status is active and returns to its normal status
  presentation when the burn expires or the target dies.

## Non-goals

- No new unique effects beyond making `everburning_wound` live.
- No new art packs or Godot addons.
- No item tooltip rewrite; v103/v104 already surface effect ids and summary lines enough for this
  first live proof.
- No balancing pass on unique drop frequency.

## Data and authority

- The Go sim remains authoritative for unique-effect activation, burn duration, tick cadence,
  damage, kill credit, and events.
- Values come from `shared/rules/unique_effects.v0.json`.
- `fire` becomes a supported damage type so future monster resistances can tune burn without
  treating it as force damage.
- The client is presentation-only and reacts to authoritative `skill_effect_started`,
  `monster_damaged`, and `skill_effect_ended` events.

## Acceptance

- Go tests prove an equipped `everburning_wound` item applies burn from a basic attack and ticks
  10% of the original hit once per second for 10 seconds.
- Go tests prove burn emits `fire` damage and can kill/credit/drop through the existing monster
  kill path.
- Protocol bot coverage proves a unique-equipped hero can burn a combat-lab monster.
- Client bot or unit coverage proves the burning visual cue is attached and removed from the
  target presentation.
- `make validate-shared`, focused Go tests, focused client checks, and final `make ci` pass.
