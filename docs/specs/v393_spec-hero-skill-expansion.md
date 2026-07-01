# v393 Spec — Hero Skill Expansion

Status: Draft
Date: 2026-06-30
Codename: hero-skill-expansion

## Purpose

Add one new **tier-2 column-3** active skill per hero class, filling an empty skill-tree column
without new skill capability types.

| Class | Skill | Kind |
|-------|-------|------|
| Barbarian | Ground Slam | cone_attack |
| Sorcerer | Arcane Orb | projectile_attack |
| Paladin | Radiant Bolt | projectile_attack |
| Rogue | Fan of Blades | cone_attack (poison) |
| Ranger | Snipe | projectile_attack |

## Non-goals

- No new skill kinds, protocol changes, or balance pass across all depths.
- No production VFX/audio; borrow existing projectile/cone presentations.

## Acceptance criteria

- [ ] Five skills in `skills.v0.json` at tier 2 column 3 with class-appropriate prerequisites.
- [ ] Presentations + i18n keys for all five.
- [ ] Go test proves tree placement and class ownership.
- [ ] Extended bot scenario casts barbarian Ground Slam in skill lab.
- [ ] `make validate-shared` passes.

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run TestTier2Column3SkillExpansion -count=1
make bot scenario=hero_skill_expansion
```
