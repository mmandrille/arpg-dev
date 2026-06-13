# Project progress & slice lifecycle

**Read this file at the start of every new task** before writing specs, plans, or code.
It is the canonical snapshot of where the project stands and what is still open. Per-slice as-built
summaries live in [`docs/as-built/`](docs/as-built/).

Last updated: 2026-06-13

---

## Current status

| Field | Value |
|-------|-------|
| **Latest completed slice** | v119 — `live-unique-drops-all-effects` |
| **Active branch** | `main` |
| **CI gate** | `make ci` green on 2026-06-13 |
| **Next slice** | v120 — `tuning-friendly-rule-tests` |
| **Last engineering review** | v110 — [`docs/reviews/20260613_v110-overview.md`](docs/reviews/20260613_v110-overview.md) (2026-06-13) |
| **Next engineering review** | v120 (~every 10 slices) |

### Slice numbering note

ADR-0001 sometimes calls the first slice **v1**; repo lifecycle labels use **v0–v9**
(`v0` = first playable). **Spec and plan filenames** use a `vN_` prefix for execution order:

```text
v1_* = first-playable    v5_* = resume-state    v8_* = equipped-weapon-damage
v2_* = equip-and-see-it  v6_* = visual-bot
v3_* = animate-and-react v7_* = gear-before-combat v9_* = solid-collision
v4_* = take-a-hit        v10_* = click-action-and-melee-range
v11_* = click-to-move-and-auto-path
v12_* = ranged-projectile-combat
v13_* = inventory-ui
v14_* = godot-client-bot
v15_* = item-visuals-and-loot-presentation
v16_* = use-consumable
v17_* = monster-chase-movement
v18_* = dungeon-levels-and-stairs
v19_* = teleporters-and-waypoint-ui
v20_* = play-session-loop
v21_* = dungeon-monster-combat
v22_* = character-scoped-persistence
v23_* = item-templates-and-rolled-drops
v24_* = main-menu-and-character-start
v25_* = treasure-classes-and-guarded-chests
v26_* = character-stats-and-leveling
v27_* = hold-click-controls
v28_* = full-equipment-and-belt-hotbar
v29_* = dungeon-equipment-drop-expansion
v30_* = monster-rarity-and-loot-scaling
v31_* = combat-stat-effects-and-feedback
v32_* = test-floor-and-resilient-scenarios
v33_* = true-coop-session
v34_* = model-reaction-polish
v35_* = boss-floor-gate
v36_* = inventory-paper-doll-capacity
v37_* = combat-control-and-boss-ai-fixes
v38_* = session-browser-and-uncapped-coop-menu
v39_* = ui-currency-and-mana-polish
v40_* = reachable-dungeon-obstacles
v41_* = town-vendor-gold-sink
v42_* = vendor-appraisal-and-item-comparison
v43_* = equipment-requirements-and-preview
v44_* = skill-points-and-magic-bolt
v45_* = menu-create-join-flow
v46_* = client-join-game-proof
v47_* = shop-stock-lifecycle
v48_* = coop-rewards-and-scaling
v49_* = gold-autopickup-and-shared-loot-rules
v50_* = account-stash-storage
v51_* = mystery-seller-core
v52_* = ranged-monster-ai
v53_* = boss-health-bar-ui
v54_* = character-select-summaries
v55_* = consolidation-and-quality-gates
v56_* = monster-attack-cadence
v57_* = boss-phase-readability
v58_* = boss-pattern-variety
v59_* = data-driven-skill-catalog
v60_* = data-driven-content-library-manifest
v61_* = rage-and-heal-skills
v62_* = monster-depth-stat-scaling
v63_* = runtime-sim-error-construction
v64_* = mystery-seller-paid-reroll
v65_* = stash-search-and-sorting
v66_* = progress-backlog-hygiene
v67_* = boss-kill-reward-polish
v68_* = market-stash-listing-foundation
v69_* = character-class-foundation
v70_* = class-skill-and-item-gates
v71_* = class-picker-and-sprites
v72_* = monster-visual-catalog
v73_* = draggable-window-foundation
v74_* = gameplay-window-chrome
v75_* = persistent-window-layout
v76_* = main-config-foundation
v77_* = main-config-derived-gameplay
v78_* = main-config-drop-profiles
v79_* = elite-pack-roles
v80_* = combat-threat-readability
v81_* = paladin-holy-shield
v82_* = realtime-fanout-level-snapshot
v83_* = defensive-client-envelope-payloads
v84_* = client-bot-step-registry
v85_* = skill-demo-catalog
v86_* = skill-visual-command
v87_* = skill-visual-matrix
v88_* = skill-visual-rank-seeding
v89_* = class-second-combat-skills
v90_* = text-catalog-foundation
v91_* = spanish-language-selector
v92_* = town-bishop-respec
v93_* = market-multi-item-offers
v94_* = item-upgrade-starter
v95_* = unique-item-catalog-seed
v96_* = town-presentation-polish
v97_* = class-starter-loadouts
v98_* = rogue-class-foundation
v99_* = rogue-skill-mechanics
v100_* = damage-types-and-resistances
v101_* = undead-skeleton-poison-immunity
v102_* = class-bot-visual-scenarios
v103_* = unique-effect-catalog-foundation
v104_* = unique-drop-roll-contract
v105_* = unique-burn-effect-live
v106_* = offensive-unique-effects
v107_* = survival-reactive-unique-effects
v108_* = resource-support-mobility-unique-effects
v109_* = permanent-death-corpse-recovery
v110_* = item-upgrade-repeat-action
v111_* = market-purchase-and-delivery
v112_* = elite-aura-foundation
v113_* = elite-aura-readability
v114_* = market-board-ui
v115_* = market-purchase-ui
v116_* = elite-aura-radius-preview
v117_* = market-active-offer-ui
v118_* = blacksmith-upgrade-ui
v119_* = live-unique-drops-all-effects
```

Pattern: `docs/specs/vN_spec-<codename>.md`, `docs/plans/vN_<YYYY-MM-DD>-<codename>.md`.

### Periodic engineering reviews

Every **~10 completed slices**, pause for a repo-wide engineering review under [`docs/reviews/`](docs/reviews/).
Use the milestone slice number in filenames and headings (e.g. v50, v60, v70 — v60 is the latest pass).

**When to write:** after the milestone slice ships and `make ci` is green — typically as part of `/finish`
close-out or as a dedicated review task before `/next` proposes the next batch.

**Minimum set** (follow the v53 pattern):

| File | Focus |
|------|-------|
| `docs/reviews/YYYYMMDD_vN-overview.md` | Executive summary, scorecard, cross-cutting themes |
| `docs/reviews/backend/YYYYMMDD_vN-backend.md` | Go server / `internal/game` |
| `docs/reviews/client/YYYYMMDD_vN-client.md` | Godot client |
| `docs/reviews/extras/YYYYMMDD_vN-shared-tooling-and-process.md` | `shared/`, `tools/`, SDD process |

Update **Last engineering review** / **Next engineering review** in the table above when a review lands.
Feed actionable findings into open gaps or the next slice briefs — reviews are input to `/next`, not shelfware.

---

## Slice lifecycle

Slices are small, end-to-end proofs. Each ships: shared contracts → Go sim → Godot client →
Python bot/smoke → golden fixtures → `make ci` green.

```text
v0 first-playable ──► v2 equip-and-see-it ──► v3 animate-and-react ──► v4 take-a-hit ──► v5 resume-state ──► v6 visual-bot-scenarios ──► v7 gear-before-combat ──► v8 equipped-weapon-damage ──► v9 solid-collision ──► v10 click-action ──► v11 auto-path ──► v12 ranged-projectiles ──► v13 inventory-ui
   (architecture)        (visual pipeline)         (skeletal anims)         (player damage)      (resume replay)      (visual replay playlist)        (world presets)              (weapon damage)             (walls + bodies)
        │                      │                        │                        │                         │                         │                              │                              │                         │
     main ✓                  main ✓                    main ✓                    main ✓              branch ✓                  branch ✓                       branch ✓                       branch ✓                  branch ✓                  branch ✓
```

