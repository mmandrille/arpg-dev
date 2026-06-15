# v177 Plan — Boss Ranged Pattern

Status: Ready for implementation
Goal: Add a deterministic Cave Warden ranged line pattern with server-owned hit detection.
Architecture: Shared rules define `stone_lance` as a line telegraph/active pattern with range and width. The Go sim captures a locked aim vector when the telegraph starts, carries it into the active phase, and applies an authoritative line hit predicate. Existing boss phase events remain the wire surface, with additive shape metadata in v8 schemas.
Tech stack: shared JSON rules/schemas, Go sim/tests, Python protocol bot scenario, existing Godot client bot regression, SDD docs.

## Baseline and Shortcut Decision

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_patterns.v0.json` | Add `stone_lance` line pattern |
| Modify | `shared/rules/boss_templates.v0.json` | Add `stone_lance` to Cave Warden deck after `charged_melee` |
| Modify | `shared/rules/boss_patterns.v0.schema.json` | Allow line width metadata |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow additive line width/range in boss phase views |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow additive line width/range in boss phase events |
| Modify | `server/internal/game/rules.go` | Load/validate line width and matching active predicate |
| Modify | `server/internal/game/types.go` | Expose optional line width/range in boss telegraph/hit views |
| Modify | `server/internal/game/sim.go` | Store boss phase aim state and keep the coordinator slim |
| Add | `server/internal/game/boss_patterns.go` | Capture line aim, advance boss phases, and apply line hit predicate |
| Modify | `server/internal/game/game_test.go` | Deck coverage |
| Add | `server/internal/game/boss_ranged_pattern_test.go` | Validation and line-hit coverage |
| Modify | `tools/bot/scenarios/24_boss_floor_gate.json` | Assert `stone_lance` phase event appears |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `sim.go` baseline after boss phase extraction |
| Add | `docs/as-built/v177_boss-ranged-pattern.md` | As-built summary |
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
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [ ] Defer extraction with rationale: `<only if needed after implementation review>`

Notes: boss phase runtime moved from `sim.go` into `server/internal/game/boss_patterns.go`, and
the `sim.go` file-size baseline dropped from 6836 to 6636 lines. `rules.go` and `game_test.go`
remain within the ratchet allowance.

Verification:
```bash
make maintainability
```

## Task 1 — Shared and Protocol Shape Data

Files:
- Modify: `shared/rules/boss_patterns.v0.json`
- Modify: `shared/rules/boss_templates.v0.json`
- Modify: `shared/rules/boss_patterns.v0.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`

- [x] Step 1.1: Add `stone_lance` with line telegraph, active damage, range, width, recovery, and cooldown.
- [x] Step 1.2: Add `stone_lance` after `charged_melee` in `cave_warden.pattern_deck`.
- [x] Step 1.3: Allow and validate line width/range metadata in rule and v8 protocol schemas.
```bash
make validate-shared
```

## Task 2 — Server Line Pattern Runtime

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add `Width` to boss pattern phases and boss telegraph/hit views.
- [x] Step 2.2: Validate active line phases match prior telegraph shape, range, and width.
- [x] Step 2.3: Capture a locked aim vector at telegraph start and preserve it into active/recovery cleanup.
- [x] Step 2.4: Implement line hit detection using forward projection within range and lateral distance within half-width plus player radius.
- [x] Step 2.5: Add tests for deck order including `stone_lance`, inside-line damage, outside-width miss, and beyond-range miss.
```bash
cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1
```

## Task 3 — Bot and Client Regression Proof

Files:
- Modify: `tools/bot/scenarios/24_boss_floor_gate.json`

- [x] Step 3.1: Add an `event_seen` assertion for `boss_phase_started` with `pattern_id: stone_lance` and `phase_kind: telegraph`.
- [x] Step 3.2: Run the protocol boss-floor gate.
- [x] Step 3.3: Run the existing client boss phase readability scenario to prove current presentation still handles the expanded deck.
```bash
make bot scenario=24_boss_floor_gate.json
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v177_boss-ranged-pattern.md`
- Modify: `docs/plans/v177_2026-06-15-boss-ranged-pattern.md`
- Modify: `docs/specs/v177_spec-boss-ranged-pattern.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark plan tasks complete and write as-built notes.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle and next-slice pointer.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- [x] `make bot scenario=24_boss_floor_gate.json`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] `make maintainability`
- [x] `make ci`
