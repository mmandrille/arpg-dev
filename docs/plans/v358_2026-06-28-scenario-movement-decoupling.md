# v358 Plan — Scenario Movement Decoupling

Status: Ready for implementation  
Goal: Audit all bot scenarios, eliminate incidental navigation setup, delete strict duplicates, and lock movement-budget authoring policy so path/town tuning stops breaking unrelated proofs.  
Architecture: No gameplay changes. Classify every scenario step against the spec allowlist, then apply the remediation playbook in order (lab relocation → `start_level` → `teleport_to_level` → `debug_progression` → semantic actions → client `click_entity` trims → delete). New depth setups prefer dedicated lab entries in `worlds.v0.json` (pattern: `boss_floor_gate_lab`, `generated_wall_lab`) over stair chains. Assertions drop `visited_levels_contain: 0` when descent is no longer the contract. Optional bot helper only if ≥5 scenarios need the same `ensure_at_level` pattern after lab pass.  
Tech stack: Python bot + pytest audit gate, shared `worlds.v0.json` lab presets, scenario JSON bulk edits, docs policy only (no protocol bump expected).

Spec: [`docs/specs/v358_spec-scenario-movement-decoupling.md`](../specs/v358_spec-scenario-movement-decoupling.md)  
Baseline: v357 `town-night-perimeter`

## Spec review (gate)

| Area | Result |
|------|--------|
| Baseline v358 / builds on v357 | OK — `PROGRESS.md` lists v357 complete |
| Scope / non-goals | OK — test hygiene only; no gameplay tuning |
| Contracts | No protocol/schema bump expected; lab `start_level` already in `worlds.v0.schema.json` |
| Determinism | Lab worlds + pinned seeds; no `game/` hot-path changes |
| Shared rules | Optional new lab world entries only |
| Server authority | N/A — scenarios consume existing sim behavior |
| Animation (ADR-0007) | N/A |
| World presets | Concrete: clone/adjust lab worlds with `start_level` where depth setup replaced |
| Bot proof | Refactors + audit TSV + `make ci-full`; no new merge-gate scenario unless delete leaves pack hole |
| Replay | Refactored scenarios that own replay must still pass; update steps, not relax determinism |
| Client assets | N/A |
| Maintainability | Avoid `run.py` growth; new helper in focused module if needed |
| As-built drift | `teleport_to_level` bot step exists; `start_level` used by `boss_floor_gate_lab` / `generated_wall_lab`; v357 town uses `interactable_def_id` in many client scenarios already |

**Assertion note:** When replacing `use_stair` with `start_level: -N`, remove `visited_levels_contain: 0` (and other transit-only level checks) unless level transition is still under test.

## Baseline and shortcut decision

Reuse:

- `boss_floor_gate_lab` / `generated_wall_lab` — `start_level` + spawn position pattern
- `character_stats_lab`, `equipment_lab`, `unique_*_lab`, `vendor_lab` — spawn-adjacent entities
- `tools/bot/run.py` — `teleport_to_level`, `discover_teleporter`, `debug_progression` seeding
- `tools/bot/ci_pack.py` + `tools/test_ci_pack.py` — pack validation after deletes
- v357 client pattern — `click_entity` + `interactable_def_id` for town services (see `client/35_market_board_ui.json`)

