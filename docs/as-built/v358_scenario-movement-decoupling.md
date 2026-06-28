# v358 as-built — scenario-movement-decoupling

## What shipped

Test-hygiene slice: audited all 203 bot scenario files, added movement-budget policy, introduced
`start_level` lab worlds, and refactored extended + pack scenarios to drop incidental navigation setup.

## Deliverables

| Artifact | Purpose |
|----------|---------|
| `tools/bot/scenario_movement_audit.py` | Discovers scenarios, counts movement steps, writes/validates audit TSV |
| `tools/test_scenario_movement_audit.py` | Pytest gate: every scenario file has an audit row; allowlist ids are `contract` |
| `docs/progress/scenario-movement-audit.tsv` | Committed inventory (`scenario_path`, before/after step counts, classification) |
| `monster_rarity_lab`, `dungeon_depth_one_lab` | `shared/rules/worlds.v0.json` labs with `start_level` for depth without stair setup |

## Refactor highlights

**Protocol (stairs / depth setup → `start_level` labs):**

- `monster_rarity_loot_scaling`, `pack_aggro_and_dungeon_packs`, `elite_minion_pack_ai`,
  `dungeon_monsters`, `dungeon_elite_side_objective`, `treasure_classes_and_guarded_chests`,
  `random_quest_reward_floor`, `dungeon_combat_perf_probe`

**Protocol (spawn-adjacent labs / trim setup movement):**

- `survival_reactive_unique_effects` (monster moved adjacent in `unique_reprisal_lab`)
- `skill_points_and_magic_bolt`, `upgrade_resource_drop` (lab layout + step removal)
- `mercenary_hiring_board`, `companion_stance_command`, `mercenary_offer_variants`
- `purple_town_unique_chest` → `vendor_lab` + direct `action_entity`

**Client (click_floor walks → `click_entity` / `click_loot_item`):**

- Blacksmith suite (`blacksmith_upgrade_ui`, recipe selector/history, material wallet, armor recipe)
- Mystery seller client scenarios
- `wall_floor_dungeon_rollout` → `generated_wall_lab` (no stair click walk)

**Pack impact:** `elite_minion_pack_ai` stair step removed; no pack scenarios gained incidental movement.

**Deletions:** None — no strict duplicates met deletion policy without coverage loss.

**Task 7 skipped:** No `ensure_at_level` helper (<5 scenarios needed identical teleporter chains).

## Policy

- `docs/progress/scenario-catalog.md` — **Movement budget** section + allowlist
- `CLAUDE.md` — incidental movement forbidden off-allowlist

## Verification

```bash
make validate-shared
.venv/bin/pytest tools/test_scenario_movement_audit.py tools/test_ci_pack.py -q
make bot scenario=monster_rarity_loot_scaling
make bot scenario=elite_minion_pack_ai
make bot-client SCENARIO=blacksmith_upgrade_ui HEADLESS=1
make maintainability
make ci  # green
make ci-full  # extended matrix: 5 pre-existing protocol failures unrelated to v358 refactors
```

## Manual

```bash
make play
# Town services (blacksmith, vendor, mystery seller) — no long approach walks in unrelated proofs.
```
