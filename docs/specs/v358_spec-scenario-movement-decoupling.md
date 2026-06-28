# v358 Spec: Scenario Movement Decoupling

Status: Draft  
Date: 2026-06-28  
Codename: `scenario-movement-decoupling`  
Baseline: v357 `town-night-perimeter`

Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`v32_spec-test-floor-and-resilient-scenarios.md`](v32_spec-test-floor-and-resilient-scenarios.md) — test locking / semantic assertion policy
- [`v120_spec-tuning-friendly-rule-tests.md`](v120_spec-tuning-friendly-rule-tests.md) — rule-derived expectations (narrower scope; movement is out of scope there)
- [`v356_spec-player-navigation-guardrails.md`](v356_spec-player-navigation-guardrails.md) — recent path-budget changes that exposed incidental-walk fragility
- [`v357_spec-town-night-perimeter.md`](v357_spec-town-night-perimeter.md) — post-v357 town layout/coordinates this slice must respect
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) — bot as integration proof layer
- [`../progress/scenario-catalog.md`](../progress/scenario-catalog.md) — scenario authoring guidance

## Purpose

Bot scenarios have accumulated **incidental navigation** — stair descents, cross-map walks,
`walk_to_loot` / `walk_to_monster` setup, client `click_floor` + `wait_player_near` hops — that are
not part of the behavior under test. When movement speed, path budgets, acceleration, obstacle
layout, or town geometry change (v356 player path horizon, v357 town hub), unrelated scenarios fail
or slow down even though their contracts (loot, UI, combat stats, mercenaries, persistence, etc.)
are unchanged.

This slice **audits every protocol and client bot scenario**, classifies movement steps, and
refactors or **deletes** incidental travel so each scenario only depends on navigation when
navigation *is* the feature.

Deliverables:

1. **Movement audit** — machine-readable inventory of all scenarios with per-step classification and
   recommended action.
2. **Remediation playbook** — documented preference order for eliminating setup travel; locked
   **movement-contract allowlist** for scenarios that must keep walking/pathing/level transitions.
3. **Scenario refactors** — apply playbook across protocol + client bot JSON; fix post-v357 town
   coordinate assumptions.
4. **Policy lock-in** — update authoring guidance so new scenarios default to zero incidental
   movement.

No new gameplay mechanics. This is test-hygiene and tooling only.

## Non-goals

- No gameplay tuning (movement speed, path horizon, aggro, damage, acceleration).
- No changes to scenarios on the **movement-contract allowlist** (see below) except post-v357
  coordinate fixes where layout changed but the movement contract is unchanged.
- No protocol schema bump unless planning finds ≥5 scenarios blocked without a minimal debug hook;
  prefer `worlds.start_level` lab worlds first.
- No replay determinism relaxation.
- No broad `tools/bot/run.py` mechanical split (v55 orchestrator freeze).
- Not a full v120 tuning-value audit across Go/GDScript unit tests — movement decoupling in bot
  scenarios and scenario-catalog policy only.
- No new merge-gate scenarios unless a deleted pack proof requires a 1:1 replacement (budget-neutral
  curation per CI pack policy).

## Movement-contract allowlist

These scenarios (by `id`) **must retain** movement, pathing, or level-transition steps as part of
their contract. Audit may still trim *redundant* steps inside them, but must not remove the core
proof:

| Area | Scenario IDs (representative; plan may extend with rationale) |
|------|----------------------------------------------------------------|
| Auto-path / click-to-move | `path_maze`, `click_to_move` (client), `town_floor_click_to_move` (client) |
| Monster chase / leash | `chase_lab`, `chase_maze`, `leash_lab` |
| Dungeon level transitions | `dungeon_levels`, `teleporter_lab` |
| Generated-route / obstacles | `reachable_dungeon_obstacles`, `collision_lab` |
| Player path budgets | `player_path_budget_lab` |
| Client movement presentation | `town_teleporter_auto_approach` (client), `attack_move_sticky_targeting` (client), `movement_visual_smoothing` (client), `entity_tick_smoothing` (client), `mobility_skill_smoothing` (client), `melee_lunge_micro_step` (client), `torch_walk_visual` (client) |
| Flying / special navigation | `flying_navigation_trait` |
| Boss floor gate (boss movement + stair descent as contract) | `boss_floor_gate` |
| Vertical slice core loop | `vertical_slice`, `gear_before_combat` (walk-to-loot is the v0/v7 loop contract) |

