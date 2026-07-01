# v392 Spec — Class Creation Features Panel

Status: Draft
Date: 2026-06-30
Codename: class-creation-features

## Purpose

When creating a hero, selecting a class should reveal that class's **special features** in a
dedicated summary panel — not only the single starter skill label and hover tooltip on the class
button. Players see signature actives (grouped by skill-tree tier), passive column skills, and
class identity stats (starting attributes, movement speed, light radius) sourced from shared rules
rather than hardcoded `CLASS_DEFS` duplicates in `character_select_panel.gd`.

## Non-goals

- No new skills, class balance changes, or server authority changes.
- No class model preview, portraits, or production art.
- No protocol/schema bump; creation still sends `character_class` only.
- No migration UI for changing class after creation.
- No full skill-tree UI on the create screen (read-only feature list only).

## Acceptance criteria

- [ ] Shared rules catalog lists per-class creation feature entries: starting stats, movement speed,
  light radius, ordered signature active skill ids by tier, and passive skill ids.
- [ ] `character_select_panel.gd` loads creation metadata from shared rules (via existing loader
  pattern); hardcoded `CLASS_DEFS` stats/skill strings are removed or reduced to fallbacks only.
- [ ] Clicking a class button updates a visible **features panel** below the class picker with:
  class name, starting stat line, movement/light identity line, tier-grouped active skill names,
  and passive skill names.
- [ ] Default selection (`barbarian`) shows its features on first open of forced-create mode.
- [ ] `get_debug_state()` exposes selected class feature lines for headless tests and client bot.
- [ ] Client unit test proves feature panel content changes when `select_class()` is called for at
  least two classes.
- [ ] Client bot scenario `20_menu_create_join_flow` (or focused extension) asserts non-empty
  feature summary for the selected class before create.
- [ ] `make validate-shared` passes.

## Scope and likely files

- `shared/rules/character_progression.v0.json` — optional `creation_features` block per class, or
  new `shared/rules/class_creation.v0.json` + schema if cleaner separation.
- `shared/rules/skills.v0.json` — read-only reference for skill display names (no skill changes).
- `client/scripts/character_select_panel.gd` — features panel UI + loader wiring.
- New `client/scripts/class_creation_loader.gd` (or extend existing rules loader) — static
  singleton with `ensure_loaded()`.
- `client/tests/test_coop_client.gd` or new `client/tests/test_character_select_panel.gd`.
- `tools/bot/scenarios/client/20_menu_create_join_flow.json` — assert feature summary.
- Docs: plan, as-built, lifecycle on finish.

## Test and bot proof

```bash
make validate-shared
make client-unit
make bot-client SCENARIO=20_menu_create_join_flow HEADLESS=1
```

Optional visual: `make bot-visual scenario=20_menu_create_join_flow`

## Asset decision

- **Adopt:** existing class icons (`class_icon.gd`, `class_presentations.v0.json`).
- **Borrow:** skill display names from `skills.v0.json` / text catalog keys.
- **Reject:** new bitmap art, external UI plugins, 3D class previews.

## Open questions and risks

| # | Question | Default |
|---|----------|---------|
| Q-1 | Separate `class_creation.v0.json` vs extend `character_progression.v0.json`? | Extend `character_progression` with `creation_summary` per class listing skill id refs; avoids new manifest rollout. |
| Q-2 | Include skill one-line descriptions on create screen? | Names + kind tags only (e.g. "Rage — self buff"); defer long descriptions. |
