# v221 As-Built - Resource Wallet Foundation

Date: 2026-06-16

## What shipped

- Added account-wide `account_resource_wallet` persistence plus session-start wallet snapshot rows.
- Moved `upgrade_shard` pickup into `resource_wallet[upgrade_shard]`; pickup removes the loot entity,
  emits `resource_picked_up`, and sends a private `resource_wallet_update`.
- Loaded account wallet balances into solo, host, guest, and late-join sim state, and persisted live
  wallet pickup increments back to the account table.
- Updated inventory blacksmith upgrades to require and spend the configured wallet resource, returning
  the remaining wallet count even when it reaches zero.
- Exposed `resource_wallet` in v8 snapshots and `resource_wallet_update` in v8 deltas, with examples
  and bot ingestion/assertions.
- Updated the Godot blacksmith panel to display and enable upgrades from the wallet count rather than
  counting shard items in the bag.

## Proof

```bash
make validate-shared
cd server && go test ./internal/store ./internal/http -run 'ResourceWallet|Upgrade' -count=1
cd server && go test ./internal/realtime ./internal/replay -run '^$' -count=1
make bot scenario=upgrade_resource_drop
make client-unit
make bot-client scenario=blacksmith_upgrade_ui
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
remains deferred until the selected feature queue is complete.

Manual visual proof, if desired:

```bash
make bot-visual scenario=blacksmith_upgrade_ui
```

## Scope limits

- Existing inventory-held upgrade shards are not migrated into the wallet.
- No standalone material wallet UI, multi-resource recipes, stash material tabs, market restrictions,
  or resource trading shipped.
- The realtime persistence path treats `resource_wallet_update` as pickup increments only; HTTP owns
  blacksmith spending for this slice.
