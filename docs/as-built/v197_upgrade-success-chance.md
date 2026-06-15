# v197 As-built: Upgrade Success Chance

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `item_upgrade_success_chance_percent` to `shared/rules/main_config.v0.json` with schema
  validation from 0 to 100.
- Loaded and validated the success chance in server rules.
- Extended atomic store upgrade calls with deterministic chance/roll inputs.
- Upgrade attempts now always spend the configured cost, but failed rolls return `success=false`
  and leave rolled stats unchanged.
- HTTP upgrade responses now include `success` for both stash and inventory upgrade flows.
- The blacksmith panel now shows success chance in its preview/debug state and displays failed
  attempts as normal blacksmith status.
- Added client bot matcher support for `success_chance_percent`.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/store ./internal/http -run 'Upgrade' -count=1`
- `make client-unit`
- `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- `make ci`

## Deferred

- Upgrade-resource consumption, per-rarity chance curves, and pity counters remain future
  itemization/economy work.
