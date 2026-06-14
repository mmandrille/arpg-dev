# v171 Spec - Sorcerer Paladin Third Skills

Status: Complete
Date: 2026-06-14
Codename: `sorcerer-paladin-third-skills`

## Purpose

Complete the current higher-row active skill pass by adding one new third-row active skill for
Sorcerer and one for Paladin. The slice should expand both classes using the existing server-owned
skill catalog, progression gates, Godot skill presentation, and protocol bot proof.

Conservative default skill choices:

- Sorcerer `arcane_barrage`: a higher-row direct projectile spell that requires Lightning.
- Paladin `sanctuary`: a higher-row area ally defense skill that requires Holy Shield.

Both skills reuse supported capability types so this slice proves player-visible class expansion
without adding a new mechanics framework.

## Non-goals

- No new skill capability type or protocol schema.
- No changes to existing skill definitions or behavior.
- No passive skill tree, mana regeneration, or final balance pass.
- No production VFX/audio or external Godot plugin dependency.
- No broad skill-tree restructuring beyond the two new prerequisite links.

## Acceptance criteria

1. `shared/rules/skills.v0.json` contains Sorcerer `arcane_barrage` and Paladin `sanctuary` with
   class, tree, prerequisite, cost, cooldown, damage/effect, and text key data.
2. `arcane_barrage` is class-gated to Sorcerer and requires `ligthing` rank 1.
3. `sanctuary` is class-gated to Paladin and requires `holy_shield` rank 1.
4. Skill presentations, English/Spanish text, and skill visual tooling include both new skills.
5. The server can learn and cast both skills authoritatively through existing skill handlers,
   spending mana and starting cooldowns.
6. Cross-class learning/casting remains rejected.
7. Protocol bot coverage proves both skills can be seeded and cast in focused per-class labs.
8. Client unit coverage proves the skill loader/panel can surface the new higher-row skills.
9. Focused validation, Go tests, bot scenarios, client tests, maintainability, and `make ci` pass.

## Scope and likely files

- `shared/rules/skills.v0.json`
- `shared/assets/skill_presentations.v0.json`
- `shared/i18n/en.json`
- `shared/i18n/es.json`
- `tools/validate_skills.py`
- `tools/bot/test_skill_demo.py`
- `server/internal/game/game_test.go`
- `server/internal/game/class_third_skills_test.go`
- `client/tests/test_skill_rules_loader.gd`
- `tools/bot/scenarios/69_sorcerer_arcane_barrage.json`
- `tools/bot/scenarios/70_paladin_sanctuary.json`
- `PROGRESS.md`, plan, and as-built docs

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestLoadRules|TestThirdClassSkillsRequirePrerequisites|TestClassSkillGates' -count=1
.venv/bin/pytest tools/bot/test_skill_demo.py -q
make client-unit
make bot scenario=69_sorcerer_arcane_barrage.json
make bot scenario=70_paladin_sanctuary.json
make maintainability
make ci
```

Manual visual verification:

```bash
make bot-visual scenario=69_sorcerer_arcane_barrage.json
make bot-visual scenario=70_paladin_sanctuary.json
```

## Open questions and risks

- No blocking questions. This slice intentionally reuses existing skill capability types; more
  distinctive Sorcerer and Paladin mechanics remain future work.
- The existing `ligthing` skill id is misspelled and already shipped. This slice uses that stable
  id for the Sorcerer prerequisite rather than renaming it.
