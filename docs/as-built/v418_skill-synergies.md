# v418 As-Built — Skill Synergies

Date: 2026-07-03

## What shipped

- Declarative `synergies[]` on 41 prerequisite skills (data + `validate_skill_synergies` CI gate landed pre-slice).
- Server `skill_synergies.go`: allocated source rank × `percent_per_source_rank`; scaled helpers for damage, cone,
  volley spread, projectile range, buff duration/power, area radius, root/mark/bleed duration, revive power,
  passive stat %, execute threshold.
- Synergy source rank uses `progression.SkillRanks` only (not gear `+N` effective rank).
- `SkillProgressionView` rows expose optional `synergy_status` with per-source display lines.
- Client `skill_synergy_tooltip.gd` + skills panel "Synergies:" section (server view or rules fallback).
- Extended bot proof `skill_synergy_lab`: snipe damage scales with piercing shot ranks.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run SkillSynergy -count=1
make client-unit
make bot scenario=skill_synergy_lab
make ci
```

## Deferred

- Synergy tuning pass on the 41-row catalog
- Reverse synergies, respec UX
- Engineering review at v418 milestone (post-loop handoff)
