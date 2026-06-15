# v178 As-built — Boss Summoned Adds

Date: 2026-06-15
Status: Complete

## What Shipped

- Added the Cave Warden `summon_wolves` pattern in shared boss rules with a telegraph, active
  summon phase, recovery, cooldown, summoned monster id, count, and spawn radius.
- Added `summon_wolves` to the deterministic Cave Warden deck after `stone_lance`, before
  `ground_slam`.
- Extended boss pattern validation for summon metadata and added `monster_def_id` to v8 event
  schemas so `boss_summoned_adds` can identify the summoned add type.
- Implemented exact-once active summon execution in `boss_patterns.go`, spawning normal non-boss
  `dungeon_wolf` adds near the boss as `entity_spawn` changes with common generated stats and
  `no_drop`.
- Tightened early boss pattern pacing and updated the charged-melee timeline golden so
  `boss_floor_gate` proves `stone_lance` and `summon_wolves` while staying under its protocol bot
  budget.
- Extracted boss pattern validation into `server/internal/game/boss_pattern_rules.go`, lowering the
  `rules.go` maintainability baseline from 3262 to 3222 lines.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|SummonedAdds|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make maintainability`
- `make ci`

## Follow-up Notes

- Summoned adds have no special despawn or post-boss cleanup yet. That keeps this slice limited to
  server-owned spawn pressure and leaves lifecycle tuning for a later boss-polish slice.
