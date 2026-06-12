# v95 As-built: Unique Item Catalog Seed

## What shipped

- Added `shared/rules/unique_items.v0.json` and its schema.
- Seeded `embercall_blade` as a disabled unique concept based on `cave_blade`.
- The seed records a future behavior hook and remains excluded from loot, shops, mystery seller,
  market eligibility rules, client presentation, and equip behavior.
- `tools/validate_shared.py` now validates unique seeds and cross-checks base item templates.

## Proof

- `make validate-shared`
- `make maintainability`
- `make test-go`
- `make ci`

## Deferred

Unique drops, unique effects, skill/build behavior changes, client presentation, mystery-seller
unique eligibility, and market restrictions remain deferred.
