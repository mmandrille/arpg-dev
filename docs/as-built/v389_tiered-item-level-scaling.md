# v389 As-built: Tiered Item Level Scaling

Date: 2026-06-30
Status: Complete — focused verification green; final batch `make ci` pending autoloop handoff

## What shipped

- Added `item_level_tiers.levels_per_tier` (default 10) to dungeon generation rules.
- Introduced unified depth scaling in `depth_scaling.go` consumed by monsters and items.
- Drops roll `item_level` uniformly from `1..max(1, floor(depth / levels_per_tier))`, then scale stats/requirements at the tier anchor depth.
- Blacksmith upgrades rescale stats proportionally to the next tier instead of +1 random stat.
- Upgrade cap uses `deepest_dungeon_depth` in addition to config `item_upgrade_max_level`.
- Client blacksmith preview disables upgrades and explains depth cap when next tier exceeds progression.
- Regenerated `shared/golden/shop_offers.json` for tiered roll drift.
- Extended bot scenario `89_tiered_item_level_drops.json`.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ItemLevel|DepthScaling|Tiered|RolledItemLevel|ShopGeneratedOfferGolden' -count=1
cd server && go test ./internal/store/... -run 'Upgrade' -count=1
cd server && go test ./internal/http/... -run 'Upgrade|BlacksmithRecipe' -count=1
make bot scenario=tiered_item_level_drops
make maintainability
```

## Deferred

- Full loot-band / treasure-class rebalance for tier bands
- Grandfathered stash item migration/rescale on load
- GDScript parity evaluator for exact tier preview numbers (server remains authoritative)
