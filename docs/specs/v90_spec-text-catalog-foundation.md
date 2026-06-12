# Spec: `text-catalog-foundation`

Status: Complete
Date: 2026-06-12
Codename: `text-catalog-foundation`
Slice: v90 - text catalog foundation
Baseline: v89 `class-second-combat-skills` plus v90 engineering review cleanup

## Purpose

Create the localization foundation for all player-visible text. English becomes the source catalog,
client code gains a small reusable text lookup service, and current high-traffic visible labels start
resolving by stable keys instead of hardcoded strings.

This is the first of two localization slices. It does not add Spanish selection yet; it creates the
catalog shape, loader, fallback behavior, and tests that the Spanish slice will reuse.

## Non-goals

- No Settings language selector in this slice.
- No Spanish catalog in this slice.
- No final migration of every tooltip, shop line, inventory line, quest line, or monster label.
- No protocol change.
- No server-side runtime localization.
- No backward-compatibility layer for old local settings beyond preserving existing settings fields.

## Acceptance criteria

1. A shared English text catalog exists under `shared/i18n/` and is treated as the source of truth
   for localized visible text.
2. The catalog contains stable keys for current core menus, Settings labels, common action labels,
   stat labels, class names, current skill names/descriptions, and current monster names.
3. A focused Godot text service loads English, supports `tr(key, fallback)` style lookup, and returns
   English fallback text when a key is missing.
4. Main menu, pause/settings entry points, Settings panel labels, stat labels, character class summary
   labels, skill names/descriptions, and monster names touched by this slice resolve through the text
   service rather than direct literals.
5. Skill and monster shared data can reference text keys without breaking existing rule loading.
6. Missing text keys do not crash the client and fall back to the supplied English literal or key.
7. Shared validation checks that the English catalog has the expected shape and no empty text keys.
8. Client unit tests cover catalog loading, fallback behavior, and at least one migrated panel/loader.
9. Future work is documented: Spanish catalog plus Settings language selector is the next slice.

## Scope and likely files

- Shared:
  - Add `shared/i18n/en.json`
  - Add `shared/i18n/i18n.v0.schema.json` if schema validation needs a dedicated schema
  - Modify selected `shared/rules/*.json` rows to add `text_key` / `name_key` where useful
  - Modify `tools/validate_shared.py` or a focused helper if practical
- Client:
  - Add `client/scripts/text_catalog.gd`
  - Modify `client/scripts/main_menu.gd`
  - Modify `client/scripts/pause_menu.gd`
  - Modify `client/scripts/settings_panel.gd`
  - Modify `client/scripts/client_settings.gd`
  - Modify `client/scripts/stat_labels.gd`
  - Modify `client/scripts/skill_rules_loader.gd`
  - Modify `client/scripts/monster_visuals_loader.gd` or name consumers if needed
  - Add focused client tests
- Docs:
  - Add `docs/plans/v90_2026-06-12-text-catalog-foundation.md`
  - Add `docs/as-built/v90_text-catalog-foundation.md`
  - Update `PROGRESS.md`

## Test and bot proof

- `make validate-shared`
- `make client-unit`
- `.venv/bin/pytest tools -q` if validation tests change
- `make bot-client scenario=08_main_menu_flow.json` or the equivalent focused client-bot scenario if
  the harness supports single-scenario selection
- `make ci` before finish

## Open questions and assumptions

- Use neutral English source text as the canonical fallback.
- Use stable dotted keys such as `menu.create_game`, `settings.title`, `skill.magic_bolt.name`, and
  `monster.training_dummy.name`.
- Existing `name` fields in gameplay JSON remain temporarily for compatibility; this slice adds keys
  and starts migrating client presentation paths without requiring one giant content rewrite.

## Shortcut decision

Godot plugin adoption: reject external plugins/assets. Localization only needs shared JSON, a small
Godot loader, and existing UI/client tests.
