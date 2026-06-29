# v369 As-built — Wayfarer's Accord Set

## What shipped

- Added `wayfarers_accord`, a third enabled five-piece set in `shared/rules/set_items.v0.json` using
  only head, chest, gloves, boots, and amulet slots so every class can wear the full set with its
  class weapon equipped.
- Piece fixed stats spread primary attributes and resource pools; partial bonuses grant all four
  primary stats, then max HP/mana, skill damage, and a full-set package of `all_skills`, cooldown
  reduction, and health regen.
- Registered all five pieces in `elite_objective_special_tc_1`; the debug unique chest picks them up
  through existing rule-derived catalog wiring.
- Added `TestWayfarersAccordSetPayloadsAndBonuses` and extended bot scenario
  `84_wayfarers_accord_set.json`.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestWayfarersAccord|TestUniqueTestChest|TestSetItem' -count=1`
- `make bot scenario=84_wayfarers_accord_set.json`
- `make client-unit`

## Deferred

- Boss-specific Wayfarer's drop rotation, mystery-seller weight tuning, production set art, and
  additional universal set packages remain future itemization work.
