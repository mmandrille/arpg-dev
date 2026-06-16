# v203 As-built: Upgrade Pity Counter

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `item_upgrade_pity_failure_threshold` to shared `main_config`, currently set to two failed
  accepted attempts.
- Persisted item-owned pity progress in rolled item metadata as `upgrade_pity.failures`.
- Failed accepted upgrade attempts now spend gold/resources, keep upgrade stats unchanged, and
  increment the item pity failure count.
- When the pre-attempt failure count meets the configured threshold, the accepted attempt succeeds
  even if the success roll would otherwise fail.
- Successful upgrades reset `upgrade_pity.failures` to zero.
- Added focused store coverage for fail/fail/guaranteed-success behavior.
- Updated blacksmith preview/debug state to show pity progress and guaranteed-next-upgrade state.
- Added a focused blacksmith panel unit test and bot matcher/scenario assertions for default pity
  fields.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/store -run 'Upgrade' -count=1`
- `cd server && go test ./internal/http -run 'Upgrade' -count=1`
- `make client-unit`
- `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- `make ci`

## Deferred

- Per-rarity pity curves, account-wide pity, resource refunds, recipe redesign, market restrictions,
  and richer upgrade UI styling remain future itemization/economy work.
