# v350 Spec — CI-Full Green

Status: Ready for planning  
Date: 2026-06-26  
Codename: `ci-full-green`  
Baseline: v349 `movement-tick-smoothing` on `main` @ `82422216`

## Purpose

Restore a **fully green `make ci-full`** by fixing every extended-scenario failure recorded in the v349
engineering review. Merge-gate `make ci` is already green; this slice is **regression recovery only** —
no new player-facing features.

Evidence: [`docs/reviews/20260626_v349-ci-full-failures.md`](../reviews/20260626_v349-ci-full-failures.md)
(15 failures: 12 protocol + 3 client bot).

## Non-goals

- No new gameplay systems, balance passes, or protocol/schema version bumps unless a fix requires a
  contract correction.
- No `main.gd` / `sim.go` coordinator extractions (v337 review paydown stays deferred).
- No CI pack membership changes unless a scenario is shortened and still merge-worthy (unlikely).
- No full-matrix reshuffle of extended scenarios beyond the 15 listed failures.
- No re-run of `make ci-full` on every intermediate fix (targeted scenario reruns during development).

## Failure inventory (authoritative)

All scenarios are `"ci_tier": "extended"`.

### Protocol bot (12)

| # | Scenario ID | Scenario JSON | v349 ci-full time | Diagnostic symptom (2026-06-26 reruns) |
|---|-------------|---------------|-------------------|--------------------------------------|
| 1 | `full_equipment` | `tools/bot/scenarios/19_full_equipment.json` | 4.96s | `walk_toward exhausted 40 ticks` during `pick_up_loot` (`equipment_lab`, target ~`{x:3,y:6}`) |
| 2 | `monster_rarity_loot_scaling` | `tools/bot/scenarios/21_monster_rarity_loot_scaling.json` | 29.53s | Scenario budget: **28.23s > 28.00s** `max_elapsed_s` (functional steps pass) |
| 3 | `inventory_capacity_and_paper_doll` | `tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json` | 4.17s | `walk_toward exhausted 40 ticks` during `pick_up_loot` |
| 4 | `coop_rewards_and_scaling` | `tools/bot/scenarios/34_coop_rewards_and_scaling.json` | 31.87s | `co-op attack did not kill dungeon_mob` |
| 5 | `gold_autopickup_shared_loot` | `tools/bot/scenarios/35_gold_autopickup_shared_loot.json` | 20.29s | Co-op wait timeout: shared gold not removed/awarded |
| 6 | `pack_aggro_and_dungeon_packs` | `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json` | 32.32s | Failed in ci-full; **passed on isolated rerun** — treat as flake/budget until reproduced |
| 7 | `offensive_unique_effects` | `tools/bot/scenarios/54_offensive_unique_effects.json` | 7.36s | `action_until_combat_event: player died` |
| 8 | `survival_reactive_unique_effects` | `tools/bot/scenarios/55_survival_reactive_unique_effects.json` | 27.18s | `move_intent rejected: player_dead` |
| 9 | `resource_support_mobility_unique_effects` | `tools/bot/scenarios/56_resource_support_mobility_unique_effects.json` | **762.62s** | `walk_to_monster` pathfind leg ~12+ min; scenario budget **762s > 15s** default |
| 10 | `dungeon_elite_side_objective` | `tools/bot/scenarios/68_dungeon_elite_side_objective.json` | 66.65s | Scenario budget: **58.00s > 52.00s** `max_elapsed_s` |
| 11 | `live_rare_combat_affixes` | `tools/bot/scenarios/79_live_rare_combat_affixes.json` | 4.16s | `walk_toward exhausted 40 ticks` during `pick_up_loot` (`equipment_lab`) |
| 12 | `mercenary_foundation` | `tools/bot/scenarios/86_mercenary_foundation.json` | 11.47s | Missing `monster_damaged` from companion `mercenary_guard` → `combat_lab_soft_target` |

### Client bot (3)

