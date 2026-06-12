# v107 As-Built: Survival Reactive Unique Effects

## Shipped

- Added server-authoritative mechanics for `veil_of_the_last_oath`, `frostglass_ward`, `mirrorsteel_skin`, and `ashen_reprisal`.
- Routed monster melee, monster projectile, retaliation, and boss active damage through shared survival unique-effect hooks.
- Reused existing `skill_effect_started`, `skill_effect_ended`, `skill_cooldown_started`, `skill_cooldown_update`, and `effect_ids` protocol surfaces.
- Added `tools/bot/scenarios/55_survival_reactive_unique_effects.json`, proving `ashen_reprisal` over the live protocol with seed `853`.

## Verification

- `cd server && go test ./internal/game/... -run 'TestSurvivalUnique|TestOffensiveUnique|TestUniqueBurn'`
- `make validate-shared`
- `ARPG_BOT_SCENARIO=survival_reactive_unique_effects VERBOSE=1 make bot`

## Notes

- `veil_of_the_last_oath` exposes its 60 second cooldown through the existing skill cooldown list so the current hotbar cooldown renderer can display it without a schema change.
- `frostglass_ward` uses the existing skill-effect stat pipeline for armor and monster movement-speed slows.
- No corpse-looting files or contracts were touched.
