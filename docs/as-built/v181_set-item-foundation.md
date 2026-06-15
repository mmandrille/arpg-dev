# v181 As-built - Set Item Foundation

## What shipped

- Added `shared/rules/set_items.v0.json` with the five-piece `Verdant Vanguard` set and schema validation.
- Added fixed set payload construction with `rarity: "set"`, durable `set_id` / `set_item_id`, and five debug-testable items.
- Extended the debug unique chest to include the set pieces alongside existing unique effect and named unique items.
- Added server-authoritative equipped-piece bonuses: 2 pieces grant armor, 3 grant max HP, 4 grant attack speed, and 5 grant `all_skills`, skill damage, and health regen.
- Added green set rarity presentation in existing loot labels and item tooltips.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestUnique|TestSet'`
- `make client-unit`
- `make ci`

## Deferred

- Random set drops, drop weights, dedicated set chest tabs, richer set tooltip breakdowns, collection tracking, and production set art remain future itemization work.