Any scenario **not** on this list that opens with `use_stair`, `walk_to_*`, `move_until_*`,
`teleport_to_level`, or client `wait_player_near` / `click_floor` setup is a **refactor or delete
candidate** unless the audit documents a feature-specific reason.

## Remediation playbook

Apply in this order when movement is **not** the contract:

1. **Relocate to an existing lab world** with entities at spawn (`equipment_lab`,
   `character_stats_lab`, `boss_floor_gate_lab`, `monster_visual_catalog_lab`, `vendor_lab`, etc.).
2. **Set `start_level`** on a dedicated lab world in `shared/rules/worlds.v0.json` when depth
   matters but stair traversal does not (pattern: `boss_floor_gate_lab` at `-5`).
3. **`teleport_to_level` bot step** for multi-floor state when fast-travel / teleporter discovery is
   not under test (allowed by product decision).
4. **`debug_progression`** (`level`, `deepest_dungeon_depth`, stats, gold) to skip grind walks and
   depth gates.
5. **Replace `walk_to_*`** with `action_until_event`, `kill_monsters`, `pick_up_loot`, or
   `click_entity` / `interactable_def_id` targeting when range is not asserted.
6. **Client bot:** remove `click_floor` + `wait_player_near` setup when not testing
   click-to-move; use spawn-adjacent labs or protocol-side setup where possible.
7. **Delete** the scenario file when it is strict duplicate coverage of a lab or lower-level test
   and does not prove a distinct feature (see deletion policy).
8. **Session start override** (debug-only `session_start_level` or equivalent) — add only if ≥5
   scenarios need ad-hoc depth without a new world entry after steps 1–2 are exhausted.

### Deletion policy

- **Delete when possible** if the scenario only exists to walk a long route to assert something
  already proven elsewhere with less movement.
- Every deletion records in the audit: `deleted` | `merged_into` | `reason`.
- Do not delete the last proof of a behavior; merge steps into the surviving scenario or add a
  focused lab first.
- Pack (`ci_pack.json`) scenarios must not be deleted without a surviving pack member that owns the
  same merge-blocking contract.

### Assertion style after refactor

- Prefer semantic selectors (`interactable_def_id`, `monster_def_id`, `event_type`, `current_level`,
  `at_least` entity counts) over exact coordinates.
- Exact positions remain only where movement/pathing is the contract or a named golden owns the
  fixture (v32 test-locking policy).

## Acceptance criteria

### Audit and policy

- [ ] `docs/progress/scenario-movement-audit.tsv` (or equivalent committed artifact) lists every
  `tools/bot/scenarios/**/*.json` file with columns at minimum:
  `scenario_id`, `runner`, `ci_tier`, `movement_class` (`contract` | `setup-eliminable` |
  `delete-candidate` | `deleted`), `movement_steps_before`, `movement_steps_after`, `action`,
  `merged_into_or_reason`.
- [ ] `docs/progress/scenario-catalog.md` gains a **Movement budget** section: playbook summary,
  allowlist reference, and rule that new scenarios must not add incidental travel.
- [ ] `CLAUDE.md` or `AGENTS.md` includes one durable paragraph extending v32 policy:
  incidental movement in bot scenarios is forbidden unless the scenario id is on the allowlist.

### Scenario refactors

- [ ] All `setup-eliminable` **extended** scenarios are refactored, deleted, or explicitly deferred
  in the audit with rationale (target: **≥80%** of eliminable extended scenarios addressed in this
  slice).
- [ ] All **pack** (`ci_pack.json`) scenarios: net count of `use_stair` / `walk_to_*` /
  `move_until_*` / incidental client `wait_player_near` steps is **unchanged or lower**; no pack
  scenario gains new incidental movement.
- [ ] Post-v357 town: scenarios touching level-0 use `interactable_def_id` targeting or updated
  coordinates; no failures from pre-v357 service positions or daylight-town assumptions.
- [ ] High-churn extended offenders from v349/v356 class are addressed (including but not limited
  to): `monster_rarity_loot_scaling`, unique-effects suite, `pack_aggro_and_dungeon_packs`,
  `dungeon_elite_side_objective`, mercenary/coop loot flows, and client scenarios that timed out on
  movement setup.

