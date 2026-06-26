# v349 review — `make ci-full` failure inventory

**Baseline:** `main` @ `82422216`  
**Run:** 2026-06-26, **37m05s**, exit 1  
**Log:** `/tmp/arpg-ci-full-review.log` (quiet mode; use `VERBOSE=1` on focused reruns for assertion text)

**Summary:** 15 failing scenarios total — **12 protocol (extended)** + **3 client bot (extended)**.  
Steps 1–8 and 11/11 passed. Replay step ran for passing scenarios only (protocol step failed before replay gate on failed matrix).

---

## Protocol bot failures (12) — step 9/11

All are `"ci_tier": "extended"` (not in merge-gate `ci_pack.json`).

| # | Scenario ID | JSON path | Elapsed | Focused rerun |
|---|-------------|-----------|---------|---------------|
| 1 | `full_equipment` | `tools/bot/scenarios/19_full_equipment.json` | 4.96s | `make db-up && make server` then `make bot scenario=full_equipment VERBOSE=1` |
| 2 | `monster_rarity_loot_scaling` | `tools/bot/scenarios/21_monster_rarity_loot_scaling.json` | 29.53s | `make bot scenario=monster_rarity_loot_scaling VERBOSE=1` |
| 3 | `inventory_capacity_and_paper_doll` | `tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json` | 4.17s | `make bot scenario=inventory_capacity_and_paper_doll VERBOSE=1` |
| 4 | `coop_rewards_and_scaling` | `tools/bot/scenarios/34_coop_rewards_and_scaling.json` | 31.87s | `make bot scenario=coop_rewards_and_scaling VERBOSE=1` |
| 5 | `gold_autopickup_shared_loot` | `tools/bot/scenarios/35_gold_autopickup_shared_loot.json` | 20.29s | `make bot scenario=gold_autopickup_shared_loot VERBOSE=1` |
| 6 | `pack_aggro_and_dungeon_packs` | `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json` | 32.32s | `make bot scenario=pack_aggro_and_dungeon_packs VERBOSE=1` |
| 7 | `offensive_unique_effects` | `tools/bot/scenarios/54_offensive_unique_effects.json` | 7.36s | `make bot scenario=offensive_unique_effects VERBOSE=1` |
| 8 | `survival_reactive_unique_effects` | `tools/bot/scenarios/55_survival_reactive_unique_effects.json` | 27.18s | `make bot scenario=survival_reactive_unique_effects VERBOSE=1` |
| 9 | `resource_support_mobility_unique_effects` | `tools/bot/scenarios/56_resource_support_mobility_unique_effects.json` | **762.62s** | `make bot scenario=resource_support_mobility_unique_effects VERBOSE=1` |
| 10 | `dungeon_elite_side_objective` | `tools/bot/scenarios/68_dungeon_elite_side_objective.json` | 66.65s | `make bot scenario=dungeon_elite_side_objective VERBOSE=1` |
| 11 | `live_rare_combat_affixes` | `tools/bot/scenarios/79_live_rare_combat_affixes.json` | 4.16s | `make bot scenario=live_rare_combat_affixes VERBOSE=1` |
| 12 | `mercenary_foundation` | `tools/bot/scenarios/86_mercenary_foundation.json` | 11.47s | `make bot scenario=mercenary_foundation VERBOSE=1` |

**Bot runner summary line (authoritative list):**

```text
full_equipment, monster_rarity_loot_scaling, inventory_capacity_and_paper_doll,
coop_rewards_and_scaling, gold_autopickup_shared_loot, pack_aggro_and_dungeon_packs,
offensive_unique_effects, survival_reactive_unique_effects,
resource_support_mobility_unique_effects, dungeon_elite_side_objective,
live_rare_combat_affixes, mercenary_foundation
```

**Note:** Quiet `make ci-full` does **not** print per-scenario assertion messages for protocol failures. Use `VERBOSE=1 make bot scenario=<id>` (server + db required) to capture the failing step/assertion.

**Domain grouping (for refactor batching):**

- **Equipment / inventory:** `full_equipment`, `inventory_capacity_and_paper_doll`
- **Loot / rarity / affixes:** `monster_rarity_loot_scaling`, `live_rare_combat_affixes`, unique suite (`offensive_*`, `survival_*`, `resource_support_*`)
- **Coop / shared loot:** `coop_rewards_and_scaling`, `gold_autopickup_shared_loot`
- **Dungeon AI / packs:** `pack_aggro_and_dungeon_packs`, `dungeon_elite_side_objective`
- **Mercenaries:** `mercenary_foundation`

---

## Client bot failures (3) — step 10/11

All are `"ci_tier": "extended"`. **87 passed**, 3 failed.

| # | Scenario ID | JSON path | Elapsed | Failing step | Symptom (from log) |
|---|-------------|-----------|---------|--------------|-------------------|
| 1 | `character_stats_panel` | `tools/bot/scenarios/client/09_character_stats_panel.json` | (budget 109s) | **step 4/31** `wait_event` | Timeout **1.0s** waiting for `monster_killed` after `click_entity_until_event` on monster |
| 2 | `blacksmith_armor_recipe` | `tools/bot/scenarios/client/70_blacksmith_armor_recipe.json` | (budget 300s) | **step 35/43** `wait_blacksmith_panel` | Timeout **15.0s** — expected blacksmith panel with `item_def_id=cave_mail`, `item_level=0`, `upgrade_enabled=true`, `resource_wallet_count=1` |
| 3 | `wall_floor_dungeon_rollout` | `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` | (budget 90s) | **step 2/4** `wait_wall_layout` | Timeout **15.0s** — expected `current_level=-1`, `at_least=8` walls, `generated_at_least=4`; saw **level=0**, `wall_count=0`, `generated=0` after teleporter click |

**Focused reruns (server on :18081 or default per `make ci`):**

```bash
make db-up && ARPG_ADDR=:18081 make server   # background
make bot-client SCENARIO=character_stats_panel HEADLESS=1 VERBOSE=1
make bot-client SCENARIO=blacksmith_armor_recipe HEADLESS=1 VERBOSE=1
make bot-client SCENARIO=wall_floor_dungeon_rollout HEADLESS=1 VERBOSE=1
```

---

## What passed (context)

- **Merge pack (`make ci`):** green (user-confirmed; 36 scenarios).
- **ci-full steps 1–8:** maintainability, validate-shared, validate-assets, determinism lint, Go test/vet, pytest, db+server start.
- **ci-full step 11:** headless `client_smoke` + slice smoke — **PASS**.
- **Client bot:** 87/90 extended client scenarios passed (including `entity_tick_smoothing`, `movement_visual_smoothing`, `client_full_equipment`).

---

## Refactor checklist (no full ci-full until these are green)

Run targeted reruns above; fix; re-run only the touched scenario. When all 15 pass individually, one confirmation `make ci-full` is enough.

```bash
# Quick protocol batch (requires server):
for s in full_equipment monster_rarity_loot_scaling inventory_capacity_and_paper_doll \
  coop_rewards_and_scaling gold_autopickup_shared_loot pack_aggro_and_dungeon_packs \
  offensive_unique_effects survival_reactive_unique_effects \
  resource_support_mobility_unique_effects dungeon_elite_side_objective \
  live_rare_combat_affixes mercenary_foundation; do
  make bot scenario=$s || break
done

# Client trio:
for s in character_stats_panel blacksmith_armor_recipe wall_floor_dungeon_rollout; do
  make bot-client SCENARIO=$s HEADLESS=1 || break
done
```
