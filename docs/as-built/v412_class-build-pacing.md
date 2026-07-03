# v412 As-Built — Class Build Pacing

Date: 2026-07-03
Spec: [`docs/specs/v412_spec-class-build-pacing.md`](../specs/v412_spec-class-build-pacing.md)
Plan: [`docs/plans/v412_2026-07-03-class-build-pacing.md`](../plans/v412_2026-07-03-class-build-pacing.md)

## What shipped

- Skill point cadence grants at level 1 and every third level (1, 3, 6, 9, …).
- Per-class automatic `level_stat_growth` on level-up, respec floor, and resume normalization (barbarian +STR, sorcerer +magic, paladin +VIT, ranger/rogue +DEX).
- Seven tier-3 active skills: Rain of Arrows, Explosive Shot, Fireball, Energy Ward, War Cry, Hammer of Light, Smoke Screen.
- `area_stat_percent_buff` extended for `evade_chance` (Smoke Screen).
- `class_progression.go` extracted for cadence and growth helpers.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestSkillPoint|TestClassLevelStatGrowth|TestClassBuildPacing|TestBishopRespec' -count=1
make bot scenario=class_progression_pacing
make bot scenario=class_build_pacing_skills
```

## Boundaries

- Rend, Retribution, and Predator's Mark deferred to v413 (`class-skill-engine-expansion`).
- Extended-only bot scenarios; not added to CI pack.
- Presentation uses existing skill icon shapes and placeholder VFX families.
