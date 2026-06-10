# v56 As Built: Monster Attack Cadence

Date: 2026-06-10
Spec: [`docs/specs/v56_spec-monster-attack-cadence.md`](../specs/v56_spec-monster-attack-cadence.md)
Plan: [`docs/plans/v56_2026-06-10-monster-attack-cadence.md`](../plans/v56_2026-06-10-monster-attack-cadence.md)

## What shipped

- `dungeon_mob.attack_cooldown_ticks` changed from `40` to `32`.
- `dungeon_archer.attack_cooldown_ticks` changed from `90` to `75`.
- Training dummies, combat-lab monsters, boss pattern timing, movement speed, damage, HP, aggro,
  and leash tuning are unchanged.
- `shared/golden/dungeon_monster_attack.json` now records the tuned melee cooldown, and both Go and
  GDScript golden checks verify it against `shared/rules/monsters.v0.json`.
- `shared/golden/item_rolls.json` regained the required `description` metadata discovered missing
  by `make validate-shared`.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestDungeonMonster(ProactiveAttackGolden|AttackCooldownAndDeterminism)|TestRangedMonsterProjectileDamagesPlayer' -count=1`
- `make client-unit`
- `make bot scenario=38_ranged_monster_ai.json`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot scenario=32_skill_points_and_magic_bolt.json`
- `make ci`

## Deferred

- Final combat balance across monster damage, HP, movement, rarity, and depth scaling.
- Boss phase readability and boss pattern variety; these remain queued as separate selected
  autoloop slices.
