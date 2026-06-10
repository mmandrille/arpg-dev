# Spec: `monster-attack-cadence`

Status: Accepted
Date: 2026-06-10
Codename: `monster-attack-cadence`
Slice: v56 - monster attack cadence
Baseline: v55 `consolidation-and-quality-gates`

## Purpose

Regular dungeon monsters currently attack slowly enough that early generated floors feel too easy
once the player understands movement and basic combat. This slice slightly increases the attack
cadence for generated dungeon melee and ranged monsters while keeping damage, movement, boss timing,
and lab fixture monsters unchanged.

The goal is a conservative feel improvement: generated monsters should pressure the player sooner
and more often, but existing training and combat proof scenarios should not become brittle.

## Non-goals

- No boss-pattern timing, boss cooldown, boss damage, or boss UI changes.
- No monster damage, HP, hit chance, crit, armor, movement speed, aggro radius, or leash tuning.
- No changes to training dummies or combat-lab monsters.
- No adaptive difficulty, depth-specific cooldown scaling, or final balance pass.
- No protocol/schema change; cooldown values remain shared rules data.

## Acceptance criteria

1. `shared/rules/monsters.v0.json` reduces `dungeon_mob.attack_cooldown_ticks` from `40` to `32`.
2. `shared/rules/monsters.v0.json` reduces `dungeon_archer.attack_cooldown_ticks` from `90` to `75`.
3. Shared validation still accepts all monster definitions and rejects invalid cooldown wiring.
4. The dungeon monster attack golden fixture records the tuned melee cooldown and the existing
   proactive damage proof remains deterministic.
5. Go tests prove that `dungeon_mob` does not attack before the tuned cooldown and does attack when
   the tuned cooldown expires.
6. Go tests prove the ranged archer still fires and damages the player with the tuned cooldown.
7. Protocol bot coverage still proves archer-sourced player damage on a generated dungeon floor.
8. Existing boss-floor, skill, combat-stat, and vertical bot scenarios remain green.

## Scope and likely files

- `shared/rules/monsters.v0.json` - tune `dungeon_mob` and `dungeon_archer` cooldowns.
- `shared/golden/dungeon_monster_attack.json` - record tuned melee cooldown.
- `shared/golden/dungeon_monster_attack.v0.schema.json` - validate the recorded cooldown.
- `server/internal/game/game_test.go` - assert golden cooldown and existing cooldown behavior.
- `client/tests/test_golden.gd` - cross-check the golden cooldown against shared monster rules.
- `docs/specs/v56_spec-monster-attack-cadence.md` - this spec.
- `docs/plans/v56_2026-06-10-monster-attack-cadence.md` - implementation plan.
- `PROGRESS.md` and `docs/as-built/v56_monster-attack-cadence.md` - lifecycle close-out.

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestDungeonMonster(ProactiveAttackGolden|AttackCooldownAndDeterminism)|TestRangedMonsterProjectileDamagesPlayer' -count=1`
- `make bot scenario=38_ranged_monster_ai.json`
- `make bot scenario=24_boss_floor_gate.json`
- `make ci`

## Open questions and risks

- Q1: How much is "a little" faster?
  - Decision: use a conservative first-pass tuning: `dungeon_mob` `40 -> 32` ticks and
    `dungeon_archer` `90 -> 75` ticks at the current 15 Hz sim rate.
- Risk: Faster generated monster attacks can make long bot scenarios more fragile. Mitigation:
  keep damage unchanged and run focused generated-dungeon, boss-floor, and full CI bot gates.
