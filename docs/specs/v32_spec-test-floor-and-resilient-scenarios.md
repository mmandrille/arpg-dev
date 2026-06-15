# Spec: `test-floor-and-resilient-scenarios`

Status: Draft
Branch: `main`
Slice: v32 - resilient test floor before more gameplay
Baseline: v31 `combat-stat-effects-and-feedback`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - deterministic authoritative sim, replay, shared rules
- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md) - generated dungeon floors and stair placement
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - dungeon monster population and combat
- [`v30_spec-monster-rarity-and-loot-scaling.md`](v30_spec-monster-rarity-and-loot-scaling.md) - generated monster rarity and loot scaling
- [`v31_spec-combat-stat-effects-and-feedback.md`](v31_spec-combat-stat-effects-and-feedback.md) - combat outcome metadata and recent balance-sensitive fixtures

## 1. Purpose

The project now has enough generated worlds, dungeon population, monster movement, combat stats,
loot tables, and client-bot scenarios that normal tuning can create test churn. Changing dungeon
floor size, generated monster count, movement speed, damage tuning, or loot weighting should not
break unrelated tests.

This slice creates a clear testing floor before more gameplay work:

- Document which values are intentional contracts and which values are tunable implementation
  details.
- Refactor brittle Go tests, Python bot scenarios, Godot smoke/golden checks, and shared validation
  where they over-specify balance-sensitive values.
- Keep deterministic replay and protocol/schema coverage strict.
- Keep golden fixtures only where exact output is the feature being protected.
- Prefer semantic selectors, derived expectations, ranges, and eventual outcomes for gameplay
  behavior.
- Make future tuning changes localized to rules, specs, and intentionally pinned fixtures.

After v32, changing dungeon dimensions, population counts, movement speed, or first-pass balance
numbers should only break tests that intentionally lock those values.

## 2. Non-goals

- No new gameplay mechanics.
- No balance pass.
- No protocol schema changes unless an existing brittle assertion requires clearer metadata to test
  behavior correctly.
- No replay determinism relaxation. Same seed, rules, and ordered inputs must still reproduce the
  same authoritative result.
- No removal of cross-language golden coverage for formula contracts.
- No broad test framework migration.

## 3. Files to create or modify

```text
docs/specs/v32_spec-test-floor-and-resilient-scenarios.md       - this slice contract
docs/plans/v32_<YYYY-MM-DD>-test-floor-and-resilient-scenarios.md - implementation plan
CLAUDE.md                                                       - durable testing policy for future agents
PROGRESS.md                                                - lifecycle update when v32 ships
server/internal/game/*_test.go                                  - convert brittle gameplay assertions where needed
server/internal/replay/*_test.go                                - preserve replay strictness, remove incidental tuning locks
tools/bot/run.py                                                - semantic/range/eventual assertion helpers
tools/bot/scenarios/*.json                                      - replace fragile counts, coordinates, and timing assumptions
tools/bot/scenarios/client/*.json                               - replace fragile client-bot assertions where possible
client/tests/test_golden.gd                                     - keep exact formula locks, avoid unrelated tuning locks
tools/validate_shared.py                                        - keep rules/schema validation strict; classify golden drift guards
shared/golden/*.json                                            - keep, rename, narrow, or regenerate only intentional contract fixtures
```

## 4. Test locking policy

### 4.1 Contract locks

Exact assertions are required for stable contracts:

- Protocol schema shape, required fields, enum values, and rejection reasons.
- Replay determinism for the same seed, rules, session-start snapshot, and ordered inputs.
- Persistence boundaries: character inventory, equipment, hotbar, progression, waypoints, and
  session-start snapshots.
- Formula fixtures where the formula is the feature: combat formulas, stat derivation, reach math,
  loot roll semantics, and deterministic RNG stream ownership.
- Cross-language evaluator parity between Go and GDScript.

Contract-lock tests may assert exact numbers, IDs, event order, and full snapshots when the test name
and fixture make that ownership explicit.

### 4.2 Behavior proofs

Gameplay tests should assert observable behavior without pinning unrelated tuning:

- Entity existence by semantic selector: type, definition ID, rarity, state, level, or inventory
  relationship.
- Reachability instead of exact generated coordinates.
- Valid bounds, no overlap, and legal placement instead of a full generated layout unless generation
  is the feature.
- `at_least`, `at_most`, `between`, `within_tolerance`, and `eventually` instead of exact counts or
  ticks when counts/timing are balance data.
- Derived expectations from current shared rules instead of duplicated constants.
- Event outcomes and state transitions instead of incidental path length or movement tick count.

Examples:

