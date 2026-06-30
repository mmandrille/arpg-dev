# v390 Spec — Leveled Upgrade Shards

Status: Approved for implementation
Date: 2026-06-30
Codename: leveled-upgrade-shards
Baseline: v389 `tiered-item-level-scaling` complete

## Purpose

Refactor the blacksmith upgrade resource from a flat account-wallet **Upgrade Badge** into visible,
**leveled Upgrade Shard** items that players pick up, stash, and spend at the blacksmith.

1. **Rename & present** — canonical ID `upgrade_shard`; display name **Upgrade Shard**; purple
   icon/ground presentation with level label.
2. **Item instances** — shards are bag/stash rows with `item_level` (shard level) in `rolled_stats`;
   not `account_resource_wallet` entries.
3. **Dungeon extra drops** — independent roll per event (not treasure-class loot):
   - enemy kill: **1%**
   - chest open: **2%**
   - boss kill: **3%**  
   Shard level uniform `1..max(1, floor(depth / levels_per_tier))` (v389 tier band).
4. **Quest badge rewards** — `upgrade_shard` remains in `badge_reward_rules` for **quest turn-in**
   only; grants a leveled shard item (level from `deepest_dungeon_depth` tier band). Boss kills do
   **not** grant `upgrade_shard` via `badge_reward` (dungeon 3% roll only).
5. **Upgrade cost** — upgrading item ilvl N → N+1 consumes one shard with `shard_level >= N+1`
   (higher qualifies) plus gold equal to the item **sell price** (shop appraisal).
6. **Blacksmith merge tab** — 5×5 slot grid + **Merge** button; first recipe: 3 shards of the same
   level → 1 shard of level +1 (no upper cap on merge output).
7. **Migration** — convert existing flat `upgrade_shard` wallet balances to level-1 shard stash rows.

## Non-goals

- Additional merge formulas beyond 3→1 same-level shards
- Affix add/improve-roll upgrades; success-chance/pity changes beyond existing recipe
- Market restrictions, item binding, upgraded-item trade rules
- Leveled wallet for other badge types (respec/stat/skill/resurrection stay flat wallet)
- Production art/plugins; code-native purple placeholder only
- Grandfathered equipment rescale (v389 deferred)
- Full loot-band / treasure-class rebalance

## Acceptance criteria

- [ ] Display name **Upgrade Shard** everywhere player-facing; ID remains `upgrade_shard`
- [ ] Purple presentation on ground loot, bag/stash rows, tooltips, blacksmith/merge UI
- [ ] Shard pickup adds inventory row (or stash transfer) with visible `item_level`; not wallet
- [ ] Enemy 1% / chest 2% / boss 3% extra-drop chances are data-driven and deterministic
- [ ] Dropped shard level uses v389 `levels_per_tier` band at event depth
- [ ] Extra drop is a second roll; normal TC loot unchanged
- [ ] Quest turn-in `badge_reward` can still grant `upgrade_shard` as leveled inventory/stash item
- [ ] Boss `badge_reward` does not grant `upgrade_shard`
- [ ] Blacksmith upgrade at ilvl N→N+1 requires shard `>= N+1` and sell-price gold
- [ ] Upgrade still respects v389 depth cap and `item_upgrade_max_level`
- [ ] Blacksmith **Merge** tab: 5×5 staging, server validates 3 same-level shards → 1 level+1
- [ ] Migration: flat wallet `upgrade_shard` → level-1 stash items on account load or boot
- [ ] `make validate-shared`, focused tests, and bot scenarios green

## Scope and likely files

- `shared/rules/items.v0.json`, `main_config.v0.json` + schema — drop rules; quest-only badge row
- `shared/assets/item_presentations.v0.json` — purple shard icon/ground + level label
- `server/internal/game/resource_wallet.go` — exclude `upgrade_shard` from wallet auto-pickup
- `server/internal/game/upgrade_shard_drops.go` (new) — extra-roll hooks
- `server/internal/game/badge_rewards.go` — quest shard items; skip boss upgrade_shard badge
- `server/internal/http/account_stash.go` — leveled shard spend; merge endpoint; sell-price gold
- `server/internal/store/repos.go` — upgrade + merge transactions
- `client/scripts/blacksmith_panel.gd`, `blacksmith_merge_panel.gd` (new)
- `client/scripts/material_wallet_panel.gd` — remove upgrade_shard wallet row (or aggregate read-only from stash)
- Bot scenarios: new/updated drop, merge, upgrade proofs

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'UpgradeShard|BadgeReward' -count=1
cd server && go test ./internal/store/... -run 'Upgrade|Merge' -count=1
cd server && go test ./internal/http/... -run 'Upgrade|Merge|Blacksmith' -count=1
make bot scenario=upgrade_shard_level_drop
make bot-client SCENARIO=blacksmith_shard_merge HEADLESS=1
```

Visual: `make bot-visual scenario=blacksmith_shard_merge`

## Asset decision

- **Borrow** — existing `badge` family icon shape from `item_presentations.v0.json`
- **Reject** — external plugins/textures; purple via data-driven colors + level label

## Open questions

None — resolved in `/next` brief and user follow-up (Q1 stash-visible items; Q5 quest badge_reward retained; Q6 separate extra roll).
