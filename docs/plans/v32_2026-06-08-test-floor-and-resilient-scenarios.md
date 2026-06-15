# v32 Plan — Test Floor and Resilient Scenarios

Status: Complete (`make ci` green)
Goal: Refactor the test suite so CI locks contracts and behavior while allowing normal dungeon, population, movement, and balance tuning.
Architecture: Keep strict deterministic replay, schema validation, and intentional golden fixtures. Convert unrelated gameplay tests to semantic selectors, ranges, derived expectations, and eventual assertions. Treat this as a test-infrastructure slice only: no new gameplay, no balance pass, and no client authority changes.
Tech stack: Go tests, shared JSON/golden validation, Python protocol bot scenarios, Godot client-bot smoke/golden tests, project lifecycle docs.

## Baseline and shortcut decision

Baseline is v31 `combat-stat-effects-and-feedback`, with scenarios currently numbered through
`tools/bot/scenarios/22_combat_stat_effects.json` and client scenarios through
`tools/bot/scenarios/client/11_combat_feedback.json`.

This slice reuses the existing Go test suite, Python bot runner, shared JSON validation, and Godot
test/scenario assertion cleanup, not new UI, camera, inventory presentation, isometric tooling, or
placeholder art.

Work stays on the current branch. The implementation must preserve unrelated dirty worktree changes
and must not create a branch.

## Spec decisions

| Question | Decision for implementation |
|----------|-----------------------------|
| Q-1 temporary tuning probes | Include local, reverted verification probes for dungeon size, population, and movement speed when practical. Do not commit tuning changes. |
| Q-2 broad generation golden split | Split or narrow only fixtures that create churn or unclear ownership during the audit. Do not churn stable formula goldens just to rename them. |
| Q-3 `tuning_sensitive` metadata | Defer metadata unless the audit shows names/comments are not enough. Prefer clear test/scenario names first. |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `CLAUDE.md` | Add durable test locking policy for future agents |
| Modify | `PROGRESS.md` | Add v32 lifecycle entry and as-built audit summary when complete |
| Create/Modify | `docs/plans/v32_2026-06-08-test-floor-and-resilient-scenarios.md` | Implementation checklist and audit record |
| Modify | `tools/bot/run.py` | Add or standardize semantic/range/eventual assertion helpers |
| Modify | `tools/bot/scenarios/*.json` | Convert brittle protocol scenarios to selectors/ranges/derived assertions |
| Modify | `tools/bot/scenarios/client/*.json` | Convert brittle client-bot waits/assertions where they depend on tunable values |
| Modify | `server/internal/game/*_test.go` | Preserve contract locks; relax incidental tuning assertions |
| Modify | `server/internal/replay/*_test.go` | Preserve replay strictness; avoid unrelated layout/population locks |
| Modify | `client/tests/test_golden.gd` | Keep cross-language formula locks; avoid generated tuning locks |
| Modify | `tools/validate_shared.py` | Classify validation as schema/contract/tuning; remove hardcoded first-pass tuning locks where inappropriate |
| Modify | `shared/golden/*.json` | Narrow or regenerate only intentional locks discovered by the audit |

## Task 1 — Audit and Classification

Files:
- Modify: `docs/plans/v32_2026-06-08-test-floor-and-resilient-scenarios.md`

- [x] Step 1.1: Audit exact assertions in Go tests, replay tests, bot scenarios, client scenarios,
  `client/tests/test_golden.gd`, `tools/validate_shared.py`, and `shared/golden/*.json`.

```bash
rg -n '"equals"|"position"|"ticks"|"timeout_s"|expected_|want |DeepEqual|expected_monster_count|expected_player_position' server/internal/game server/internal/replay tools/bot client/tests shared/golden tools/validate_shared.py
```

- [x] Step 1.2: Add an "Audit Record" section near the bottom of this plan with each candidate
  grouped as **contract lock**, **behavior proof**, or **tuning detail**.
- [x] Step 1.3: Mark non-actionable contract locks explicitly so later implementation does not
  weaken replay, protocol, formula, or persistence coverage.
- [x] Step 1.4: Identify a short high-value refactor list instead of trying to rewrite every test
  style in one pass.

## Task 2 — Document the Testing Policy

Files:
- Modify: `CLAUDE.md`

