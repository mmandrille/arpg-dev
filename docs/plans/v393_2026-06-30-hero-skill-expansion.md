# v393 Plan — Hero Skill Expansion

Date: 2026-06-30
Spec: `docs/specs/v393_spec-hero-skill-expansion.md`

## Tasks

- [x] Shared: five tier-2 column-3 actives in `skills.v0.json` with class prerequisites
- [x] Shared: presentations + en/es i18n keys; borrow existing projectile/cone visuals
- [x] Server: `TestTier2Column3SkillExpansion` locks tree placement and class ownership
- [x] Bot: extended `hero_skill_expansion` proves barbarian Ground Slam damages a lab target

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run TestTier2Column3SkillExpansion -count=1
make bot scenario=hero_skill_expansion
```

## Deferred

- Per-class bot casts for all five skills, production VFX, balance pass, tier-3 column gaps
