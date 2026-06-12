# v91 As-Built - Spanish Language Selector

## What Shipped

- Added `shared/i18n/es.json` as the first non-English locale catalog.
- Extended shared validation so every locale catalog must declare its locale, contain only keys that exist in English, and provide non-empty values.
- Added selected-locale overlay loading in `TextCatalog`, with English as the key-by-key fallback.
- Persisted the selected language in `ClientSettings`.
- Added a Settings language selector for English and Spanish.
- Refreshed menu, pause, and Settings text immediately after language changes.
- Added client tests for Spanish lookup, fallback, Settings selection state, and settings persistence.

## Verification

- `make validate-shared`
- `.venv/bin/pytest tools -q`
- `make client-unit`
- `make bot-client scenario=20_menu_create_join_flow.json HEADLESS=1`
- `make ci`

## Follow-Ups

- Continue migrating remaining visible legacy strings into `shared/i18n/en.json`.
- Add server-side localized labels when protocol-facing text becomes user-visible.
