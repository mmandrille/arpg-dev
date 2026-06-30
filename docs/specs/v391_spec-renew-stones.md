# v391 Spec — Renew Stones

Status: Complete
Date: 2026-06-30
Codename: renew-stones
Baseline: v390 `leveled-upgrade-shards` complete

## Purpose

Add **Renew Stone** leveled consumables and blacksmith **Renew Item** rerolls, unify dungeon
resource loot into a weighted pool, retier drop chances by monster/chest type, simplify blacksmith
recipes, and align blacksmith actions with bag-only staging.

1. **`renew_stone` item** — leveled bag consumable; shiny light-blue badge presentation; merge
   3 same-level → 1 level+1.
2. **Unified `resource_loot_drops`** — on trigger success, pick from pool (`upgrade_shard`,
   `renew_stone`), then roll level from depth tier.
3. **Tiered drop chances** (data-driven):
   - 1% common/rare monster kills
   - 2% champion monster kills; regular chest opens
   - 3% unique monster kills; boss chest opens
   - 5% boss kills
4. **Blacksmith Renew Item** — consumes renew stone with `level >= item_level` + sell-price gold;
   rerolls random stat affixes; preserves template, rarity, item level, unique fixed effects.
5. **Bag-only blacksmith** — upgrade, renew, and merge use character bag items dragged to the
   staging slot (no direct stash staging).
6. **Remove** Hone Weapon and Reinforce Armor recipes.
7. **Bishop debug** — `ARPG_GAMEPLAY_DEBUG=1` drops level-1 renew stone on the floor near player.

## Non-goals

- Quest badge grants for renew stones
- Affix add/improve-roll (ADR-0012)
- Market/trade restrictions on renewed items
- Production art; code-native light-blue placeholder
- Mystery seller debug grant

## Acceptance criteria

- [ ] `renew_stone` in shared items; display name **Renew Stone**; leveled `item_level` payload
- [ ] Light-blue shiny presentation on ground, bag, tooltips, blacksmith UI
- [ ] `resource_loot_drops` pool + tier chances in `main_config`; old per-shard drop fields removed
- [ ] Monster/chest/boss hooks use correct tier chance; pool pick + level roll deterministic
- [ ] Renew: stone `level >= item_level`, sell-price gold, affixes change, unique fixed effects unchanged
- [ ] Blacksmith shows Upgrade Item + Renew Item only; hone/reinforce removed server + client
- [ ] Upgrade, renew, merge operate on bag items only from blacksmith UI
- [ ] Merge supports upgrade shards and renew stones from bag (3 same-level → +1)
- [ ] Bishop debug drops renew stone on floor when gameplay debug enabled
- [ ] Extended bot scenarios prove pool drop and bag renew; stale hone scenarios removed/updated
- [ ] `make validate-shared`, focused tests green

## Scope and likely files

- `shared/rules/main_config.v0.json` + schema — `resource_loot_drops`
- `shared/rules/items.v0.json`, `shared/assets/item_presentations.v0.json`
- `server/internal/game/resource_loot_drops.go`, `renew_stone_items.go`, `item_reroll.go`
- `server/internal/game/bishop_debug.go`, `handlers.go`, `sim.go`, `interactables.go`
- `server/internal/http/account_stash.go`, `account_stash_merge.go`
- `server/internal/store/upgrade_shard_store.go` (generalize leveled consumables)
- `client/scripts/blacksmith_*.gd`, `bishop_panel.gd`, `main.gd`, `net_client.gd`
- Bot scenarios: `resource_loot_pool_drop`, `client/blacksmith_renew_item`

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ResourceLoot|RenewStone|ItemReroll|BishopDebug' -count=1
cd server && go test ./internal/store/... -run 'Merge|Renew' -count=1
cd server && go test ./internal/http/... -run 'Renew|Blacksmith|Merge' -count=1
make bot scenario=resource_loot_pool_drop
make bot-client SCENARIO=blacksmith_renew_item HEADLESS=1
```

## Open questions

None — resolved in brief (pool pick, `>=` level match, sell-price gold, bag-only, bishop debug).
