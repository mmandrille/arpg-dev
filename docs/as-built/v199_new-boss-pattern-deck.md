# v199 As-built - New Boss Pattern Deck

Date: 2026-06-15
Status: Complete

## What Shipped

- Added the Cave Warden `shard_fan` boss pattern in shared boss rules with a cone telegraph,
  active cone damage, recovery, cooldown, range, width-in-degrees, and data-driven damage.
- Inserted `shard_fan` into the deterministic Cave Warden deck after `summon_wolves`, before
  `ground_slam`.
- Tightened `summon_wolves` recovery/cooldown pacing in shared rules so the expanded boss-floor
  proof observes summon pressure and the new cone pattern within the protocol bot budget.
- Extended boss pattern validation so cone phases require positive width and active damage phases
  must match the telegraphed radius/width predicate.
- Extended server-owned boss hit detection with locked-aim cone predicates that account for range,
  angular width, and player radius.
- Updated the boss-floor protocol proof to observe `shard_fan` before killing the boss.

## Proof

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestBoss(PatternDeckCycles|ShardFan|StoneLance|SummonedAdds|GroundSlam|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make ci`

## Follow-up Notes

- The client still renders this through the existing generic boss telegraph/readability path. A
  production cone decal, VFX, animation, or audio pass remains a separate presentation slice.
