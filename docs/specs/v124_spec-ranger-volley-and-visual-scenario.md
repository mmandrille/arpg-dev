# v124 Spec - Ranger Volley And Visual Scenario

Status: Complete
Date: 2026-06-13
Codename: `ranger-volley-and-visual-scenario`

## Purpose

Finish the first Ranger skill set by adding `Volley`, a server-authoritative bow skill that fires a
fan of physical arrows, and add a visual showcase scenario covering Ranger's model, icon/loadout,
Piercing Shot, Pinning Shot, and Volley.

## Non-goals

- No production animation/audio set.
- No new client-owned hit selection or non-authoritative projectile simulation.
- No broad Ranger balance pass outside shared skill data.
- No PvP behavior.

## Acceptance Criteria

1. `shared/rules/skills.v0.json` contains Ranger-only `volley` with data-owned mana cost,
   cooldown, rank requirements, projectile range/speed/visual, physical damage, arrow count, and
   spread angle.
2. Skill schemas and Go rule validation reject malformed volley payloads.
3. Volley resolves server-side: each arrow ray is derived from the cast direction, hits at most one
   live monster, avoids duplicate target damage within one cast, emits deterministic damage events,
   spends mana, starts cooldown, and emits a projectile presentation event.
4. Ranger class coverage includes all three skills; other classes cannot learn or cast Volley.
5. Godot renders a distinct green multi-arrow Volley icon/projectile cue.
6. A protocol bot scenario proves a Ranger can cast Pinning Shot, Piercing Shot, and Volley in one
   visual-friendly lab, with semantic assertions for root, pierce, and multi-target Volley damage.
7. Focused Go tests, shared validation, bot tests, client tests, maintainability, and `make ci` pass.

## Likely Files

- `shared/rules/skills.v0.schema.json`
- `shared/rules/skills.v0.json`
- `shared/assets/skill_presentations.v0.schema.json`
- `shared/assets/skill_presentations.v0.json`
- `shared/i18n/en.json`, `shared/i18n/es.json`
- `server/internal/game/rules.go`
- `server/internal/game/ranger_skills.go`
- `server/internal/game/ranger_skills_test.go`
- `server/internal/game/handlers.go`
- `client/scripts/projectile_visuals.gd`
- `client/scripts/skill_icon.gd`
- `client/tests/test_projectile_visuals.gd`
- `client/tests/test_skill_rules_loader.gd`
- `tools/bot/scenarios/58_ranger_class_foundation.json`
- `tools/bot/scenarios/60_ranger_volley_and_visual_showcase.json`
- `tools/bot/test_protocol.py`
- `tools/validate_shared.py`
- `PROGRESS.md`, plan, and as-built docs

## Test And Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=60_ranger_volley_and_visual_showcase
make client-unit
make maintainability
make ci
```

Manual visual verification:

```bash
make bot-visual scenario=60_ranger_volley_and_visual_showcase
```

