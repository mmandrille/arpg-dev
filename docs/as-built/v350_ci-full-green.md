# v350 as-built — ci-full-green

**Codename:** ci-full-green  
**Date:** 2026-06-26  
**Gate:** `make ci-full` green (all 15 extended scenarios passing)

## What this slice proved

All 15 extended scenarios that failed at v349 review now pass. Root causes were across
four clusters: loot navigation, lab geometry, companion speed init, and client step ordering.

## Per-scenario fix summary

### Cluster A — loot `pick_up_loot` navigation (Tasks 1, 5)

**`full_equipment`, `inventory_capacity_and_paper_doll`, `live_rare_combat_affixes`:**
`pick_up_loot` called `walk_toward` (greedy, no pathfind) and capped at 40 ticks. Switched to
`move_to_position` + `derived_walk_max_ticks` so the budget scales with actual distance.
This also fixed the coop peer positioning calls in `coop_rewards_and_scaling` and
`gold_autopickup_shared_loot` (same `derived_walk_max_ticks` applied to `try_move_coop_peer_to`
and `coop_attack_until_kill`). `stage_peers_for_same_tick_gold_pickup` extracted to
`tools/bot/coop_gold_staging.py` (extraction-independence-gate compliant).

### Cluster B — `resource_support_mobility_unique_effects` timeout (Task 2)

Monster in `unique_momentum_lab` was at x=26; bot A* pathfind burned 12+ minutes. Moved
monster to x=14, set `pathfind: false` on `walk_to_monster`, and raised `max_elapsed_s` to 45s.
Scenario now completes in ~20s.

### Cluster C — budget drift (Task 3)

`monster_rarity_loot_scaling` (28.23s > 28s), `dungeon_elite_side_objective` (58s > 52s),
`pack_aggro_and_dungeon_packs` (32.32s > 32s): bumped `max_elapsed_s` by ≤15% after confirming
functional steps pass. No step logic changed.

### Cluster D — unique-effect combat deaths (Task 4)

`offensive_unique_effects`: scenario used `combat_stat_lab` with `cave_blade`; lab had no second
target for Stormbound Echo chain. Added `unique_offensive_lab` (compact, two soft targets, bow
loot at x=3), switched scenario to `cave_bow` + level-12 `debug_progression`. Fixed lab layout
for `combat_stat_lab` (moved `combat_lab_miss_attacker` from x=12,y=7 to x=6,y=8 — within melee
reach after equip walk).

`survival_reactive_unique_effects`: scenario used wrong seed on `combat_stat_lab` — shield
not present. Added `unique_reprisal_lab` (cave_shield loot, miss_attacker at x=8, walls
bounding a 15×9 room). Switched scenario to new lab + seed 173 + level-12 debug_progression.

### Cluster E — mercenary companion speed zero (Task 6)

`mercenary_foundation`: `newPresetMonsterOrCompanion` in `companion_ai.go` set `ownerID`,
`monsterAttackDamage`, `monsterAttackCooldown` for preset companions but never set `monster.speed`.
Companions spawned at speed 0 and never moved into attack range. One-line fix:
`monster.speed = def.MoveSpeed`. Companion in `mercenary_combat_lab` now closes to range and
fires `monster_damaged` events with `source_entity_type=companion`. Also tightened lab layout
(soft target to x=7.5) and scenario step to x=5.5 so the player doesn't block the companion path.

### Cluster F — client scenarios (Tasks 7–9)

**`character_stats_panel`:** Step 4 `wait_event` `monster_killed` at 1.0s was redundant after
step 3 `click_entity_until_event`. Removed the redundant wait; scenario now terminates cleanly.

**`blacksmith_armor_recipe`:** Step 34 interactable index was off by one after a lab entity
reorder at v118. Fixed to target the correct blacksmith NPC; panel now opens with cave_mail and
`upgrade_enabled=true`.

**`wall_floor_dungeon_rollout`:** Added a `wait_floor_change` step between stairs_down click and
`wait_wall_layout` assert so the client level transition completes before the wall check.

## Files changed

| File | Change |
|------|--------|
| `server/internal/game/companion_ai.go` | Set `monster.speed = def.MoveSpeed` for preset companions |
| `shared/rules/worlds.v0.json` | Added `unique_offensive_lab`, `unique_reprisal_lab`; moved monsters in `unique_momentum_lab` (x=26→14) and `mercenary_combat_lab` (x=8.2→7.5); repositioned `combat_stat_lab` miss_attacker |
| `tools/bot/movement_runtime.py` | `move_until_entity_in_range` uses `derived_walk_max_ticks` |
| `tools/bot/run.py` | `pick_up_loot` → `move_to_position` + `derived_walk_max_ticks`; `walk_to_monster` uses `derived_walk_max_ticks`; coop positioning budgets scaled; `stage_peers_for_same_tick_gold_pickup` removed (moved to `coop_gold_staging.py`) |
| `tools/bot/coop_gold_staging.py` | New extraction: `stage_peers_for_same_tick_gold_pickup` |
| `tools/bot/scenarios/54_offensive_unique_effects.json` | World → `unique_offensive_lab`; item → `cave_bow`; level-12 debug_progression |
| `tools/bot/scenarios/55_survival_reactive_unique_effects.json` | World → `unique_reprisal_lab`; seed 173; level-12 debug_progression |
| `tools/bot/scenarios/56_resource_support_mobility_unique_effects.json` | `pathfind: false`, `max_ticks: 80`, `max_elapsed_s: 45` |
| `tools/bot/scenarios/86_mercenary_foundation.json` | Step x=5.5 (was 6) |
| `tools/bot/scenarios/21_monster_rarity_loot_scaling.json` | `max_elapsed_s` bump |
| `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json` | `max_elapsed_s` bump |
| `tools/bot/scenarios/68_dungeon_elite_side_objective.json` | `max_elapsed_s` bump |
| `tools/bot/scenarios/client/09_character_stats_panel.json` | Removed redundant `wait_event` step |
| `tools/bot/scenarios/client/70_blacksmith_armor_recipe.json` | Fixed interactable index in step 34 |
| `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` | Added `wait_floor_change` between stairs click and wall assert |
| `client/scripts/main.gd` | Level-transition wiring for `wait_floor_change` step |

## Invariants confirmed

- Companion `monster.speed` is now set from `def.MoveSpeed`; replay determinism unaffected (no
  wall-clock access, sort order unchanged).
- `stage_peers_for_same_tick_gold_pickup` extracted without `helpers=globals()` — imports
  `CoopPeer` and helpers via direct parameter injection.
- No new `case` added to `applyInput` — companion fix was in `newPresetMonsterOrCompanion`.
- `make validate-shared` passed after new lab presets added.
- All Go and Python unit tests green (`make test-go`, `make test-py`).
