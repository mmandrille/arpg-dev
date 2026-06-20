# v309 Plan — Passive Skill Column

Status: Complete
Goal: Add a server-authoritative right-side passive skill chain for every class.
Architecture: Introduce a declarative `passive_stat_bonus` skill kind with rank-scaled stat bonuses. Skill rules remain shared data, the Go sim owns derived-stat effects, and the Godot client renders summaries/icons from shared presentation metadata.
Tech stack: Shared JSON schemas/rules, Go sim, GDScript skill panel, Python validation, lifecycle docs.

## Baseline and shortcut decision

Builds on v308 with the existing class-gated skill tree, skill presentations, and server-owned allocate/cast validation. Asset/plugin decision: reject external assets/plugins; borrow existing code-native skill icon shapes and shared `skill_presentations.v0.json` metadata for the passive logos.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Add passive-stat skill contract |
| Modify | `shared/rules/skills.v0.json` | Add 15 class passive skill definitions |
| Modify | `shared/assets/skill_presentations.v0.schema.json` | Allow reused icon shapes only if needed |
| Modify | `shared/assets/skill_presentations.v0.json` | Add 15 passive logos/summaries |
| Modify | `server/internal/game/rules.go` | Load and validate passive stats |
| Modify | `server/internal/game/sim.go` | Apply passive stat bonuses to derived stats |
| Create | `server/internal/game/passive_skill_rules.go` | Focused passive rule validation helpers |
| Create | `server/internal/game/passive_skill_stats.go` | Focused passive stat application helpers |
| Create | `server/internal/game/passive_skill_column_test.go` | Focused passive skill authority tests |
| Modify | `tools/validate_skills.py` | Shared passive-chain cross-check |
| Modify | `client/scripts/skills_panel.gd` | Passive effect tooltip lines |
| Create | `client/scripts/skill_passive_tooltip.gd` | Focused passive tooltip helpers |
| Modify | `client/tests/test_skills_panel.gd` | Passive tooltip/presentation test |
| Modify | `tools/bot/class_foundation_coverage.py` | Keep active class scenario coverage scoped to scenario-relevant skills |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle status |
| Create | `docs/as-built/v309_passive-skill-column.md` | Slice proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/skills_panel.gd`
- [x] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `server/internal/game/sim.go`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Defer extraction with rationale: changes are narrow additions to existing stat/progression surfaces; splitting during this slice would risk unrelated architecture churn.

Verification:
```bash
make maintainability
```

## Task 1 — Shared passive contracts and catalog

Files:
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/assets/skill_presentations.v0.json`
- Modify: `tools/validate_skills.py`

- [x] Add `passive_stat_bonus` schema with `passive_stats.stats` rank-linear values.
- [x] Add the 15 one-rank class passives at tier rows 1, 2, and 3 in the right-side column.
- [x] Add icon metadata and summaries for every passive.
- [x] Extend shared validation to require the passive chain levels, prerequisites, no stat requirements, class ownership, and presentations.

```bash
make validate-shared
```

## Task 2 — Server authority for passive stat effects

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Load and validate passive stat definitions.
- [x] Add learned passive stat totals to effective base stats, item-like derived stat totals, and stat breakdowns.
- [x] Reject passive casts through the existing non-castable path.
- [x] Add focused tests for unlock chain, level gates, derived stat effects, class visibility, and cast rejection.

```bash
cd server && go test ./internal/game -run 'TestPassiveSkillColumn|TestClassSkillAccessGatesSpendabilityAndLearning'
```

## Task 3 — Client passive tooltip presentation

Files:
- Modify: `client/scripts/skills_panel.gd`
- Modify: `client/tests/test_skills_panel.gd`

- [x] Show passive stat bonuses in tooltip bodies and next-rank lines.
- [x] Add a headless test proving a passive summary/effect and icon metadata can be read from shared data.

```bash
make client-unit
```

## Task 4 — Lifecycle docs and focused verification

Files:
- Modify: `docs/plans/v309_2026-06-20-passive-skill-column.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Create: `docs/as-built/v309_passive-skill-column.md`

- [x] Mark completed plan tasks.
- [x] Update lifecycle docs and as-built proof.
- [x] Run focused verification for the touched surfaces.

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestPassiveSkillColumn|TestClassSkillAccessGatesSpendabilityAndLearning'
make client-unit
make maintainability
.venv/bin/pytest tools -q
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestPassiveSkillColumn|TestClassSkillAccessGatesSpendabilityAndLearning'`
- [x] `make client-unit`
- [x] `.venv/bin/pytest tools -q`
- [x] `make maintainability`
- [x] Batch-level `make ci` after transferring verified changes back to `main`
