# v110 As-Built: Item Upgrade Repeat Action

**Status:** Complete on `main`

## What shipped

- Added `item_upgrade_cost_growth_per_level` to `shared/rules/main_config.v0.json`.
- The existing account-stash upgrade route now charges
  `item_upgrade_cost_gold + current_item_level * item_upgrade_cost_growth_per_level`.
- Account-stash equipment can upgrade repeatedly until `item_upgrade_max_level`.
- Store mutation still preserves unrelated rolled payload keys, increments `item_level`, and improves
  one deterministic numeric stat.
- The route returns the actual charged cost for the upgrade attempt.
- The maintainability baseline was caught up to current post-v109 file sizes with the exception
  recorded in the v110 plan.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/store -run TestAccountStashItemUpgrade -count=1`
- `cd server && go test ./internal/http -run TestAccountStashItemUpgrade -count=1`
- `make maintainability`
- `make test-go`
- `make ci`

## Deferred

Blacksmith UI/NPC, inventory/equipped upgrades, resource costs, failure chances, item bricking,
recipe tiers, upgrade audit history, and market restrictions for upgraded items remain deferred.