```text
Prefer: "the current level has a down stair reachable from spawn"
Avoid:  "the down stair is at x=27, z=-8" unless a named golden owns that fixture.

Prefer: "kill one dungeon_mob and observe XP/loot behavior"
Avoid:  "the floor has exactly 8 dungeon_mob entities" unless population tuning is under test.

Prefer: "monster moves closer within a derived timeout"
Avoid:  "monster position equals (4, 0, 2) after exactly 3 ticks" unless movement stepping is the feature.
```

### 4.3 Tuning details

These values are tunable during active development and should not be exact locks in unrelated tests:

- Dungeon floor dimensions.
- Generated monster population.
- Generated chest or shrine chance.
- Movement speed and pathing tick budgets.
- Damage ranges and combat probabilities outside formula-specific fixtures.
- Loot weights, rarity rates, and item distribution outside loot-specific fixtures.
- UI text placement or presentation timing outside focused client UI tests.

If a test needs one of these values, it should derive it from shared rules or mark the fixture as a
deliberate golden contract.

## 5. Audit and classification

The implementation plan must audit existing tests and scenarios before refactoring. Each brittle
assertion should be classified as:

| Class | Meaning | Action |
|-------|---------|--------|
| Contract lock | Exact value is part of architecture or feature contract | Keep exact; rename or document if unclear |
| Behavior proof | Test should prove a player/system outcome | Convert to selector/range/eventual assertion |
| Tuning detail | Test duplicates balance or generated layout details | Remove, derive from rules, or move to named golden |

The audit should prioritize:

1. Dungeon generation and stair fixtures.
2. Dungeon monster population and rarity scenarios.
3. Monster chase/leash movement tests.
4. Bot scenarios that use exact counts, coordinates, or tick counts.
5. Godot golden/smoke checks that read generated rules.
6. Replay tests that compare more of a snapshot than needed for the behavior under test.

## 6. Bot scenario assertion model

The Python protocol bot should grow or standardize assertion helpers that make semantic tests easy:

```text
assert_entity_count(entity_type, filters..., equals|at_least|at_most|between)
assert_entity_exists(filters...)
assert_entity_reachable(filters..., from=player|spawn, timeout_s?)
assert_player_within(entity_selector|position, tolerance)
assert_eventually_event(event_type, filters..., timeout_s)
assert_monster_moved_closer(monster_selector, target=player, timeout_s derived from rules)
assert_current_level(level)
assert_rules_value(path, comparator, expected?) only for explicit rules validation
```

Scenario JSON should prefer selectors over fixed entity IDs. Fixed IDs are acceptable only for
static hand-authored worlds or fixtures that explicitly document ID stability.

## 7. Golden fixture policy

Golden fixtures remain valuable, but they must be clearly scoped:

- Keep cross-language formula fixtures.
- Keep replay fixtures where exact state equality is the subject.
- Keep generation fixtures only when proving deterministic generation, RNG stream isolation, or
  rules/schema drift.
- Do not use a broad golden fixture as a convenience snapshot for unrelated gameplay.
- When a golden includes tunable values, name the fixture and test so future agents know changing it
  is a conscious contract update.

The plan should identify any golden files that are currently doing too much and split or narrow them
where practical.

## 8. Acceptance criteria

1. `CLAUDE.md` documents the test locking policy for future agents.
2. Existing brittle tests and scenarios are audited and classified in the v32 plan or an as-built
   section of `PROGRESS.md`.
3. Dungeon size tuning does not break unrelated bot scenarios or Go tests.
4. Dungeon population tuning does not break unrelated bot scenarios or Go tests.
5. Movement speed tuning does not break unrelated bot scenarios or Go tests.
6. Replay determinism remains strict for identical seed, rules, session-start snapshot, and ordered
   inputs.
7. Protocol schema and shared rules validation remain strict.
8. Golden fixtures are kept only as intentional locks and are named/documented accordingly.
9. `make ci` passes.

## 9. Testing plan

1. Run focused Go tests while refactoring brittle server assertions:

```bash
make test-go
```

2. Run protocol bot scenarios after scenario helper changes:

```bash
make bot
```

3. Run Godot client smoke/unit coverage if client scenario or golden assertions change:

```bash
make client-unit
make client-smoke
```

4. Run shared validation after touching rules, schemas, or golden fixtures:

```bash
make validate-shared
```

5. Final gate:

```bash
make ci
```

## 10. Open questions

| # | Question | Status |
|---|----------|--------|
| Q-1 | Should v32 include temporary tuning-change probes, such as changing dungeon size or speed locally and reverting, to prove resilience? | Proposed: yes in plan verification, but do not commit tuning changes. |
| Q-2 | Should broad generation golden fixtures be split into layout-contract and gameplay-behavior fixtures? | Proposed: split only if an existing fixture causes churn or unclear ownership. |
| Q-3 | Should bot scenarios support explicit `tuning_sensitive: true` metadata for intentional exact locks? | Proposed: optional; use test/scenario names first unless metadata improves clarity. |
