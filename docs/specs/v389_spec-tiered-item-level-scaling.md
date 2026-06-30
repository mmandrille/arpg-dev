# v389 Spec — Tiered Item Level Scaling

Status: Approved for implementation
Date: 2026-06-30
Codename: tiered-item-level-scaling
Baseline: v388 `armor-slot-families` complete

## Purpose

Replace the v196 **1:1 source-depth → item_level** model with a **10-level tier band** system:

1. **Drop tier gating** — at dungeon depth `D`, loot rolls `item_level` uniformly from `1..max(1, floor(D / 10))`.
2. **Unified depth scaling** — one shared depth-scaling implementation drives monster combat stats and item stat/requirement scaling. Item ilvl `N` anchors to representative depth `(N-1)*10 + 1`.
3. **Roll-time scaling** — after rarity/affix rolls at ilvl-1 baselines, apply tier scaling to damage, armor, rolled numeric affixes, and requirements.
4. **Blacksmith alignment** — successful upgrade increments `item_level` and **re-scales** stats/requirements proportionally (preserve affix identity). Upgrade cap: `item_level + 1 <= max(1, floor(deepest_dungeon_depth / 10))`, in addition to `item_upgrade_max_level`.

## Non-goals

- Full v29 loot-band / treasure-class rebalance
- Affix grammar, procedural names, new unique/set catalogs
- Market restrictions, item binding, migration/rescale of grandfathered stash rows
- Client art; tooltip copy beyond existing item level footer + blacksmith preview/disable reason
- Protocol schema version bump (reuse existing `item_level` + rolled stats fields)

## Acceptance criteria

- [ ] Shared rules document `levels_per_tier` (default 10) under dungeon generation; schema validated
- [ ] Single Go depth-scaling module used by `generatedMonsterStats` and item scaling (no duplicate `depthFactor` helpers)
- [ ] `max_item_level(D) = max(1, floor(D / levels_per_tier))` for drops and upgrade caps
- [ ] Drop roll picks ilvl uniformly in range; ilvl 1 items match template base at depth 1
- [ ] ilvl 2 item stats align with tier-2 representative depth using the same scaling curve as monsters
- [ ] Blacksmith upgrade: depth cap enforced via `deepest_dungeon_depth`; success rescales stats (not +1 random stat)
- [ ] Client blacksmith preview disables upgrade when next ilvl exceeds depth cap
- [ ] Updated protocol bot scenario proves tiered drop; focused Go tests for scaling parity and upgrade cap
- [ ] `make validate-shared`, focused tests, and autoloop focused verification green

## Scope and likely files

- `shared/rules/dungeon_generation.v0.json` + schema — `item_level_tiers.levels_per_tier`
- `server/internal/game/depth_scaling.go` — unified depth factor + stat scaling helpers (new)
- `server/internal/game/item_level_scaling.go` — tier pick, roll finalize, upgrade rescale (new)
- `server/internal/game/item_rolls.go`, `dungeon_population.go`, `unique_chest.go`, `set_items.go`, `sim_load.go`
- `server/internal/store/repos.go` — upgrade uses rescale + depth cap parameter
- `server/internal/http/account_stash.go` — load `deepest_dungeon_depth`, pass cap
- `client/scripts/blacksmith_upgrade_preview.gd`, `blacksmith_panel.gd`
- `tools/bot/scenarios/89_tiered_item_level_drops.json` (extended); update `85_item_level_progression.json`
- Docs: plan, as-built, lifecycle

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ItemLevel|DepthScaling|Tiered' -count=1
cd server && go test ./internal/store/... -run 'Upgrade' -count=1
make bot scenario=tiered_item_level_drops
```

Visual (optional): `make bot-visual scenario=blacksmith_upgrade_ui`

## Asset decision

- **Reject** external plugins/assets — server/rules + existing blacksmith UI only

## Open questions

None — resolved in `/next` brief and user follow-up (Q3 unified scaling, Q8 deepest-depth upgrade cap, Q1 `floor(D/10)` formula).