| Slice | Codename | Status | Spec | Plan | As-built |
|-------|----------|--------|------|------|----------|
| **v0** | `first-playable-vertical-slice` | Complete (on `main`) | [`v1_spec-first-playable-vertical-slice.md`](docs/specs/v1_spec-first-playable-vertical-slice.md) | [`v1_2026-06-05-first-playable-vertical-slice.md`](docs/plans/v1_2026-06-05-first-playable-vertical-slice.md) | [`as-built`](docs/as-built/v0_first-playable-vertical-slice.md) |
| **v2** | `equip-and-see-it` | Complete (on `main`) | [`v2_spec-equip-and-see-it.md`](docs/specs/v2_spec-equip-and-see-it.md) | [`v2_2026-06-05-equip-and-see-it.md`](docs/plans/v2_2026-06-05-equip-and-see-it.md) | [`as-built`](docs/as-built/v2_equip-and-see-it.md) |
| **v3** | `animate-and-react` | Complete (on `main`) | [`v3_spec-animate-and-react.md`](docs/specs/v3_spec-animate-and-react.md) | [`v3_2026-06-05-animate-and-react.md`](docs/plans/v3_2026-06-05-animate-and-react.md) | [`as-built`](docs/as-built/v3_animate-and-react.md) |
| **v4** | `take-a-hit` | Complete (on `main`) | [`v4_spec-take-a-hit.md`](docs/specs/v4_spec-take-a-hit.md) | [`v4_2026-06-05-take-a-hit.md`](docs/plans/v4_2026-06-05-take-a-hit.md) | [`as-built`](docs/as-built/v4_take-a-hit.md) |
| **v5** | `resume-authoritative-state` | Complete (`make ci` green) | [`v5_spec-resume-authoritative-state.md`](docs/specs/v5_spec-resume-authoritative-state.md) | [`v5_2026-06-05-resume-authoritative-state.md`](docs/plans/v5_2026-06-05-resume-authoritative-state.md) | [`as-built`](docs/as-built/v5_resume-authoritative-state.md) |
| **v6** | `visual-bot-scenario-runner` | Complete (`make ci` green) | [`v6_spec-visual-bot-scenario-runner.md`](docs/specs/v6_spec-visual-bot-scenario-runner.md) | [`v6_2026-06-05-visual-bot-scenario-runner.md`](docs/plans/v6_2026-06-05-visual-bot-scenario-runner.md) | [`as-built`](docs/as-built/v6_visual-bot-scenario-runner.md) |
| **v7** | `gear-before-combat-scenario` | Complete (`make ci` green) | [`v7_spec-gear-before-combat-scenario.md`](docs/specs/v7_spec-gear-before-combat-scenario.md) | [`v7_2026-06-05-gear-before-combat-scenario.md`](docs/plans/v7_2026-06-05-gear-before-combat-scenario.md) | [`as-built`](docs/as-built/v7_gear-before-combat-scenario.md) |
| **v8** | `equipped-weapon-damage` | Complete (`make ci` green) | [`v8_spec-equipped-weapon-damage.md`](docs/specs/v8_spec-equipped-weapon-damage.md) | [`v8_2026-06-05-equipped-weapon-damage.md`](docs/plans/v8_2026-06-05-equipped-weapon-damage.md) | [`as-built`](docs/as-built/v8_equipped-weapon-damage.md) |
| **v9** | `solid-collision-and-obstacles` | Complete (`make ci` green) | [`v9_spec-solid-collision-and-obstacles.md`](docs/specs/v9_spec-solid-collision-and-obstacles.md) | [`v9_2026-06-05-solid-collision-and-obstacles.md`](docs/plans/v9_2026-06-05-solid-collision-and-obstacles.md) | [`as-built`](docs/as-built/v9_solid-collision-and-obstacles.md) |
| **v10** | `click-action-and-melee-range` | Complete (`make ci` green) | [`v10_spec-click-action-and-melee-range.md`](docs/specs/v10_spec-click-action-and-melee-range.md) | [`v10_2026-06-05-click-action-and-melee-range.md`](docs/plans/v10_2026-06-05-click-action-and-melee-range.md) | [`as-built`](docs/as-built/v10_click-action-and-melee-range.md) |
| **v11** | `click-to-move-and-auto-path` | Complete (`make ci` green) | [`v11_spec-click-to-move-and-auto-path.md`](docs/specs/v11_spec-click-to-move-and-auto-path.md) | [`v11_2026-06-05-click-to-move-and-auto-path.md`](docs/plans/v11_2026-06-05-click-to-move-and-auto-path.md) | [`as-built`](docs/as-built/v11_click-to-move-and-auto-path.md) |
| **v12** | `ranged-projectile-combat` | Complete (`make ci` green) | [`v12_spec-ranged-projectile-combat.md`](docs/specs/v12_spec-ranged-projectile-combat.md) | [`v12_2026-06-05-ranged-projectile-combat.md`](docs/plans/v12_2026-06-05-ranged-projectile-combat.md) | [`as-built`](docs/as-built/v12_ranged-projectile-combat.md) |
| **v13** | `inventory-ui` | Complete (`make ci` green) | [`v13_spec-inventory-ui.md`](docs/specs/v13_spec-inventory-ui.md) | [`v13_2026-06-05-inventory-ui.md`](docs/plans/v13_2026-06-05-inventory-ui.md) | [`as-built`](docs/as-built/v13_inventory-ui.md) |
| **v14** | `godot-client-bot` | Complete (`make ci` green) | [`v14_spec-godot-client-bot.md`](docs/specs/v14_spec-godot-client-bot.md) | [`v14_2026-06-02-godot-client-bot.md`](docs/plans/v14_2026-06-02-godot-client-bot.md) | [`as-built`](docs/as-built/v14_godot-client-bot.md) |
| **v15** | `item-visuals-and-loot-presentation` | Complete (`make ci` green) | [`v15_spec-item-visuals-and-loot-presentation.md`](docs/specs/v15_spec-item-visuals-and-loot-presentation.md) | [`v15_2026-06-06-item-visuals-and-loot-presentation.md`](docs/plans/v15_2026-06-06-item-visuals-and-loot-presentation.md) | [`as-built`](docs/as-built/v15_item-visuals-and-loot-presentation.md) |
| **v16** | `use-consumable` | Complete (`make ci` green) | [`v16_spec-use-consumable.md`](docs/specs/v16_spec-use-consumable.md) | [`v16_2026-06-06-use-consumable.md`](docs/plans/v16_2026-06-06-use-consumable.md) | [`as-built`](docs/as-built/v16_use-consumable.md) |
| **v17** | `monster-chase-movement` | Complete (`make ci` green) | [`v17_spec-monster-chase-movement.md`](docs/specs/v17_spec-monster-chase-movement.md) | [`v17_2026-06-06-monster-chase-movement.md`](docs/plans/v17_2026-06-06-monster-chase-movement.md) | [`as-built`](docs/as-built/v17_monster-chase-movement.md) |
| **v18** | `dungeon-levels-and-stairs` | Complete (`make ci` green) | [`v18_spec-dungeon-levels-and-stairs.md`](docs/specs/v18_spec-dungeon-levels-and-stairs.md) | [`v18_2026-06-06-dungeon-levels-and-stairs.md`](docs/plans/v18_2026-06-06-dungeon-levels-and-stairs.md) | [`as-built`](docs/as-built/v18_dungeon-levels-and-stairs.md) |
| **v19** | `teleporters-and-waypoint-ui` | Complete (`make ci` green) | [`v19_spec-teleporters-and-waypoint-ui.md`](docs/specs/v19_spec-teleporters-and-waypoint-ui.md) | [`v19_2026-06-06-teleporters-and-waypoint-ui.md`](docs/plans/v19_2026-06-06-teleporters-and-waypoint-ui.md) | [`as-built`](docs/as-built/v19_teleporters-and-waypoint-ui.md) |
| **v20** | `play-session-loop` | Complete (`make ci` green) | [`v20_spec-play-session-loop.md`](docs/specs/v20_spec-play-session-loop.md) | [`v20_2026-06-06-play-session-loop.md`](docs/plans/v20_2026-06-06-play-session-loop.md) | [`as-built`](docs/as-built/v20_play-session-loop.md) |
| **v21** | `dungeon-monster-combat` | Complete (`make ci` green) | [`v21_spec-dungeon-monster-combat.md`](docs/specs/v21_spec-dungeon-monster-combat.md) | [`v21_2026-06-06-dungeon-monster-combat.md`](docs/plans/v21_2026-06-06-dungeon-monster-combat.md) | [`as-built`](docs/as-built/v21_dungeon-monster-combat.md) |
| **v22** | `character-scoped-persistence` | Complete (`make ci` green) | [`v22_spec-character-scoped-persistence.md`](docs/specs/v22_spec-character-scoped-persistence.md) | [`v22_2026-06-07-character-scoped-persistence.md`](docs/plans/v22_2026-06-07-character-scoped-persistence.md) | [`as-built`](docs/as-built/v22_character-scoped-persistence.md) |
| **v23** | `item-templates-and-rolled-drops` | Complete (`make ci` green) | [`v23_spec-item-templates-and-rolled-drops.md`](docs/specs/v23_spec-item-templates-and-rolled-drops.md) | [`v23_2026-06-07-item-templates-and-rolled-drops.md`](docs/plans/v23_2026-06-07-item-templates-and-rolled-drops.md) | [`as-built`](docs/as-built/v23_item-templates-and-rolled-drops.md) |
| **v24** | `main-menu-and-character-start` | Complete (`make ci` green) | [`v24_spec-main-menu-and-character-start.md`](docs/specs/v24_spec-main-menu-and-character-start.md) | [`v24_2026-06-07-main-menu-and-character-start.md`](docs/plans/v24_2026-06-07-main-menu-and-character-start.md) | [`as-built`](docs/as-built/v24_main-menu-and-character-start.md) |
| **v25** | `treasure-classes-and-guarded-chests` | Complete (`make ci` green) | [`v25_spec-treasure-classes-and-guarded-chests.md`](docs/specs/v25_spec-treasure-classes-and-guarded-chests.md) | [`v25_2026-06-07-treasure-classes-and-guarded-chests.md`](docs/plans/v25_2026-06-07-treasure-classes-and-guarded-chests.md) | [`as-built`](docs/as-built/v25_treasure-classes-and-guarded-chests.md) |
| **v26** | `character-stats-and-leveling` | Complete (`make ci` green) | [`v26_spec-character-stats-and-leveling.md`](docs/specs/v26_spec-character-stats-and-leveling.md) | [`v26_2026-06-07-character-stats-and-leveling.md`](docs/plans/v26_2026-06-07-character-stats-and-leveling.md) | [`as-built`](docs/as-built/v26_character-stats-and-leveling.md) |
| **v27** | `hold-click-controls` | Complete (`make ci` green) | [`v27_spec-hold-click-controls.md`](docs/specs/v27_spec-hold-click-controls.md) | [`v27_2026-06-07-hold-click-controls.md`](docs/plans/v27_2026-06-07-hold-click-controls.md) | [`as-built`](docs/as-built/v27_hold-click-controls.md) |
| **v28** | `full-equipment-and-belt-hotbar` | Complete (`make ci` green) | [`v28_spec-full-equipment-and-belt-hotbar.md`](docs/specs/v28_spec-full-equipment-and-belt-hotbar.md) | [`v28_2026-06-07-full-equipment-and-belt-hotbar.md`](docs/plans/v28_2026-06-07-full-equipment-and-belt-hotbar.md) | [`as-built`](docs/as-built/v28_full-equipment-and-belt-hotbar.md) |
| **v29** | `dungeon-equipment-drop-expansion` | Complete (`make ci` green) | [`v29_spec-dungeon-equipment-drop-expansion.md`](docs/specs/v29_spec-dungeon-equipment-drop-expansion.md) | [`v29_2026-06-07-dungeon-equipment-drop-expansion.md`](docs/plans/v29_2026-06-07-dungeon-equipment-drop-expansion.md) | [`as-built`](docs/as-built/v29_dungeon-equipment-drop-expansion.md) |
| **v30** | `monster-rarity-and-loot-scaling` | Complete (`make ci` green) | [`v30_spec-monster-rarity-and-loot-scaling.md`](docs/specs/v30_spec-monster-rarity-and-loot-scaling.md) | [`v30_2026-06-07-monster-rarity-and-loot-scaling.md`](docs/plans/v30_2026-06-07-monster-rarity-and-loot-scaling.md) | [`as-built`](docs/as-built/v30_monster-rarity-and-loot-scaling.md) |
| **v31** | `combat-stat-effects-and-feedback` | Complete (`make ci` green) | [`v31_spec-combat-stat-effects-and-feedback.md`](docs/specs/v31_spec-combat-stat-effects-and-feedback.md) | [`v31_2026-06-07-combat-stat-effects-and-feedback.md`](docs/plans/v31_2026-06-07-combat-stat-effects-and-feedback.md) | [`as-built`](docs/as-built/v31_combat-stat-effects-and-feedback.md) |
| **v32** | `test-floor-and-resilient-scenarios` | Complete (`make ci` green) | [`v32_spec-test-floor-and-resilient-scenarios.md`](docs/specs/v32_spec-test-floor-and-resilient-scenarios.md) | [`v32_2026-06-08-test-floor-and-resilient-scenarios.md`](docs/plans/v32_2026-06-08-test-floor-and-resilient-scenarios.md) | [`as-built`](docs/as-built/v32_test-floor-and-resilient-scenarios.md) |
| **v33** | `true-coop-session` | Complete (`make ci` green) | [`v33_spec-true-coop-session.md`](docs/specs/v33_spec-true-coop-session.md) | [`v33_2026-06-08-true-coop-session.md`](docs/plans/v33_2026-06-08-true-coop-session.md) | [`as-built`](docs/as-built/v33_true-coop-session.md) |
| **v34** | `model-reaction-polish` | Complete (`make ci` green) | [`v34_spec-model-reaction-polish.md`](docs/specs/v34_spec-model-reaction-polish.md) | [`v34_2026-06-08-model-reaction-polish.md`](docs/plans/v34_2026-06-08-model-reaction-polish.md) | [`as-built`](docs/as-built/v34_model-reaction-polish.md) |
| **v35** | `boss-floor-gate` | Complete (`make ci` green) | [`v35_spec-boss-floor-gate.md`](docs/specs/v35_spec-boss-floor-gate.md) | [`v35_2026-06-08-boss-floor-gate.md`](docs/plans/v35_2026-06-08-boss-floor-gate.md) | [`as-built`](docs/as-built/v35_boss-floor-gate.md) |
| **v36** | `inventory-paper-doll-capacity` | Complete (`make ci` green) | [`v36_spec-inventory-paper-doll-capacity.md`](docs/specs/v36_spec-inventory-paper-doll-capacity.md) | [`v36_2026-06-08-inventory-paper-doll-capacity.md`](docs/plans/v36_2026-06-08-inventory-paper-doll-capacity.md) | [`as-built`](docs/as-built/v36_inventory-paper-doll-capacity.md) |
| **v37** | `combat-control-and-boss-ai-fixes` | Complete (`make ci` green) | [`v37_spec-combat-control-and-boss-ai-fixes.md`](docs/specs/v37_spec-combat-control-and-boss-ai-fixes.md) | [`v37_2026-06-08-combat-control-and-boss-ai-fixes.md`](docs/plans/v37_2026-06-08-combat-control-and-boss-ai-fixes.md) | [`as-built`](docs/as-built/v37_combat-control-and-boss-ai-fixes.md) |
| **v38** | `session-browser-and-uncapped-coop-menu` | Complete (`make ci` green) | [`v38_spec-session-browser-and-uncapped-coop-menu.md`](docs/specs/v38_spec-session-browser-and-uncapped-coop-menu.md) | [`v38_2026-06-08-session-browser-and-uncapped-coop-menu.md`](docs/plans/v38_2026-06-08-session-browser-and-uncapped-coop-menu.md) | [`as-built`](docs/as-built/v38_session-browser-and-uncapped-coop-menu.md) |
| **v39** | `ui-currency-and-mana-polish` | Complete (`make ci` green) | [`v39_spec-ui-currency-and-mana-polish.md`](docs/specs/v39_spec-ui-currency-and-mana-polish.md) | [`v39_2026-06-09-ui-currency-and-mana-polish.md`](docs/plans/v39_2026-06-09-ui-currency-and-mana-polish.md) | [`as-built`](docs/as-built/v39_ui-currency-and-mana-polish.md) |
| **v40** | `reachable-dungeon-obstacles` | Complete (`make ci` green) | [`v40_spec-reachable-dungeon-obstacles.md`](docs/specs/v40_spec-reachable-dungeon-obstacles.md) | [`v40_2026-06-09-reachable-dungeon-obstacles.md`](docs/plans/v40_2026-06-09-reachable-dungeon-obstacles.md) | [`as-built`](docs/as-built/v40_reachable-dungeon-obstacles.md) |
| **v41** | `town-vendor-gold-sink` | Complete (`make ci` green) | [`v41_spec-town-vendor-gold-sink.md`](docs/specs/v41_spec-town-vendor-gold-sink.md) | [`v41_2026-06-09-town-vendor-gold-sink.md`](docs/plans/v41_2026-06-09-town-vendor-gold-sink.md) | [`as-built`](docs/as-built/v41_town-vendor-gold-sink.md) |
| **v42** | `vendor-appraisal-and-item-comparison` | Complete (`make ci` green) | [`v42_spec-vendor-appraisal-and-item-comparison.md`](docs/specs/v42_spec-vendor-appraisal-and-item-comparison.md) | [`v42_2026-06-09-vendor-appraisal-and-item-comparison.md`](docs/plans/v42_2026-06-09-vendor-appraisal-and-item-comparison.md) | [`as-built`](docs/as-built/v42_vendor-appraisal-and-item-comparison.md) |
| **v43** | `equipment-requirements-and-preview` | Complete (`make ci` green) | [`v43_spec-equipment-requirements-and-preview.md`](docs/specs/v43_spec-equipment-requirements-and-preview.md) | [`v43_2026-06-09-equipment-requirements-and-preview.md`](docs/plans/v43_2026-06-09-equipment-requirements-and-preview.md) | [`as-built`](docs/as-built/v43_equipment-requirements-and-preview.md) |
| **v44** | `skill-points-and-magic-bolt` | Complete (`make ci` green) | [`v44_spec-skill-points-and-magic-bolt.md`](docs/specs/v44_spec-skill-points-and-magic-bolt.md) | [`v44_2026-06-09-skill-points-and-magic-bolt.md`](docs/plans/v44_2026-06-09-skill-points-and-magic-bolt.md) | [`as-built`](docs/as-built/v44_skill-points-and-magic-bolt.md) |
| **v45** | `menu-create-join-flow` | Complete (`make ci` green) | [`v45_spec-menu-create-join-flow.md`](docs/specs/v45_spec-menu-create-join-flow.md) | [`v45_2026-06-09-menu-create-join-flow.md`](docs/plans/v45_2026-06-09-menu-create-join-flow.md) | [`as-built`](docs/as-built/v45_menu-create-join-flow.md) |
| **v46** | `client-join-game-proof` | Complete (`make ci` green) | [`v46_spec-client-join-game-proof.md`](docs/specs/v46_spec-client-join-game-proof.md) | [`v46_2026-06-09-client-join-game-proof.md`](docs/plans/v46_2026-06-09-client-join-game-proof.md) | [`as-built`](docs/as-built/v46_client-join-game-proof.md) |
| **v47** | `shop-stock-lifecycle` | Complete (`make ci` green) | [`v47_spec-shop-stock-lifecycle.md`](docs/specs/v47_spec-shop-stock-lifecycle.md) | [`v47_2026-06-09-shop-stock-lifecycle.md`](docs/plans/v47_2026-06-09-shop-stock-lifecycle.md) | [`as-built`](docs/as-built/v47_shop-stock-lifecycle.md) |
| **v48** | `coop-rewards-and-scaling` | Complete (`make ci` green) | [`v48_spec-coop-rewards-and-scaling.md`](docs/specs/v48_spec-coop-rewards-and-scaling.md) | [`v48_2026-06-09-coop-rewards-and-scaling.md`](docs/plans/v48_2026-06-09-coop-rewards-and-scaling.md) | [`as-built`](docs/as-built/v48_coop-rewards-and-scaling.md) |
| **v49** | `gold-autopickup-and-shared-loot-rules` | Complete (`make ci` green) | [`v49_spec-gold-autopickup-and-shared-loot-rules.md`](docs/specs/v49_spec-gold-autopickup-and-shared-loot-rules.md) | [`v49_2026-06-10-gold-autopickup-and-shared-loot-rules.md`](docs/plans/v49_2026-06-10-gold-autopickup-and-shared-loot-rules.md) | [`as-built`](docs/as-built/v49_gold-autopickup-and-shared-loot-rules.md) |
| **v50** | `account-stash-storage` | Complete (`make ci` green) | [`v50_spec-account-stash-storage.md`](docs/specs/v50_spec-account-stash-storage.md) | [`v50_2026-06-10-account-stash-storage.md`](docs/plans/v50_2026-06-10-account-stash-storage.md) | [`as-built`](docs/as-built/v50_account-stash-storage.md) |
| **v51** | `mystery-seller-core` | Complete (`make ci` green) | [`v51_spec-mystery-seller-core.md`](docs/specs/v51_spec-mystery-seller-core.md) | [`v51_2026-06-10-mystery-seller-core.md`](docs/plans/v51_2026-06-10-mystery-seller-core.md) | [`as-built`](docs/as-built/v51_mystery-seller-core.md) |
| **v52** | `ranged-monster-ai` | Complete (`make ci` green) | [`v52_spec-ranged-monster-ai.md`](docs/specs/v52_spec-ranged-monster-ai.md) | [`v52_2026-06-10-ranged-monster-ai.md`](docs/plans/v52_2026-06-10-ranged-monster-ai.md) | [`as-built`](docs/as-built/v52_ranged-monster-ai.md) |
| **v53** | `boss-health-bar-ui` | Complete (`make ci` green) | [`v53_spec-boss-health-bar-ui.md`](docs/specs/v53_spec-boss-health-bar-ui.md) | [`v53_2026-06-10-boss-health-bar-ui.md`](docs/plans/v53_2026-06-10-boss-health-bar-ui.md) | [`as-built`](docs/as-built/v53_boss-health-bar-ui.md) |
| **v54** | `character-select-summaries` | Complete (`make ci` green) | [`v54_spec-character-select-summaries.md`](docs/specs/v54_spec-character-select-summaries.md) | [`v54_2026-06-10-character-select-summaries.md`](docs/plans/v54_2026-06-10-character-select-summaries.md) | [`as-built`](docs/as-built/v54_character-select-summaries.md) |
| **v55** | `consolidation-and-quality-gates` | Complete (`make ci` green) | [`v55_spec-consolidation-and-quality-gates.md`](docs/specs/v55_spec-consolidation-and-quality-gates.md) | [`v55_2026-06-10-consolidation-and-quality-gates.md`](docs/plans/v55_2026-06-10-consolidation-and-quality-gates.md) | — |
| **v56** | `monster-attack-cadence` | Complete (`make ci` green) | [`v56_spec-monster-attack-cadence.md`](docs/specs/v56_spec-monster-attack-cadence.md) | [`v56_2026-06-10-monster-attack-cadence.md`](docs/plans/v56_2026-06-10-monster-attack-cadence.md) | [`as-built`](docs/as-built/v56_monster-attack-cadence.md) |
| **v57** | `boss-phase-readability` | Complete (`make ci` green) | [`v57_spec-boss-phase-readability.md`](docs/specs/v57_spec-boss-phase-readability.md) | [`v57_2026-06-10-boss-phase-readability.md`](docs/plans/v57_2026-06-10-boss-phase-readability.md) | [`as-built`](docs/as-built/v57_boss-phase-readability.md) |
| **v58** | `boss-pattern-variety` | Complete (`make ci` green) | [`v58_spec-boss-pattern-variety.md`](docs/specs/v58_spec-boss-pattern-variety.md) | [`v58_2026-06-10-boss-pattern-variety.md`](docs/plans/v58_2026-06-10-boss-pattern-variety.md) | [`as-built`](docs/as-built/v58_boss-pattern-variety.md) |
| **v59** | `data-driven-skill-catalog` | Complete (`make ci` green) | [`v59_spec-data-driven-skill-catalog.md`](docs/specs/v59_spec-data-driven-skill-catalog.md) | [`v59_2026-06-10-data-driven-skill-catalog.md`](docs/plans/v59_2026-06-10-data-driven-skill-catalog.md) | [`as-built`](docs/as-built/v59_data-driven-skill-catalog.md) |
| **v60** | `data-driven-content-library-manifest` | Complete (`make ci` green) | [`v60_spec-data-driven-content-library-manifest.md`](docs/specs/v60_spec-data-driven-content-library-manifest.md) | [`v60_2026-06-10-data-driven-content-library-manifest.md`](docs/plans/v60_2026-06-10-data-driven-content-library-manifest.md) | [`as-built`](docs/as-built/v60_data-driven-content-library-manifest.md) |
| **v61** | `rage-and-heal-skills` | Complete (`make ci` green) | [`v61_spec-rage-and-heal-skills.md`](docs/specs/v61_spec-rage-and-heal-skills.md) | [`v61_2026-06-10-rage-and-heal-skills.md`](docs/plans/v61_2026-06-10-rage-and-heal-skills.md) | [`as-built`](docs/as-built/v61_rage-and-heal-skills.md) |
| **v62** | `monster-depth-stat-scaling` | Complete (`make ci` green) | [`v62_spec-monster-depth-stat-scaling.md`](docs/specs/v62_spec-monster-depth-stat-scaling.md) | [`v62_2026-06-11-monster-depth-stat-scaling.md`](docs/plans/v62_2026-06-11-monster-depth-stat-scaling.md) | [`as-built`](docs/as-built/v62_monster-depth-stat-scaling.md) |
| **v63** | `runtime-sim-error-construction` | Complete (`make ci` green) | [`v63_spec-runtime-sim-error-construction.md`](docs/specs/v63_spec-runtime-sim-error-construction.md) | [`v63_2026-06-11-runtime-sim-error-construction.md`](docs/plans/v63_2026-06-11-runtime-sim-error-construction.md) | [`as-built`](docs/as-built/v63_runtime-sim-error-construction.md) |
| **v64** | `mystery-seller-paid-reroll` | Complete (`make ci` green) | [`v64_spec-mystery-seller-paid-reroll.md`](docs/specs/v64_spec-mystery-seller-paid-reroll.md) | [`v64_2026-06-11-mystery-seller-paid-reroll.md`](docs/plans/v64_2026-06-11-mystery-seller-paid-reroll.md) | [`as-built`](docs/as-built/v64_mystery-seller-paid-reroll.md) |
| **v65** | `stash-search-and-sorting` | Complete (`make ci` green) | [`v65_spec-stash-search-and-sorting.md`](docs/specs/v65_spec-stash-search-and-sorting.md) | [`v65_2026-06-11-stash-search-and-sorting.md`](docs/plans/v65_2026-06-11-stash-search-and-sorting.md) | [`as-built`](docs/as-built/v65_stash-search-and-sorting.md) |
| **v66** | `progress-backlog-hygiene` | Complete (`make ci` green) | [`v66_spec-progress-backlog-hygiene.md`](docs/specs/v66_spec-progress-backlog-hygiene.md) | [`v66_2026-06-11-progress-backlog-hygiene.md`](docs/plans/v66_2026-06-11-progress-backlog-hygiene.md) | [`as-built`](docs/as-built/v66_progress-backlog-hygiene.md) |
| **v67** | `boss-kill-reward-polish` | Complete (`make ci` green) | [`v67_spec-boss-kill-reward-polish.md`](docs/specs/v67_spec-boss-kill-reward-polish.md) | [`v67_2026-06-11-boss-kill-reward-polish.md`](docs/plans/v67_2026-06-11-boss-kill-reward-polish.md) | [`as-built`](docs/as-built/v67_boss-kill-reward-polish.md) |
| **v68** | `market-stash-listing-foundation` | Complete (`make ci` green) | [`v68_spec-market-stash-listing-foundation.md`](docs/specs/v68_spec-market-stash-listing-foundation.md) | [`v68_2026-06-11-market-stash-listing-foundation.md`](docs/plans/v68_2026-06-11-market-stash-listing-foundation.md) | [`as-built`](docs/as-built/v68_market-stash-listing-foundation.md) |
| **v69** | `character-class-foundation` | Complete (`make ci` green) | [`v69_spec-character-class-foundation.md`](docs/specs/v69_spec-character-class-foundation.md) | [`v69_2026-06-11-character-class-foundation.md`](docs/plans/v69_2026-06-11-character-class-foundation.md) | [`as-built`](docs/as-built/v69_character-class-foundation.md) |
| **v70** | `class-skill-and-item-gates` | Complete (`make ci` green) | [`v70_spec-class-skill-and-item-gates.md`](docs/specs/v70_spec-class-skill-and-item-gates.md) | [`v70_2026-06-11-class-skill-and-item-gates.md`](docs/plans/v70_2026-06-11-class-skill-and-item-gates.md) | [`as-built`](docs/as-built/v70_class-skill-and-item-gates.md) |
| **v71** | `class-picker-and-sprites` | Complete (`make ci` green) | [`v71_spec-class-picker-and-sprites.md`](docs/specs/v71_spec-class-picker-and-sprites.md) | [`v71_2026-06-11-class-picker-and-sprites.md`](docs/plans/v71_2026-06-11-class-picker-and-sprites.md) | [`as-built`](docs/as-built/v71_class-picker-and-sprites.md) |
| **v72** | `monster-visual-catalog` | Complete (`make ci` green) | [`v72_spec-monster-visual-catalog.md`](docs/specs/v72_spec-monster-visual-catalog.md) | [`v72_2026-06-11-monster-visual-catalog.md`](docs/plans/v72_2026-06-11-monster-visual-catalog.md) | [`as-built`](docs/as-built/v72_monster-visual-catalog.md) |
| **v73** | `draggable-window-foundation` | Complete (`make client-unit` green) | [`v73_spec-draggable-window-foundation.md`](docs/specs/v73_spec-draggable-window-foundation.md) | [`v73_2026-06-11-draggable-window-foundation.md`](docs/plans/v73_2026-06-11-draggable-window-foundation.md) | [`as-built`](docs/as-built/v73_draggable-window-foundation.md) |
| **v74** | `gameplay-window-chrome` | Complete (`make client-unit` green) | [`v74_spec-gameplay-window-chrome.md`](docs/specs/v74_spec-gameplay-window-chrome.md) | [`v74_2026-06-11-gameplay-window-chrome.md`](docs/plans/v74_2026-06-11-gameplay-window-chrome.md) | [`as-built`](docs/as-built/v74_gameplay-window-chrome.md) |
| **v75** | `persistent-window-layout` | Complete (`make client-unit` green) | [`v75_spec-persistent-window-layout.md`](docs/specs/v75_spec-persistent-window-layout.md) | [`v75_2026-06-11-persistent-window-layout.md`](docs/plans/v75_2026-06-11-persistent-window-layout.md) | [`as-built`](docs/as-built/v75_persistent-window-layout.md) |
| **v76** | `main-config-foundation` | Complete (`make ci` green) | [`v76_spec-main-config-foundation.md`](docs/specs/v76_spec-main-config-foundation.md) | [`v76_2026-06-11-main-config-foundation.md`](docs/plans/v76_2026-06-11-main-config-foundation.md) | [`as-built`](docs/as-built/v76_main-config-foundation.md) |
| **v77** | `main-config-derived-gameplay` | Complete (`make ci` green) | [`v77_spec-main-config-derived-gameplay.md`](docs/specs/v77_spec-main-config-derived-gameplay.md) | [`v77_2026-06-11-main-config-derived-gameplay.md`](docs/plans/v77_2026-06-11-main-config-derived-gameplay.md) | [`as-built`](docs/as-built/v77_main-config-derived-gameplay.md) |
| **v78** | `main-config-drop-profiles` | Complete (`make ci` green) | [`v78_spec-main-config-drop-profiles.md`](docs/specs/v78_spec-main-config-drop-profiles.md) | [`v78_2026-06-11-main-config-drop-profiles.md`](docs/plans/v78_2026-06-11-main-config-drop-profiles.md) | [`as-built`](docs/as-built/v78_main-config-drop-profiles.md) |
| **v79** | `elite-pack-roles` | Complete (`make ci` green) | [`v79_spec-elite-pack-roles.md`](docs/specs/v79_spec-elite-pack-roles.md) | [`v79_2026-06-11-elite-pack-roles.md`](docs/plans/v79_2026-06-11-elite-pack-roles.md) | [`as-built`](docs/as-built/v79_elite-pack-roles.md) |
| **v80** | `combat-threat-readability` | Complete (`make ci` green) | [`v80_spec-combat-threat-readability.md`](docs/specs/v80_spec-combat-threat-readability.md) | [`v80_2026-06-11-combat-threat-readability.md`](docs/plans/v80_2026-06-11-combat-threat-readability.md) | [`as-built`](docs/as-built/v80_combat-threat-readability.md) |
| **v81** | `paladin-holy-shield` | Complete (`make ci` green) | [`v81_spec-paladin-holy-shield.md`](docs/specs/v81_spec-paladin-holy-shield.md) | [`v81_2026-06-11-paladin-holy-shield.md`](docs/plans/v81_2026-06-11-paladin-holy-shield.md) | [`as-built`](docs/as-built/v81_paladin-holy-shield.md) |
| **v82** | `realtime-fanout-level-snapshot` | Complete (`make ci` green) | [`v82_spec-realtime-fanout-level-snapshot.md`](docs/specs/v82_spec-realtime-fanout-level-snapshot.md) | [`v82_2026-06-11-realtime-fanout-level-snapshot.md`](docs/plans/v82_2026-06-11-realtime-fanout-level-snapshot.md) | [`as-built`](docs/as-built/v82_realtime-fanout-level-snapshot.md) |
| **v83** | `defensive-client-envelope-payloads` | Complete (`make ci` green) | [`v83_spec-defensive-client-envelope-payloads.md`](docs/specs/v83_spec-defensive-client-envelope-payloads.md) | [`v83_2026-06-11-defensive-client-envelope-payloads.md`](docs/plans/v83_2026-06-11-defensive-client-envelope-payloads.md) | [`as-built`](docs/as-built/v83_defensive-client-envelope-payloads.md) |
| **v84** | `client-bot-step-registry` | Complete (`make ci` green) | [`v84_spec-client-bot-step-registry.md`](docs/specs/v84_spec-client-bot-step-registry.md) | [`v84_2026-06-11-client-bot-step-registry.md`](docs/plans/v84_2026-06-11-client-bot-step-registry.md) | [`as-built`](docs/as-built/v84_client-bot-step-registry.md) |
| **v85** | `skill-demo-catalog` | Complete (`make ci` green) | [`v85_spec-skill-demo-catalog.md`](docs/specs/v85_spec-skill-demo-catalog.md) | [`v85_2026-06-11-skill-demo-catalog.md`](docs/plans/v85_2026-06-11-skill-demo-catalog.md) | [`as-built`](docs/as-built/v85_skill-demo-catalog.md) |
| **v86** | `skill-visual-command` | Complete (`make ci` green) | [`v86_spec-skill-visual-command.md`](docs/specs/v86_spec-skill-visual-command.md) | [`v86_2026-06-11-skill-visual-command.md`](docs/plans/v86_2026-06-11-skill-visual-command.md) | [`as-built`](docs/as-built/v86_skill-visual-command.md) |
| **v87** | `skill-visual-matrix` | Complete (`make ci` green) | [`v87_spec-skill-visual-matrix.md`](docs/specs/v87_spec-skill-visual-matrix.md) | [`v87_2026-06-11-skill-visual-matrix.md`](docs/plans/v87_2026-06-11-skill-visual-matrix.md) | [`as-built`](docs/as-built/v87_skill-visual-matrix.md) |
| **v88** | `skill-visual-rank-seeding` | Complete (`make ci` green) | — | — | [`as-built`](docs/as-built/v88_skill-visual-rank-seeding.md) |
| **v89** | `class-second-combat-skills` | Complete (`make ci` green) | [`v89_spec-class-second-combat-skills.md`](docs/specs/v89_spec-class-second-combat-skills.md) | [`v89_2026-06-12-class-second-combat-skills.md`](docs/plans/v89_2026-06-12-class-second-combat-skills.md) | [`as-built`](docs/as-built/v89_class-second-combat-skills.md) |
| **v90** | `text-catalog-foundation` | Complete (`make ci` green) | [`v90_spec-text-catalog-foundation.md`](docs/specs/v90_spec-text-catalog-foundation.md) | [`v90_2026-06-12-text-catalog-foundation.md`](docs/plans/v90_2026-06-12-text-catalog-foundation.md) | [`as-built`](docs/as-built/v90_text-catalog-foundation.md) |
| **v91** | `spanish-language-selector` | Complete (`make ci` green) | [`v91_spec-spanish-language-selector.md`](docs/specs/v91_spec-spanish-language-selector.md) | [`v91_2026-06-12-spanish-language-selector.md`](docs/plans/v91_2026-06-12-spanish-language-selector.md) | [`as-built`](docs/as-built/v91_spanish-language-selector.md) |
| **v92** | `town-bishop-respec` | Complete (`make ci` green) | [`v92_spec-town-bishop-respec.md`](docs/specs/v92_spec-town-bishop-respec.md) | [`v92_2026-06-12-town-bishop-respec.md`](docs/plans/v92_2026-06-12-town-bishop-respec.md) | [`as-built`](docs/as-built/v92_town-bishop-respec.md) |
| **v93** | `market-multi-item-offers` | Complete (`make ci` green) | [`v93_spec-market-multi-item-offers.md`](docs/specs/v93_spec-market-multi-item-offers.md) | [`v93_2026-06-12-market-multi-item-offers.md`](docs/plans/v93_2026-06-12-market-multi-item-offers.md) | [`as-built`](docs/as-built/v93_market-multi-item-offers.md) |
| **v94** | `item-upgrade-starter` | Complete (`make ci` green) | [`v94_spec-item-upgrade-starter.md`](docs/specs/v94_spec-item-upgrade-starter.md) | [`v94_2026-06-12-item-upgrade-starter.md`](docs/plans/v94_2026-06-12-item-upgrade-starter.md) | [`as-built`](docs/as-built/v94_item-upgrade-starter.md) |
| **v95** | `unique-item-catalog-seed` | Complete (`make ci` green) | [`v95_spec-unique-item-catalog-seed.md`](docs/specs/v95_spec-unique-item-catalog-seed.md) | [`v95_2026-06-12-unique-item-catalog-seed.md`](docs/plans/v95_2026-06-12-unique-item-catalog-seed.md) | [`as-built`](docs/as-built/v95_unique-item-catalog-seed.md) |
| **v96** | `town-presentation-polish` | Complete (`make ci` green) | [`v96_spec-town-presentation-polish.md`](docs/specs/v96_spec-town-presentation-polish.md) | [`v96_2026-06-12-town-presentation-polish.md`](docs/plans/v96_2026-06-12-town-presentation-polish.md) | [`as-built`](docs/as-built/v96_town-presentation-polish.md) |
| **v97** | `class-starter-loadouts` | Complete (`make ci` green) | [`v97_spec-class-starter-loadouts.md`](docs/specs/v97_spec-class-starter-loadouts.md) | [`v97_2026-06-12-class-starter-loadouts.md`](docs/plans/v97_2026-06-12-class-starter-loadouts.md) | [`as-built`](docs/as-built/v97_class-starter-loadouts.md) |
| **v98** | `rogue-class-foundation` | Complete (`make ci` green) | [`v98_spec-rogue-class-foundation.md`](docs/specs/v98_spec-rogue-class-foundation.md) | [`v98_2026-06-12-rogue-class-foundation.md`](docs/plans/v98_2026-06-12-rogue-class-foundation.md) | [`as-built`](docs/as-built/v98_rogue-class-foundation.md) |
| **v99** | `rogue-skill-mechanics` | Complete (`make ci` green) | [`v99_spec-rogue-skill-mechanics.md`](docs/specs/v99_spec-rogue-skill-mechanics.md) | [`v99_2026-06-12-rogue-skill-mechanics.md`](docs/plans/v99_2026-06-12-rogue-skill-mechanics.md) | [`as-built`](docs/as-built/v99_rogue-skill-mechanics.md) |
| **v100** | `damage-types-and-resistances` | Complete (`make ci` green) | [`v100_spec-damage-types-and-resistances.md`](docs/specs/v100_spec-damage-types-and-resistances.md) | [`v100_2026-06-12-damage-types-and-resistances.md`](docs/plans/v100_2026-06-12-damage-types-and-resistances.md) | [`as-built`](docs/as-built/v100_damage-types-and-resistances.md) |
| **v101** | `undead-skeleton-poison-immunity` | Complete (`make ci` green) | [`v101_spec-undead-skeleton-poison-immunity.md`](docs/specs/v101_spec-undead-skeleton-poison-immunity.md) | [`v101_2026-06-12-undead-skeleton-poison-immunity.md`](docs/plans/v101_2026-06-12-undead-skeleton-poison-immunity.md) | [`as-built`](docs/as-built/v101_undead-skeleton-poison-immunity.md) |
| **v102** | `class-bot-visual-scenarios` | Complete (`make ci` green) | [`v102_spec-class-bot-visual-scenarios.md`](docs/specs/v102_spec-class-bot-visual-scenarios.md) | [`v102_2026-06-12-class-bot-visual-scenarios.md`](docs/plans/v102_2026-06-12-class-bot-visual-scenarios.md) | [`as-built`](docs/as-built/v102_class-bot-visual-scenarios.md) |
| **v103** | `unique-effect-catalog-foundation` | Complete (`make ci` green) | [`v103_spec-unique-effect-catalog-foundation.md`](docs/specs/v103_spec-unique-effect-catalog-foundation.md) | [`v103_2026-06-12-unique-effect-catalog-foundation.md`](docs/plans/v103_2026-06-12-unique-effect-catalog-foundation.md) | [`as-built`](docs/as-built/v103_unique-effect-catalog-foundation.md) |
| **v104** | `unique-drop-roll-contract` | Complete (`make ci` green) | [`v104_spec-unique-drop-roll-contract.md`](docs/specs/v104_spec-unique-drop-roll-contract.md) | [`v104_2026-06-12-unique-drop-roll-contract.md`](docs/plans/v104_2026-06-12-unique-drop-roll-contract.md) | [`as-built`](docs/as-built/v104_unique-drop-roll-contract.md) |
| **v105** | `unique-burn-effect-live` | Complete (`make ci` green) | [`v105_spec-unique-burn-effect-live.md`](docs/specs/v105_spec-unique-burn-effect-live.md) | [`v105_2026-06-12-unique-burn-effect-live.md`](docs/plans/v105_2026-06-12-unique-burn-effect-live.md) | [`as-built`](docs/as-built/v105_unique-burn-effect-live.md) |
| **v106** | `offensive-unique-effects` | Complete (`make ci` green) | [`v106_spec-offensive-unique-effects.md`](docs/specs/v106_spec-offensive-unique-effects.md) | [`v106_2026-06-12-offensive-unique-effects.md`](docs/plans/v106_2026-06-12-offensive-unique-effects.md) | [`as-built`](docs/as-built/v106_offensive-unique-effects.md) |
| **v107** | `survival-reactive-unique-effects` | Complete (`make ci` green) | [`v107_spec-survival-reactive-unique-effects.md`](docs/specs/v107_spec-survival-reactive-unique-effects.md) | [`v107_2026-06-12-survival-reactive-unique-effects.md`](docs/plans/v107_2026-06-12-survival-reactive-unique-effects.md) | [`as-built`](docs/as-built/v107_survival-reactive-unique-effects.md) |
| **v108** | `resource-support-mobility-unique-effects` | Complete (`make ci` green) | [`v108_spec-resource-support-mobility-unique-effects.md`](docs/specs/v108_spec-resource-support-mobility-unique-effects.md) | [`v108_2026-06-12-resource-support-mobility-unique-effects.md`](docs/plans/v108_2026-06-12-resource-support-mobility-unique-effects.md) | [`as-built`](docs/as-built/v108_resource-support-mobility-unique-effects.md) |
| **v109** | `permanent-death-corpse-recovery` | Complete (`make ci` green) | — | — | [`as-built`](docs/as-built/v109_permanent-death-corpse-recovery.md) |
| **v110** | `item-upgrade-repeat-action` | Complete (`make ci` green) | [`v110_spec-item-upgrade-repeat-action.md`](docs/specs/v110_spec-item-upgrade-repeat-action.md) | [`v110_2026-06-13-item-upgrade-repeat-action.md`](docs/plans/v110_2026-06-13-item-upgrade-repeat-action.md) | [`as-built`](docs/as-built/v110_item-upgrade-repeat-action.md) |
| **v111** | `market-purchase-and-delivery` | Complete (`make ci` green) | [`v111_spec-market-purchase-and-delivery.md`](docs/specs/v111_spec-market-purchase-and-delivery.md) | [`v111_2026-06-13-market-purchase-and-delivery.md`](docs/plans/v111_2026-06-13-market-purchase-and-delivery.md) | [`as-built`](docs/as-built/v111_market-purchase-and-delivery.md) |
| **v112** | `elite-aura-foundation` | Complete (`make ci` green) | [`v112_spec-elite-aura-foundation.md`](docs/specs/v112_spec-elite-aura-foundation.md) | [`v112_2026-06-13-elite-aura-foundation.md`](docs/plans/v112_2026-06-13-elite-aura-foundation.md) | [`as-built`](docs/as-built/v112_elite-aura-foundation.md) |
| **v113** | `elite-aura-readability` | Complete (`make ci` green) | [`v113_spec-elite-aura-readability.md`](docs/specs/v113_spec-elite-aura-readability.md) | [`v113_2026-06-13-elite-aura-readability.md`](docs/plans/v113_2026-06-13-elite-aura-readability.md) | [`as-built`](docs/as-built/v113_elite-aura-readability.md) |
| **v114** | `market-board-ui` | Complete (`make ci` green) | [`v114_spec-market-board-ui.md`](docs/specs/v114_spec-market-board-ui.md) | [`v114_2026-06-13-market-board-ui.md`](docs/plans/v114_2026-06-13-market-board-ui.md) | [`as-built`](docs/as-built/v114_market-board-ui.md) |
| **v115** | `market-purchase-ui` | Complete (`make ci` green) | [`v115_spec-market-purchase-ui.md`](docs/specs/v115_spec-market-purchase-ui.md) | [`v115_2026-06-13-market-purchase-ui.md`](docs/plans/v115_2026-06-13-market-purchase-ui.md) | [`as-built`](docs/as-built/v115_market-purchase-ui.md) |
| **v116** | `elite-aura-radius-preview` | Complete (`make ci` green) | [`v116_spec-elite-aura-radius-preview.md`](docs/specs/v116_spec-elite-aura-radius-preview.md) | [`v116_2026-06-13-elite-aura-radius-preview.md`](docs/plans/v116_2026-06-13-elite-aura-radius-preview.md) | [`as-built`](docs/as-built/v116_elite-aura-radius-preview.md) |
| **v117** | `market-active-offer-ui` | Complete (`make ci` green) | [`v117_spec-market-active-offer-ui.md`](docs/specs/v117_spec-market-active-offer-ui.md) | [`v117_2026-06-13-market-active-offer-ui.md`](docs/plans/v117_2026-06-13-market-active-offer-ui.md) | [`as-built`](docs/as-built/v117_market-active-offer-ui.md) |
| **v118** | `blacksmith-upgrade-ui` | Complete (`make ci` green) | [`v118_spec-blacksmith-upgrade-ui.md`](docs/specs/v118_spec-blacksmith-upgrade-ui.md) | [`v118_2026-06-13-blacksmith-upgrade-ui.md`](docs/plans/v118_2026-06-13-blacksmith-upgrade-ui.md) | [`as-built`](docs/as-built/v118_blacksmith-upgrade-ui.md) |
| **v119** | `live-unique-drops-all-effects` | Complete (`make ci` green) | [`v119_spec-live-unique-drops-all-effects.md`](docs/specs/v119_spec-live-unique-drops-all-effects.md) | [`v119_2026-06-13-live-unique-drops-all-effects.md`](docs/plans/v119_2026-06-13-live-unique-drops-all-effects.md) | [`as-built`](docs/as-built/v119_live-unique-drops-all-effects.md) |

