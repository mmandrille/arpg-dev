# v178 Plan — Boss Summoned Adds

Status: Ready for implementation
Goal: Add a deterministic Cave Warden summon pattern that spawns ordinary server-owned add monsters.
Architecture: Shared boss pattern rules define `summon_wolves` as a telegraphed active summon phase.
The Go boss phase runtime executes summon phases exactly once, places adds deterministically near
the boss, emits normal entity-add changes, and emits a `boss_summoned_adds` event with additive
metadata. Existing Godot entity rendering and boss phase presentation remain the client surface.
Tech stack: shared JSON rules/schemas, Go sim/tests, Python protocol bot scenario, existing Godot
client bot regression, SDD docs.

## Baseline and Shortcut Decision

Builds on v177 boss ranged pattern and the extracted `server/internal/game/boss_patterns.go`
runtime. Godot plugin adoption decision: reject/not applicable; this slice has no new client UI,
art, camera, or inventory presentation and reuses existing monster rendering.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_patterns.v0.json` | Add `summon_wolves` telegraph/active/recovery pattern |
| Modify | `shared/rules/boss_templates.v0.json` | Add `summon_wolves` to Cave Warden deck after `stone_lance` |
| Modify | `shared/rules/boss_patterns.v0.schema.json` | Allow summon metadata on active phases |
| Modify | `shared/golden/boss_pattern_timeline.json` | Update charged-melee timeline after early-deck pacing changes |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow `monster_def_id` on summon events |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Keep event schema parity for replay/example validation |
| Modify | `server/internal/game/rules.go` | Load summon metadata and delegate boss pattern validation |
| Add | `server/internal/game/boss_pattern_rules.go` | Focused boss pattern validation, including summon metadata |
| Modify | `server/internal/game/types.go` | Add event-level `monster_def_id` field if needed |
| Modify | `server/internal/game/sim.go` | Add per-phase execution flag |
| Modify | `server/internal/game/boss_patterns.go` | Execute summon active phases and emit add entities/events |
| Modify | `server/internal/game/game_test.go` | Update deterministic boss deck coverage |
| Add | `server/internal/game/boss_summon_pattern_test.go` | Focused summon validation and spawn coverage |
| Modify | `tools/bot/scenarios/24_boss_floor_gate.json` | Assert summon phase, event, and live wolf add count |
| Modify | `tools/bot/scenarios/client/28_boss_phase_readability.json` | Update charged-melee telegraph duration expectation |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `rules.go` baseline after extraction |
| Add | `docs/as-built/v178_boss-summoned-adds.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `client/scripts/bot_scenario_runner.gd`
- [x] `server/internal/game/game_test.go`
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/rules.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected beyond `game_test.go` and `rules.go`
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [ ] Defer extraction with rationale: `<only if needed after implementation review>`

Notes: new summon-specific tests go in `boss_summon_pattern_test.go`; boss pattern validation moved
from `rules.go` into `boss_pattern_rules.go`, lowering the `rules.go` baseline from 3262 to 3222.
`sim.go` only gained the per-phase execution flag and remains within baseline allowance.

Verification:
```bash
make maintainability
```

## Task 1 — Shared and Protocol Shape Data

Files:
- Modify: `shared/rules/boss_patterns.v0.json`
- Modify: `shared/rules/boss_templates.v0.json`
- Modify: `shared/rules/boss_patterns.v0.schema.json`
- Modify: `shared/golden/boss_pattern_timeline.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`

- [x] Step 1.1: Add `summon_wolves` with telegraph, active summon count/radius/monster id, recovery, and cooldown.
- [x] Step 1.2: Add `summon_wolves` after `stone_lance` in `cave_warden.pattern_deck`.
- [x] Step 1.3: Allow summon metadata in the boss pattern schema.
- [x] Step 1.4: Add event-level `monster_def_id` schema support for `boss_summoned_adds`.
- [x] Step 1.5: Tighten early boss pattern timings and update the charged-melee timeline golden so the scenario remains under budget.
```bash
make validate-shared
```

## Task 2 — Server Summon Runtime

Files:
- Modify: `server/internal/game/rules.go`
- Add: `server/internal/game/boss_pattern_rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/boss_patterns.go`
- Modify: `server/internal/game/game_test.go`
- Add: `server/internal/game/boss_summon_pattern_test.go`

- [x] Step 2.1: Add summon fields to `BossPatternPhase` and validate active summon phases reference a known monster with positive count/radius.
- [x] Step 2.2: Track per-phase active execution so summon phases fire exactly once.
- [x] Step 2.3: Place summoned adds deterministically near the boss, avoiding blocked positions where current placement helpers allow.
- [x] Step 2.4: Spawn add entities as normal non-boss monsters using common generated stats and `no_drop`.
- [x] Step 2.5: Emit `boss_summoned_adds` with boss id, pattern id, phase index/kind, monster def id, amount, and boss position.
- [x] Step 2.6: Add tests for validation, exact-once add spawning, entity-add changes, event metadata, and deck order.
```bash
cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|SummonedAdds|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1
```

## Task 3 — Bot and Client Regression Proof

Files:
- Modify: `tools/bot/scenarios/24_boss_floor_gate.json`
- Modify: `tools/bot/scenarios/client/28_boss_phase_readability.json`

- [x] Step 3.1: Add bounded waits for `summon_wolves` telegraph and `boss_summoned_adds`.
- [x] Step 3.2: Assert at least the configured number of live `dungeon_wolf` add entities appear.
- [x] Step 3.3: Run the protocol boss-floor gate.
- [x] Step 3.4: Run the existing client boss phase readability scenario to prove current presentation still handles the expanded deck.
```bash
make bot scenario=24_boss_floor_gate.json
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v178_boss-summoned-adds.md`
- Modify: `docs/plans/v178_2026-06-15-boss-summoned-adds.md`
- Modify: `docs/specs/v178_spec-boss-summoned-adds.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark plan tasks complete and write as-built notes.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle and next-slice pointer.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|SummonedAdds|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- [x] `make bot scenario=24_boss_floor_gate.json`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] `make maintainability`
- [x] `make ci`
