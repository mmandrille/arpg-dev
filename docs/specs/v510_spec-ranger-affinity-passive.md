# v510 Spec — ranger-affinity-passive

- **Status:** Complete; focused implementation and combined batch `make ci` passed.
- **Date:** 2026-10-01
- **Batch base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Dependency:** v509 transfer and conflict check passed on 2026-10-01

## Purpose

Continue v451's equipped class-affinity passive pilot for Ranger. Allocating **Deadeye** gives a bounded, data-driven ranged-damage bonus based on beneficial Ranger affinity rolls on equipped gear and the Ranger's effective **Dexterity**. The owner confirmed Dexterity on 2026-10-01. The server owns the count, formula inputs, damage, and replay outcome.

## Scope and exact rule

The existing `deadeye` rank-1 passive retains its crit-chance bonus. Its new optional shared-rules payload defines `affinity_stat: "dex"`, `base_percent_per_affinity: 2`, `stat_points_per_extra_percent: 10`, `max_active_affinities: 2`, and `max_bonus_percent: 12`. The owner confirmed the stat choice, and the coordinator approved these coefficients for this slice on 2026-10-01. All remain editable shared-rule values, validated and owned in `shared/rules/skills.v0.json` rather than literals in Go or GDScript. The schema and runtime validator bound the base at 0–10, the denominator at 1–100, affinity count at 1–3, and bonus cap at 1–25%; only the four character base-stat keys are allowed.

Let `C` be the number of **beneficial, class-matched Ranger** affinity rolls on equipped items. Each roll counts once, regardless of its rolled numeric value. Inactive off-class rolls, penalties active because the wearer is the wrong class, and unequipped inventory do not count. Let `D` be the server's current effective Dexterity (base, level growth, and legitimate equipment/passive bonuses), floored at zero. With Deadeye allocated and a Ranger-owned eligible hit:

`bonus_percent = min(12, min(C, 2) × (2 + floor(D / 10)))`

Otherwise the bonus is zero. At `D=17`, `C=0/1/2/3` gives `0/3/6/6%`; at `D=40`, `C=2` gives the capped `12%`. A selected attack's minimum and maximum raw damage are each transformed once as `floor(raw_bound × (100 + bonus_percent) / 100)` before the existing hit/armor/resistance/crit pipeline. The result is snapshotted when a persistent projectile launches, or once when a cast-tick Ranger ray/volley resolves. Later gear or stat changes do not alter an in-flight hit. The rule consumes no RNG and must not change roll order.

Eligible damage is the Ranger's bow basic projectile and direct-hit Ranger attack projectiles explicitly tagged as eligible in shared skill rules. At the batch base, the planned skill IDs are `piercing_shot`, `pinning_shot`, `volley`, `snipe`, `rain_of_arrows`, `explosive_shot`, `pinning_volley`, `hunters_volley`, `meteor_shot`, and `arrow_storm`. The first, second, and volley-family shots resolve direct hits on the cast tick; `snipe`, `explosive_shot`, and `meteor_shot` create persistent server projectiles. Recheck these paths after v509 is integrated. Ineligible damage includes melee attacks, Sorcerer/staff projectiles, companion attacks, damage-over-time ticks, follow-on elemental procs, and any skill without that tag. Apply the multiplier once to the attack's source range, even when a piercing shot or volley hits multiple targets. Preserve existing skill-specific multipliers and their order in a documented server helper. Place the existing Ranger affinity bow plus a new `affinity_ranger_quiver` in the existing skill progression lab so two class-matched rolls are available. No new loot source outside the lab is required in this slice.

The new quiver uses the belt slot because the affinity bow already occupies both hands.

Expose the server-calculated `ranged_damage_bonus_percent` as an additive field in character progression derived stats and show Deadeye's scaling rule/current bonus in the existing skills/stats UI. Generic `damage_min`/`damage_max` remain generic values; do not disguise the conditional ranged bonus as universal damage. Use the existing protocol version and update both v8 schemas/examples/tests for the additive field.

## Acceptance

