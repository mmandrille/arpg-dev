# v452 As-Built — resource-wallet-bag

Date: 2026-07-08  
Spec: [`docs/specs/v452_spec-resource-wallet-bag.md`](../specs/v452_spec-resource-wallet-bag.md)  
Plan: [`docs/plans/v452_2026-07-08-resource-wallet-bag.md`](../plans/v452_2026-07-08-resource-wallet-bag.md)

## Shipped

- Account-wide **resource bag** (`account_resource_bag_items`) for slot-based currency and quest resources.
- Pickup routing sends `upgrade_shard`, `renew_stone`, `quest_leaf`, and quest trophies into the resource bag instead of character gear inventory.
- Flat badge counts (`respec_badge`, etc.) remain in `resource_wallet` with existing summary-row presentation.
- Resources wallet window: 5-column scrollable grid with auto-growing rows when the last slot fills.
- Drag transfer between resource bag and inventory via `resource_bag_deposit_item_intent` / `resource_bag_withdraw_item_intent`.
- Session snapshot and deltas expose `resource_bag_items`; persistence mirrors account stash transfer pattern.
- Session start migrates legacy inventory-held resource items into the account resource bag.
- Blacksmith leveled-consumable spend/count includes resource-bag-held shards.

## Authority model

- Server owns pickup routing, bag mutations, and persistence. Client renders bag grid and issues transfer intents only.
- Resource bag is account-scoped (like stash gold/items), not per-character gear inventory.

## Verification

```bash
cd server && go test ./internal/game -run ResourceBag -count=1
make validate-shared
godot --headless --path client --script res://tests/test_material_wallet_panel.gd
```

## Deferred

- Auto-stacking identical `(def_id, level)` bag entries.
- Moving badge wallet counts into per-slot bag rows.
- Dedicated stash tab for resources beyond the wallet window grid.
