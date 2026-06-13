# v124 Plan - Ranger Volley And Visual Scenario

Status: Complete
Goal: Add Ranger `Volley` plus a visual-friendly protocol scenario covering all three Ranger skills.
Architecture: Shared skill data owns Volley count/spread/projectile/cost tuning. The Go Ranger skill helper resolves the fan deterministically from one cast direction. Godot renders data-named projectile/icon variants only.
Tech stack: shared JSON/schema, Go sim, Python protocol bot, Godot presentation, SDD docs.

## Baseline And Shortcut Decision

Baseline is v123 `ranger-piercing-and-pinning-shots` on `main`, committed as `107537b7`.

Godot plugin adoption checklist: reject external plugins/assets. Volley uses existing code-native
skill icon and projectile presentation helpers.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Add closed `volley` payload. |
| Modify | `shared/rules/skills.v0.json` | Add Ranger `volley` skill row. |
| Modify | `shared/assets/skill_presentations.v0.schema.json`, `.json` | Add volley icon/projectile metadata. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add localized Volley name/summary. |
| Modify | `tools/validate_shared.py` | Cross-check Ranger skill set and projectile presentations. |
| Modify | `server/internal/game/rules.go` | Decode/validate volley rules. |
| Modify | `server/internal/game/ranger_skills.go` | Add deterministic fan resolution. |
| Modify | `server/internal/game/ranger_skills_test.go` | Add Volley focused tests. |
| Modify | `client/scripts/projectile_visuals.gd`, `skill_icon.gd` | Add Volley visual variants. |
| Modify | `client/tests/test_projectile_visuals.gd`, `test_skill_rules_loader.gd` | Add client proof. |
| Modify | `tools/bot/scenarios/58_ranger_class_foundation.json` | Reference all Ranger skills for class coverage. |
| Add | `tools/bot/scenarios/60_ranger_volley_and_visual_showcase.json` | Visual-friendly all-skills proof. |
| Modify | `tools/bot/test_protocol.py` | Scenario discovery proof. |
| Add | `docs/as-built/v124_ranger-volley-and-visual-scenario.md` | As-built proof. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `server/internal/game/rules.go`
- [x] `client/scripts/main.gd`
- [x] `tools/bot/test_protocol.py`
- [x] `tools/validate_shared.py`

Decision:
- [x] Keep mechanics in `ranger_skills.go` and tests in `ranger_skills_test.go`.
- [x] Defer large-file extraction only for thin schema/validator/scenario discovery wiring.

## Task 1 - Shared Volley Content

- [x] Step 1.1: Add schema-backed `volley` payload with arrow count and spread degrees.
- [x] Step 1.2: Add `volley` Ranger skill row with physical damage, projectile visual, requirements,
  mana cost, and cooldown.
- [x] Step 1.3: Add presentation metadata and localized text.
- [x] Step 1.4: Update validators for the three-skill Ranger set and projectile visual coverage.

## Task 2 - Server Authority

- [x] Step 2.1: Decode and validate Volley rules.
- [x] Step 2.2: Dispatch Volley through the existing Ranger projectile helper path.
- [x] Step 2.3: Resolve a deterministic fan of arrow rays from the cast direction.
- [x] Step 2.4: Damage at most one target per arrow and avoid duplicate targets within one cast.
- [x] Step 2.5: Add tests for fan hit ordering, duplicate prevention, class gates, and current
  Ranger skill regressions.

## Task 3 - Bot And Visual Scenario

- [x] Step 3.1: Seed a Ranger visual showcase scenario with all three skill ranks.
- [x] Step 3.2: Prove Pinning Shot root, Piercing Shot multi-hit, and Volley multi-target damage.
- [x] Step 3.3: Keep the scenario visually framed and runnable with
  `make bot-visual scenario=60_ranger_volley_and_visual_showcase`.
- [x] Step 3.4: Add protocol discovery tests.

## Task 4 - Godot Presentation

- [x] Step 4.1: Render `volley_arrow_projectile` as a distinct green arrow cue.
- [x] Step 4.2: Render the Volley icon as a multi-arrow/fan shape.
- [x] Step 4.3: Add focused GDScript tests for the projectile and loader presentation.

## Task 5 - Lifecycle Docs And CI

- [x] Step 5.1: Mark spec/plan complete and add as-built notes.
- [x] Step 5.2: Update `PROGRESS.md` latest completed slice, next slice, lifecycle row, and
  recently closed notes.
- [x] Step 5.3: Run final verification and commit after green CI.

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=60_ranger_volley_and_visual_showcase`
- [x] `make client-unit`
- [x] `make ci`