| # | Scenario ID | Scenario JSON | v349 failing step | Symptom |
|---|-------------|---------------|-------------------|---------|
| 1 | `character_stats_panel` | `tools/bot/scenarios/client/09_character_stats_panel.json` | step 4 `wait_event` (1.0s) | `monster_killed` not seen within 1s after `click_entity_until_event` |
| 2 | `blacksmith_armor_recipe` | `tools/bot/scenarios/client/70_blacksmith_armor_recipe.json` | step 35 `wait_blacksmith_panel` (15s) | Blacksmith panel state never matched (`cave_mail`, upgrade enabled) |
| 3 | `wall_floor_dungeon_rollout` | `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` | step 2 `wait_wall_layout` (15s) | Stayed level **0**, `wall_count=0` after `stairs_down` click (expected level **-1**) |

## Acceptance criteria

1. **All 15 scenarios pass** individually via focused reruns (see plan).
2. **`make ci-full` exits 0** on `main` after fixes (single confirmation run at slice end).
3. **`make ci` remains green** (merge pack unchanged unless a fix requires pack scenario touch).
4. Fixes prefer **root-cause** corrections (movement, combat, coop, client wiring) over permanent
   budget inflation; budget bumps are allowed only when the scenario contract is correct but
   wall-clock drift is environmental (document in as-built).
5. No determinism lint regressions; no new `helpers=globals()` extraction coupling.
6. Touched grandfathered files stay within maintainability baseline.

## Scope and likely files

| Cluster | Likely cause | Likely files |
|---------|--------------|--------------|
| Loot walk failures (4 protocol) | Bot `walk_toward` / `pick_up_loot` vs movement tuning or lab layout | `tools/bot/movement_runtime.py`, `tools/bot/run.py`, `shared/rules/worlds.v0.json`, `shared/rules/main_config.v0.json` (movement speed) |
| Pathfind hang (`resource_support`) | `walk_to_monster` default `pathfind=true` + `move_until_entity_in_range` on long lab | `tools/bot/run.py`, `tools/bot/movement_runtime.py`, scenario `56_*.json`, `unique_momentum_lab` preset |
| Scenario budgets (3 protocol) | `max_elapsed_s` too tight vs stair/dungeon traversal | `21_*.json`, `42_*.json`, `68_*.json`, `56_*.json` |
| Unique combat deaths (2 protocol) | Unique effect / combat tuning regression | `server/internal/game/` combat + unique handlers, `shared/rules/` uniques |
| Coop loot (2 protocol) | Shared gold / kill credit in multi-peer sessions | `server/internal/game/` coop loot, `tools/bot/` coop helpers |
| Mercenary combat (1 protocol) | Companion damage events not emitted | `server/internal/game/` mercenary/companion combat |
| Stats panel combat (1 client) | Client click-to-kill / event timing vs 1s `wait_event` | `client/scripts/main.gd`, bot scenario `09_*.json`, combat presentation |
| Blacksmith panel (1 client) | Blacksmith UI state / interactable routing | `client/scripts/` blacksmith panel, stash/shop flow |
| Wall layout (1 client) | Floor transition or wall renderer not updating on descent | `client/scripts/wall_renderer.gd`, `main.gd` level transition, `79_*.json` |

## Test and bot proof

Per-scenario targeted verification during implementation:

```bash
# Protocol (server + db required)
make bot scenario=<id> VERBOSE=1

# Client
make bot-client SCENARIO=<id> HEADLESS=1 VERBOSE=1
```

Final gate:

```bash
make ci-full
```

## Open questions and risks

| Item | Risk | Default for plan |
|------|------|------------------|
| `pack_aggro` passed on isolated rerun | May be flake vs real bug | Re-run 3× in plan; fix only if reproducible |
| Movement regression vs scenario drift | v332+ pathfinding / v349 client smoothing should not affect protocol bot | Investigate server `move_intent` acceptance first |
| `resource_support` 762s | Blocks full matrix for ~13 min per failure | Highest priority; likely scenario `pathfind: false` or lab compaction |
| Budget-only failures | Easy to “fix” by inflating `max_elapsed_s` | Allow small bump (+10–15%) only after proving functional pass |
| `character_stats_panel` 1s `wait_event` | May be redundant after `click_entity_until_event` | Prefer scenario step fix before client combat changes |

## Asset decision

Not applicable — recovery slice; reuse existing labs, UI, and bot patterns. **Adopt** existing scenario JSON structure; **reject** new external assets.
