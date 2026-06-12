# v94 As-built: Item Upgrade Starter

## What shipped

- Added `item_upgrade_cost_gold` and `item_upgrade_max_level` to `main_config.v0.json`.
- Added an authenticated account-stash upgrade route:
  `POST /v0/account-stash/items/{stash_item_id}/upgrade`.
- The store owns the upgrade transaction: lock stash gold, lock the target stash item, spend gold,
  increment `item_level`, and increase the first deterministic numeric rolled stat by 1.
- Upgraded items remain ordinary account-stash items and can still be listed/traded by existing
  market flows.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/store -run TestAccountStashItemUpgrade -count=1`
- `cd server && go test ./internal/http -run TestAccountStashItemUpgrade -count=1`
- `make test-go`
- `make ci`

## Deferred

Advanced resource costs, failure chances, recipe tiers, random affix addition, blacksmith UI/NPC,
upgrade audit history, and market restrictions for upgraded items remain deferred.
