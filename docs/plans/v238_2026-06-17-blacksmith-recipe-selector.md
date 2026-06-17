# v238 Plan - Blacksmith Recipe Selector

Status: Complete
Goal: Add an explicit active-recipe selector to the existing blacksmith panel.
Architecture: Keep the existing item-upgrade recipe as the only option; expose selected recipe
metadata client-side and reuse current upgrade execution.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v118 blacksmith upgrade UI, v221 wallet-backed upgrade resources, and v222 upgrade result
preview. Asset/plugin decision: reject external assets/plugins; this is an `OptionButton` selector
inside the existing blacksmith panel.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/blacksmith_panel.gd` | Render recipe selector and expose selected recipe state |
| Modify | `client/tests/test_blacksmith_panel.gd` | Prove selector metadata and preview text |
| Add | `tools/bot/scenarios/client/55_blacksmith_recipe_selector.json` | Client proof |
| Add | `docs/as-built/v238_blacksmith-recipe-selector.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] None expected; focused blacksmith files are below 600 lines.

Decision:
- [x] Reuse existing blacksmith bot preview assertions rather than adding bot runner fields.
- [x] Keep the selector single-option and non-semantic until additional recipes exist.

Verification:
```bash
make maintainability
```

## Task 1 - Selector UI/debug state

Files:
- Modify: `client/scripts/blacksmith_panel.gd`

- [x] Add a recipe selector control with `Upgrade Item` selected.
- [x] Expose `recipe_selector_visible`, `selected_recipe_id`, `selected_recipe_label`, and
  `recipe_options` in debug state.
- [x] Add `Recipe: Upgrade Item` to staged preview lines.
- [x] Preserve existing upgrade emission, resource checks, and gold checks.

## Task 2 - Focused tests and bot proof

Files:
- Modify: `client/tests/test_blacksmith_panel.gd`
- Add: `tools/bot/scenarios/client/55_blacksmith_recipe_selector.json`

- [x] Assert the selector/debug metadata in the focused blacksmith unit.
- [x] Assert existing preview lines still include cost/resource/pity information.
- [x] Add a bot scenario that stages an item and verifies `Recipe: Upgrade Item` via
  `preview_contains`.

```bash
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
make bot-client scenario=55_blacksmith_recipe_selector.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Modify: `docs/progress/slice-codename-index.md`
- Add: `docs/as-built/v238_blacksmith-recipe-selector.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=55_blacksmith_recipe_selector.json HEADLESS=1`
- [x] `make maintainability`