---

## Slice as-built summaries

Per-slice **what it proved** notes live in [`docs/as-built/`](docs/as-built/) — one file per
completed slice. Specs record intent; plans record execution; as-built records what shipped.

On `/finish`, add or update `docs/as-built/vN_<codename>.md` instead of growing this file.
Use the **As-built** column in the slice lifecycle table above for links.

## Architecture decisions (ADRs)

| ADR | Topic | Status |
|-----|-------|--------|
| [0001](docs/adr/0001-technology-stack.md) | Foundational stack (Go server, Godot client, shared rules, replay, bot) | Accepted |
| [0006](docs/adr/0006-asset-pipeline.md) | glTF-first assets, manifests, sockets, validation | Accepted; v3 as-built for rigged GLBs |
| [0007](docs/adr/0007-animation-state-model.md) | Client-only animation; event-driven reactions | Accepted; v4 as-built for player reactions |
| [0008](docs/adr/0008-world-structure-and-dungeon-progression.md) | World structure: infinite inverted-tower dungeon, multi-level Sim, character-scoped persistence, waypoints, co-op | Accepted |
| [0009](docs/adr/0009-boss-floors-and-timing-mechanics.md) | Boss floors, telegraphed timing mechanics, and progression gates | Proposed; v35 as-built covers first boss-floor gate |
| [0010](docs/adr/0010-mercenaries-from-player-characters.md) | Hired mercenaries derived from other players' characters | Proposed |
| [0011](docs/adr/0011-player-market-and-multi-item-trade-offers.md) | Player market listings and multi-item trade offers | Proposed |
| [0012](docs/adr/0012-item-upgrades-and-item-levels.md) | Item upgrades, item levels, and advanced dungeon resources | Proposed |
| [0013](docs/adr/0013-mystery-seller-and-unidentified-item-offers.md) | Mystery seller with expensive unidentified equipment offers | Proposed |
| [0014](docs/adr/0014-core-progression-and-endgame-design-rules.md) | Core progression, itemization, economy, endgame, co-op, and PvP design rules | Proposed |

