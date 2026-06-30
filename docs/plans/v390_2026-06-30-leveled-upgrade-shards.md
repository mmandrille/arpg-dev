# v390 Plan — Leveled Upgrade Shards

Date: 2026-06-30
Spec: `docs/specs/v390_spec-leveled-upgrade-shards.md`

## Tasks

- [x] 1. Plan + shared rules (`items`, `main_config` drop %, schema), purple `item_presentations`, remove TC `upgrade_shard`
- [x] 2. Server game: `upgrade_shard_drops.go`, `upgrade_shard_items.go`, wallet exclude, badge quest items / boss skip, kill + chest hooks
- [x] 3. Server store + HTTP: leveled shard spend, sell-price gold, merge endpoint, wallet→stash migration
- [x] 4. Client: purple shards, blacksmith preview (inventory shards, sell price), merge tab, hide wallet `upgrade_shard`
- [x] 5. Bot scenarios + broken test/scenario updates
- [x] 6. `docs/as-built/v390_leveled-upgrade-shards.md`

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'UpgradeShard|BadgeReward' -count=1
cd server && go test ./internal/store/... -run 'Upgrade|Merge|Shard' -count=1
cd server && go test ./internal/http/... -run 'Upgrade|Merge|Blacksmith' -count=1
make bot scenario=upgrade_shard_level_drop
make bot-client SCENARIO=blacksmith_shard_merge HEADLESS=1
```
