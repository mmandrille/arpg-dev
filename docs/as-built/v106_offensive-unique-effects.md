# v106 As-Built: Offensive Unique Effects

Date: 2026-06-12

## What Shipped

- Made `stormbound_echo` live for basic attacks from equipped unique-effect items. It uses seeded
  proc rolls, sorted nearby target selection, lightning damage, and existing `monster_damaged`
  events with `skill_id: "stormbound_echo"`.
- Made `executioners_mark` live for low-health damaged monsters. It starts and expires
  deterministically, marks the target through `effect_ids`, and pulses nearby monsters when the
  marked monster dies before expiration.
- Made `hunger_of_the_deep` live as same-target stacking damage. Stacks update after successful
  hero damage, apply before the next hit, reset on target change, and expire deterministically.
- Added persistent sim/player state for offensive marks and hunger stacks.
- Added protocol bot proof for `stormbound_echo` and adjusted the existing burn scenario seed so
  v105 continues to prove `everburning_wound` after the larger unique-effect catalog.

## Key Decisions

- No protocol schema bump was needed; existing combat/effect event fields cover all three effects.
- No client presentation work shipped in this slice. Dedicated lightning, mark, and hunger visuals
  remain deferred to a presentation slice.
- The bot proof covers one offensive effect through the full WebSocket path; Go tests own the full
  deterministic matrix for all three offensive effects.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestUniqueBurn|TestOffensiveUnique|TestUniqueEffect'`
- `ARPG_BOT_SCENARIO=offensive_unique_effects VERBOSE=1 make bot`
- `ARPG_BOT_SCENARIO=unique_burn_effect_live VERBOSE=1 make bot`
- `make maintainability`
- `make ci`

## Deferred

- Defensive/reactive unique effects: `veil_of_the_last_oath`, `frostglass_ward`,
  `mirrorsteel_skin`, and `ashen_reprisal`.
- Resource/support/mobility unique effects: `grave_pact`, `blood_price`, `pilgrims_momentum`, and
  `lantern_of_the_fallen`.
- Client-specific presentation for offensive unique effects.
