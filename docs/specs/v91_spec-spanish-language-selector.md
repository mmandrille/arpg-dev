# Spec: `spanish-language-selector`

Status: Complete
Date: 2026-06-12
Codename: `spanish-language-selector`
Slice: v91 - Spanish language selector
Baseline: v90 `text-catalog-foundation`

## Purpose

Add Spanish localization for the current text catalog and expose a Settings language selector. The
client must persist the selected language locally, update catalog-backed labels, and fall back to
English key-by-key when a locale is missing a string.

## Acceptance criteria

1. `shared/i18n/es.json` exists and translates the current English catalog keys.
2. Shared validation accepts locale catalogs and verifies Spanish shape.
3. `TextCatalog` can switch between `en` and `es`, reload labels, and fall back to English for
   missing Spanish entries.
4. `ClientSettings` persists a normalized `language` setting with default `en`.
5. Settings shows a language selector with English and Spanish options.
6. Changing language in Settings updates catalog-backed menu/settings labels without restarting.
7. Client unit tests cover Spanish lookup, missing-key English fallback, settings persistence, and
   the Settings language buttons.

## Non-goals

- No full migration of remaining legacy tooltip/shop/inventory/debug literals.
- No server-side localization.
- No locale auto-detection.