Anticipated but **not written:** netcode timing, Protobuf migration, production auth, multiplayer split,
quest system design, NPC interaction protocol, character progression formulas
(see ADR-0001 follow-up list and ADR-0008 deferred items). Future mercenaries, player market,
item upgrades, and mystery seller economy are captured separately in ADR-0010, ADR-0011, ADR-0012,
and ADR-0013.

---

## Scripted vertical slice flow (bot + smoke)

Every slice keeps this loop working unless the spec explicitly changes it:

```text
dev-login → create session → move → attack training dummy → pick up loot → equip rusty_sword
```

After v4 the player **survives with reduced HP** (`hp < 10`). Monster dies; player may take retaliation
each successful hit. After v7 this flow lives in `tools/bot/scenarios/01_vertical_slice.json`; additional
scenario JSON files are automatically included in filename order in `make bot` and `make bot-visual`.
Every protocol bot scenario has a hard **10.0 second** full-run budget. When a proof grows past
that, shorten setup to the behavior under test with compact lab worlds or focused lower-level tests
instead of waiting through unrelated traversal, farming, or natural timing loops.

The scenario catalog also includes:

```text
gear_before_combat: walk to rusty_sword → pick up → equip → one-shot reward dummy → pick up training_badge
collision_lab: pass through middle wall gap → kill monster on far side
inventory_lab: pick up rusty_sword → equip → unequip → drop → re-pickup → re-equip
heal_lab: pick up red_potion x2 → take damage → use potion twice → full HP
chase_lab / chase_maze / leash_lab: wait while chase monster closes; kite beyond leash and return
dungeon_levels / teleporter_lab: start in town, descend/ascend generated floors; discover teleporters and fast-travel back
character_persistence: same-account fresh sessions retain gear/equipment and discovered waypoint access
rolled_drops: kill dungeon mob → pick up/equip rolled cave_blade → prove rolled metadata persists
main_menu_flow: Create Game root flow → settings → listed co-op session → pause input lock → return → existing-character fresh session
treasure_classes_and_guarded_chests: pinned chest floor → kill guarded mob → open chest once → pick up chest loot
character_stats_and_leveling: descend to dungeon → kill mobs for XP → level up → spend VIT → prove persistence
full_equipment: pick up/equip paper-doll gear → prove hand occupancy → assign belt-gated hotbar → prove persistence
dungeon_equipment_drops: compact equipment lab → pick up/equip rolled equipment → prove persistence; depth-banded generation stays in lower-level tests/goldens
monster_rarity_loot_scaling: descend to generated dungeon → assert champion rarity → kill → pick up rolled loot → prove persistence
combat_stat_effects: combat lab proofs for miss, crit, armor floor, block, monster crit/block, projectile impact, and stat breakdowns
client_combat_feedback: equip gear → assert stat breakdowns → prove normal/crit/miss/block floating text and settings toggle
true_coop_session: host creates co-op → guest joins → shared-level visibility → independent movement → disconnect/reconnect → replay proof
model_reaction_polish: attack training dummy → prove monster hit reaction → prove local player hit reaction → kill dummy → prove terminal corpse reaction
boss_floor_gate: start on compact boss floor → assert locked exits → observe boss phase telegraph → kill boss → unlock exits → descend to -6
boss_kill_reward_polish: compact boss floor → kill Cave Warden → observe `boss_killed` with `boss_template_id` and client reward status
paladin_class_foundation / barbarian_class_foundation / sorcerer_class_foundation / rogue_class_foundation: class starter gear → movement → at least three basic attacks → all current class skills
inventory_capacity_and_paper_doll: fill base 15-capacity bag → reject full pickup → equip capacity belt → fill expanded 20-capacity bag
combat_control_and_boss_ai_fixes: equip training bow → fire directional free shot → prove damage, group aggro, and monster movement
session_browser_uncapped_coop: host creates listed co-op → two peers join from active list → prove three-player visibility, disconnect/reconnect, and replay
ui_currency_and_mana_polish: pick up gold instead of reward badges, persist character wallet, and use/reject blue mana potions
reachable_dungeon_obstacles: descend through generated dungeon floors → assert generated interior wall layout → route to loot beyond obstacles → prove replay
dungeon_wall_rendering: headless Godot client descends to generated floors → assert authoritative non-perimeter wall rendering state
vendor_appraisal_quotes: open compact vendor lab with rolled loot → assert server-authored offer summaries, comparisons, sell appraisals, buy, sell, replay
vendor_item_comparison: headless Godot client opens vendor → assert visible offer/sell details, comparison rows, buy, and sell
shop_stock_lifecycle: compact vendor generated stock → sell-to-buyback → rebuy → fresh-session buyback cleared and generated stock retained
client_shop_stock_lifecycle: headless Godot client opens vendor → sell to buyback → fixed purchase refresh → sell/rebuy buyback → assert fixed/generated rows remain visible
equipment_requirements_and_preview: pick up requirement-gated gear → reject unmet equip → level and allocate STR → equip → prove persistence
client_equipment_requirements_and_preview: headless Godot client opens inventory → assert requirement-status and equip-preview rows
skill_points_and_magic_bolt: level to 5 → learn Magic Bolt at baseline Magic 5 → cast → reject rank 2/cooldown recast → recover → prove replay/fresh persistence
client_skill_points_and_magic_bolt: headless Godot client opens skill panel → proves baseline Magic 5 availability and rank 2 Magic 8 gating → observes skill bar cooldown and recovery
rage_and_heal_skills: level to the second skill-point grant → learn Rage and Heal → cast Rage → fresh heal_lab session casts Heal and proves skill-sourced healing
menu_create_join_flow: Join Game empty state → Settings Create Game Type Solo → solo Create Game → existing-character fresh session
join_game_listed_session: protocol host holds active listed co-op session → Godot guest joins via Join Game → remote host visible
coop_rewards_and_scaling: compact three-account co-op → nearby host/guest share full XP → out-of-range guest excluded → replay/fresh persistence; different-level exclusion stays in lower-level tests
gold_autopickup_shared_loot: compact co-op loot lab → shared floor gold race → lowest player id wins private wallet update → item loot still requires click
account_stash_storage: acquire dungeon loot/gold → open town stash → deposit/withdraw item and gold → replay/reconnect/state/fresh session persistence
market_stash_listing_foundation: HTTP/store proof creates active market listing from stash item → browse active listings → reject foreign cancel → cancel back to stash
client_account_stash_panel: headless Godot client opens stash → verifies bag/stash item sync → deposits/withdraws item and gold
blacksmith_upgrade_ui: headless Godot client funds stash gold, deposits a rolled stash item, opens town blacksmith, upgrades once, and asserts item level/gold changes
live_unique_drops_all_effects: compact protocol lab picks up a deterministic unique rolled item and asserts its live effect_ids payload
ranged_monster_ai: compact archer lab → assert dungeon_archer → observe archer-sourced ranged player damage; generated archer placement stays in lower-level/client coverage
client_ranged_monster_ai: headless Godot client descends to generated dungeon → asserts bow marker → observes ranged player damage
client_boss_health_bar_ui: headless Godot client descends to first boss floor → asserts Cave Warden boss health bar
client_boss_phase_readability: headless Godot client descends to first boss floor → asserts boss phase countdown and telegraph marker
character_select_summaries: headless Godot client opens Create Game → asserts character row level/gold/depth/status summaries
mystery_seller_paid_reroll: open mystery seller → spend 50 gold to reroll concealed stock → prove old offers are replaced, gold persists, and replay/fresh stock remain deterministic
stash_search_and_sorting: headless Godot client opens stash → searches and sorts bag/stash rows → deposits/withdraws by stable server IDs
```