1. Shared skill, item, world, and protocol data validate; invalid stat key, negative or zero denominator, unsafe cap, or unsupported skill eligibility fails validation. Rule values are editable without changing evaluator code.
2. Deadeye rank zero grants no bonus. At rank one, one and two beneficial Ranger affinities increase eligible ranged raw damage according to the formula; a third is capped. Raising effective Dexterity changes the bonus at a threshold. Wrong-class/penalty/unequipped rolls do not contribute.
3. The same configured bonus affects eligible Ranger basic and tagged direct skill shots once, and never affects melee, staff, pets, DoT, or elemental follow-ons. Damage is frozen at persistent projectile launch or cast-tick ray/volley resolution and remains server-authored through hit resolution.
4. A shared golden covers `D=17` with `C=0..3`, `D=40/C=2`, below-threshold/threshold Dexterity, and changed rule values. Go and GDScript formula tests consume the same fixture; server integration tests assert actual projectile outcomes and replay equality with the same seed/inputs.
5. `make bot scenario=affinity_passive_ranger` allocates Deadeye, equips one then two affinities, observes authoritative bonus and a ranged damage event, and proves the bonus disappears on unequip. The scenario uses a deterministic lab setup and stays in the extended tier unless CI-pack curation justifies promotion.
6. Existing Ranger combat and the v451 Barbarian/Rogue affinity passives remain valid. No new client assets or external dependencies are needed.

## Non-goals

- Wider class rollout, full Ranger passive-tree rebalance, general ranged damage stat for other classes, PvP tuning, new VFX, or production loot-table changes.
- Client-authoritative damage or a new network protocol version.

## Likely surfaces and ownership

- **Shared:** `skills.v0.json` and schema own the passive payload and eligible skill tags; `item_templates.v0.json` and `worlds.v0.json` own lab equipment; a dedicated golden/schema own formula examples; v8 protocol schemas own the additive display field.
- **Server:** parse/validate the new rules, count only beneficial Ranger affinities, evaluate the bounded formula, snapshot it at persistent projectile creation or cast-tick shot resolution, apply once to eligible damage, and expose the authoritative display value. Prefer a focused helper outside `sim.go`.
- **Client:** reuse the current skill tooltip/stats panel and shared-rule loader to show the formula and server value; a small pure evaluator consumes the golden for parity.
- **Bot/tests:** extend focused Go tests and add one protocol scenario based on the v451 affinity-passive labs. Verify replay without new inputs or RNG.
- **Docs:** plan and as-built evidence in the slice; coordinator owns lifecycle and progress closeout.

## Dependencies and integration risk

- v451 provides the affinity item model and passive counting, but its general `activeClassAffinityCount()` also counts active wrong-class penalties. v510 needs a beneficial class-matched count for this bonus without silently changing v451 behavior.
- v509 may touch `rules.go` and both v8 protocol schemas. Recompare its integrated paths and behavior before implementation; resolve these shared-file edits and any changed gameplay resolver paths with the coordinator.
- Any new `derived_stats` field requires both v8 protocol schemas and example/contract coverage. Both client and server must read the same golden formula.
- ADR-0001 requires server authority, shared data, and deterministic replay. ADR-0014 requires meaningful stat/passive/gear investment; the cap limits multiplicative growth.

## Adopt / borrow / reject

- **Adopt:** v451 affinity-count pilot, current Ranger Deadeye node, existing in-repo skills UI, training lab, and bot framework.
- **Borrow:** v451 affinity gear/scenario shapes and existing shared golden parity pattern.
- **Reject:** external assets/plugins and a new asset pipeline; this slice changes combat math and existing UI text only.

## Review decision and prerequisite resolution

The owner selected Dexterity as Ranger's affinity stat on 2026-10-01. The coordinator reviewed and approved this spec and its plan, including the data-owned coefficients, on the same date. The coordinator transferred v509 into this detached worktree without losing these documents; the overlap review found no Ranger rule or resolver conflict and preserved the v509 boss-lane schema additions. Revisit the formula examples and golden if the owner later changes the stat or coefficients.
