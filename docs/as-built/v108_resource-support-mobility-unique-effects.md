# v108 As-Built: Resource Support Mobility Unique Effects

## Shipped

- Added server-authoritative mechanics for `grave_pact`, `blood_price`, `pilgrims_momentum`, and `lantern_of_the_fallen`.
- Added per-player Pilgrim's Momentum state that charges from continuous movement and is consumed by the next successful hero hit.
- Added Blood Price as a skill-cast resource fallback before `not_enough_mana` rejection.
- Extended monster-kill unique hooks with Grave Pact and Lantern healing.
- Added `unique_momentum_lab` and `tools/bot/scenarios/56_resource_support_mobility_unique_effects.json`, proving Pilgrim's Momentum over the live protocol with seed `1469`.

## Verification

- `cd server && go test ./internal/game/... -run 'TestResourceUnique|TestSurvivalUnique|TestOffensiveUnique|TestUniqueBurn'`
- `make validate-shared`
- `ARPG_BOT_SCENARIO=resource_support_mobility_unique_effects VERBOSE=1 make bot`

## Notes

- The momentum bot uses a dedicated long corridor lab so the movement charge threshold is reached without brittle timing.
- No corpse-looting files or contracts were touched.
