# v154 Spec - Class Third Skill Trio

Status: Complete
Date: 2026-06-14
Codename: `class-third-skill-trio`

## Purpose

Add one new active higher-row skill for Barbarian, Rogue, and Ranger in a single vertical slice.
The slice should expand those classes without changing the behavior or tuning of their existing
skills.

Conservative default skill choices:

- Barbarian `earthbreaker`: a weapon-sourced cone smash that requires Cleave.
- Rogue `shadow_flurry`: a quick close cone attack that requires Dash.
- Ranger `split_arrow`: a direct projectile attack that requires Volley.

All three reuse currently supported skill capability types so the slice proves catalog expansion,
class gates, requirements, skill panel visibility, bot proof, and visual tooling without opening a
new mechanics framework.

## Non-goals

- No modifications to existing skill definitions or behavior.
- No Sorcerer or Paladin additions in this slice.
- No new skill kind or protocol schema.
- No final class balance pass.

## Acceptance criteria

1. `shared/rules/skills.v0.json` contains Barbarian `earthbreaker`, Rogue `shadow_flurry`, and
   Ranger `split_arrow` with class, tree, prerequisite, cost, cooldown, and damage data.
2. Each new skill is class-gated and requires a shipped same-class prerequisite skill.
3. Skill presentations, i18n strings, and skill visual tooling list all three new skills.
4. The server can learn and cast each new skill authoritatively, spending mana and starting
   cooldowns through the existing skill pipeline.
5. Cross-class learning/casting remains rejected.
6. Protocol bot coverage proves all three new skills can be seeded or learned and cast in focused
   per-class labs.
7. Client unit coverage proves the skills panel can surface the new higher-row skills.
8. Focused validation, Go tests, bot scenario, client tests, maintainability, and `make ci` pass.

## Scope and likely files

- `shared/rules/skills.v0.json`
- `shared/assets/skill_presentations.v0.json`
- `shared/i18n/en.json`
- `shared/i18n/es.json`
- `tools/validate_skills.py`
- `server/internal/game/game_test.go`
- `client/tests/test_skill_rules_loader.gd`
- `tools/bot/scenarios/62_barbarian_earthbreaker.json`
- `tools/bot/scenarios/63_rogue_shadow_flurry.json`
- `tools/bot/scenarios/64_ranger_split_arrow.json`
- `tools/bot/test_skill_demo.py`
- `PROGRESS.md`, plan, and as-built docs

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestLoadRules|TestSkill'
.venv/bin/pytest tools/bot/test_skill_demo.py -q
make bot scenario=62_barbarian_earthbreaker.json
make bot scenario=63_rogue_shadow_flurry.json
make bot scenario=64_ranger_split_arrow.json
make client-unit
make maintainability
make ci
```

Manual visual verification:

```bash
make bot-visual scenario=64_ranger_split_arrow.json
```

## Open questions and risks

- This slice intentionally reuses existing skill kinds. More distinctive mechanics for these
  classes remain future work.
- Existing large files such as `server/internal/game/game_test.go` should only receive focused
  assertions; no broad coordinator growth is needed.