**Verify:**

```bash
make db-up && make server    # terminal 1
make bot                     # terminal 2 — all protocol bot scenarios
make client-unit             # headless Godot unit gates (no server required)
make client-smoke            # headless Godot gates + slice smoke
make bot-client              # Godot client bot scenarios; requires live server
make ci                      # full suite
make bot-visual              # optional — record all bot scenarios and watch replay playlist in Godot
make bot-visual scenario=07_inventory_lab.json  # optional — replay one scenario by file name or id
```

---

## Open gaps & deferred work

Do **not** assume these are the next slice — they are documented backlog items agents should know about.

### Recently closed

**Live unique drops now expose the full enabled effect catalog.** v119 marks named unique metadata
ready, keeps live behavior on rolled equipment `effect_ids`, proves every enabled unique effect can
be selected by at least one compatible template, and adds protocol scenario
`57_live_unique_drops_all_effects` for a deterministic unique drop.

**Item upgrades are now usable from a town blacksmith.** v118 adds a server-authored
`town_blacksmith` service in town and vendor lab, a focused Godot upgrade panel for account-stash
items, and client bot scenario `39_blacksmith_upgrade_ui`. The upgrade route remains authoritative,
and the store now supports both legacy flat rolled stats and current generated rolled-stat payloads.

**Elite command aura radius is now previewed in Godot.** v116 adds optional generated-pack metadata
to monster entity views, renders a display-only radius ring around visible pack leaders whose
followers are server-marked with `elite_command`, and proves the marker plus shared-radius debug
state with client bot scenario `37_elite_aura_radius_preview`.

**Market active offers are now inspectable and acceptable from Godot.** v117 adds seller-side offer
inspection to the market board, lets sellers accept an active item offer through the existing HTTP
contract, refreshes the listing list after acceptance, and proves the flow with client bot scenario
`38_market_active_offer_ui`.

