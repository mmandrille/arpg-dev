# v452 Plan — Resource Wallet Bag

Date: 2026-07-08  
Spec: `docs/specs/v452_spec-resource-wallet-bag.md`

## Tasks

1. Migration + store (`0030_account_resource_bag.sql`, `resource_bag_store.go`, session snapshot hooks)
2. Sim routing (`resource_bag.go`, pickup branch, deposit/withdraw intents, badge shard grants)
3. Protocol v8 optional fields (`resource_bag_items`, delta ops, intents)
4. Client wallet grid (`material_wallet_panel.gd`, drag transfer, `main.gd` + delta runtime wiring)
5. Blacksmith consumers (HTTP leveled consumable spend/count includes resource bag; client merges bag into blacksmith resource list)
6. Tests + scenario updates (`resource_bag_test.go`, `test_material_wallet_panel.gd`, bot `resource_bag_count` assertions)

## Verification

```bash
cd server && go test ./internal/game -run ResourceBag -count=1
make validate-shared
godot --headless --path client --script res://tests/test_material_wallet_panel.gd
make bot scenario=upgrade_resource_drop
make bot-client SCENARIO=material_wallet_window HEADLESS=1
```
