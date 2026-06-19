# v283 As-Built - Boss Summon Bats and HP

Date: 2026-06-19
Spec: [`docs/specs/v283_spec-boss-summon-bats-and-hp.md`](../specs/v283_spec-boss-summon-bats-and-hp.md)
Plan: [`docs/plans/v283_2026-06-19-boss-summon-bats-and-hp.md`](../plans/v283_2026-06-19-boss-summon-bats-and-hp.md)

## Shipped

- Added `summon_bats`, a Cave Warden boss pattern with telegraph, active summon, and recovery
  phases.
- Added `summon_bats` to the Cave Warden pattern deck after `crystal_wall`.
- Doubled Cave Warden `hp_multiplier` from `8.0` to `16.0`.
- Extended boss summon tests so both `summon_wolves` and `summon_bats` prove their active summon
  metadata, spawn the configured add count once, and emit `boss_summoned_adds`.
- Extended dungeon boss population coverage to derive expected boss HP from the base monster max HP
  and the active boss template multiplier.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/...
```

All focused checks above passed on 2026-06-19.

Post-loop `$refactor` paid down the ratchet debt and the selected batch passed full `make ci` on
2026-06-19.

Manual visual verification command:

```bash
make bot-visual scenario=24_boss_floor_gate
```

## Boundaries

- No new summon AI behavior; summoned bats use the existing `dungeon_bat` definition and dive attack
  behavior.
- No boss damage multiplier, loot table, boss model, or boss-floor progression changed.
