# v417 Plan — Skill Rank Scaling and Pacing

Status: Ready for implementation
Goal: Reshape skill-point cadence, raise max buyable rank to 10, and unify compound rank scaling across server and client tooltips.
Architecture: Extend `character_progression.v0.json` with cadence + global rank curves. New `skill_rank_scaling.go` centralizes `rankScaledInt`/`rankScaledFloat`; sim paths call it instead of inline linear math. Client `SkillRankScaling` mirrors for tooltips. Bulk-update `max_rank` 5→10 in `skills.v0.json`.

Tech stack: shared JSON, Go sim, GDScript client, Python validator/bot.

## Baseline and shortcut decision

Builds on v416 effective rank (gear stacks without use clamp). **Borrow** existing skill UI/VFX.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/character_progression.v0.json` + schema | Cadence, rank curves |
| Modify | `shared/rules/skills.v0.json` | `max_rank: 10` for rankable skills |
| Create | `server/internal/game/skill_rank_scaling.go` | Rank curve evaluators |
| Modify | `server/internal/game/class_progression.go` | New cadence logic |
| Modify | `server/internal/game/rules.go` | Load new progression fields |
| Modify | `server/internal/game/sim.go`, `skill_weapon_damage.go`, `passive_skill_stats.go`, etc. | Wire scaling |
| Create | `client/scripts/skill_rank_scaling.gd` | Tooltip parity |
| Modify | `client/scripts/skills_panel.gd`, `skill_passive_tooltip.gd`, `skill_bar.gd` | Use evaluator |
| Modify | `tools/validate_skills.py` | Cadence cross-check |
| Modify | `tools/bot/scenarios/class_progression_pacing.json` | Cadence proof |
| Create | `tools/bot/scenarios/skill_rank_scaling_lab.json` | Rank scaling proof |
| Create | `client/tests/test_skill_rank_scaling.gd` | GDScript parity |
| Create | `server/internal/game/skill_rank_scaling_test.go` | Go curve tests |
| Update | `docs/as-built/v417_skill-rank-scaling-and-pacing.md`, `PROGRESS.md`, lifecycle |

## Maintenance ratchet

- [x] New files under 600 lines
- [x] `make maintainability`

## Task 1 — Shared progression + schema

- [x] Extend `skill_points` with `second_grant_level`, `grant_every_min_level`
- [x] Add `skill_rank_scaling`, `skill_mana_scaling` blocks
- [x] Update schema; `make validate-shared`

## Task 2 — Go rank scaling core

- [x] `skill_rank_scaling.go` + tests
- [x] Update `skillPointGrantLevel`, `skillRequirementsForRank`

## Task 3 — Wire sim rank scaling

- [x] Replace linear rank math in damage, effects, mana, passives, mobility, rogue/companion helpers
- [x] `cd server && go test ./internal/game -run SkillRank -count=1`

## Task 4 — Skills catalog

- [x] Bulk `max_rank` 5 → 10 (keep `max_rank: 1` utilities)
- [x] `make validate-shared`

## Task 5 — Client tooltips

- [x] `skill_rank_scaling.gd` + `test_skill_rank_scaling.gd`
- [x] Register in `client_smoke.sh`
- [x] Update `skills_panel.gd`, `skill_passive_tooltip.gd`, `skill_bar.gd`

## Task 6 — Bot scenarios

- [x] Update `class_progression_pacing.json` (4 grants at L6)
- [x] Create `skill_rank_scaling_lab.json` (`ci_tier: extended`)
- [x] `make bot scenario=class_progression_pacing scenario=skill_rank_scaling_lab`

## Task 7 — Lifecycle docs

- [x] As-built, PROGRESS, lifecycle row

## Final verification

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run 'SkillPoint|SkillRank' -count=1
make bot scenario=class_progression_pacing scenario=skill_rank_scaling_lab
```
