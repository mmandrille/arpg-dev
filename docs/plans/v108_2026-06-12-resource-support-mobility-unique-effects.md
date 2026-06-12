# v108 Plan: Resource Support Mobility Unique Effects

## Scope

Implement server-authoritative mechanics for `grave_pact`, `blood_price`, `pilgrims_momentum`, and `lantern_of_the_fallen`.

## Tasks

- [x] Add resource/support/mobility unique-effect state to player/session cloning and save/load paths.
- [x] Track continuous player movement and expose a charged/expiry state for `pilgrims_momentum`.
- [x] Apply `pilgrims_momentum` bonus damage before the next attack and knock the target back after a successful hit.
- [x] Add `blood_price` resource fallback to skill casting before rejecting for missing mana.
- [x] Extend the monster-kill unique hook with `grave_pact` and `lantern_of_the_fallen` healing.
- [x] Add focused Go tests for all four effects.
- [x] Add one protocol bot scenario covering a v108 effect.
- [x] Run targeted tests and finish with `make ci`.

## Verification

- `cd server && go test ./internal/game/... -run 'TestResourceUnique|TestSurvivalUnique|TestOffensiveUnique|TestUniqueBurn'`
- `make validate-shared`
- `ARPG_BOT_SCENARIO=resource_support_mobility_unique_effects VERBOSE=1 make bot`
- `make ci`

## Coordination Notes

- Do not touch corpse looting files or contracts.
- No new branch.
- No new Godot plugin adoption; reuse current protocol/status presentation.
