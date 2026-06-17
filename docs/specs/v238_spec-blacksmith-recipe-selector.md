# v238 Spec - Blacksmith Recipe Selector

Status: Approved for autoloop
Date: 2026-06-17
Codename: blacksmith-recipe-selector

## Purpose

Make the blacksmith panel explicitly show which recipe is active before the player stages or upgrades
an item. The current game has one configured recipe, item upgrade, so this slice adds a selector
surface without changing upgrade behavior.

## Non-goals

- No new recipes, server/protocol changes, recipe persistence, crafting categories, material tuning,
  cost formulas, success formulas, or item upgrade semantics.
- No recipe icons, external assets, or standalone crafting window.

## Acceptance Criteria

- The blacksmith panel renders a recipe selector with the configured item-upgrade recipe selected.
- The selected recipe label and recipe option metadata are available in debug state.
- Staged item previews include the selected recipe label while keeping existing cost/resource/result
  preview lines intact.
- The existing upgrade request flow and resource/gold gates continue to behave the same.
- A focused blacksmith panel unit test proves selector visibility, selected recipe metadata, and
  preview text.
- A client bot scenario opens the blacksmith, stages an item, and verifies the active recipe label in
  the preview using existing blacksmith assertions.

## Scope and Likely Files

- Client: `client/scripts/blacksmith_panel.gd`.
- Unit tests: `client/tests/test_blacksmith_panel.gd`.
- Bot/scenario: `tools/bot/scenarios/client/55_blacksmith_recipe_selector.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- `make bot-client scenario=55_blacksmith_recipe_selector.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. The selector is intentionally single-option until future slices add more
  recipes.
- Risk: a selector with one option can look decorative. Debug state and preview text make the active
  recipe explicit, while behavior remains unchanged.
