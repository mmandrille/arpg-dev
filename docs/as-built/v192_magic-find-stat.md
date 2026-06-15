# v192 As-built: Magic Find Stat

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `magic_find_percent` as a bounded rollable equipment stat with shared schema, validation,
  text catalog, client label, and protocol derived-stat coverage.
- Added rare Magic Find roll candidates to cave jewelry and shop pricing weights so generated offers
  and sale/appraisal math stay data-driven.
- Exposed equipped Magic Find through `derived_stats.magic_find_percent` and stat breakdowns with
  equipment-roll sources.
- Moved item-template rolling into a focused helper and added a Magic-Find-aware rarity-weight path
  used only by monster item-template loot rolls.
- Kept shop, authored loot, fixed drops, gold, resources, chests, and zero-Magic-Find rolls on the
  baseline rarity path.
- Added protocol bot scenario `81_magic_find_stat.json`, proving deterministic Magic Find pickup,
  equip, derived stat, and stat breakdown payloads.
- Lowered the `server/internal/game/shop.go` maintainability baseline after extracting the roll body.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'MagicFind|ItemRollsGolden|ShopGeneratedOfferGolden' -count=1`
- `make bot scenario=81_magic_find_stat.json`
- `make ci`

## Deferred

- Economy tuning for Magic Find caps, diminishing returns, or rarity-specific curves.
- Client-specific visual treatment beyond existing stat labels and stat panel rendering.
