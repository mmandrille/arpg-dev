# v386 Spec — Skill Weapon Damage Scaling

Status: Ready for implementation
Date: 2026-06-29
Codename: skill-weapon-damage-scaling

## Purpose

Make every offensive projectile skill derive its damage range from the character's authoritative
basic-attack `damage_min` / `damage_max` via data-driven weapon multipliers (e.g. 300% min / 225%
max at rank 1 for Magic Bolt). Equipping a stronger weapon increases skill damage the same way it
increases basic attacks. Existing magic scaling and `skill_damage_percent` modifiers apply after the
weapon-based base range is computed.

## Non-goals

- Class weapon affinities (v387).
- Changing cone/mobility skills that already use weapon damage.
- Heals, buffs, passives, or non-damage skills.
- Balance retune beyond preserving approximate rank-1 output at the reference unarmed 2–4 band.

## Acceptance criteria

- `skills.v0.schema.json` supports `damage.type: weapon_multiplier_range` (percent fields reuse
  `min_base`, `max_base`, `min_per_rank`, `max_per_rank`).
- All eight `projectile_attack` skills with absolute damage migrate to `weapon_multiplier_range`.
- Go sim `skillDamageRange` uses `resolvePlayerAttackDamage()` for weapon-multiplier skills.
- Focused Go test proves Magic Bolt damage increases after equipping a higher-damage weapon.
- Golden `skill_points_and_magic_bolt.json` stores multiplier expectations, not absolute damage.
- GDScript golden test validates multipliers against shared rules.
- Extended bot scenario proves skill hit damage rises with a stronger equipped weapon.
- `make validate-shared`, focused Go tests, `make client-unit`, and the new bot scenario pass.

## Scope and files

- `shared/rules/skills.v0.json`, `skills.v0.schema.json`
- `shared/golden/skill_points_and_magic_bolt.json` + schema
- `server/internal/game/skill_weapon_damage.go`, `sim.go`, `ranger_skills.go`, `rules.go`, tests
- `client/tests/test_golden.gd`
- `tools/bot/scenarios/108_skill_weapon_damage_scaling.json`

## Bot proof

Extended scenario `108_skill_weapon_damage_scaling.json` in `skill_progression_lab`: sorcerer learns
Magic Bolt, damages a soft target unarmed, equips a stronger staff, casts again, asserts higher
`monster_damaged` damage.

## Open questions

None — prerequisite for v387 class affinities per session decisions.
