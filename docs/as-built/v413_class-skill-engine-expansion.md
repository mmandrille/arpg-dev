# v413 As-Built — Class Skill Engine Expansion

Date: 2026-07-03
Spec: [`docs/specs/v413_spec-class-skill-engine-expansion.md`](../specs/v413_spec-class-skill-engine-expansion.md)
Plan: [`docs/plans/v413_2026-07-03-class-skill-engine-expansion.md`](../plans/v413_2026-07-03-class-skill-engine-expansion.md)

## What shipped

- Skill schema/engine support for `bleed`, `mark`, and `reflect_on_block_buff`.
- **Rend** — cone bleed after `ground_slam`.
- **Retribution** — paladin self-buff reflecting damage on block.
- **Predator's Mark** — rogue projectile applies `rogue_mark` damage amp.
- Extracted `class_skill_engine.go` for bleed/mark/reflect helpers.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestClassSkillEngine|TestRend|TestRetribution|TestPredatorsMark' -count=1
make bot scenario=class_skill_engine_expansion
```

## Boundaries

- Completes the ten-skill class build variety batch started in v412.
- Bot scenario proves Rend bleed; Retribution and Predator's Mark rely on focused Go tests.
- Extended-only scenarios; not added to CI pack.
