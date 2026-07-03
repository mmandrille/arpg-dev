# v412 Plan — Class build pacing and seven skills

Status: Complete  
Goal: Ship progression pacing plus seven data-driven skills with evade area buff support.  
Architecture: Rules-first cadence and class growth in `character_progression.v0.json`; sim applies growth on level-up and respec; seven skills reuse existing kinds; Smoke Screen extends `area_stat_percent_buff` for `evade_chance`.  
Tech stack: shared JSON, Go sim, Python bot, Godot client (presentations/i18n only).

## Baseline and shortcut decision

Reuses v411 skill-tree layout (data-driven). **Borrow** existing skill VFX/audio families.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/character_progression.v0.json` | Cadence + `level_stat_growth` |
| Modify | `shared/rules/character_progression.v0.schema.json` | Schema for growth |
| Modify | `shared/rules/skills.v0.json` | Seven skills |
| Modify | `shared/rules/skills.v0.schema.json` | `evade_chance` on area buff |
| Modify | `server/internal/game/sim.go` | Cadence, growth, evade derived |
| Modify | `server/internal/game/rules.go` | Load growth |
| Modify | `server/internal/game/rules_area_buffs.go` | Validate evade |
| Modify | `shared/i18n/en.json`, `skill_presentations.v0.json` | Copy |
| Create | `tools/bot/scenarios/class_progression_pacing.json` | Pacing proof |
| Create | `tools/bot/scenarios/class_build_pacing_skills.json` | Multi-skill proof |
| Create | `server/internal/game/class_build_pacing_test.go` | Unit tests |

## Maintenance ratchet

Hotspot: `sim.go` (grandfathered) — touch-to-shrink; no new domains in coordinator.

## Task 1 — Progression rules

- [x] Update skill cadence semantics in JSON (document in comment if needed)
- [x] Add `level_stat_growth` per class
```bash
make validate-shared
```

## Task 2 — Go sim pacing

- [x] `skillPointGrantLevel` / `totalEarnedSkillPoints` for 1,3,6,9…
- [x] Apply growth on level-up; respec + resume floor
- [x] Tests `TestSkillPointCadenceAndSpend` (level 6 unspent skills), new growth test
```bash
cd server && go test ./internal/game -run 'TestSkillPoint|TestClassLevelStatGrowth' -count=1
```

## Task 3 — Evade area buff engine

- [x] Schema + validator + derived stats for `evade_chance` skill effects
```bash
cd server && go test ./internal/game -run TestSmokeScreen -count=1
```

## Task 4 — Seven skills + copy

- [x] `skills.v0.json`, i18n, presentations
- [x] `TestClassBuildPacingSkills` tree/prereq smoke
```bash
make validate-shared
```

## Task 5 — Bot scenarios

- [x] `class_progression_pacing.json`, `class_build_pacing_skills.json` (`ci_tier: extended`)
```bash
make bot scenario=class_progression_pacing
make bot scenario=class_build_pacing_skills
```

## Task 6 — Docs

- [x] `docs/as-built/v412_class-build-pacing.md`
- [x] Lifecycle row on finish

## Final verification

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run 'TestSkillPoint|TestClassLevelStatGrowth|TestClassBuildPacing' -count=1
make bot scenario=class_progression_pacing
make bot scenario=class_build_pacing_skills
```
