# v218 Spec - Rogue Duelist Mark

Codename: `rogue-duelist-mark`
Status: Draft
Baseline: v217 `paladin-charge-channeling-protocol`

## Purpose

Improve Rogue's 1v1 identity without making it the best tank or pack clearer. `Poison Stab`
should create a duel target that takes increased player damage while marked, `Dash` should become
a stronger opener by stunning crossed monsters, Rogue gains a learnable execute passive, and DEX
should improve critical-damage scaling for every class through the shared stat formula.

This advances ADR-0014 D1 because stats, active skills, and passive skill investment all contribute
to the build. It avoids ADR-0014 D7 unfair spike risk by keeping execute gated behind low target HP,
a rank-scaled chance roll, and existing deterministic combat events.

## Non-goals

- No full passive tree framework beyond one learnable passive skill kind.
- No protocol schema version bump; existing skill progression and combat events carry the proof.
- No new external client visual assets or production asset pipeline; the mark cue is procedural.
- No broad class DPS rebalance beyond the DEX-derived critical-damage formula change.
- No PvP-specific execute rules.

## Adopt / Borrow / Reject

- **Adopt:** Existing shared skill catalog/schema and server-authoritative skill ranks.
- **Borrow:** Existing `skill_effect_started` / `skill_effect_ended`, stun effect, and combat event
  patterns for mark, dash stun, and execute proof.
- **Reject:** External assets/plugins and a separate passive-tree content pipeline for this slice.

## Data Shapes

- `poison_stab.poison` gains mark tuning:
  - `mark_damage_bonus_percent`
  - `mark_duration_ticks`
  - `mark_effect_id`
- `dash.dash` gains stun tuning:
  - `stun_effect_id`
  - `stun_duration_ticks`
- A new Rogue skill `executioner` uses a passive execute payload:
  - `kind: "passive_execute"`
  - `execute.threshold_percent_base: 10`
  - `execute.threshold_percent_per_rank: 5`
  - `execute.chance_percent: 35`
- `character_progression.derived_stats.crit_damage` gains standard DEX scaling.

## Architecture

`Poison Stab` starts the existing poison DOT and a server-owned mark for the target. All player
damage paths consult the mark before resistance/armor resolution so basic attacks, off-hand hits,
skills, and poison DOT ticks all benefit from the mark. Mark state is stored in the player sim state
alongside poison DOTs so it survives player-state swaps and remains replay-stable.

`Dash` reuses existing crossed-target resolution, then applies the configured stun effect to each
damaged target.

The client reads authoritative `effect_ids` from monster entity updates. When `rogue_mark` is
present on a living monster, it shows a red skull marker above the monster; removing the effect id
removes the marker.

The `executioner` passive is not castable. Once learned, each successful player damage event checks
the target's current HP ratio after normal damage. If the target is alive and below the rank-scaled
threshold, the deterministic session RNG rolls the passive chance; success kills the target and
emits a `monster_damaged` event with `skill_id: "executioner"`.

## Acceptance Criteria

1. `Poison Stab` starts a mark that increases all player damage against the marked target until the
   mark expires or the target dies.
2. Poison DOT ticks also benefit from the mark.
3. `Dash` applies a stun effect to crossed monsters.
4. A learned rank-1 `executioner` can execute an enemy at or below 10% HP; higher ranks increase the
   threshold by 5 percentage points per rank.
5. `executioner` cannot be cast as an active skill.
6. DEX contributes to standard `crit_damage` derived stats for all classes.
7. A marked living enemy shows a red skull over its head, and the skull disappears when `rogue_mark`
   leaves the enemy's `effect_ids`.

## Testing Plan

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRogue|TestLoadRules|TestCritDamageUsesDexterityAsStandardDerivedStat|TestDerivedStats|TestEffectiveAttackSpeedUsesWeaponAndItemPercent' -count=1`
- `.venv/bin/pytest tools/bot/test_skill_demo.py tools/bot/test_protocol.py::test_load_scenarios_discovers_rogue_class_foundation tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=rogue_class_foundation`
- `make client-unit`

Manual visual check, if desired:

```bash
make bot-visual scenario=rogue_class_foundation
```
