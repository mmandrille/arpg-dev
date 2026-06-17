# v238 As-Built - Blacksmith Recipe Selector

Date: 2026-06-17

## What shipped

- Added an explicit recipe selector to the blacksmith panel with the current configured `Upgrade
  Item` recipe selected.
- Exposed `recipe_selector_visible`, `selected_recipe_id`, `selected_recipe_label`, and
  `recipe_options` in blacksmith debug state.
- Added `Recipe: Upgrade Item` to staged preview lines while preserving existing gold/resource,
  success/failure, and pity preview lines.
- Increased the blacksmith panel height to fit the selector without crowding the staged item and
  preview area.
- Added `55_blacksmith_recipe_selector.json`, which stages an item and verifies the active recipe
  appears in the blacksmith preview.

## Proof

```bash
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=55_blacksmith_recipe_selector.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=55_blacksmith_recipe_selector.json
```

## Scope limits

- No new recipes, server/protocol changes, recipe persistence, crafting categories, material tuning,
  cost formulas, success formulas, item upgrade semantics, icons, or external assets shipped.
- The selector intentionally has one option until a future slice adds more blacksmith recipes.
