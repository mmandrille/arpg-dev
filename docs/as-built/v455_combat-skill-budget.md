# v455 As-Built — combat-skill-budget

Date: 2026-07-08  
Spec: [`docs/specs/v455_spec-combat-skill-budget.md`](../specs/v455_spec-combat-skill-budget.md)

## Shipped

- `combat_processing` caps in `shared/rules/combat.v0.json` (12 skill resolutions, 8 projectile spawns, 48 damage soft cap metadata).
- `server/internal/game/skill_budget.go`: defer queue, counters, prepend on next tick.
- `handleCastSkill` defers overflow casts; projectile spawn defer in `handleProjectileSkillCast`.
- `TestSkillBudgetDefersOverflowCastsDeterministically`.
- Extended scenario `crowded_skill_overlap_lab`.

## Verification

```bash
cd server && go test ./internal/game/... -run SkillBudget -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_skill_overlap_lab
```