| Option | Decision |
|--------|----------|
| New lab world per depth pattern | **Adopt** when ≥2 scenarios share the same depth+seed contract |
| `teleport_to_level` for setup | **Adopt** when multi-floor state needed and teleporter contract not under test |
| Debug `session_start_level` on session create | **Defer** unless audit finds ≥5 blocked scenarios after lab pass |
| `helpers=globals()` extraction from `run.py` | **Reject** (v55 freeze) |
| Delete redundant long-route scenarios | **Adopt** per spec with `merged_into` in audit |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/progress/scenario-movement-audit.tsv` | Committed audit inventory |
| Create | `tools/bot/scenario_movement_audit.py` | Generate + validate audit TSV against discovered scenarios |
| Create | `tools/test_scenario_movement_audit.py` | Pytest: every scenario file has audit row; allowlist ids classified `contract` |
| Modify | `tools/bot/scenarios/*.json` | Protocol refactors + deletions |
| Modify | `tools/bot/scenarios/client/*.json` | Client setup trims |
| Modify | `shared/rules/worlds.v0.json` | New/adjusted labs (`start_level`, spawn-adjacent layout) |
| Modify | `tools/bot/ci_pack.json` | Only if deletion removes pack coverage |
| Create (optional) | `tools/bot/scenario_setup.py` | `ensure_at_level` helper if audit warrants |
| Create (optional) | `tools/test_scenario_setup.py` | Direct import test for helper |
| Modify | `docs/progress/scenario-catalog.md` | Movement budget section + allowlist |
| Modify | `CLAUDE.md` | One paragraph: incidental movement forbidden off-allowlist |
| Modify | `docs/CODEMAP.md` | Audit script entry |
| Create | `docs/as-built/v358_scenario-movement-decoupling.md` | On `/finish` |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | On `/finish` |

### Priority refactor targets (from spec + grep)

**Protocol — stair descent setup (replace with `start_level` lab or world clone):**

| Scenario | Current setup | Planned fix |
|----------|---------------|-------------|
| `monster_rarity_loot_scaling` | 2× `use_stair` on `dungeon_levels` | `monster_rarity_lab` with `start_level: -2`, seed `v30_monster_rarity`; drop level-0 visit assertion |
| `pack_aggro_and_dungeon_packs` | `use_stair` + `move_until_in_range` | `start_level: -1` on pinned seed world; keep combat pull steps only |
| `elite_minion_pack_ai` (**pack**) | `use_stair` | `elite_minion_pack_lab` or `dungeon_levels` + `start_level: -1`, seed unchanged; drop level-0 visit |
| `dungeon_monsters` | `use_stair` + range walk | `start_level: -1` lab; attack at spawn range |
| `treasure_classes_and_guarded_chests` | `use_stair` | Existing chest lab / `start_level` on pinned floor |
| `skill_points_and_magic_bolt` | `move_until_in_range` after level grind | `debug_progression` + compact lab; remove range walk if dummy adjacent |
| `17_treasure_classes…`, `72_upgrade_resource_drop`, `65_random_quest_reward_floor` | audit each | lab / `start_level` / delete if duplicate |

**Protocol — `move_until_*` setup (replace with spawn-adjacent lab or `action_until_*`):**

| Scenario | Planned fix |
|----------|-------------|
| `survival_reactive_unique_effects` | Place `combat_lab_miss_attacker` in melee range in `unique_reprisal_lab`; remove `move_until_in_range` |
| `resource_support_mobility_unique_effects` | Same pattern — audit lab layout |
| `mercenary_foundation`, `mercenary_hiring_board`, `companion_stance_command` | Town service adjacent spawn in `vendor_lab` / mercenary lab; `click_entity` / `action_entity` only |
| `fog_of_war_radius` (**pack**) | Audit whether scout move is contract; if not, spawn mob in fog boundary lab |
| `companion_ai_foundation` (**pack**) | Trim `move_until_player_position` if companion proof works at spawn |
| `line_of_sight_blockers` (**pack**) | Keep if LOS movement is contract; else shorten to pre-positioned lab |

**Client — replace `click_floor` + `wait_player_near` with `click_entity` when UI is the contract:**

| Scenario | Planned fix |
|----------|-------------|
| `blacksmith_upgrade_ui`, `blacksmith_recipe_selector`, `blacksmith_second_recipe`, `blacksmith_upgrade_history`, `blacksmith_armor_recipe`, `material_wallet_window` | `click_entity` `town_blacksmith` / `town_vendor` / `town_stash` at `vendor_lab` spawn |
| `mystery_seller_core` (client), `mystery_seller_paid_reroll`, `mystery_seller_silhouettes` | `click_entity` `town_mystery_seller` / `town_vendor` |
| `character_stats_panel` | Already uses `click_entity_until_event`; verify timeout not movement-related (increase combat timeout only if needed) |
| `wall_floor_dungeon_rollout` | Use `click_entity` `stairs_down` or protocol `teleport_to_level` setup in client world with `start_level` |

**Movement-contract allowlist — do not strip core proof:**

`vertical_slice`, `gear_before_combat`, `path_maze`, `chase_lab`, `chase_maze`, `leash_lab`, `dungeon_levels`, `teleporter_lab`, `collision_lab`, `reachable_dungeon_obstacles`, `player_path_budget_lab`, `boss_floor_gate`, `flying_navigation_trait`, client: `click_to_move`, `town_floor_click_to_move`, `town_teleporter_auto_approach`, `attack_move_sticky_targeting`, `movement_visual_smoothing`, `entity_tick_smoothing`, `mobility_skill_smoothing`, `melee_lunge_micro_step`, `torch_walk_visual`.

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `tools/bot/run.py` (baseline 4237) — **no new domain logic**; optional one-line dispatch to `scenario_setup.py` only
- [x] Other over-limit file: none expected

Decision:

- [x] Extract `tools/bot/scenario_movement_audit.py` + `tools/bot/scenario_setup.py` (if needed) as focused modules with direct pytest imports
- [x] Defer any `run.py` shrink — touch-to-shrink not required this slice

Verification:

```bash
make maintainability
```

## Task 1 — Audit generator and policy docs

Files:

- Create: `tools/bot/scenario_movement_audit.py`
- Create: `tools/test_scenario_movement_audit.py`
- Create: `docs/progress/scenario-movement-audit.tsv` (generated then committed)
- Modify: `docs/progress/scenario-catalog.md`
- Modify: `CLAUDE.md`

- [x] Step 1.1: Implement audit script that discovers all `tools/bot/scenarios/**/*.json`, counts movement steps (`use_stair`, `walk_to_*`, `move_until_*`, `teleport_to_level`, client `click_floor`, `wait_player_near`), classifies against allowlist, outputs TSV columns per spec.
- [x] Step 1.2: Pytest asserts every scenario file has exactly one audit row; allowlist ids are `contract`; deleted files marked `deleted` with `merged_into`.
- [x] Step 1.3: Add **Movement budget** section to scenario catalog (playbook, allowlist, “no incidental travel” rule).
- [x] Step 1.4: Add CLAUDE.md paragraph extending v32 test-locking for movement setup.

```bash
.venv/bin/pytest tools/test_scenario_movement_audit.py -q
```

## Task 2 — Lab worlds (`start_level` / spawn-adjacent)

Files:

- Modify: `shared/rules/worlds.v0.json`

- [x] Step 2.1: Add `monster_rarity_lab` — `mode: multi_level`, `start_level: -2`, player spawn clear of mobs, reuse seed contract from `21_monster_rarity_loot_scaling.json`.
- [x] Step 2.2: Add or extend labs for pack aggro / elite minion (**pack**) — `start_level: -1`, pinned seeds matching current scenarios.
- [x] Step 2.3: Audit `unique_*_lab` / `vendor_lab` layouts — ensure combat targets and town services within interaction range of spawn (adjust entity positions only, no tuning).
- [x] Step 2.4: Prefer extending existing lab world ids over proliferating one-off worlds; document each new world in audit TSV.

```bash
make validate-shared
```

## Task 3 — Protocol scenario refactors (extended batch)

Files:

- Modify: priority protocol scenarios listed in file map (batch A: v349/v356 class)

- [x] Step 3.1: Refactor `monster_rarity_loot_scaling` → `monster_rarity_lab`, remove stairs, fix assertions.
- [x] Step 3.2: Refactor `pack_aggro_and_dungeon_packs`, `dungeon_elite_side_objective`, unique-effects trio (`survival_*`, `resource_support_*` if still walking), coop/mercenary extended offenders.
- [x] Step 3.3: Run each refactored scenario individually before batch commit.

```bash
make db-up && make server   # background
make bot scenario=monster_rarity_loot_scaling VERBOSE=1
make bot scenario=pack_aggro_and_dungeon_packs VERBOSE=1
make bot scenario=survival_reactive_unique_effects VERBOSE=1
make bot scenario=dungeon_elite_side_objective VERBOSE=1
```

## Task 4 — Protocol scenario refactors (pack + remaining)

Files:

- Modify: `tools/bot/scenarios/77_elite_minion_pack_ai.json` and other pack members with incidental movement
- Modify: remaining `setup-eliminable` protocol scenarios from audit

- [x] Step 4.1: Refactor `elite_minion_pack_ai` (**pack**) — `start_level: -1`, no new stair steps; verify pack still green in `make ci`.
- [x] Step 4.2: Audit pack members: `combat_stat_effects`, `companion_ai_foundation`, `fog_of_war_radius`, `line_of_sight_blockers`, `quest_town_turn_in`, `archer_retreat_ai` — trim only setup movement; net step count ≤ before.
- [x] Step 4.3: Complete remaining extended protocol scenarios until ≥80% of `setup-eliminable` extended rows show `movement_steps_after < movement_steps_before` or `deleted`.

```bash
make bot scenario=elite_minion_pack_ai VERBOSE=1
make bot scenario=ci VERBOSE=1
```

## Task 5 — Client bot refactors

Files:

- Modify: `tools/bot/scenarios/client/*.json` flagged `setup-eliminable` in audit (blacksmith, mystery seller, market flows, etc.)

- [x] Step 5.1: Replace incidental `click_floor` + `wait_player_near` approach hops with `click_entity` + `interactable_def_id` where auto-approach is not the contract (pattern: `client/35_market_board_ui.json`).
- [x] Step 5.2: For dungeon-depth client proofs (`wall_floor_dungeon_rollout`, elite HUD scenarios), use `generated_wall_lab` / `start_level` worlds or `click_entity` `stairs_down` instead of long floor walks.
- [x] Step 5.3: Leave allowlist client movement scenarios unchanged except v357 coordinate fixes.
- [x] Step 5.4: Re-run v349 client failure class individually.

```bash
make bot-client SCENARIO=blacksmith_upgrade_ui HEADLESS=1 VERBOSE=1
make bot-client SCENARIO=character_stats_panel HEADLESS=1 VERBOSE=1
make bot-client SCENARIO=wall_floor_dungeon_rollout HEADLESS=1 VERBOSE=1
make bot-client SCENARIO=ci HEADLESS=1
```

## Task 6 — Deletion pass and CI pack hygiene

Files:

- Delete: redundant scenario JSON files per audit
- Modify: `tools/bot/ci_pack.json` (only if required)
- Modify: `docs/progress/scenario-movement-audit.tsv`

- [x] Step 6.1: Delete scenarios that are strict duplicates; record `merged_into` + `reason` in audit.
- [x] Step 6.2: If a deleted file was in `ci_pack.json`, promote surviving lab proof or demote consciously (budget-neutral).
- [x] Step 6.3: Re-run audit script; confirm no orphan scenario files.

```bash
python3 -c "from tools.bot.ci_pack import validate_ci_pack; validate_ci_pack()"
.venv/bin/pytest tools/test_ci_pack.py tools/test_scenario_movement_audit.py -q
```

## Task 7 — Optional bot setup helper (only if audit triggers)

Files:

- Create (conditional): `tools/bot/scenario_setup.py`, `tools/test_scenario_setup.py`
- Modify (conditional): `tools/bot/run.py` — single dispatch line for `ensure_at_level` action

- [ ] Step 7.1: If ≥5 scenarios need identical `discover_teleporter` + `teleport_to_level` chains, add `ensure_at_level` bot step wrapping existing intents.
- [ ] Step 7.2: Unit test imports `scenario_setup` directly without `run.py`.

Skip entire task if audit completes with lab/`start_level` only. **Skipped** — lab/`start_level` covered all refactors.

## Task 8 — Lifecycle docs and CI

Files:

- Modify: `docs/CODEMAP.md`
- Modify: `PROGRESS.md`, `docs/progress/slice-lifecycle.md` (on `/finish`)
- Create: `docs/as-built/v358_scenario-movement-decoupling.md` (on `/finish`)

- [x] Step 8.1: Update CODEMAP with audit module.
- [x] Step 8.2: Final audit TSV committed with before/after movement step totals.
- [x] Step 8.3: Summarize deletions and pack impact in as-built.

```bash
make maintainability
make ci
make ci-full
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared` (if `shared/rules/worlds.v0.json` touched)
- [x] `.venv/bin/pytest tools/test_scenario_movement_audit.py tools/test_ci_pack.py -q`
- [x] `make ci`
- [x] `make ci-full` (required before `/finish`) — **red**: pre-existing extended failures (`purple_town_unique_chest` v357 coords, `barbarian_class_foundation`, unique-skill trio); all v358 refactors pass individually and in `make ci`

Manual:

```bash
make play
# Spot-check town services after v357 — blacksmith/vendor/mystery flows without long walks.
```

## Deferred scope

- v120 full Go/GDScript tuning-pin audit (separate slice).
- Debug `session_start_level` on session create (only if Task 7 triggers).
- `run.py` orchestrator split / maintainability paydown on grandfathered coordinators.
- New merge-gate scenarios (none unless pack hole from deletion).

## Handoff

Work stays on current branch (`main`). Next step:

```
/execute docs/plans/v358_2026-06-28-scenario-movement-decoupling.md
```
