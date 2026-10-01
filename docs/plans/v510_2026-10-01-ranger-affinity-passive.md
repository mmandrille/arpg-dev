# v510 Plan — ranger-affinity-passive

- **Status:** Complete; integrated and combined batch `make ci` passed in 11m41s.
- **Date:** 2026-10-01
- **Spec:** [`v510_spec-ranger-affinity-passive.md`](../specs/v510_spec-ranger-affinity-passive.md)
- **Recorded batch base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Assigned checkout:** `/Users/mmandrille/git/arpg-dev-batch/v510-ranger-affinity-passive` (detached)

## Review gate and sequencing

The spec is internally consistent with the owner's 2026-10-01 Dexterity choice and preserves v451 affinity semantics, ADR-0001 server authority/shared golden/replay, and ADR-0014's stat/passive/gear investments. On 2026-10-01 the coordinator passed the spec/plan review and approved the bounded, data-owned coefficients for this slice. The v509 prerequisite and overlap review passed. If the owner changes the stat or mechanic, revise the spec, formula examples, golden expectations, and this plan before further implementation.

v509 was a prerequisite for implementation because the coordinator reserved that sequence. Its integrated working state was transferred into this assigned detached worktree and checked against the shared file map. No branch was created.

## File map and ownership

| Owner | Expected paths | Purpose / conflict note |
|---|---|
| Shared skill rules | `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json` | Put Deadeye's bounded affinity/Dexterity coefficients and explicit Ranger projectile eligibility in data; validate after the v509 prerequisite transfer. |
| Shared gear/world | `shared/rules/item_templates.v0.json`, `shared/rules/worlds.v0.json` | Add lab-only Ranger quiver and place it beside the existing Ranger affinity bow in `skill_progression_lab`; preserve existing loot. |
| Shared golden/protocol | `shared/golden/ranger_affinity_damage.json`, matching golden schema, `shared/protocol/session_snapshot.v8.schema.json`, `shared/protocol/state_delta.v8.schema.json`, examples if required | Formula parity and additive `ranged_damage_bonus_percent` field. Keep v8; reconcile both schema edits with v509's lane descriptor. |
| Server rules/formula | `server/internal/game/rules.go`, `server/internal/game/passive_skill_rules.go`, `server/internal/game/skill_rules_validation.go`, new focused `ranger_affinity_damage.go` and tests | Parse/validate bounded payload, beneficial Ranger affinity count, pure formula, and one-time raw damage-range scaling. |
| Server attack/replay | `server/internal/game/sim.go`, `server/internal/game/ranger_skills.go`, `server/internal/game/derived_stats.go`, `server/internal/game/class_item_affinities.go` plus focused tests | Snapshot bonus at persistent basic/skill projectile launch or cast-tick Ranger shot resolution, keep generic stats unchanged, expose authoritative conditional bonus. Avoid broad edits to the already-large `sim.go`; use small call sites and helper file. Check replay tests/fixtures after protocol update. |
| Client | `client/scripts/skill_passive_tooltip.gd`, `client/scripts/character_stats_panel.gd`, a small shared-rule evaluator if needed, `client/tests/test_golden.gd`, focused panel/tooltip tests | Show configured rule and server-computed current percent. Reuse current widgets; no new plugin, asset, or scene. |
| Bot/validation | `tools/validate_shared.py`, `tools/validate_skills.py`, `tools/bot/scenarios/<next-id>_affinity_passive_ranger.json` and focused tests | Register golden, reject invalid rules, and prove allocate/equip/unequip plus actual authoritative hit. Keep scenario `ci_tier: extended`. |
| Documentation | `docs/as-built/v510_ranger-affinity-passive.md` (after implementation) | Record exact files, tests, replay proof, and limitations for coordinator handoff. Coordinator owns `PROGRESS.md` and lifecycle closeout. |

At the recorded base, the direct-hit Ranger `projectile_attack` IDs are `piercing_shot`, `pinning_shot`, `volley`, `snipe`, `rain_of_arrows`, `explosive_shot`, `pinning_volley`, `hunters_volley`, `meteor_shot`, and `arrow_storm`. The `ranger_skills.go` ray/volley path resolves the first two and volley-family shots on the cast tick; `snipe`, `explosive_shot`, and `meteor_shot` use persistent `spawnSkillProjectile` entities. A skill with a separate AoE, DoT, or companion damage path is excluded unless its direct projectile damage can be isolated and tested. Recheck this list and its resolver paths after v509 integration; never infer eligibility from visuals or the player's currently equipped weapon at impact.