- [x] Step 2.1: Add a concise "Test Locking Policy" section near the testing commands or key
  invariants.
- [x] Step 2.2: State that exact values are for protocol/schema, replay determinism, persistence
  boundaries, formula goldens, and cross-language evaluator parity.
- [x] Step 2.3: State that dungeon size, generated population, movement speed, timing budgets,
  damage tuning, loot weights, and generated coordinates should use semantic/range/derived/eventual
  assertions unless a named golden intentionally owns them.
- [x] Step 2.4: Include one short example for dungeon placement, population, and movement speed.

```bash
sed -n '1,240p' CLAUDE.md
```

## Task 3 — Bot Assertion Helpers

Files:
- Modify: `tools/bot/run.py`

- [x] Step 3.1: Inspect existing scenario assertion helpers before adding new ones; reuse current
  helper patterns and error style.
- [x] Step 3.2: Add missing comparator support where needed: `at_least`, `at_most`, `between`, and
  tolerance-based checks for existing entity/count/player assertions.
- [x] Step 3.3: Add or standardize semantic selectors for entities by type, definition ID, rarity,
  state, and level.
- [x] Step 3.4: Add eventual helpers only where scenarios currently fake timing with exact
  `wait_ticks` or hardcoded movement positions.
- [x] Step 3.5: Add focused Python unit tests if `tools/bot/test_protocol.py` or nearby test files
  already cover helper behavior; otherwise keep validation through scenario runs.

```bash
python -m py_compile tools/bot/run.py
```

## Task 4 — Protocol Bot Scenario Cleanup

Files:
- Modify: `tools/bot/scenarios/*.json`

- [x] Step 4.1: Update dungeon-level scenarios to assert stair discovery, level transition, valid
  bounds, and reachability instead of exact generated coordinates unless a named golden owns them.
- [x] Step 4.2: Update dungeon monster and rarity scenarios to assert required behavior through
  semantic counts such as `at_least` or selected rarity existence, not incidental total population.
- [x] Step 4.3: Update chase/leash scenarios to assert movement direction, leash outcome, or event
  behavior with derived/eventual timeouts instead of exact tick positions when possible.
- [x] Step 4.4: Review `16_rolled_drops.json`, `17_treasure_classes_and_guarded_chests.json`,
  `20_dungeon_equipment_drops.json`, `21_monster_rarity_loot_scaling.json`, and
  `22_combat_stat_effects.json` for brittle exact counts or entity indexes introduced by recent
  slices.
- [x] Step 4.5: Keep exact assertions when they prove persistence, equipment slot behavior, hotbar
  capacity, progression formulas, combat outcome metadata, or loot formula contracts.

```bash
make bot
```

## Task 5 — Go and Replay Test Cleanup

Files:
- Modify: `server/internal/game/*_test.go`
- Modify: `server/internal/replay/*_test.go`

- [x] Step 5.1: Keep exact `reflect.DeepEqual` snapshot checks only where replay equality is the
  subject of the test.
- [x] Step 5.2: Replace incidental generated coordinate or population assertions with rule-derived
  bounds, non-overlap, reachability, entity existence, or semantic state checks.
- [x] Step 5.3: Keep exact formula, RNG stream, event ordering, persistence snapshot, duplicate
  sequence, and schema-shaped behavior assertions.
- [x] Step 5.4: Add helper functions for generated dungeon invariants if that reduces repeated
  brittle checks.

```bash
make test-go
```

## Task 6 — Shared Validation and Golden Fixture Cleanup

Files:
- Modify: `tools/validate_shared.py`
- Modify: `shared/golden/*.json`
- Modify: `client/tests/test_golden.gd`

- [x] Step 6.1: Classify current validation checks as schema/contract/tuning. Keep schema and
  contract validation strict.
- [x] Step 6.2: Remove or relax hardcoded first-pass tuning checks that duplicate mutable rules,
  especially rarity weights/colors/offsets and generated population counts, unless the fixture name
  clearly owns that contract.
- [x] Step 6.3: Keep cross-language formula goldens exact: damage, retaliation, equipped weapon
  damage, progression formulas, combat stat effects, loot rolls, and item rolls.
- [x] Step 6.4: Narrow generation goldens only where they currently lock unrelated gameplay. Keep
  deterministic generation/RNG isolation checks where they are intentional.
