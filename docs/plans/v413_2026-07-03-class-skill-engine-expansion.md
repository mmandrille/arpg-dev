# v413 Plan — Class skill engine expansion

Status: Complete  
Goal: Ship Rend bleed cone, Retribution reflect-on-block, and Predator's Mark.

## Task 1 — Schema + rules types

- [x] `bleed`, `mark`, `reflect_on_block_buff` in skills schema
- [x] Go `SkillBleedDef`, `SkillMarkDef`, validation

## Task 2 — Sim engine

- [x] Cone bleed in `applyConeSkill`
- [x] Projectile mark in `resolveSkillProjectileMonsterHit`
- [x] Reflect on block in `damagePlayerByMonsterWithSource` (+ retaliate block path)
- [x] `applySkillBuff` handles `reflect_on_block_buff`

## Task 3 — Skills data + copy

- [x] `skills.v0.json`, i18n, presentations
- [x] Update `validate_skills.py` active counts (+1 per class)

## Task 4 — Tests + bot

- [x] `class_skill_engine_test.go`
- [x] `tools/bot/scenarios/class_skill_engine_expansion.json` (`ci_tier: extended`)

## Task 5 — Docs

- [x] `docs/as-built/v413_class-skill-engine-expansion.md`
