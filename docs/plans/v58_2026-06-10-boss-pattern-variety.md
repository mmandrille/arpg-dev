# v58 Plan - Boss Pattern Variety

Status: Ready for implementation
Goal: Add one more authoritative Cave Warden pattern and deterministic pattern-deck cycling.
Architecture: Shared JSON defines the new pattern and deck order. The Go sim remains authoritative:
it cycles patterns in declared deck order after full pattern completion/cooldown and resolves the
new circle hit predicate server-side. Existing boss phase events and the v57 client readability
layer render the new pattern without a protocol shape change.
Tech stack: shared JSON rules/validation, Go sim/tests, Python protocol bot, existing Godot client
bot regression, SDD docs.

## Baseline and shortcut decision

Baseline is v57 `boss-phase-readability` on `main`.
## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_patterns.v0.json` | Add `ground_slam` pattern |
| Modify | `shared/rules/boss_templates.v0.json` | Add `ground_slam` to Cave Warden deck |
| Modify | `server/internal/game/sim.go` | Deck cycling and circle hit predicate |
| Modify | `server/internal/game/game_test.go` | Deck-order and circle-hit coverage |
| Modify | `tools/bot/bot_types.py` | Store full event rows for payload assertions |
| Modify | `tools/bot/run.py` | Match `event_seen` assertions with optional payload fields |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for payload-aware event assertions |
| Modify | `tools/bot/scenarios/24_boss_floor_gate.json` | Assert both boss pattern ids appear |
| Add | `docs/specs/v58_spec-boss-pattern-variety.md` | Slice spec |
| Add | `docs/plans/v58_2026-06-10-boss-pattern-variety.md` | This plan |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Add | `docs/as-built/v58_boss-pattern-variety.md` | As-built proof |

## Task 1 - Shared boss rules

Files:
- Modify: `shared/rules/boss_patterns.v0.json`
- Modify: `shared/rules/boss_templates.v0.json`

- [x] Step 1.1: Add `ground_slam` with `circle` telegraph/hit shape, matching radius, damage,
  recovery, and cooldown.
- [x] Step 1.2: Add `ground_slam` after `charged_melee` in `cave_warden.pattern_deck`.
- [x] Step 1.3: Validate shared rule references and telegraph guarantee.

```bash
make validate-shared
```

## Task 2 - Server pattern cycling and circle hits

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add internal boss deck index state initialized from the template deck.
- [x] Step 2.2: Advance the deck index after a full pattern completes, preserving in-pattern phase
  sequencing and cooldown behavior.
- [x] Step 2.3: Implement the `circle` active-phase hit predicate as distance from boss center to
  player center within the configured radius.
- [x] Step 2.4: Add tests proving deck order starts `charged_melee`, advances to `ground_slam`,
  and wraps back to `charged_melee`.
- [x] Step 2.5: Add tests proving `ground_slam` damages inside radius and misses outside radius.

```bash
cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|GroundSlamCircleHit|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1
```

## Task 3 - Protocol bot event payload assertions

Files:
- Modify: `tools/bot/bot_types.py`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 3.1: Store full event dictionaries in runtime state while keeping existing
  `seen_events` behavior.
- [x] Step 3.2: Extend `event_seen` assertions so optional scalar fields such as `pattern_id`,
  `phase_kind`, and `entity_id` must match when present.
- [x] Step 3.3: Add Python unit tests for positive and negative payload-aware event assertions.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

## Task 4 - Bot and client regression proof

Files:
- Modify: `tools/bot/scenarios/24_boss_floor_gate.json`

- [x] Step 4.1: Add runtime/final assertions that observe `boss_phase_started` for both
  `charged_melee` and `ground_slam`.
- [x] Step 4.2: Run the protocol boss-floor gate.
- [x] Step 4.3: Run the v57 client phase readability scenario to prove the existing boss bar and
  telegraph marker still work with the expanded deck.

```bash
make bot scenario=24_boss_floor_gate.json
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 5 - Lifecycle docs and CI

Files:
- Modify: `docs/plans/v58_2026-06-10-boss-pattern-variety.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v58_boss-pattern-variety.md`

- [x] Step 5.1: Add v58 as-built notes.
- [x] Step 5.2: Update `PROGRESS.md` latest slice, numbering note, lifecycle row, recent summary,
  deferred backlog, and autoloop candidate status.
- [x] Step 5.3: Run full CI.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|GroundSlamCircleHit|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=24_boss_floor_gate.json`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] `make ci`

## Deferred scope

- Weighted or random pattern selection.
- Ranged boss patterns, projectiles, adds, enrage phases, and additional boss templates.
- Production shape-specific telegraph decals, boss art, VFX, audio, and animation clips.
