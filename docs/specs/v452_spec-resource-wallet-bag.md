# v452 Spec — Resource Wallet Bag

Status: Approved  
Date: 2026-07-08  
Codename: `resource-wallet-bag`

## Purpose

Route currency and quest resource pickups into an account-wide **resource bag** instead of the character gear inventory. Present the bag in the existing Resources wallet window as a scrollable grid that auto-grows rows when the last slot fills. Allow drag transfer between the resource bag and the regular inventory bag. Flat badge counts (`respec_badge`, etc.) stay in `resource_wallet` until stacking ships.

## Non-goals

- Auto-stacking identical `(def_id, level)` entries (deferred).
- Moving badge wallet counts into per-slot bag rows.
- Stash tab for resources; blacksmith merge UI changes beyond reading bag-held shards.
- Protocol version bump beyond optional v8 fields.

## Acceptance criteria

1. Picking up `upgrade_shard`, `renew_stone`, `quest_leaf`, and quest trophies lands items in `resource_bag_items`, not character `inventory`.
2. `resource_wallet` flat badge counts still auto-pickup and display as summary rows.
3. Resources wallet window shows a 5-column grid inside a scroll area; filling the last visible slot grows `wallet_rows` by one.
4. Drag from resource bag slot → inventory bag deposits via `resource_bag_withdraw_item_intent`; reverse uses `resource_bag_deposit_item_intent`.
5. Session snapshot and deltas include `resource_bag_items`; persistence mirrors account stash transfer pattern.
6. Existing inventory-held resource items are migrated to the account resource bag on session start.
7. Headless unit tests cover wallet grid row growth and a Go pickup routing test.

## Scope

| Area | Files |
|------|-------|
| Migration | `server/migrations/0030_account_resource_bag.sql` |
| Store | `server/internal/store/resource_bag_store.go`, `interfaces.go`, `models.go`, `repos.go` |
| Sim | `server/internal/game/resource_bag.go`, `sim.go`, `sim_players.go`, `sim_load.go`, `handlers.go`, `badge_rewards.go`, `types.go` |
| Realtime | `server/internal/realtime/session_loop.go` |
| HTTP session | `server/internal/http/session.go` |
| Protocol | `shared/protocol/session_snapshot.v8.schema.json`, `state_delta.v8.schema.json`, `messages.v8.schema.json` |
| Client | `client/scripts/material_wallet_panel.gd`, `inventory_panel.gd`, `inventory_transfer_router.gd`, `inventory_wallet_delta_runtime.gd`, `main.gd` |
| Tests | `server/internal/game/resource_bag_test.go`, `client/tests/test_material_wallet_panel.gd` |

## Test proof

```bash
cd server && go test ./internal/game -run ResourceBag -count=1
godot --headless --path client --script res://tests/test_material_wallet_panel.gd
make validate-shared
```

## Open questions

None — quest trophies included in auto-route per user confirmation.
