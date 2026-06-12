# v107 Plan: Survival Reactive Unique Effects

## Scope

Implement server-authoritative mechanics for `veil_of_the_last_oath`, `frostglass_ward`, `mirrorsteel_skin`, and `ashen_reprisal`.

## Tasks

- [x] Add survival/reactive unique-effect state to player/session cloning and save/load paths.
- [x] Route monster melee, monster projectile, retaliation, and boss active damage through a common incoming-player-damage hook.
- [x] Implement `veil_of_the_last_oath` lethal prevention, cloak `effect_ids`, status events, and hotbar-visible cooldown using existing `skill_cooldowns`.
- [x] Implement `frostglass_ward` large-hit detection, monster slow, player armor buff, status events, and cooldown.
- [x] Implement `mirrorsteel_skin` projectile damage reduction, reflection damage, and cooldown.
- [x] Implement `ashen_reprisal` block/evade priming and next-hit bonus fire damage plus burn.
- [x] Add focused Go tests for all four effects and their cooldown/expiry behavior.
- [x] Add one protocol bot scenario covering a v107 effect.
- [x] Run targeted tests and finish with `make ci`.

## Verification

- `cd server && go test ./internal/game/... -run 'TestSurvivalUnique|TestOffensiveUnique|TestUniqueBurn'`
- `make validate-shared`
- `ARPG_BOT_SCENARIO=survival_reactive_unique_effects VERBOSE=1 make bot`
- `make ci`

## Coordination Notes

- Do not touch corpse looting files or contracts.
- No new branch.
- No new Godot plugin adoption; reuse current status/cooldown presentation.
