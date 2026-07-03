# v420 As-Built — Class Skill Build Branches

Date: 2026-07-03

## What shipped

- **30 new actives** (6 per class): all slot-A forks and slot-B capstones from the v420 brief.
- Per-class active count **7 → 13** (unchanged: 1 mobility, 4 passives, 1 survival).
- Generator merged all five classes into `skills.v0.json` (one-shot data entry).
- Direct-prerequisite synergies, layout-hint fixes, buff cooldown validation fixes.
- Go tests: registration smoke (10 skills) + `gore_strike` bleed cast.
- Extended bot: `class_build_branches_lab`.

## Proof

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run 'TestClassBuildBranch|TestGoreStrike' -count=1
make bot scenario=class_build_branches_lab
```

## Boundaries

- `pinning_volley` uses volley-only (no volley+root combo — engine constraint).
- Single v420 slice for the full 30-skill batch (user-approved all options at once).
- Extended-only bot scenario; not CI pack.