**Market purchase is now usable from the Godot board.** v115 adds a buyer-only `Buy` action for
priced listings, calls the existing v111 purchase route, refreshes active listings, and proves the
flow with a seller-listing preflight plus client bot scenario `36_market_purchase_ui`.

**Market board priced listing UI is now proven in Godot.** v114 adds a deterministic publish price
control, sends `price_gold` through the existing listing-create HTTP route, renders listing prices in
browse rows, and proves stash-item publication through client bot scenario `35_market_board_ui`.

**Elite command aura is now client-readable.** v113 exposes server-owned `elite_command` aura state
through existing monster `effect_ids` when a generated pack follower is actively buffed, then renders
a compact Godot marker on those monsters with a focused client bot proof.

**Elite packs now have an authoritative aura foundation.** v112 adds a data-driven `elite_command`
aura under dungeon monster placement rules, preserves generated pack metadata on live monsters, and
applies a nearby living leader damage bonus to same-pack followers. Aura radius previews, additional
aura types, and richer production VFX remain deferred.

**Market listings can now be purchased for stash gold.** v111 adds optional `price_gold` on
market listings and a direct purchase route that atomically debits buyer stash gold, credits seller
stash gold, delivers the listed item to the buyer stash, marks the listing accepted, and refunds
active item offers. The first purchase proof stays store/HTTP-only; player-facing market UI,
notifications, expiration, fees, and listing edits remain deferred.

**Item upgrades can now repeat with scaling gold costs.** v110 extends the account-stash upgrade
route so equipment can upgrade through `item_upgrade_max_level`, charging
`item_upgrade_cost_gold + current_item_level * item_upgrade_cost_growth_per_level` from stash gold
and preserving deterministic stat mutation. The v110 review also catches the maintainability
baseline up to the post-v109 file sizes so the ratchet enforces from the current repo state.

**Barbarian and Sorcerer now have second combat skills.** v89 adds Cleave as a
server-owned cone weapon attack with pushback and Ice Shard as a cold projectile with stackable
slow plus deterministic shard fan-out; skill visuals now list and replay both skills from the
shared catalog.

**English text now has a shared catalog foundation.** v90 adds `shared/i18n/en.json`, schema and
validator coverage, skill/monster text keys, and a Godot text lookup service with fallback behavior.
Menu, pause, Settings, stat, class summary, skill, skill-bar, and status-effect helpers now resolve
through the catalog so v91 can add Spanish plus the Settings language selector.

**Spanish localization is selectable from Settings.** v91 adds `shared/i18n/es.json`, validates locale
catalogs against English, persists the selected language, and refreshes menu, pause, and Settings
labels immediately while falling back to English for missing keys.

**Town bishop respec service is live.** v92 adds a red `town_bishop` interactable that heals HP/mana
on activation and opens a compact service panel with a 250 gold Respec action. The server owns gold
deduction, stat reset/refund, skill rank refund, cooldown clearing, resource refill, and rejection
when the player cannot afford the service.

**Market listings now accept multi-item offers.** v93 adds active/accepted/rejected market offers
backed by 1-10 bidder stash items. Sellers can inspect offers, accept one offer to atomically swap
the listed item for offered items through account stashes, or cancel the listing and refund all
active offers.

**The first item upgrade action is server-owned.** v94 adds main-config tuning for starter upgrade
cost/max level plus an authenticated account-stash upgrade route. The store spends stash gold,
increments `item_level`, and increases one existing rolled stat deterministically while preserving
market eligibility for upgraded items.

**The unique item catalog has a disabled seed.** v95 adds schema-backed `unique_items.v0.json` with
`embercall_blade` as a non-player-facing unique concept. Validation cross-checks the base template
and keeps the seed disabled until a future behavior-changing unique effect path exists.

**Town now reads as a wider hub.** v96 distributes town services at least 5 tiles from the central
campfire, adds two procedural wood cabins, improves the town ground texture, and adds
`$showme --focus town` for focused visual feedback without changing server authority.

**New heroes now get class starter kits.** v97 seeds explicitly created paladins with sword/shield,
sorcerers with a two-handed staff, and barbarians with a slower harder-hitting axe, plus one health
and one mana potion. The starter staff also introduces item-backed max mana and skill damage scaling.
Follow-up minor improvements after v97 added dedicated starter staff/axe models, item-family
presentation assets, class-specific character models, magic scaling for existing skill effects, and
floor-loot presentation fixes. These are considered unversioned polish/consolidation commits, not
separate numbered slices; the next gameplay slice remains v98.

**Rogue class foundation is playable.** v98 adds Rogue as the fourth selectable class with a slimmer
deterministic character model, dagger class icon, dexterity-leaning starting stats, and a durable
starter kit of two common swords plus one health and one mana potion. Rogues can equip one-handed
melee weapons in `off_hand`; non-Rogue classes still cannot.

**Rogue starter skills are authoritative.** v99 makes Poison Stab deal weapon damage plus
rank-scaled poison ticks, makes Dash move through and damage crossed monsters from shared skill
data, and gives Rogues independent off-hand basic attacks at 1.5x main-hand cadence. The Rogue
foundation bot scenario now learns Dash and Poison Stab, dashes through a target, poisons it, and
observes two main-hand attacks plus one off-hand attack.

**Damage types and monster resistances are authoritative.** v100 adds canonical `force`, `cold`,
`poison`, and `lightning` damage types, skill/item fallback to `force`, monster resistance maps,
and `damage_type` on combat events. Lightning now deals half damage to flying lab/bat targets and
50% bonus damage to quadruped/wolf targets, proven through focused Go tests and the
`damage_types_and_resistances` protocol bot scenario.

**Undead poison immunity is playable and visible.** v101 adds a localized `dungeon_undead` monster
with full poison resistance, a generated skeleton GLB/scene wired through the monster visual
catalog, and a compact lab scenario. Poison Stab now applies poisoned status on a connected hit
even when resistance mitigates the hit to zero, and poison ticks against undead emit authoritative
zero-damage poison events instead of lowering HP.

**Every playable class has a foundation visual scenario.** v102 adds Paladin, Barbarian, and
Sorcerer class-foundation protocol/visual scenarios alongside the existing Rogue scenario. Each
scenario proves starter gear, movement, at least three basic attacks, and every current class skill;
Python coverage now fails if a playable class lacks a foundation scenario or a class skill is not
referenced by that class scenario.

**Skill visual replays now seed requested rank directly.** v88 lets `make skill-visual
skill=<id> rank=<n>` start from the requested class, minimum level/stats, and skill rank without
first killing an XP dummy or allocating a skill point during the replay.

**Combat/world state now persists on same-session resume.** v5 replays recorded
inputs before the WebSocket `session_snapshot`, so monster death, player HP,
inventory, equipped state, and ID continuity are restored authoritatively.

**World preset identity now persists on sessions.** v7 stores `world_id`, so fresh WebSocket attach,
resume, `/state`, replay verification, and replay timeline all reconstruct the same initial layout.

**Equipped weapon damage now changes authoritative combat.** v8 resolves `rusty_sword.damage`
from equipped server state at attack time and proves the equipped gear scenario kills the reward
dummy in one acknowledged attack.

**Solid collision now blocks movement through bodies and walls.** v9 resolves player movement
against live monsters and static world walls, while collision lab proves routed movement and
deterministic replay.

**Click action and melee reach are now authoritative.** v10 unifies combat/pickup/door activation
behind `action_intent`, enforces reach from shared rules, and proves a replayable opening door.

**Click-to-move and action auto-approach are now authoritative.** v11 adds shared navigation
rules, deterministic server A*, `move_to_intent`, and a path-maze bot proof.

**Ranged projectile combat is now authoritative.** v12 adds ranged weapon rules, projectile
entities, swept collision, impact-time hit/damage, and a ranged-lab bot/replay proof.

**Inventory UI, unequip, and player drop are now authoritative.** v13 adds protocol-backed
unequip/drop intents, deterministic adjacent loot placement, persisted inventory removal, and a
display-only Godot panel that mirrors server snapshots/deltas.

**Current item presentation is now shared-data-driven.** v15 adds presentation metadata for all
current item definitions and uses it for inventory icons and ground loot silhouettes without
server/protocol changes.

**Consumable healing is now authoritative.** v16 adds `use_intent`, red potion heal rules, HP cap
goldens, server-owned inventory removal, and a client-only hotbar that sends use intents.

**Monster chase movement is now authoritative.** v17 adds opt-in chase behavior, deterministic
monster pathing around solids, leash return, chase/lab bot scenarios, and client walk presentation
from position deltas.

**Dungeon levels, stairs, teleporters, town entry, and dungeon monster threat are now authoritative.**
v18 adds multi-level dungeon state and generated stairs; v19 adds generated teleporters, session
discovery, and server-owned fast travel with a client-only waypoint panel; v20 makes town level `0`
the fresh play-session entry and keeps dungeon floors lazy; v21 spawns deterministic hostile dungeon
mobs that chase and proactively damage the player.

**Character inventory/equipment and waypoint unlocks now persist across fresh sessions.** v22 moves
durable item instances and discovered waypoint levels to the default character, while preserving
session-start snapshots for deterministic replay and keeping HP, dungeon maps, monsters, corpses,
opened doors, and floor drops session-scoped.

**Dungeon mobs now drop rolled weapon gear.** v23 adds server-authoritative item templates,
deterministic rarity/stat rolls, rolled weapon damage, rolled payload persistence, and tooltip
presentation for the first rolled weapon template.

**The client now has a player-facing menu shell.** v24 adds named character list/create APIs,
fresh-session Continue/New Game flows, local window-size settings, ESC pause, Return to Main Menu,
and a Godot client bot proof for the complete menu path.

**Treasure classes and guarded chests are now authoritative.** v25 adds data-driven multi-attempt
monster/chest loot, deterministic rare chest generation with guarded monster bonus, open-once chest
loot, and bot/golden coverage for the complete path.

**Character stats and leveling are now authoritative.** v26 adds durable XP, levels, stat points,
base stats, derived substats, stat allocation, VIT max-HP effects, STR damage contribution, a
Godot character sheet, an XP bar, and protocol/client bot proofs.

**Sustained left-click controls are now client-side.** v27 adds hold-to-attack on monsters and
hold-to-move on floor by repeating existing `action_intent` / `move_to_intent` at `SEND_INTERVAL`,
with sticky targets, move epsilon, and headless unit coverage — no protocol or server changes.

**Full paper-doll equipment and belt-gated hotbar are now authoritative.** v28 replaces the single
weapon slot with full equipment slots, two-hand occupancy, droppable gear templates, persisted
character hotbar layout, replay-safe session hotbar snapshots, and protocol/client bot proofs for
server-synced paper-doll and belt capacity behavior.

**Generated dungeon drops now reach the expanded equipment catalog.** v29 adds temporary depth
bands, depth-specific monster/chest treasure classes, validation for full v28 template reachability
by depth `3+`, golden fixtures for varied equipment outcomes, and a real generated dungeon bot proof.

**Generated dungeon monster rarity now scales challenge and loot depth.** v30 adds deterministic
generated monster rarity tiers, scaled HP/damage/XP, effective monster loot depth offsets,
monster rarity in protocol/replay state, player/enemy tinting, and a real generated dungeon bot
proof for non-common rarity.

**Combat stats now affect authoritative outcomes.** v31 applies hit, crit, armor, block, minimum
damage, and effective stat breakdowns across player and monster combat, then renders normal, crit,
miss, and block feedback from protocol events in Godot.

**The test floor now separates contracts from tuning details.** v32 keeps exact locks for replay,
schema, formula parity, persistence boundaries, and named UI/protocol contracts, while converting
brittle dungeon size, generated population, movement timing, rarity tuning, and selector-index
assumptions to semantic, range, derived, or eventual checks.

**True two-player co-op sessions are now authoritative.** v33 adds server-owned co-op session
membership, hashed join codes, actor-tagged inputs, per-player sim state, recipient-scoped realtime
snapshots/deltas, remote-player Godot rendering, and a protocol bot proof for join, movement,
disconnect/reconnect, and replay.

**Character-like model reactions are now unified in the Godot client.** v34 adds client-only
hit/death transform and tint reactions for local players, remote co-op players, and monsters;
remote co-op players now reuse the humanoid character model with a distinct dark tint.

**The first boss floor gate is now authoritative.** v35 adds a compact level `-5` boss arena,
telegraphed boss phases, locked down-stair/teleporter exits until boss death, boss visual scale
metadata, and protocol/replay/bot proof for unlock and descent.

**Inventory capacity and the paper-doll bag grid are now authoritative.** v36 adds server-derived
`inventory_rows` / `inventory_capacity`, an item-granted row source, full-bag and overflow rejection
guards, a 5-column capacity grid, and protocol/client bot proofs.

**Combat control and boss AI fixes are now authoritative.** v37 adds server-owned directional attacks,
authoritative stop movement, aggro-on-hit with nearby contagious group aggro, boss chase/damage repair,
and protocol/client unit proofs.

**Session browser and uncapped co-op are now authoritative.** v38 adds persisted listed co-op sessions,
active session summaries, listed join without join code, three-plus-member realtime/replay proofs, a
Godot Multiplayer menu path, and local/remote multi-client menu launchers. Empty listed sessions are
hidden from discovery, and a listed session is ended when its last connected player disconnects.

**Character gold, mana, and related UI polish are now authoritative.** v39 adds durable character
gold, currency loot pickup, generated gold scaling, snapshot/delta/replay wallet coverage, player
mana, blue mana potions, DEX-sourced armor, and Godot HUD/inventory/menu polish.

**Reachable generated dungeon obstacles are now authoritative.** v40 adds deterministic generated
interior dungeon walls, obstacle reachability retries, authoritative protocol wall layouts,
Godot server-layout rendering, and protocol/client bot proofs that generated walls exist without
blocking generated targets.

