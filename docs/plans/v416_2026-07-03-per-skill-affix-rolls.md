# v416 Plan — Per-Skill Affix Rolls

Status: Complete
Goal: Roll +N random skill levels on helm/amulet and class gear; uncapped effective rank from gear.
Architecture: `SkillLevelBonusRoll` on `ItemRollPayload`; roll kinds `random_skill_level` /
`random_class_skill_level`; `effectiveSkillRank` adds gear without MaxRank/requirement clamp;
`skill_bonus_status` for tooltips. Reuses `ClassAffinityTooltip` color pattern.
Tech stack: shared JSON, Go sim, Godot client, Python bot.

## Baseline and shortcut decision

Builds on v189 skill affixes and v405/v415 class specialist gear. **Borrow** class affinity
tooltip colors; **reject** new art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | helm/amulet + class specialist roll pools |
| Modify | `shared/rules/item_templates.v0.schema.json` | roll stat kinds |
| Modify | `shared/golden/item_rolls.v0.schema.json` | payload shape |
| Modify | `shared/protocol/state_delta.v8.schema.json` | `skill_bonus_status` |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | `skill_bonus_status` |
| Create | `server/internal/game/item_skill_level_bonuses.go` | roll, status, aggregation |
| Modify | `server/internal/game/item_skill_stats.go` | move rank helpers if needed |
| Modify | `server/internal/game/item_rolls.go`, `item_reroll.go`, `weapon_elemental.go` | roll path |
| Modify | `server/internal/game/sim.go`, `handlers.go`, mobility/survival | rank + cast gates |
| Modify | `server/internal/game/types.go`, `sim_load.go`, `shop_views.go` | views + clone |
| Create | `server/internal/game/item_skill_level_bonuses_test.go` | focused tests |
| Create | `client/scripts/skill_bonus_tooltip.gd` | colored lines |
| Create | `client/tests/test_skill_bonus_tooltip.gd` | headless test |
| Modify | inventory/stash/shop/market panels | wire tooltip |
| Create | `tools/bot/scenarios/skill_level_affix_lab.json` | extended proof |
| Modify | `shared/rules/worlds.v0.json` | lab loot spawn |
| Modify | `tools/validate_shared.py` | roll kind validation |

## Maintenance ratchet

New `item_skill_level_bonuses.go` ≤600 lines. Touch `sim.go` only for annotate hooks + slim
`effectiveSkillRank` delegation.

## Task 1 — Shared contracts

- [x] Schema roll kinds + golden `skill_level_bonuses`
- [x] Template weights: helm/amulet ~5%, class gear ~10% class-only skill roll

```bash
make validate-shared
```

## Task 2 — Server rolls and effective rank

- [x] `item_skill_level_bonuses.go` roll + status + equipped sum
- [x] Wire item_rolls / reroll / rollAffixStatsOntoMap
- [x] `effectiveSkillRank` uncapped; remove cast requirement gate
- [x] `skill_bonus_status` on item views
- [x] Equip emits `skill_progression_update` for gear rank changes

```bash
cd server && go test ./internal/game/... -run 'SkillLevelBonus' -count=1
```

## Task 3 — Client tooltips

- [x] `skill_bonus_tooltip.gd` + panel wiring + headless test

```bash
godot --headless --path client --script res://tests/test_skill_bonus_tooltip.gd
```

## Task 4 — Bot scenario

- [x] `skill_level_affix_lab` world + extended scenario

```bash
make bot scenario=skill_level_affix_lab
```

## Task 5 — Lifecycle docs

- [x] as-built, lifecycle, PROGRESS

## Final verification

```bash
make maintainability
make validate-shared
cd server && go test ./internal/game/... -run 'SkillLevelBonus|SkillAffix' -count=1
godot --headless --path client --script res://tests/test_skill_bonus_tooltip.gd
make bot scenario=skill_level_affix_lab
```
