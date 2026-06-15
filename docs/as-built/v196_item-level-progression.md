# v196 As-built: Item Level Progression

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `item_level` to the durable rolled-item payload.
- Generated template rolls now set item level from source depth, clamped to at least 1.
- Named unique and set package payloads now set item level from their effective minimum level.
- Surfaced `item_level` through floor loot, inventory, stash, shop offer, and shop appraisal views.
- Updated current v8 protocol snapshot and delta schemas for the new rolled-item field.
- Added bot assertion support for exact/minimum item-level checks.
- Added protocol bot scenario `85_item_level_progression.json`.
- Moved `weightedRollableStat` out of `sim.go` into `item_roll_helpers.go` to keep the sim
  coordinator under the file-size ratchet.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'ItemLevel|RolledTemplateLootTransfersToInventory' -count=1`
- `make bot scenario=85_item_level_progression.json`
- `make ci`

## Deferred

- Item-level display in client tooltips remains future UI work.
- Item-level-driven stat scaling, upgrade odds, and loot-band rebalance remain future itemization
  slices.
