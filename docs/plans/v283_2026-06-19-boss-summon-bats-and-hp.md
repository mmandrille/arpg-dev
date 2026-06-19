# v283 Plan - Boss Summon Bats and HP

Status: Complete
Goal: Add a second Cave Warden summon pattern and double Cave Warden HP through shared rules data.
Architecture: Reuse the existing boss pattern summon phase machinery. Add `summon_bats`, append it
to the boss deck, and update focused tests to derive boss HP from `hp_multiplier`.
Tech stack: Shared boss rules, Go simulation/rules tests, docs.

## Baseline and shortcut decision

Builds on existing `summon_wolves` pattern support and v280 bat behavior. Adopt existing bat visuals
and server summon machinery, reject new bespoke scripting.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_patterns.v0.json` | Add `summon_bats` pattern |
| Modify | `shared/rules/boss_templates.v0.json` | Add `summon_bats` to deck and double HP multiplier |
| Modify | `server/internal/game/boss_summon_pattern_test.go` | Prove bat summon pattern and spawn event |
| Modify | `server/internal/game/dungeon_population_test.go` | Prove boss HP derives from updated multiplier |
| Modify | `docs/specs/v283_spec-boss-summon-bats-and-hp.md` | Mark complete |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row |
| Modify | `PROGRESS.md` | Update current status |
| Add | `docs/as-built/v283_boss-summon-bats-and-hp.md` | Record proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Decision:
- [x] Keep changes in focused boss test files; do not add new logic to large coordinators.

Verification:

```bash
make maintainability
```

Expected known issue: `make maintainability` is currently blocked by unrelated pre-existing ratchet
debt and will be paid down in the post-loop `$refactor`.

## Task 1 - Boss data

Files:
- Modify: `shared/rules/boss_patterns.v0.json`
- Modify: `shared/rules/boss_templates.v0.json`

- [x] Add `summon_bats` pattern using `dungeon_bat`.
- [x] Add `summon_bats` to Cave Warden deck.
- [x] Change Cave Warden `hp_multiplier` from `8.0` to `16.0`.

```bash
make validate-shared
```

## Task 2 - Boss summon and HP proof

Files:
- Modify: `server/internal/game/boss_summon_pattern_test.go`
- Modify: `server/internal/game/dungeon_population_test.go`

- [x] Prove `summon_bats` validation metadata.
- [x] Prove `summon_bats` spawns configured bats once and emits `boss_summoned_adds`.
- [x] Prove generated boss HP equals base monster max HP times template HP multiplier.

```bash
cd server && go test ./internal/game/...
```

## Task 3 - Lifecycle docs

Files:
- Modify: `docs/specs/v283_spec-boss-summon-bats-and-hp.md`
- Modify: `docs/plans/v283_2026-06-19-boss-summon-bats-and-hp.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v283_boss-summon-bats-and-hp.md`

- [x] Mark the spec and plan complete.
- [x] Record focused verification and batch-CI-pending status.
- [x] Add lifecycle/as-built notes.

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `make maintainability` deferred with known unrelated ratchet debt

Manual visual command:

```bash
make bot-visual scenario=24_boss_floor_gate
```

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