- [x] Step 6.5: Update `client/tests/test_golden.gd` to match any narrowed fixture scope while
  preserving Go/GDScript parity.

```bash
make validate-shared
make client-unit
```

## Task 7 — Client Scenario Cleanup

Files:
- Modify: `tools/bot/scenarios/client/*.json`
- Modify: `client/tests/test_golden.gd`

- [x] Step 7.1: Replace client-bot exact generated positions, entity indexes, or count assumptions
  with selector-based waits and tolerance checks.
- [x] Step 7.2: Keep exact UI assertions that prove stable user-facing contracts, such as hotbar
  capacity after belt equip, settings persistence, and expected panel visibility.
- [x] Step 7.3: Preserve combat feedback assertions for event variants, but avoid relying on
  incidental monster indexes when a selector can express the target.

```bash
make client-smoke
```

## Task 8 — Tuning-Change Probe

Files:
- Temporarily modify and revert: `shared/rules/dungeon_generation.v0.json`
- Temporarily modify and revert: movement-speed source rules identified by the audit

- [x] Step 8.1: Temporarily change dungeon floor dimensions or placement bounds locally, then run
  the focused tests/scenarios that previously over-locked dungeon size.
- [x] Step 8.2: Revert the local tuning change immediately after the probe.
- [x] Step 8.3: Temporarily change generated monster population locally, then run focused dungeon
  population/ranking scenarios.
- [x] Step 8.4: Revert the local tuning change immediately after the probe.
- [x] Step 8.5: Temporarily change movement speed or a movement timing rule locally, then run
  focused movement/chase scenarios.
- [x] Step 8.6: Revert the local tuning change immediately after the probe.
- [x] Step 8.7: Record probe results in this plan's Audit Record or `PROGRESS.md` as-built
  summary. Do not commit probe tuning changes.

```bash
git diff -- shared/rules/dungeon_generation.v0.json shared/rules/monsters.v0.json shared/rules/character_progression.v0.json
```

## Task 9 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v32_2026-06-08-test-floor-and-resilient-scenarios.md`

- [x] Step 9.1: Update `PROGRESS.md` with v32 as complete only after implementation and CI
  pass.
- [x] Step 9.2: Add an as-built summary of which exact locks remain intentional and which brittle
  areas were converted.
- [x] Step 9.3: Ensure no temporary tuning probe changes remain in the worktree.
- [x] Step 9.4: Run the final gate.

```bash
make ci
```

## Final verification

- [x] `python -m py_compile tools/bot/run.py`
- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `make ci`
- [x] `git diff --check`

## Deferred scope

- No broad test framework migration.
- No gameplay, balance, protocol, or UI feature work.
- No committed tuning changes from the resilience probes.
- No `tuning_sensitive` scenario metadata unless implementation proves names/comments are
  insufficient.

## Audit Record

Task 1 audit command:

```bash
rg -n '"equals"|"position"|"ticks"|"timeout_s"|expected_|want |DeepEqual|expected_monster_count|expected_player_position' server/internal/game server/internal/replay tools/bot client/tests shared/golden tools/validate_shared.py
```

