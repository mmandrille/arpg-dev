# v90 Plan - Text Catalog Foundation

Status: Complete
Goal: Add the English localization catalog and client lookup path so visible text can move out of scripts/content incrementally.
Architecture: English source text lives in shared JSON. Godot uses a focused text service with key lookup and fallback. Shared validation checks catalog shape. Existing gameplay data keeps its current fields while presentation gains text keys.
Tech stack: shared JSON/schema, Python shared validator, Godot client scripts/tests, client bot smoke.

## Baseline and shortcut decision

Baseline is v89 `class-second-combat-skills` plus the v90 engineering review cleanup commit on
`main`. This slice starts the localization work requested for menus, skills, stats, future quests,
monster names, and other visible text.
## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `shared/i18n/en.json` | English source catalog for current visible text keys. |
| Add | `shared/i18n/i18n.v0.schema.json` | Catalog shape documentation/schema if useful. |
| Modify | `shared/rules/skills.v0.json` | Add text keys for current skill names/summaries. |
| Modify | `shared/rules/monsters.v0.json` | Add text keys for current monster names. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add or consume text keys for presentation summaries. |
| Modify | `tools/validate_shared.py` | Validate English catalog shape and referenced keys. |
| Add | `client/scripts/text_catalog.gd` | Focused Godot text lookup/fallback service. |
| Modify | `client/scripts/main_menu.gd` | Resolve root menu labels through text keys. |
| Modify | `client/scripts/pause_menu.gd` | Resolve Settings label through text keys. |
| Modify | `client/scripts/settings_panel.gd` | Resolve title/field labels through text keys. |
| Modify | `client/scripts/client_settings.gd` | Resolve session-type labels through text keys. |
| Modify | `client/scripts/stat_labels.gd` | Resolve stat labels through text keys. |
| Modify | `client/scripts/skill_rules_loader.gd` | Expose localized skill name/summary helpers. |
| Add/Modify | `client/tests/*text*` | Unit coverage for loader, fallback, and migrated labels. |
| Add | `docs/as-built/v90_text-catalog-foundation.md` | As-built proof. |
| Modify | `PROGRESS.md` | Mark v90 complete and steer v91 Spanish selector. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `tools/validate_shared.py`

Decision:
- [x] Prefer a focused helper/module for catalog validation if validator changes become non-trivial.
- [x] Keep `main.gd` untouched unless integration discovers unavoidable small wiring.

Rationale: validator growth was a small cross-check block plus schema mapping, and `main.gd` did not
need to change. New client functionality lives in `text_catalog.gd`. The implementation also split
bot debug progression seeding and i18n validation into focused helper modules. The final baseline
update covers already-committed v89 backend file growth that blocked the ratchet before this slice
could finish; v90 did not add backend code.

## Task 1 - Shared English catalog

Files:
- Add: `shared/i18n/en.json`
- Add/Modify: `shared/i18n/i18n.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/assets/skill_presentations.v0.json`

- [x] Step 1.1: Define the English catalog structure and key naming convention.
- [x] Step 1.2: Add keys for menu, settings, common actions, stats, classes, skills, and current monsters.
- [x] Step 1.3: Add `name_key` / `summary_key` style fields where shared content currently owns visible names.

```bash
make validate-shared
```

## Task 2 - Shared validation

Files:
- Modify: `tools/validate_shared.py` or add a focused helper

- [x] Step 2.1: Validate that `shared/i18n/en.json` is valid JSON with non-empty string keys/values.
- [x] Step 2.2: Validate referenced skill and monster text keys exist in English.
- [x] Step 2.3: Add or update Python tests if the validation surface changes.

```bash
.venv/bin/pytest tools -q
make validate-shared
```

## Task 3 - Godot text service and migrated labels

Files:
- Add: `client/scripts/text_catalog.gd`
- Modify: `client/scripts/main_menu.gd`
- Modify: `client/scripts/pause_menu.gd`
- Modify: `client/scripts/settings_panel.gd`
- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/stat_labels.gd`
- Modify: `client/scripts/skill_rules_loader.gd`

- [x] Step 3.1: Implement English catalog loading from `shared/i18n/en.json`.
- [x] Step 3.2: Implement key lookup with supplied fallback and non-crashing missing-key behavior.
- [x] Step 3.3: Migrate root menu, pause/settings labels, Settings panel labels, create-game type labels, stat labels, and skill display helpers to the text service.
- [x] Step 3.4: Keep any untouched legacy literals documented as deferred migration, not new pattern.

```bash
make client-unit
```

## Task 4 - Client tests and focused bot proof

Files:
- Add/Modify: `client/tests/*text*`
- Modify client-bot expectations only if migrated visible labels require it

- [x] Step 4.1: Test catalog load and fallback.
- [x] Step 4.2: Test migrated skill display helper uses the English catalog.
- [x] Step 4.3: Run a focused menu/settings client bot scenario when practical.

```bash
make client-unit
make bot-client scenario=08_main_menu_flow.json
```

## Task 5 - Lifecycle docs and finish

Files:
- Modify: `docs/specs/v90_spec-text-catalog-foundation.md`
- Modify: `docs/plans/v90_2026-06-12-text-catalog-foundation.md`
- Add: `docs/as-built/v90_text-catalog-foundation.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and add as-built.
- [x] Step 5.2: Update `PROGRESS.md` latest completed slice, lifecycle row, recently closed notes, and next slice.
- [x] Step 5.3: Run final verification and commit.

## Final verification

- [x] `make validate-shared`
- [x] `.venv/bin/pytest tools -q`
- [x] `make client-unit`
- [x] `make bot-client scenario=08_main_menu_flow.json HEADLESS=1`
- [x] `make maintainability`
- [x] `make bot-client scenario=20_menu_create_join_flow.json HEADLESS=1`
- [x] `make bot scenario=45_class_second_combat_skills.json VERBOSE=1`
- [x] `make ci`

```bash
make ci
```

## Deferred scope

- Spanish catalog and Settings language selector.
- Full migration of every tooltip/shop/inventory/future quest line.
- Server runtime localization.
- Content id spelling migrations such as `ligthing`.
