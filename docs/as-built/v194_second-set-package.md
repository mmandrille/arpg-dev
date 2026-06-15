# v194 As-built: Second Set Package

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `Stormrunner Covenant` as the second enabled five-piece set package in
  `shared/rules/set_items.v0.json`.
- Reused existing bow, helm, gloves, boots, and ring templates with fixed set stats and level 5
  requirements.
- Added server-authoritative 2/3/4/full-set bonuses for dexterity, crit chance, attack speed,
  all skills, skill damage, and Magic Find.
- Kept debug unique chest contents rule-derived so both enabled set packages are offered
  deterministically.
- Added focused Go coverage for the new set payload, partial bonuses, full-set bonus, summary
  lines, all-skills rank flow, and Magic Find stat breakdown.
- Added protocol bot scenario `83_second_set_package.json`, proving a Stormrunner set piece can be
  taken from the unique chest.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestUniqueTestChest|TestSetItem' -count=1`
- `make bot scenario=83_second_set_package.json`
- `make ci`

## Deferred

- Random set drops, boss/elite set rewards, set collection UI, and new set art remain future
  itemization work.