## Ordered implementation tasks

- [x] **0. Dependency check.** Dexterity and the spec/plan, including coefficients, were approved on 2026-10-01. The coordinator transferred integrated v509 into this worktree byte-for-byte. Its `rules.go` additions are boss-only, the v8 schema additions are boss-lane fields/definitions, and its `sim.go` edit adds boss lane state; the Ranger rule and damage paths are unchanged. The v510 documents survived unchanged. Coordinator checks on this prerequisite passed: `make validate-shared` (2,213), focused boss Go tests, and `git diff --check`. No implementation conflict remains at this gate.
- [x] **1. Shared contract.** Added bounded Deadeye payload, eligible-skill tags, Ranger quiver/lab loot, golden/schema, and additive v8 derived-stat field. Invalid rule bounds and unsupported tags are rejected. Shared validation passed with 2,231 checks.
- [x] **2. Server rule evaluator.** Added the pure formula, stable equipment traversal, no RNG, golden table tests, and mutated-rule rejection tests. Focused Go tests and determinism lint passed.
- [x] **3. Authoritative damage.** Basic bow and tagged persistent skill projectiles snapshot the scaled range at launch; tagged Ranger ray/volley resolvers compute it once on the cast tick. Server tests cover rank zero, zero/one/two/three affinities, inactive and wrong-class rolls, unequipped gear, Dexterity threshold/cap, launch snapshot, tagged and untagged skill paths, bow gate, piercing multi-hit, and same-input replay. The full game package passed after the new quiver changed the class-specialist catalog count. Final protocol bot proof is in Task 5.
- [x] **4. Display/parity.** The server field, both v8 schemas/examples, client stats/Deadeye tooltip, and GDScript golden test are implemented. Shared validation and a clean full `make client-unit` rerun passed. An earlier run was intentionally interrupted for v508's renewed runner reservation; it was not a test failure.
- [x] **5. Protocol bot and replay.** The extended `affinity_passive_ranger` scenario passed at level 10 with effective Dexterity 17, one/two equipped affinities, a server-observed Piercing Shot hit, unequip reversal, fixed seed, reconnect, and the bot's replay check. A separate retained session `sess_01M3W6WVSZZQKSWF9YZ455KNVP` passed `make replay SESSION_ID=<id> DATABASE_URL="$(make -s test-db-url)"`: 11 inputs and 9/9 matching authoritative events.
- [x] **6. Slice evidence/handoff.** The full game package, v451 Barbarian/Rogue protocol bots, determinism lint, maintainability, shared validation, client unit, scenario audit, and diff check pass. Exact changed/untracked/ignored evidence is in the as-built. The coordinator integrates the batch, runs combined `make ci`, performs `/finish`, and handles commits/review/refactor.

## Acceptance mapping

| Spec acceptance | Proof |
|---|---|
| Shared schema/data and bounds | Task 1 validation plus invalid-rule tests in Task 2 |
| Allocation, affinity count, Dexterity threshold/cap | Shared golden, Task 2 table test, Task 3 server integration tests |
| Eligible/ineligible damage, one-time launch snapshot | Task 3 projectile tests and Task 5 bot event |
| Go/GDScript parity and replay | Tasks 2, 4, 5 |
| Player-visible bonus and v451 regression | Tasks 4, 6 |

## Integration and verification limits

This is a gameplay math slice. No new VFX, asset, camera, or smoothness claim is made, so renderer captures and frame-time comparisons are outside the acceptance gate. The bot checks authoritative behavior; it cannot establish balance across all builds. Retune only via shared coefficients after examining the bounded cases. Do not add the bot scenario to the fast CI pack without a curation tradeoff.

The assigned worktree contains the transferred v501–v509 batch state plus this slice's changes. Before handoff, list every changed and untracked path and any ignored evidence explicitly. Do not run `make ci` or `make ci-full`, commit, push, alter coordinator `main`, or clean up the worktree from this slice session.
