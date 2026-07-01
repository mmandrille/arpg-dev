# v393 As-built — Hero Skill Expansion

## What shipped

- Five tier-2 column-3 actives: `ground_slam`, `arcane_orb`, `radiant_bolt`, `fan_of_blades`, `snipe`.
- Skill presentations and en/es i18n keys; borrowed existing projectile/cone visuals.
- Go test `TestTier2Column3SkillExpansion` locks tree placement and class ownership.
- Extended bot `hero_skill_expansion` proves barbarian Ground Slam damages a lab target.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run TestTier2Column3SkillExpansion -count=1
make bot scenario=hero_skill_expansion
```

## Deferred

- Per-class bot casts for all five skills, production VFX, balance pass, tier-3 column gaps.
