# v91 Plan - Spanish Language Selector

Status: Complete
Goal: Add Spanish catalog support and a persisted Settings language selector with English fallback.
Architecture: Shared `en` remains the fallback catalog. `TextCatalog` overlays the selected locale
on top of English. `ClientSettings` persists the selected language. Settings emits language changes
and `main.gd` refreshes menu/settings labels.

## Tasks

- [x] Add `shared/i18n/es.json`.
- [x] Extend shared validation to include all locale catalogs.
- [x] Extend `TextCatalog` with `set_locale`, `current_locale`, and key-by-key English fallback.
- [x] Persist language in `ClientSettings`.
- [x] Add Settings language buttons and signal.
- [x] Wire main menu Settings language changes and refresh labels.
- [x] Add client tests for Spanish lookup, fallback, settings persistence, and Settings language buttons.
- [x] Update lifecycle docs.

## Verification

- [x] `make validate-shared`
- [x] `.venv/bin/pytest tools -q`
- [x] `make client-unit`
- [x] `make bot-client scenario=20_menu_create_join_flow.json HEADLESS=1`
- [x] `make ci`

## Deferred

- Remaining legacy text migration.
- Server-side localization.
- Locale auto-detection.
