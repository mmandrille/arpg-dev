# v283 Spec: Boss Summon Bats and HP

Status: Complete
Date: 2026-06-19
Codename: boss-summon-bats-and-hp

## Purpose

Make the Cave Warden last longer and add another add-summon beat. The boss HP increase must live in
shared boss-template data, and the new summon phase must use the existing boss pattern system.

## Non-goals

- No new summon AI behavior; summoned adds use existing monster definitions.
- No boss damage multiplier or loot tuning.
- No client-only boss behavior.

## Asset Decision

- Adopt: existing v277 bat visual/animation and v280 bat dive behavior for summoned `dungeon_bat`
  adds.
- Borrow: existing `summon_wolves` boss pattern structure and event proof.
- Reject: new summon-specific assets or a bespoke boss scripting path.

## Acceptance Criteria

- Cave Warden `hp_multiplier` doubles from `8.0` to `16.0`.
- New `summon_bats` pattern exists with telegraph, active summon, and recovery phases.
- Cave Warden pattern deck includes `summon_bats`.
- Server tests prove boss max HP derives from base monster HP and the updated template multiplier.
- Server tests prove `summon_bats` spawns the configured `dungeon_bat` count once and emits
  `boss_summoned_adds`.
- Shared validation and game tests remain green.

## Scope and Likely Files

- Shared data: `shared/rules/boss_patterns.v0.json`, `shared/rules/boss_templates.v0.json`.
- Server tests: `server/internal/game/boss_summon_pattern_test.go`,
  `server/internal/game/dungeon_population_test.go` or focused existing boss test coverage.
- Docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, and as-built summary at finish.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/...`

Visual scenario for manual verification:

```bash
make bot-visual scenario=24_boss_floor_gate
```

## Open Questions and Risks

- No blocking questions.
- Risk: existing bot scenarios that kill the boss quickly may need more attacks. The user explicitly
  requested the 2x HP increase because bosses were too easy to kill.
