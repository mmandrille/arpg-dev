# v418 Plan — Skill Synergies

Status: Complete
Goal: Apply declarative skill synergies in the Go sim and show accumulated bonuses in skill tooltips.
Architecture: `skill_synergies.go` aggregates allocated source ranks × `percent_per_source_rank`;
cast/stat hooks call `synergyScaledInt/Float`; optional `synergy_status` on skill progression rows.
Tech stack: shared JSON (done), Go sim, Godot client, Python bot.

## Baseline and shortcut decision

Data + CI gate landed pre-slice (`synergies[]` on 41 skills, `validate_skill_synergies`). **Borrow**
skill tooltip patterns from v416 `skill_bonus_tooltip.gd`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/internal/game/skill_synergies.go` | rank lookup, bonus math, scaled helpers, status views |
| Create | `server/internal/game/skill_synergies_test.go` | snipe, rend, ward, passive tests |
| Modify | `server/internal/game/rules.go` | `SkillSynergyDef`, `SkillDef.Synergies` |
| Modify | `server/internal/game/skill_weapon_damage.go` | damage synergy wrapper |
| Modify | `server/internal/game/ranger_skills.go` | volley spread, root, damage |
| Modify | `server/internal/game/handlers.go` | cone + projectile range |
| Modify | `server/internal/game/skill_buffs.go` | buff duration/power, area radius |
| Modify | `server/internal/game/class_skill_engine.go` | bleed/mark duration |
| Modify | `server/internal/game/passive_skill_stats.go` | passive_stat_percent |
| Modify | `server/internal/game/rogue_skills.go` | execute threshold |
| Modify | `server/internal/game/rules_companion.go` | revive power |
| Modify | `server/internal/game/ranger_skills.go` | applyMonsterRoot duration |
| Modify | `server/internal/game/sim.go` | `SkillProgressionView` synergy_status |
| Modify | `server/internal/game/types.go` | view structs |
| Modify | `shared/protocol/state_delta.v8.schema.json` | optional `synergy_status` |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | optional `synergy_status` |
| Create | `client/scripts/skill_synergy_tooltip.gd` | tooltip lines |
| Create | `client/tests/test_skill_synergy_tooltip.gd` | headless test |
| Modify | `client/scripts/skills_panel.gd` | wire tooltip |
| Modify | `scripts/client_smoke.sh` | register GDScript test |
| Create | `tools/bot/scenarios/skill_synergy_lab.json` | extended proof |

## Maintenance ratchet

New `skill_synergies.go` ≤600 lines. Delegate from `sim.go` / `handlers.go` only.

## Task 1 — Server core

- [x] Types + `skill_synergies.go` helpers
- [x] `skillDamageRangeForSkill`, cone/volley/buff/passive hooks

```bash
cd server && go test ./internal/game -run SkillSynergy -count=1
```

## Task 2 — Protocol views

- [x] `synergy_status` on `skill_progression_skill`
- [x] `make validate-shared`

## Task 3 — Client tooltips

- [x] `skill_synergy_tooltip.gd` + `skills_panel.gd`
- [x] `make client-unit`

## Task 4 — Bot proof

- [x] `skill_synergy_lab.json` (`ci_tier: extended`)

```bash
make bot scenario=skill_synergy_lab
```

## Task 5 — Finish

- [x] `docs/as-built/v418_skill-synergies.md`, lifecycle row, `PROGRESS.md`
