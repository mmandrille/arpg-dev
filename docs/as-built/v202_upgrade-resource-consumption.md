# v202 As-built: Upgrade Resource Consumption

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added shared `main_config` fields for the blacksmith upgrade resource item and per-attempt count,
  currently one `upgrade_shard`.
- Validated the configured resource item through the server rules loader and shared schema checks.
- Added an `upgrade_shard` pickup to the compact `vendor_lab` blacksmith proof map.
- Made the player-facing inventory blacksmith upgrade route reject missing shards and consume the
  configured shard after an accepted attempt.
- Returned consumed resource metadata in upgrade responses so the client can reconcile local
  inventory immediately.
- Updated the blacksmith panel to show required shard availability, disable upgrades when the shard
  is missing, and report the missing-resource status.
- Extended the blacksmith client bot matcher and scenario to pick up a shard, upgrade an item, and
  observe the shard count dropping to zero.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/http -run 'Upgrade' -count=1`
- `make client-unit`
- `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- `make maintainability`
- `make ci`

## Maintenance Notes

- Refreshed file-size baselines for `client/scripts/inventory_panel.gd` and
  `client/scripts/stash_panel.gd` to match the already-committed `main` state from
  `70d5a4de fix: color unique chest item tooltips`; v202 did not modify those files.

## Deferred

- Resource wallets, stash material storage, multi-resource recipes, rarity recipe curves, binding
  rules, and richer upgrade economy loops remain future itemization/economy work.
