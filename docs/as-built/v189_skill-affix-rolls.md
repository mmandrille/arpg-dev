# v189 As-built: Skill Affix Rolls

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added shared support for `skill_cooldown_reduction_percent` and `skill_mana_cost_reduction`
  across item-template schemas, golden schemas, shop pricing, and validation bounds.
- Added rare-gated skill affix roll candidates to the starter Sorcerer staff, cave ring, and cave
  amulet so rare, unique, and set rarity pools can inherit them.
- Applied equipped flat skill mana-cost reduction before authoritative cast spend and emitted
  `skill_cast.mana`, clamped at zero.
- Applied equipped percent skill cooldown reduction to authoritative cooldown starts, clamped to a
  minimum of one tick and capped at 75 percent reduction.
- Added client stat labels/order handling for inventory, shop, stash, and market panels.
- Tightened the protocol bot rolled-item assertion helper so it can select the matching rolled item
  by template, rarity, and required stat keys when a class starter item has the same item id.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'SkillAffix|SkillDamage|AllSkills|MagicBolt'`
- `make bot scenario=80_skill_affix_rolls.json`
- `make ci`

## Deferred

- Per-skill affix targeting.
- Procedural affix names and prefix/suffix item labels.
- Crafting or blacksmith routes that add skill affix rolls.
