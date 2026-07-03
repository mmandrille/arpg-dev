# v420 Plan — Class skill build branches (full batch)

Status: Complete  
Goal: Add 30 build-branch actives (6 per class) reusing existing skill kinds.  
Architecture: Data-only `skills.v0.json` + i18n + presentations; `tools/gen_class_build_branch_skills.py` generator; validator active count 13/class.

## File map

| Action | Path |
|--------|------|
| Create | `docs/specs/v420_spec-class-skill-build-branches.md` |
| Create | `tools/gen_class_build_branch_skills.py` |
| Modify | `shared/rules/skills.v0.json` |
| Modify | `shared/i18n/en.json` |
| Modify | `shared/assets/skill_presentations.v0.json` |
| Modify | `tools/validate_skills.py` |
| Create | `server/internal/game/class_build_branches_test.go` |
| Create | `tools/bot/scenarios/class_build_branches_lab.json` |

## Tasks

- [x] Umbrella spec with full 30-skill catalog
- [x] Generator + merge all five classes
- [x] Fix i18n nesting, synergies (direct prereq), layout hints, buff cooldowns, volley cap
- [x] Go registration + Gore Strike bleed test
- [x] Extended bot lab
- [x] `make validate-shared` green

## Final verification

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run 'TestClassBuildBranch|TestGoreStrike' -count=1
make bot scenario=class_build_branches_lab
```

Visual: `/showme skills` with each class selected.
