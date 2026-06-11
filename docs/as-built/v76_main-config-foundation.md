# v76 As-built — Main Config Foundation

Date: 2026-06-11

## What Shipped

- Added `shared/rules/main_config.v0.json` and schema with top-level gameplay tuning values for base attack interval, movement speed, and base dungeon monster drop rate.
- Loaded the new config into the Go `Rules` view as `Rules.MainConfig`.
- Added loader validation for invalid attack interval, movement speed, and drop-rate bounds.
- Added shared validation drift guards that keep the new config mirrored with the current combat, navigation, and dungeon monster drop defaults until later slices consume it directly.

## Proof

- `make validate-shared` passed.
- `cd server && go test ./internal/game -run 'TestLoadRules'` passed.
- `make ci` passed.

## Deferred

- v77 will make combat and movement consumers derive from main config.
- v78 will replace repeated dungeon monster drop-rate weights with reusable config-driven profiles.
- Class, skill, monster, and broader content unification remains outside the MVP.