**The town vendor and first gold sink are now authoritative.** v41 adds the `town_vendor`, protocol
v4 shop buy/sell contracts, fixed potion stock, deterministic generated offers based on deepest
dungeon depth, durable gold mutations, deepest-depth persistence, and protocol/client bot proofs for
shop open, buy, sell, reconnect, replay, and fresh-session persistence.

**Vendor appraisals and direct item comparison are now authoritative.** v42 extends `shop_opened`
with server-authored summary, appraisal, and comparison views, and proves the richer protocol plus
Godot panel through protocol/client bot scenarios.

**Equipment requirements and equip previews are now authoritative.** v43 expands item-template
requirements to level/base stats, rejects unmet equips before mutation, annotates loot/inventory/shop
views with server-authored requirement status and equip-preview deltas, and proves the path through
protocol and Godot client bot scenarios.

**Skill points and Magic Bolt are now authoritative.** v44 adds durable skill points/ranks,
protocol v5 skill state, attack-speed-derived cooldowns, a server-owned Magic Bolt cast/reject/recover
loop, and protocol/client bot proofs through replay, reconnect, and fresh-session persistence.

**Menu Create Game and Join Game flows now match the backend session model.** v45 replaces the
player-facing Continue/New Game/Multiplayer root menu with Create Game, Join Game, Settings, and
Exit; persists the Create Game Type setting; and proves co-op/solo create plus Join Game empty-state
behavior through client bot scenarios.

**The real Godot Join Game path now has a multi-account listed-session proof.** v46 adds a
client-bot preflight host that holds an active listed co-op backend session, then drives a separate
Godot guest through Join Game, character selection, listed join, WebSocket connect, and remote-player
presence assertions.

**Town vendor stock is now finite and refresh-gated.** v47 persists per-character generated stock,
consumes purchased generated offers, refreshes stock only on newly unlocked non-town waypoints,
limits shop rarity to `rare`, and keeps buyback rows session-local and cleared when the actor leaves
town.

**Co-op rewards and monster scaling are now authoritative.** v48 grants full monster XP to nearby
eligible party members, excludes dead/disconnected/far/different-level members, scales monster
HP/damage logarithmically with active same-level party count, and routes private progression by
recipient owner.

**Gold is now auto-pickable, but loot stays shared.** v49 keeps one shared floor entity per drop,
adds passive gold pickup for the first eligible player in deterministic order, and leaves non-gold
items click-required. Personal loot, reservations, hidden/duplicated drops, shared/split gold, and
item auto-pickup remain deferred/non-goals.

**Account stash storage is now authoritative.** v50 adds a town stash interactable, protocol v7
stash contracts, account-owned item/gold persistence, replay-safe session-start stash snapshots,
server-owned item/gold transfers, owner-private realtime fanout, and protocol/client bot proofs for
item and gold storage across fresh sessions.

**Mystery seller core is now authoritative.** v51 adds a town mystery seller, protocol v8 concealed
shop rows, deterministic per-character hidden stock, reveal-on-purchase events, and protocol/client
bot proofs for hidden offers, purchase reveal, replay, and fresh-session consumed stock.

**Ranged monster AI is now authoritative.** v52 adds generated dungeon archers, data-driven
melee/ranged monster composition, server-owned monster projectiles that respect walls and target
players, and a minimal Godot bow marker with protocol/client bot proofs.

**Boss health bar UI is now client-visible.** v53 adds a top-center Godot boss health bar driven by
existing authoritative boss entity hp/max hp and metadata, plus client unit and bot scenario proof
for the first `cave_warden` boss floor.

**Character select summaries are now server-authored.** v54 extends `GET /v0/characters` with
level, gold, and deepest-depth summary fields, renders them in the Godot character picker, and
proves the menu path with focused store, HTTP, client-unit, and client-bot coverage.

**Monolith decomposition and quality gates are now in place.** v55 proves that the god-file
tax from the v53 review can be paid down without behavior change: the sim.go handler registry
(handlers.go, −1,056 LOC from sim.go) means new intents never touch the dispatcher; ItemRulesLoader
eliminates ×5 GDScript item-loader duplication; ShopRNG and bot_types.py are now importable
independently. The determinism lint (`make lint-determinism`) converts the core sim invariant from
CLAUDE.md prose to a failing CI step; `make regen-golden` closes the manual-edit correctness
hazard on golden fixtures; and `test_delta_apply.gd` adds the first unit coverage to the
highest-risk zero-tested client code. All 265 Go tests, 59 Python tests, and 15 GDScript unit
tests pass; CI is now 9 phases.

**Generated monster attacks are a little faster.** v56 tunes regular generated dungeon monsters
without changing damage, movement, bosses, or lab fixtures: `dungeon_mob` cooldown is now 32 ticks
and `dungeon_archer` cooldown is now 75 ticks. The dungeon monster attack golden owns the melee
cooldown, Go/GDScript golden checks cross-check it against shared rules, protocol bot scenarios
prove archer damage, boss-floor traversal, and skill-progression combat, and a missing
`item_rolls.json` description field found by `make validate-shared` is restored.

**Boss phase readability is now client-visible.** v57 keeps server combat unchanged and adds
display-only phase state to the Godot boss health bar: phase kind, pattern id, phase index,
duration, remaining ticks, and phase ratio. Telegraph phases now attach a primitive
`BossTelegraphMarker` under the boss using server-authored radius/color, and the client bot runner
can assert both the bar countdown and the in-world marker.

**Cave Warden has a second boss pattern.** v58 adds `ground_slam`, a data-driven circle telegraph
and active hit shape, to the Cave Warden deck after `charged_melee`. The Go sim now cycles boss
pattern decks deterministically in declared order, resolves circle boss hits server-side, and the
protocol bot can assert phase events by payload fields such as `pattern_id`.

**Magic Bolt is now catalog-driven.** v59 moves Magic Bolt into a schema-backed skill catalog with
class/tree metadata, bounded requirement/cost/damage/projectile/cooldown helpers, and a separate
skill presentation catalog. The server now validates supported skills generically and enforces
`magic >= 15` for both learning and casting; Godot resolves skill panel and hotbar labels/tooltips
from shared skill data while server progression and cooldown state remain authoritative.

**The first content-library manifest is live for skills.** v60 adds a schema-backed
`shared/content/content_libraries.v0.json` index for skill rules and skill presentations. Go and
Godot loaders now resolve skill content through manifest paths while runtime state, protocol,
replay, goldens, and UI state keep stable skill IDs such as `magic_bolt`. Validation and focused
tests prove relative path resolution, duplicate-ID rejection, and unknown manifest group rejection.

**Rage and Heal are now authoritative active skills.** v61 expands the data-driven skill catalog
with closed declarative effect rows for a self stat-percent buff and an allied area percent heal.
The server owns mana, cooldowns, buff expiry, max-HP sync, visual scale metadata, and skill-sourced
healing events; the Godot skill tree and hotbar now select among multiple first-row skills.

**Generated monster stats now scale by dungeon depth.** v62 moves regular generated dungeon monster
HP, damage, XP, and related combat pressure onto depth-aware shared rules while keeping boss
templates bespoke. Go and GDScript golden checks prove the scaling rules, and protocol bot coverage
keeps generated dungeon combat replayable.

**Default sim construction now returns errors instead of panicking.** v63 changes the exported
`game.NewSim` default-world constructor to return `(*Sim, error)`, adds explicit `MustNewSim`
panic behavior for tests with known-valid fixtures, and covers invalid default-world construction
without crashing. This closes the v60 backend review's runtime sim construction finding.

**Mystery seller paid reroll is now authoritative.** v64 adds the `shop_reroll_intent`,
a server-owned 50 gold spend, deterministic `|reroll:N` stock refresh keys, complete replacement
of concealed mystery stock, and protocol/client bot proofs for reroll, replay, and fresh-session
persistence.

**Stash search and sorting are now client-visible.** v65 adds display-only Godot controls for
searching and sorting bag/stash rows by acquired order, name, rarity, or slot, while keeping all
deposit/withdraw mutation keyed by server-authored `stash_item_id` / inventory item IDs.

**Progress backlog hygiene is current through v66.** v66 corrects the canonical discovery metadata
after v64/v65 by marking shipped candidates complete, adding their scenario catalog entries, and
narrowing deferred backlog text to still-open adjacent work.

**Boss kill reward status is now explicit.** v67 emits a dedicated `boss_killed` event with
`boss_template_id` for boss deaths while preserving the existing `monster_killed`, loot, XP, and
exit-unlock flow. The Godot client now exposes a `Cave Warden defeated` reward status, and protocol
plus client bot coverage prove the boss-specific signal.

**Market listings now have a stash-backed foundation.** v68 adds active/canceled market listing
persistence and authenticated HTTP routes to create a listing from an account stash item, browse
active listings, and cancel an owned listing back to stash. Offers, purchases, pricing, expiration,
and Godot market UI remain deferred.

**Character class identity is now authoritative.** v69 persists `character_class`, exposes it in
character APIs, validates create requests against shared progression class rules, and uses class
rules to seed fresh progression. The default `barbarian` preserves the prior baseline stats while
`sorcerer` and `paladin` prove divergent starts.

**Class gates now affect gameplay.** v70 maps each skill to its class, rejects cross-class skill
spend/cast attempts, adds one fixed class-required weapon per class, and rejects wrong-class weapon
equips. Session-start snapshots and realtime reconstruction carry class identity so restrictions
survive the authoritative boundary.

**v70 engineering review steering.** The v70 review recommends a small maintenance follow-up for
realtime fanout level snapshots and defensive `equipped_update.slot` handling in Godot/Python bot,
plus keeping the v71 class picker contained in character-select UI with shared class presentation
metadata where practical.

**Class creation is now player-facing.** v71 adds class picker blocks with code-native class
sprites, hover tooltips for class stats/skills, selected-class create request plumbing, and class
icons at the start of character rows. Client unit and bot coverage prove Sorcerer selection without
changing server authority.

**Monster visuals are now catalog-driven.** v72 adds shared monster visual metadata, deterministic
quadruped/flyer placeholder assets, wolf/bat monster definitions with unchanged chase mechanics,
and deterministic boss model pools across dummy, quadruped, and tiny flyer visuals. Godot resolves
monster scenes through the catalog, and the showme monster lineup was approved before final CI.

**Stats and skills now use draggable window chrome.** v73 adds a reusable Godot titlebar shell with
close button, titlebar-only dragging, viewport clamping, and debug proof, then migrates the
character stats and skills panels without changing server authority or gameplay protocol.

**Gameplay item panels now share draggable chrome.** v74 migrates inventory, shop, and stash onto
the reusable titlebar shell while preserving item drag/drop, buy/sell, reroll, stash search/sort,
and existing gameplay-panel APIs.

**Custom gameplay window layout now persists locally.** v75 saves and restores clamped positions
for character stats, skills, inventory, shop, and stash through `user://window_layout.cfg`, while
disabling normal persistence during client unit tests.

**Main gameplay config foundation is now validated.** v76 adds `main_config.v0.json`, exposes it
through the Go rules loader, and adds drift guards against current combat, movement, and dungeon
monster drop defaults until follow-up slices consume those values directly.

**Main gameplay config now drives attack cadence and movement.** v77 makes
`base_attack_interval_ticks` and `base_movement_speed` operational server gameplay inputs, with
focused tests proving edits to `main_config.v0.json` take effect without touching older rule files.

**Main gameplay config now drives dungeon monster drop chance.** v78 applies
`base_drop_rate_percent` to dungeon monster treasure-class primary attempts during rules loading,
so the global drop chance can be tuned from `main_config.v0.json` without hand-editing each depth
class.

**Generated dungeon fights now form packs.** The pack-aggro slice adds data-driven pack sizing,
monster assist radius, deterministic close pack placement, and a protocol bot proof that damaging
one generated monster can emit multiple `monster_aggro` events. This landed after the already-used
v76-v78 main-config slice numbers, so the as-built/spec files retain the requested v76 label while
the canonical lifecycle continues through v79/v80.

**Generated packs now have role and leader foundations.** v79 adds internal pack roles, pack
composition constraints, and deterministic elite leader markers so future elite behavior can build
on structured encounters without exposing new protocol fields yet.

**Combat threat readability is now visible.** v80 maps existing authoritative `monster_aggro`
events to display-only `AGGRO` floating text in the Godot client, adds a `threat` damage-number
variant, and proves it with client unit, focused client-bot, and protocol pack-aggro coverage.

**v80 engineering review steering.** The v80 review keeps the repo at 8.4/10 overall and recommends
small follow-ups for combat event presenter extraction and splitting the largest Python validation/bot
files by domain. The realtime fanout level snapshot finding was closed in v82, defensive client
payload parsing was closed in v83, and client bot step registry duplication was closed in v84.

**v90 engineering review steering.** The v90 review keeps the repo at 8.5/10 overall, confirms
`make ci` green after blocker cleanup, and steers the next batch toward localization: central English
text keys, Spanish translations, Settings language selection, and English fallback. It also keeps
the standing rule that touched large files should be split or shrunk rather than extended.

**v100 engineering review steering.** The v100 review keeps the repo at 8.5/10 overall and steers
the next gameplay batch toward shared damage types, data-driven monster resistances, and focused
combat helpers/tests. Undead poison immunity should consume the same resistance contract rather
than introducing a one-off immunity flag. The review gate also records a maintenance exception for
the already-landed v99 growth in `main.gd`, `test_item_visuals.gd`, `sim.go`, and `tools/bot/run.py`;
future slices touching those files should split or shrink them before adding more behavior.

**Maintainability ratchet is now explicit.** New source/test/tool files target a 600-line maximum,
existing over-limit files are grandfathered in `.maintainability/file-size-baseline.tsv`, and
`make maintainability` enforces that new files do not exceed the target while grandfathered files
do not grow by more than 25 lines without a documented maintenance exception.

