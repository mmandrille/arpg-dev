# v389 Plan — Tiered Item Level Scaling

Status: Ready for implementation
Goal: Tiered item levels every 10 dungeon depths with unified depth scaling for monsters and items, plus deepest-depth-gated blacksmith upgrades.
Architecture: Extract `depth_scaling.go` as the single depth-factor implementation. Item rolls pick ilvl uniformly, roll affixes at representative tier depth, then scale stats/requirements. Store upgrade unscale→increment→rescale using the same module. HTTP passes `deepest_dungeon_depth` cap into store.
Tech stack: Go sim + store + HTTP, shared JSON rules, Godot blacksmith preview, Python protocol bot.

## Baseline and shortcut decision

- Builds on v196 `item_level` metadata and v110 repeat upgrades
- Asset/plugin: **reject** — reuse blacksmith panel + tooltip patterns

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | `item_level_tiers.levels_per_tier` |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Schema for tiers |
| Create | `server/internal/game/depth_scaling.go` | Unified depth index/factor + per-stat scaling |
| Create | `server/internal/game/item_level_scaling.go` | Tier max, roll pick, finalize, upgrade |
| Modify | `server/internal/game/item_rolls.go` | Tier roll + finalize scaling |
| Modify | `server/internal/game/dungeon_population.go` | Use shared depth scaling |
| Modify | `server/internal/store/repos.go` | Rescale upgrade + depth cap |
| Modify | `server/internal/http/account_stash.go` | Deepest depth cap |
| Modify | `client/scripts/blacksmith_upgrade_preview.gd` | Depth cap preview/disable |
| Create | `tools/bot/scenarios/89_tiered_item_level_drops.json` | Drop proof |
| Modify | `tools/bot/scenarios/85_item_level_progression.json` | Tier expectations |

## Maintenance ratchet

Hotspot files: `dungeon_population.go`, `repos.go`, `shop.go` — extract new modules; touch-to-shrink on edited grandfathered files.

## Task 1 — Shared rules

- [ ] Add `item_level_tiers.levels_per_tier: 10` to dungeon generation rules + schema

```bash
make validate-shared
```

## Task 2 — Unified depth scaling (Go)

- [ ] Create `depth_scaling.go` with `DepthIndex`, `DepthFactor`, stat scaling helpers
- [ ] Refactor `generatedMonsterStats` to call shared helpers
- [ ] Create `item_level_scaling.go` with tier max, roll pick, finalize, upgrade rescale
- [ ] Unit tests for tier max, scaling parity, upgrade rescale

```bash
cd server && go test ./internal/game/... -run 'DepthScaling|ItemLevel|Tiered' -count=1
```

## Task 3 — Roll pipeline + named payloads

- [ ] Update `item_rolls.go` tier pick + finalize
- [ ] Update `unique_chest.go`, `set_items.go`, `sim_load.go` for tier scaling

```bash
cd server && go test ./internal/game/... -run 'RolledItemLevel|ItemLevel' -count=1
```

## Task 4 — Blacksmith upgrade + depth cap

- [ ] HTTP loads deepest depth, passes cap to store upgrade
- [ ] Replace `upgradedRolledStats` +1-stat with rescale
- [ ] Store/HTTP tests for cap and rescale

```bash
cd server && go test ./internal/store/... ./internal/http/... -run 'Upgrade' -count=1
```

## Task 5 — Client blacksmith preview

- [ ] Disable upgrade + preview reason when next ilvl exceeds depth cap

```bash
make client-unit
```

## Task 6 — Bot scenarios

- [ ] `89_tiered_item_level_drops.json` extended lab proof
- [ ] Update `85_item_level_progression.json`

```bash
make bot scenario=tiered_item_level_drops
```

## Task 7 — Lifecycle docs

- [ ] `docs/as-built/v389_tiered-item-level-scaling.md`
- [ ] `PROGRESS.md` + lifecycle row

## Final verification

```bash
make maintainability
make validate-shared
cd server && go test ./internal/game/... -run 'ItemLevel|DepthScaling|Tiered|Upgrade' -count=1
make bot scenario=tiered_item_level_drops
make client-unit
```
