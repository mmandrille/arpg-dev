# v199 Plan - New Boss Pattern Deck

Status: Ready for implementation
Goal: Add a deterministic Cave Warden cone pattern to broaden the existing boss deck.
Architecture: Shared boss rules define `shard_fan`; the Go sim captures aim at telegraph start and
resolves active cone hits authoritatively. Existing boss phase events carry shape/radius/width
metadata, and the client keeps using generic boss phase presentation.
Tech stack: shared JSON rules, Go sim/tests, Python protocol bot, Godot client bot regression, SDD
docs.

## Baseline and Shortcut Decision

Baseline is v198 `mercenary-foundation` on `main`. The requested work is server/shared boss combat,
not client UI, inventory presentation, or placeholder art. `docs/researchs/godot-plugins-and-shortcuts.md`
is absent in this checkout; plugin adoption is rejected for v199 because no Godot presentation
surface is changed.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_patterns.v0.json` | Add `shard_fan` cone telegraph/active/recovery pattern |
| Modify | `shared/rules/boss_templates.v0.json` | Insert `shard_fan` into the Cave Warden deck |
| Modify | `server/internal/game/boss_pattern_rules.go` | Validate cone width and active/telegraph shape parity |
| Modify | `server/internal/game/boss_patterns.go` | Capture cone aim and resolve cone hit predicates |
| Add | `server/internal/game/boss_cone_pattern_test.go` | Focused cone hit and validation coverage |
| Modify | `server/internal/game/game_test.go` | Expand deterministic boss deck coverage |
| Modify | `tools/bot/scenarios/24_boss_floor_gate.json` | Observe `shard_fan` during the boss-floor proof |
| Add | `docs/as-built/v199_new-boss-pattern-deck.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/game_test.go`
- [ ] `server/internal/game/sim.go`
- [ ] `client/scripts/main.gd`
- [ ] `tools/bot/run.py`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Add focused new test file for cone-specific assertions.
- [ ] Defer extraction with rationale: not needed.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Boss Deck Data

Files:
- Modify: `shared/rules/boss_patterns.v0.json`
- Modify: `shared/rules/boss_templates.v0.json`

- [x] Step 1.1: Add `shard_fan` with cone telegraph, active cone damage, recovery, and cooldown.
- [x] Step 1.2: Insert `shard_fan` after `summon_wolves` and before `ground_slam`.
- [x] Step 1.3: Tighten summon recovery/cooldown pacing in shared data so the expanded boss-floor
  proof remains bounded.
```bash
make validate-shared
```

## Task 2 - Server Cone Runtime

Files:
- Modify: `server/internal/game/boss_pattern_rules.go`
- Modify: `server/internal/game/boss_patterns.go`
- Add: `server/internal/game/boss_cone_pattern_test.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Require positive `width` for cone telegraphs.
- [x] Step 2.2: Reuse active/telegraph shape parity validation for cone radius and width.
- [x] Step 2.3: Capture cone aim at telegraph start.
- [x] Step 2.4: Resolve active cone hits using range, angular width in degrees, and player radius.
- [x] Step 2.5: Add focused cone hit/miss and validation tests.
- [x] Step 2.6: Extend deck-order test coverage.
```bash
cd server && go test ./internal/game -run 'TestBoss(PatternDeckCycles|ShardFan|StoneLance|SummonedAdds|GroundSlam|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1
```

## Task 3 - Bot and Client Regression Proof

Files:
- Modify: `tools/bot/scenarios/24_boss_floor_gate.json`

- [x] Step 3.1: Add bounded wait/assertions for `shard_fan` telegraph.
- [x] Step 3.2: Run the boss-floor protocol scenario.
- [x] Step 3.3: Run the existing client boss phase readability scenario.
```bash
make bot scenario=24_boss_floor_gate.json
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 4 - Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v199_new-boss-pattern-deck.md`
- Modify: `docs/plans/v199_2026-06-15-new-boss-pattern-deck.md`
- Modify: `docs/specs/v199_spec-new-boss-pattern-deck.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Write as-built notes and update `PROGRESS.md`.
- [x] Step 4.2: Run final CI.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestBoss(PatternDeckCycles|ShardFan|StoneLance|SummonedAdds|GroundSlam|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- [x] `make maintainability`
- [x] `make bot scenario=24_boss_floor_gate.json`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] `make ci`
