# v386 As-Built — skill weapon damage scaling

## What shipped

- Added `weapon_multiplier_range` skill damage type in `skills.v0.json` / schema. Percent fields reuse
  `min_base`, `max_base`, `min_per_rank`, `max_per_rank`.
- Migrated all eight projectile-attack skills from absolute `rank_linear_range` damage to weapon
  multipliers.
- Go sim resolves projectile skill damage via `skill_weapon_damage.go` →
  `resolvePlayerAttackDamage()` before magic scaling and `skill_damage_percent`.
- Golden `skill_points_and_magic_bolt.json` now owns multiplier percents, not absolute damage.
- Skills panel tooltip shows `Weapon damage: N%-M% -> …` for multiplier skills.
- Extended bot `108_skill_weapon_damage_scaling` proves Magic Bolt hits harder after equipping staff.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run TestSkillProjectileDamageScalesWithWeapon`
- `make client-unit`
- `make bot scenario=108_skill_weapon_damage_scaling`

## Deferred

- Production art for affinity weapon families.
