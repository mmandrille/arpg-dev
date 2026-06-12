# v90 As-Built - Text Catalog Foundation

Date: 2026-06-12
Status: Complete

## What shipped

- Added `shared/i18n/en.json` as the English source catalog for current visible menu, settings,
  stat, class, skill, and monster text.
- Added `shared/i18n/i18n.v0.schema.json` and validator coverage for catalog shape.
- Added `name_key` fields for current skill and monster rows, plus `summary_key` fields for current
  skill presentation rows.
- Added `client/scripts/text_catalog.gd`, a small Godot text lookup service that loads English and
  falls back to a supplied literal or the key without crashing.
- Migrated first client presentation surfaces to catalog-backed helpers: main menu, pause menu,
  Settings panel, create-game session type labels, stat labels, character class summaries, skill
  display names/summaries, skill bar tooltips, skills panel labels, and status-effect names.
- Registered `client/tests/test_text_catalog.gd` in `scripts/client_smoke.sh`.
- Split bot debug progression seeding and i18n validation into small helper modules so touched tool
  files stay within the maintainability ratchet.
- Updated the grandfathered backend file-size baseline for already-committed v89 backend growth that
  was blocking the ratchet before this slice could finish; v90 did not add backend code.

## Verification

```bash
make validate-shared
.venv/bin/pytest tools -q
make client-unit
make bot-client scenario=08_main_menu_flow.json HEADLESS=1
make maintainability
make bot-client scenario=20_menu_create_join_flow.json HEADLESS=1
make bot scenario=45_class_second_combat_skills.json VERBOSE=1
make ci
```

All listed checks passed during implementation. Final `make ci` ended with `CI OK`.

## Deferred

- Spanish catalog and Settings language selector.
- Full migration of every shop, stash, inventory, tooltip, quest, and debug label.
- Server-side runtime localization.
- Content id spelling migrations such as `ligthing`.
