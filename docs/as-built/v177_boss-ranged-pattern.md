# v177 As-built — Boss Ranged Pattern

Date: 2026-06-15
Status: Complete

## What Shipped

- Added the Cave Warden `stone_lance` pattern in shared boss rules: line telegraph, short active
  damage window, recovery, cooldown, range, width, and data-driven damage.
- Inserted `stone_lance` into the deterministic Cave Warden deck after `charged_melee`, before
  `ground_slam`, so the protocol boss-floor scenario observes it within its runtime budget.
- Extended boss pattern validation and v8 protocol schemas with additive line `width` metadata.
- Added server-owned line aim capture at telegraph start and authoritative active-phase hit
  detection using forward projection, range, line width, and player radius.
- Extracted boss phase runtime from `sim.go` into `server/internal/game/boss_patterns.go`, lowering
  the `sim.go` maintainability baseline from 6836 to 6636 lines.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make maintainability`
- `make ci`

## Follow-up Notes

- The client still uses the existing generic boss telegraph presentation. A richer line decal or
  bespoke `stone_lance` visual remains a presentation slice, now backed by server-authored line
  metadata.
