# v390 As-built — Leveled Upgrade Shards

Codename: `leveled-upgrade-shards`

## What shipped

- **Upgrade Shard** is a leveled bag/stash item (`item_level` in `rolled_stats`), not wallet currency.
- Purple badge presentation in `item_presentations.v0.json`; display name **Upgrade Shard**.
- Data-driven extra drops: enemy 1%, chest 2%, boss 3% (`main_config.v0.json`).
- Quest `badge_reward` still grants leveled shards; boss badge rewards skip `upgrade_shard`.
- Blacksmith upgrade consumes one shard with `level >= target ilvl + 1` and gold equal to item sell price.
- Blacksmith **Merge** tab (5×5 staging grid): 3 same-level shards → 1 shard at level +1.
- Wallet migration: flat `upgrade_shard` wallet balances become level-1 stash rows on resource load.

## Key files

- `server/internal/game/upgrade_shard_drops.go`, `upgrade_shard_items.go`
- `server/internal/game/badge_rewards.go`, `resource_wallet.go`
- `server/internal/store/upgrade_shard_store.go`
- `server/internal/http/account_stash_merge.go`, `account_stash.go`
- `client/scripts/blacksmith_merge_panel.gd`, `blacksmith_panel.gd`
- Bot: `upgrade_shard_level_drop`, `client/blacksmith_shard_merge` (extended)

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'UpgradeShard|BadgeReward' -count=1
cd server && go test ./internal/store/... -run 'Upgrade|Merge|Shard' -count=1
make bot scenario=upgrade_shard_level_drop
make bot-client SCENARIO=blacksmith_shard_merge HEADLESS=1
make bot-visual scenario=blacksmith_shard_merge
```
