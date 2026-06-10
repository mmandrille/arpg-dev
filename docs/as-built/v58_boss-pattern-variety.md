# v58 As Built: Boss Pattern Variety

Date: 2026-06-10
Spec: [`docs/specs/v58_spec-boss-pattern-variety.md`](../specs/v58_spec-boss-pattern-variety.md)
Plan: [`docs/plans/v58_2026-06-10-boss-pattern-variety.md`](../plans/v58_2026-06-10-boss-pattern-variety.md)

## What shipped

- Cave Warden now has a two-entry pattern deck: `charged_melee` followed by the new `ground_slam`.
- `ground_slam` is fully data-driven in `shared/rules/boss_patterns.v0.json`: circle telegraph,
  matching circle active hit shape, 35-tick windup, 5-tick active damage, 24-tick recovery, and
  50-tick cooldown.
- The Go sim now cycles boss pattern decks deterministically in declared order after each completed
  pattern and cooldown.
- The authoritative boss hit predicate supports `circle` phases in addition to `melee_contact`.
- The protocol bot runtime now stores full event rows and can assert `event_seen` with payload
  filters such as `pattern_id` and `phase_kind`.
- `24_boss_floor_gate.json` proves both `charged_melee` and `ground_slam` phase starts during the
  first boss-floor flow.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|GroundSlamCircleHit|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make ci`

## Deferred

- Weighted or random pattern selection.
- Ranged boss patterns, projectiles, adds, enrage phases, and additional boss templates.
- Production shape-specific telegraph decals, boss art, VFX, audio, and animation clips.