| Area | File/test/scenario | Classification | Planned action |
|------|--------------------|----------------|----------------|
| Replay determinism | `server/internal/replay/replay_test.go` `reflect.DeepEqual`, sequence/tick/snapshot checks | Contract lock | Keep exact. These are replay equality and persistence-boundary tests, not tuning locks. |
| Formula/evaluator goldens | `server/internal/game/game_test.go`, `client/tests/test_golden.gd`, `shared/golden/damage_formula.json`, `retaliation_damage.json`, `equipped_weapon_damage.json`, progression/combat-stat fixtures | Contract lock | Keep exact. Formula parity and RNG stream ownership remain intentionally pinned. |
| Protocol/persistence scenarios | `tools/bot/scenarios/15_character_persistence.json`, `19_full_equipment.json`, hotbar/equipment assertions in protocol and client scenarios | Contract lock | Keep exact where they prove inventory counts, equipped slots, hotbar capacity, and persisted state. |
| Static authored worlds | World preset entity/wall counts in `server/internal/game/game_test.go` and static client scenarios | Contract lock | Keep exact when the world is hand-authored and the test names the preset contract. |
| Dungeon stairs/teleporters | `shared/golden/dungeon_stairs.json`, `server/internal/game/game_test.go`, `server/internal/game/game_replay_test.go` | Mixed: contract lock plus tuning detail risk | Keep exact transition/discovery/replay parity. Avoid adding new unrelated gameplay expectations to the broad stair fixture. |
| Dungeon monsters | `tools/bot/scenarios/14_dungeon_monsters.json` | Behavior proof | Replace fixed `wait_ticks` movement timing with eventual movement/leash behavior where practical. |
| Chase/leash labs | `tools/bot/scenarios/09_chase_lab.json`, `11_leash_lab.json` | Behavior proof | Keep movement/aggro/leash outcomes; avoid making exact wait ticks the proof by adding an eventual helper. |
| Guarded chest bot scenario | `tools/bot/scenarios/17_treasure_classes_and_guarded_chests.json` | Tuning detail | Convert exact dungeon mob count `10` to semantic/range assertions. Keep exact one chest/open-once and persistence checks. |
| Dungeon equipment bot scenario | `tools/bot/scenarios/20_dungeon_equipment_drops.json` | Behavior proof | Keep exact depth and chest state for the scenario goal; no population lock found. |
| Rarity and rolled drop scenarios | `tools/bot/scenarios/16_rolled_drops.json`, `21_monster_rarity_loot_scaling.json` | Behavior proof | Already use rarity selectors plus `at_least`; keep and rely on expanded helper comparators if needed. |
| Combat stat feedback scenarios | `tools/bot/scenarios/22_combat_stat_effects.json`, `tools/bot/scenarios/client/11_combat_feedback.json` | Contract lock plus selector cleanup | Keep exact outcome metadata. Replace brittle client `entity_index` targets with semantic `monster_def_id` selectors if the client bot supports it or can be cheaply extended. |
| Client movement/teleporter scenarios | `tools/bot/scenarios/client/05_click_to_move.json`, `07_town_teleporter_auto_approach.json` | Behavior proof | Keep tolerance-based `wait_player_near`; prefer interactable selector over positional/index assumptions where practical. |
| Bot helper model | `tools/bot/run.py` | Behavior proof | Add common comparator support (`at_most`, `between`, optional tolerance) to entity/count assertions and standardize selector filters by type/definition/rarity/state/level. |
| Shared validation rarity first-pass checks | `tools/validate_shared.py` `expected_rarity_weights`, `expected_rarity_offsets`, `expected_rarity_colors` | Tuning detail | Relax hardcoded weights/colors/offsets to structural validation. Keep stable rarity IDs, positive multipliers, loot table reachability, and golden formula checks. |
| Guarded chest generation golden | `shared/golden/guarded_chest_generation.json`, `tools/validate_shared.py`, `server/internal/game/game_test.go` | Mixed: generation contract plus tuning detail | Keep deterministic chest/no-chest cases and valid chest fields. Derive monster count from current rules or assert bounds instead of pinning duplicated first-pass counts. |
| Pathfinding unit tests | `server/internal/game/pathfind_test.go` | Contract lock | Keep exact. These lock algorithm stepping, not balance tuning. |

### Tuning Probe Results

| Probe | Temporary change | Focused verification | Result |
|-------|------------------|----------------------|--------|
| Dungeon floor size | `shared/rules/dungeon_generation.v0.json` `floor_size` `100x50 -> 110x55` | `make bot scenario=dungeon_levels` | Passed; dynamic stair walking handled farther generated coordinates. Change reverted. |
| Generated population | `shared/rules/dungeon_generation.v0.json` `monster_placement.count` `8 -> 6` | `make bot scenario=treasure_classes_and_guarded_chests`; `cd server && go test ./internal/game -run TestGuardedChestGenerationGolden -count=1` | Passed; bot uses semantic lower bound and Go derives count from rules. Change reverted. |
| Movement speed | `shared/rules/monsters.v0.json` `training_dummy_chase` and `dungeon_mob` `move_speed` `0.7 -> 0.5` | `make bot scenario=dungeon_monsters`; `make bot scenario=chase_lab` | Passed; eventual movement assertions handled slower chase. Change reverted. |
| Probe cleanup | Rules files | `git diff -- shared/rules/dungeon_generation.v0.json shared/rules/monsters.v0.json shared/rules/character_progression.v0.json` | Clean; no temporary tuning edits remain. |
