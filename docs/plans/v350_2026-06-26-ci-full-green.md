# v350 Plan — CI-Full Green

Status: Ready for implementation  
Goal: Fix all 15 extended scenarios from v349 review so `make ci-full` is fully green.  
Architecture: Triage by failure cluster; fix root causes in server sim, bot movement helpers, client presentation, or scenario contracts; confirm each scenario individually before one final `make ci-full`.  
Tech stack: Go authoritative sim, shared rules/world presets, Python protocol bot, Godot client bot.

## Spec review

| Check | Result |
|-------|--------|
| Baseline v349 / codename `ci-full-green` | OK |
| Scope / non-goals | OK — recovery only |
| Contracts | OK — only if fix requires |
| Determinism | OK — no wall-clock in sim hot path |
| Bot proof | OK — all 15 scenarios listed |
| Client asset decision | N/A |

**Proceeding to implementation.**

## Baseline and shortcut decision

- Builds on v349 (`movement-tick-smoothing`, `forward-plus-renderer`).
- Authoritative failure list: [`docs/reviews/20260626_v349-ci-full-failures.md`](../reviews/20260626_v349-ci-full-failures.md).
- **Do not** run full `make ci-full` until all 15 pass in isolation.
- Reuse existing labs; no new world presets unless layout fix requires entity reposition only.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/bot/movement_runtime.py` | `walk_toward`, pathfind caps, `greedy_walk_max_ticks` |
| Modify | `tools/bot/run.py` | `pick_up_loot`, `walk_to_monster`, coop waits (minimal) |
| Modify | `tools/bot/scenarios/*.json` | Budget/step tuning for 15 scenarios as needed |
| Modify | `shared/rules/worlds.v0.json` | Lab layout if loot unreachable |
| Modify | `shared/rules/main_config.v0.json` | Movement tuning only if rules-owned regression |
| Modify | `server/internal/game/*.go` | Combat, coop loot, mercenary companion damage |
| Modify | `client/scripts/main.gd` | Level transition, combat click, panel routing (minimal) |
| Modify | `client/scripts/wall_renderer.gd` | Wall layout debug / rollout |
| Modify | `client/scripts/*blacksmith*` | Blacksmith panel state |
| Modify | `tools/bot/test_*.py` | Movement/coop unit tests if helpers change |
| Modify | `docs/as-built/v350_ci-full-green.md` | Slice proof |
| Modify | `PROGRESS.md` | Lifecycle on `/finish` |

## Maintenance ratchet

Hotspot files possibly touched:

- [ ] `client/scripts/main.gd` — minimal wiring only; stay within baseline
- [ ] `server/internal/game/sim.go` — only if combat/coop fix requires; touch-to-shrink
- [ ] `tools/bot/run.py` — grandfathered; no new `helpers=globals()` extractions

```bash
make maintainability
```

---

## Task 0 — Reproduce and log each failure

Capture `VERBOSE=1` assertion text for all 15 before editing (baseline for regressions).

```bash
make db-up
# Protocol batch
for s in full_equipment monster_rarity_loot_scaling inventory_capacity_and_paper_doll \
  coop_rewards_and_scaling gold_autopickup_shared_loot pack_aggro_and_dungeon_packs \
  offensive_unique_effects survival_reactive_unique_effects \
  resource_support_mobility_unique_effects dungeon_elite_side_objective \
  live_rare_combat_affixes mercenary_foundation; do
  echo "=== $s ===" | tee -a /tmp/v350-protocol-baseline.log
  VERBOSE=1 make bot scenario=$s 2>&1 | tee -a /tmp/v350-protocol-baseline.log || true
done

# Client trio
for s in character_stats_panel blacksmith_armor_recipe wall_floor_dungeon_rollout; do
  echo "=== $s ===" | tee -a /tmp/v350-client-baseline.log
  HEADLESS=1 VERBOSE=1 make bot-client SCENARIO=$s 2>&1 | tee -a /tmp/v350-client-baseline.log || true
done
```

- [x] Record per-scenario: failing step, error class, reproducible Y/N

---

## Task 1 — Protocol cluster A: loot `walk_toward` (4 scenarios)

**Scenarios:** `full_equipment`, `live_rare_combat_affixes`, `inventory_capacity_and_paper_doll` (+ any others failing `walk_toward exhausted 40 ticks`).

**Hypothesis:** Protocol bot greedy walk cannot reach loot in `equipment_lab` row transitions; may need `pathfind: true` on `pick_up_loot`, increased `max_ticks`, lab entity reposition, or server movement acceptance fix.

Files:
- Modify: `tools/bot/movement_runtime.py`, `tools/bot/run.py`
- Maybe: `shared/rules/worlds.v0.json` (`equipment_lab` loot positions)

- [x] Step 1.1: Reproduce `full_equipment` — confirm failing `pick_up_loot` step index and target coords
- [x] Step 1.2: Fix root cause (prefer server movement or lab layout over blind tick inflation)
- [x] Step 1.3: Verify all four scenarios

```bash
for s in full_equipment live_rare_combat_affixes inventory_capacity_and_paper_doll; do
  make bot scenario=$s VERBOSE=1 || exit 1
done
```

---

## Task 2 — Protocol cluster B: `resource_support_mobility_unique_effects`

**Symptom:** `walk_to_monster` with default pathfind runs ~12+ minutes in `unique_momentum_lab` (monster at x=26).

Files:
- Modify: `tools/bot/scenarios/56_resource_support_mobility_unique_effects.json`
- Modify: `tools/bot/movement_runtime.py` and/or `unique_momentum_lab` preset

- [x] Step 2.1: Set `pathfind: false` on `walk_to_monster` **or** compact lab / cap pathfind `max_ticks` with clear timeout
- [x] Step 2.2: Set explicit `max_elapsed_s` only if functional proof needs dungeon walk (prefer compact lab)
- [x] Step 2.3: Verify scenario completes in <30s

```bash
make bot scenario=resource_support_mobility_unique_effects VERBOSE=1
```

---

## Task 3 — Protocol cluster C: scenario budget drift (3 scenarios)

**Scenarios:** `monster_rarity_loot_scaling` (28.23s > 28s), `dungeon_elite_side_objective` (58s > 52s), possibly `pack_aggro_and_dungeon_packs` (32.32s > 32s).

- [x] Step 3.1: Confirm functional steps pass before bumping budget
- [x] Step 3.2: Shorten steps (`max_ticks` on stairs, remove idle waits) **or** bump `max_elapsed_s` by ≤15% with as-built note
- [x] Step 3.3: Re-run each 3× to check flake

```bash
for s in monster_rarity_loot_scaling dungeon_elite_side_objective pack_aggro_and_dungeon_packs; do
  make bot scenario=$s VERBOSE=1 || exit 1
done
```

---

## Task 4 — Protocol cluster D: unique combat deaths (2 scenarios)

**Scenarios:** `offensive_unique_effects`, `survival_reactive_unique_effects`

**Symptoms:** `player died` / `player_dead` during unique-effect combat proofs.

Files:
- Modify: `server/internal/game/` unique effect + combat handlers
- Maybe: `shared/rules/` unique definitions, scenario setup steps

- [x] Step 4.1: Reproduce with `VERBOSE=1`; identify whether death is from reflected damage, missing proc, or wrong lab monster
- [x] Step 4.2: Fix authoritative combat outcome (not bot attack spam)
- [x] Step 4.3: Verify both scenarios + replay phase

```bash
make bot scenario=offensive_unique_effects VERBOSE=1
make bot scenario=survival_reactive_unique_effects VERBOSE=1
```

---

## Task 5 — Protocol cluster E: coop loot (2 scenarios)

**Scenarios:** `coop_rewards_and_scaling`, `gold_autopickup_shared_loot`

**Symptoms:** Co-op kill credit / shared gold award timeouts.

Files:
- Modify: `server/internal/game/` coop reward + shared loot paths
- Maybe: `tools/bot/` coop peer helpers

- [x] Step 5.1: Reproduce two-peer flow; log combat events and gold ledger deltas
- [x] Step 5.2: Fix kill attribution or shared-gold removal/award
- [x] Step 5.3: Verify both scenarios

```bash
make bot scenario=coop_rewards_and_scaling VERBOSE=1
make bot scenario=gold_autopickup_shared_loot VERBOSE=1
```

---

## Task 6 — Protocol cluster F: mercenary foundation (1 scenario)

**Scenario:** `mercenary_foundation`

**Symptom:** No `monster_damaged` event with `source_entity_type=companion`, `source_monster_def_id=mercenary_guard`.

Files:
- Modify: `server/internal/game/` mercenary/companion combat event emission

- [x] Step 6.1: Reproduce; confirm companion engages `combat_lab_soft_target`
- [x] Step 6.2: Fix companion damage event emission (ADR-0010 foundation)
- [x] Step 6.3: Verify scenario + `mercenary_hiring_board` / `companion_stance_command` smoke

```bash
make bot scenario=mercenary_foundation VERBOSE=1
```

---

## Task 7 — Client cluster: `character_stats_panel`

**Symptom:** Step 4 `wait_event` `monster_killed` times out at **1.0s** after step 3 `click_entity_until_event` already targets same event.

Files:
- Modify: `tools/bot/scenarios/client/09_character_stats_panel.json` (remove redundant wait or raise timeout)
- Maybe: `client/scripts/main.gd` click-to-attack / event delivery

- [x] Step 7.1: Run `HEADLESS=1 VERBOSE=1 make bot-client SCENARIO=character_stats_panel`
- [x] Step 7.2: Fix scenario step ordering/timeouts first; client fix only if kill genuinely fails
- [x] Step 7.3: Verify scenario

```bash
HEADLESS=1 make bot-client SCENARIO=character_stats_panel VERBOSE=1
```

---

## Task 8 — Client cluster: `blacksmith_armor_recipe`

**Symptom:** Step 35 `wait_blacksmith_panel` — expected `cave_mail`, `upgrade_enabled=true`, `resource_wallet_count=1`.

Files:
- Modify: `client/scripts/` blacksmith panel + interactable routing
- Maybe: scenario step 34 interactable index

- [x] Step 8.1: Reproduce; log blacksmith debug state at failure
- [x] Step 8.2: Fix panel open binding or wallet sync after stash deposit
- [x] Step 8.3: Verify `blacksmith_second_recipe` / `blacksmith_upgrade_history` still pass

```bash
HEADLESS=1 make bot-client SCENARIO=blacksmith_armor_recipe VERBOSE=1
```

---

## Task 9 — Client cluster: `wall_floor_dungeon_rollout`

**Symptom:** After `stairs_down` click, `wait_wall_layout` sees level **0**, zero walls (expected level **-1**).

Related spec: v338 `wall-floor-dungeon-rollout`.

Files:
- Modify: `client/scripts/wall_renderer.gd`, `client/scripts/main.gd` (level transition)
- Maybe: `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` (wait for floor change step)

- [x] Step 9.1: Reproduce; inspect bot debug `wall_layout` state
- [x] Step 9.2: Ensure client applies dungeon descent + wall generation before assert
- [x] Step 9.3: Verify `test_factories.gd` + scenario

```bash
make client-unit
HEADLESS=1 make bot-client SCENARIO=wall_floor_dungeon_rollout VERBOSE=1
```

---

## Task 10 — Protocol matrix confirmation

Run all 12 protocol failures in one pass (still not full CI):

```bash
for s in full_equipment monster_rarity_loot_scaling inventory_capacity_and_paper_doll \
  coop_rewards_and_scaling gold_autopickup_shared_loot pack_aggro_and_dungeon_packs \
  offensive_unique_effects survival_reactive_unique_effects \
  resource_support_mobility_unique_effects dungeon_elite_side_objective \
  live_rare_combat_affixes mercenary_foundation; do
  make bot scenario=$s || exit 1
done
```

- [x] All 12 green

---

## Task 11 — Client matrix confirmation

```bash
for s in character_stats_panel blacksmith_armor_recipe wall_floor_dungeon_rollout; do
  HEADLESS=1 make bot-client SCENARIO=$s || exit 1
done
```

- [x] All 3 green

---

## Task 12 — Lifecycle docs

- [x] Update `docs/progress/slice-lifecycle.md` and `docs/progress/slice-codename-index.md`
- [x] Write `docs/as-built/v350_ci-full-green.md` (per-scenario fix summary)
- [x] Update `PROGRESS.md`: CI gate → `make ci-full` green; clear v349 failure inventory gap

---

## Final verification

```bash
make maintainability
make validate-shared    # if shared/ touched
make test-go            # if server/ touched
make test-py            # if tools/bot touched
make client-unit        # if client/ touched
make ci                 # merge pack still green
make ci-full            # authoritative gate (~37 min)
```

**Success criterion:** `make ci-full` exit 0; all 15 scenarios listed in spec pass.

## Deferred (explicit)

- v337 coordinator extractions (`main.gd` attack-move, `sim.go` phase paydown, `runner.go` delete)
- Registering 8 orphan client unit tests in `client_smoke.sh`
- Promoting any fixed extended scenario into merge pack