**Paladin Holy Shield is now authoritative and visible.** v81 adds a data-driven Paladin area
defensive buff, per-player active effect ids, armor/block stat application, server-owned expiry,
Rage-style status UI presentation, and a gold shield/shine around every affected hero.

**Realtime fanout now uses tick-time client level snapshots.** v82 captures connected client levels
under the session loop mutex and passes that snapshot into fanout, closing the v70/v80 review
finding without changing protocol, gameplay, or client presentation. The slice also removes stale
attack-interval-derived exact expectations from Magic Bolt, Rage, Heal, Holy Shield, and matching
client bot scenarios that CI surfaced; exact cooldown math remains owned by shared golden tests.
The model-reaction client scenario now uses a safe low-HP lab dummy for terminal reaction proof
instead of depending on a long basic-attack sequence against combat-stat targets.

**Client envelope payload parsing is now defensive.** v83 routes central Godot `_handle_message`
payload access through a dictionary guard, so missing/null/non-dictionary payloads on accepted,
rejected, error, and delta envelopes no longer crash the client message boundary.

**Client bot step registration now has one source of truth.** v84 derives `ALL_STEP_TYPES` from
the wait/assert/action category arrays in `bot_scenario_runner.gd`, preserving unknown-step
validation while removing a duplicated maintenance list.

### Other deferred items (from specs / ADRs)

| Area | Deferred item | Source |
|------|---------------|--------|
| Persistence | Player-facing old-session resume, delete/rename characters, class selection, visual customization, portraits, richer character detail panels, stash tabs/capacity upgrades, town stash delivery/market receipts, quest progress, passive skills, respec/refund, respawn/checkpoints, durable dungeon map snapshots, durable buyback history, starter loadout backfill for existing or compatibility-default characters | v22/v24/v26/v39/v40/v41/v44/v45/v47/v50/v54/v59/v97 non-goals, ADR-0008 deferred, ADR-0011, ADR-0014 |
| Combat | Basic-attack cooldown rebalance, animation-speed scaling, mana regeneration, respawn, richer spell systems, piercing/AoE/homing projectiles, debuffs/DOT/status effects, summons/traps/auras, richer ranged monster AI, quadruped pounce, bat dive/swarm behavior, true flying gameplay/pathing, ranged boss patterns, elite archer packs, retreat/cover seeking, predictive leading, final ranged monster damage/range/cooldown balance, final combat balance across damage/HP/movement/rarity/depth, depth scaling beyond loot bands, offhand abilities/dual-wield, named elite packs/minions/aura modifiers, additional boss templates/pattern decks beyond the v58 Cave Warden deck, enrage phases, summoned adds, monster population-count scaling, weighted/random boss pattern selection, final skill tree and active ability catalog, additional active skills beyond Rage/Heal/Magic Bolt/Holy Shield, free-form skill formulas, class-locked skill trees, skill capability expansion beyond projectile/self-buff/area-heal/area-stat-buff, PvP/friendly fire | v0/v4/v12/v17/v21/v23/v26/v28/v29/v30/v31/v32/v35/v37/v39/v40/v44/v48/v52/v56/v57/v58/v59/v61/v72/v81 non-goals |
| Itemization | Affix grammar, procedural item names, special-effect execution, loot filters, crafting, richer gold sinks, Magic Find, unique/set catalogs, unique items that change skill/build behavior, unique monster special drops, final item-level/depth progression, item upgrade resources, item-owned levels, success-chance add/improve-roll upgrades, richer boss drop economy, richer dungeon drop economy, expanded shop depth economy bands, item sorting/filtering, multi-cell item footprints, passive skill sources for inventory rows and equipment requirements, item auto-pickup | v23/v25/v26/v28/v29/v30/v35/v36/v39/v41/v42/v43/v47/v49/v51 non-goals, ADR-0009 deferred, ADR-0012, ADR-0013, ADR-0014 |
| Economy / trade | Gold/resource pricing beyond direct stash-gold listing prices, 24-hour expiration/delisting, active-offer item locking/reservations beyond listing-row removal, trade audit records, market restrictions for upgraded/bound/equipped/hotbar-assigned items, clock/timer/daily mystery refresh, account-wide mystery stock, stash overflow delivery for purchases, mystery refunds/binding/special resale, final mystery price tuning against visible vendor prices, clock-based shop refresh, long-term market endgame loops for advanced players | v33/v38/v41/v42/v47/v51/v64/v68/v111 non-goals, ADR-0011, ADR-0012, ADR-0013, ADR-0014 |
| Content | Production item art/icons, production menu art/audio, production town/vendor/stash/mystery-seller art, production imported town building assets, collision-aware town decorations, ambient NPC movement, production dungeon art/lighting/sound, production chest art/animation/audio, production archer/bow model and attack animation, production monster art/VFX/audio, production boss art/VFX/audio, generalized ranged-monster equipment overlays, production combat/skill VFX/audio beyond code-native placeholders, production paper-doll art/model preview, colorblind/accessibility-safe rarity presentation, additional NPCs/vendors, mystery seller presentation polish, additional item families beyond current rules, full content-library manifest/index rollout beyond skills for items, classes, and broader presentation assets | v15/v20/v23/v24/v25/v28/v29/v30/v31/v32/v35/v36/v37/v39/v40/v41/v42/v43/v44/v45/v47/v50/v51/v52/v57/v58/v59/v60/v72/v81/v96/v97 non-goals, ADR-0013 |
| Client presentation | Boss portraits, multi-boss layouts, exact authoritative boss countdown sync, production shape-specific telegraph decals/VFX/audio, production boss health bar art/audio, draggable titlebar migration for waypoint/menu windows, reset-layout UI, server/account-synced UI layout | v53/v57/v58/v73/v74/v75 non-goals, ADR-0009 |
| Dungeon generation | Generated doors in obstacle walls, full room/corridor PCG, rotated/polygon/destructible/secret obstacles, boss-floor obstacle generation, final obstacle density/biome/difficulty balance | v40 non-goals |
| Client controls | Reliable full-scene headless modifier/mouse proof for `SHIFT+LMB` stationary attack; v37 covers the behavior with Godot unit helpers and protocol bot coverage instead | v37 deferred |
| Testing / tooling | Tuning-friendly rule tests: audit hardcoded values copied from `shared/rules/*.json` across Go/GDScript/Python/bot scenarios, classify each as contract/golden/accidental tuning pin, and convert accidental pins to rule-derived, semantic, range, or eventual assertions. Goal: balance changes such as `training_dummy.max_hp`, skill mana costs, monster cooldowns, loot weights, and generated population tuning should not require unrelated test edits; exact values remain only where a named golden or protocol/schema contract intentionally owns them. | v32 test-locking policy follow-up, v76/v77/v78 deferred |
| Settings | Fullscreen, audio, controls remapping, accessibility options, graphics quality, language selection | v24 non-goals |
| Assets | Blender export pipeline, texture budget, remote patcher | ADR-0006 |
| Platform | Production auth provider, dashboards, historical inspect API | v0 §8, ADR-0001 |
| Protocol | Protobuf / `godobuf` migration | ADR-0001 |
| Multiplayer | Matchmaking/lobby beyond backend-listed sessions, active-session filters/search/sorting controls, Steam lobby/invites, friend flows, richer party UI, chat/emotes/ready checks, richer party reward bonuses beyond full shared XP and HP/damage scaling, loot allocation, personal/hidden/reserved loot, shared/split gold, friendly fire/PvP, production remote-player art, load-aware capacity limits, split deployables / cross-process session ownership, co-op roles/encounters that change the solo experience, PvP rules that preserve skill expression while respecting builds | v0/v33/v38/v45/v46/v48/v49 non-goals, ADR-0001, ADR-0014 |
| Companions / AI | Hired mercenaries derived from other players' characters, mercenary follow/aggro/combat AI, mercenary death/loss rules, pricing/listing model, gear snapshot refresh rules, limits per player/party, mercenary loot/XP/potion behavior | ADR-0010 |

### Curated autoloop candidates

These candidates were curated during `$autoloop 1` on 2026-06-10 and should be considered first by
the next autoloop pass unless code changes make them stale.

| Candidate | Status | Value | Size | Touch surfaces | Main risk / dependency |
|-----------|--------|-------|------|----------------|------------------------|
| `boss-phase-timer-ui` | Completed in v57 | Add boss phase/windup timing cues to the existing boss health bar. | S | client, bot, docs | Kept display-only from existing `boss_phase` state/events. |
| `boss-pattern-variety` | Completed in v58 | Add one more server-authored boss attack pattern so Cave Warden is less repetitive. | M | shared, server, bot, docs | Implemented deterministic deck-order cycling and server-owned circle hit shape. |
| `data-driven-content-library-manifest` | Completed in v60 | Introduce a manifest/index loader for skills first, preserving stable gameplay IDs and deterministic merge validation. | M | shared, server, client loader, validation, docs | Shipped as skills-only; item/class rollout remains deferred. |
| `mystery-seller-paid-reroll` | Completed in v64 | Let players spend gold to reroll concealed mystery seller stock. | M | shared/protocol, server, store, client, bot, docs | Shipped with a 50 gold server-owned reroll and deterministic stock replacement. |
| `stash-search-and-sorting` | Completed in v65 | Add search/sort controls to stash and bag views without changing item authority. | S/M | client, bot, docs | Shipped as display-only Godot controls with server-ID mutation safety. |
| `character-select-summaries` | Completed in v54 | Show level, gold, deepest depth, and status in character selection. | M | store, HTTP, client, bot, docs | Needs careful aggregate/query shape; rename/delete already exists. |
| `session-browser-filters` | Open | Add Join Game search/filter/sort controls for listed sessions. | S/M | client, HTTP tests maybe, bot, docs | Headless proof needs stable multi-session setup. |
| `loot-label-filter-core` | Open | Add display-only loot label filtering/highlighting for rarity/category. | M | client, bot, docs | Presentation-only; avoid changing shared loot ownership. |
| `tuning-friendly-rule-tests` | Open | Make shared-rule balance tuning less brittle by replacing accidental hardcoded rule values in tests/scenarios with rule-derived or semantic assertions. | M | shared, server tests, client tests, bot scenarios, validation docs | Must preserve exact locks for schemas, replay determinism, persistence boundaries, and named goldens. |
| `client-boss-telegraph-polish` | Completed in v57 | Improve boss telegraph readability with a clearer in-world warning marker. | S/M | client, bot, docs | Reused in-repo primitive marker patterns; external plugins/assets rejected. |

---

## Starting a new task (agent checklist)

1. **Read this file** (`PROGRESS.md`) — confirm baseline slice and open gaps.
2. **Read ADR-0001** and any feature-specific ADRs listed above.
3. **Spec first** — create or read `docs/specs/vN_spec-<feature>.md` (SDD; `N` = next execution order).
4. **Plan second** — create `docs/plans/vN_<YYYY-MM-DD>-<feature>.md` with file map + verification commands.
5. **Branch** — stay on the current checkout; do not create branches (user creates them before development if needed).
6. **Implement** shared → server → client → bot/smoke → docs; keep `make ci` green.
7. **Update this file** when the slice completes: lifecycle table row, open gaps, and status fields.
   Write the as-built summary in `docs/as-built/vN_<codename>.md` (not inline here).
8. **Engineering review cadence** — when the latest completed slice hits the next ~10-slice milestone
   (see **Next engineering review** above), write or refresh the review set under [`docs/reviews/`](docs/reviews/)
   before piling on more feature slices.

### Invariants (do not break)

- Go sim determinism: seeded RNG only, no wall-clock in `game/`, stable ordering.
  **Enforced by CI gate:** `make lint-determinism` (step 3/9) — fails on `time.Now()`,
  `math/rand` import, or bare map range (key+value) in `sim.go` / `handlers.go`.
- New intents: register one entry in `handlers.go inputHandlers` map — do **not** edit
  `applyInput` in `sim.go`. The dispatcher is a registry lookup now.
- Shared rules are **data**; formulas evaluated in Go + GDScript from the same golden fixtures.
  After intentional formula changes: `make regen-golden` → `make ci` to keep goldens current.
- ADR-0014 progression/endgame rules are challenge rules. If a requested direction conflicts with
  stats/skills/passives mattering, loot hope, economy value, low complexity, meaningful uniques,
  endless progression, fair deaths, survival passives, all-level endgame, co-op differentiation,
  skill-based PvP, or market-as-endgame, pause and ask the owner to justify the exception before
  speccing or implementing it. Record accepted exceptions in the spec or plan.
- Animation is client-only; new reactions need a **server event** first, then client mapping.
- Golden changes require Go tests **and** GDScript `test_golden.gd` / `validate_shared.py` updates.
- GDScript shared data singletons: use `class_name Foo extends RefCounted` with `static var`
  and `ensure_loaded()` guard. Do **not** use Godot autoload for anything that headless tests
  `preload()` — autoload names are not resolvable at GDScript compile time without `--import`.

---

## Repo map (quick reference)

```text
client/          Godot 4.6.3 — main.gd, animation_controller.gd, net_client.gd, smoke.gd
server/          Go — internal/game (sim), internal/realtime (WS), internal/store (Postgres)
shared/          protocol schemas, rules JSON, golden fixtures
tools/           bot, replay, validate_shared.py, assets/
assets/          manifests + gen scripts
docs/            ADRs, specs, plans, as-built, reviews (periodic ~every 10 slices)
```

**Agent entrypoints:** [`CLAUDE.md`](CLAUDE.md) (commands + architecture), this file (progress),
[`README.md`](README.md) (human onboarding), [`docs/reviews/`](docs/reviews/) (periodic engineering audits).