### Tooling (only if needed)

- [ ] If a repeated bot pattern emerges (e.g. `ensure_at_level`), add a focused helper in
  `tools/bot/` with a direct unit test import — no `helpers=globals()` extraction from `run.py`.
- [ ] If `session_start_level` (or equivalent) is added, it is debug/bot-only, deterministic, and
  covered by a focused test; document in audit why lab worlds were insufficient.

### Regression / CI

- [ ] `python3 -c "from tools.bot.ci_pack import validate_ci_pack; validate_ci_pack()"` green.
- [ ] `.venv/bin/pytest tools/test_ci_pack.py -q` green.
- [ ] `make ci` green.
- [ ] `make ci-full` green before `/finish`.

## Scope and likely files

| Area | Files |
|------|-------|
| Audit artifact | `docs/progress/scenario-movement-audit.tsv` (new) |
| Bot scenarios | `tools/bot/scenarios/*.json`, `tools/bot/scenarios/client/*.json` (bulk) |
| CI pack | `tools/bot/ci_pack.json` — only if a deleted scenario requires 1:1 replacement |
| Bot runtime | `tools/bot/run.py`, `tools/bot/movement_runtime.py`, optional new focused helper module |
| Debug setup | `tools/bot/debug_progression.py`; optional server debug hook only if ≥5 scenarios blocked |
| Lab worlds | `shared/rules/worlds.v0.json`, `worlds.v0.schema.json` — new/adjusted labs with `start_level` |
| Docs | `docs/progress/scenario-catalog.md`, `CLAUDE.md` or `AGENTS.md`, plan + as-built on `/finish` |
| Lifecycle | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` on `/finish` |

Unlikely to touch: `server/internal/game/sim.go` hot path, protocol schemas, client presentation,
golden combat formulas — unless a minimal debug spawn hook is justified.

## Client asset / plugin decision

Not applicable — no new client assets or plugins. Presentation-only client bot scenarios may be
shortened; movement-contract client scenarios stay as-is.

## Test and bot proof

```bash
# Audit validation (plan should add a small checker or pytest that audit TSV covers all scenario files)
python3 -c "from tools.bot.ci_pack import validate_ci_pack; validate_ci_pack()"
.venv/bin/pytest tools/test_ci_pack.py -q

# Spot-check refactored scenarios (examples; plan lists full matrix)
make bot scenario=monster_rarity_loot_scaling VERBOSE=1
make bot-client SCENARIO=character_stats_panel HEADLESS=1 VERBOSE=1

make maintainability
make ci
make ci-full
```

Manual sanity after town-touching refactors:

```bash
make play
# Open town services after v357 layout — no reliance on obsolete walk routes in unrelated proofs.
```

## ADR alignment

- **ADR-0001:** Strengthens the bot integration layer without weakening determinism; replay proofs on
  refactored scenarios must still pass where the scenario owns replay.
- **ADR-0008:** Preserves dedicated dungeon progression / teleporter contracts on the allowlist;
  decouples unrelated proofs from full tower walks.
- **v32 test-locking:** Extends behavior-proof guidance to **movement setup** — travel is not a
  contract unless the scenario name/allowlist says so.

## Resolved decisions (for planning)

| Item | Decision |
|------|----------|
| Sequencing | After v357 `town-night-perimeter` ships |
| `teleport_to_level` for setup | **Allowed** when teleporter discovery/fast-travel is not the test |
| Depth without new world | **Labs / `start_level` first**; session override only if ≥5 scenarios blocked |
| Redundant long-route scenarios | **Delete** when duplicate; document `merged_into` |
| Client bot setup trims | **In scope** — trim non-contract `wait_player_near` / `click_floor` |

## Open questions and risks

| Risk | Mitigation |
|------|------------|
| Deleting the only E2E proof of a feature | Audit requires `merged_into`; no deletion without surviving proof |
| Pack budget / coverage hole after delete | Replace 1:1 in `ci_pack.json` or demote consciously per pack policy |
| New lab worlds proliferate | Prefer extending existing labs; one world per distinct depth/setup pattern |
| `run.py` growth from helpers | New helpers in focused modules with direct tests; no globals() extraction |
| v357 coordinate churn | Dedicated town pass in implementation plan before `make ci-full` |

No blocking questions remain for `/plan` under the resolved decisions above.
