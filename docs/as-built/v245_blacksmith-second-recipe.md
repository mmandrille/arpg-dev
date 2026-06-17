# v245 As-Built - Blacksmith Second Recipe

Date: 2026-06-17

## What shipped

- Added `Hone Weapon` as the second blacksmith recipe selector option beside `Upgrade Item`.
- Sent the selected `recipe_id` through stash and inventory upgrade HTTP requests.
- Kept missing/empty `recipe_id` compatible with the default `item_upgrade` recipe.
- Added server-owned validation for `weapon_honing`: only weapon templates with damage stats are
  eligible, unknown recipe IDs return bad request, and non-weapons return conflict.
- Updated blacksmith preview/debug state so selected recipe label and eligibility text are visible.
- Disabled non-weapon items for `Hone Weapon` in the client and kept the default upgrade flow green.
- Added `select_blacksmith_recipe` to the client bot and `62_blacksmith_second_recipe.json` as the
  end-to-end proof.

## Proof

```bash
cd server && go test ./internal/http -run Upgrade -count=1
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=62_blacksmith_second_recipe.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` remains deferred until the feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=62_blacksmith_second_recipe.json
```

## Scope limits

- No new material types, cost tuning fields, success formulas, durability, recipe unlocks,
  production icons, external assets, or external plugins shipped.
- `Hone Weapon` reuses the existing upgrade cost/resource/chance tuning and narrows eligibility
  only; broader crafting categories and multi-resource recipes remain deferred.
