# v290 As Built: Blacksmith Armor Recipe

Date: 2026-06-19
Spec: [`docs/specs/v290_spec-blacksmith-armor-recipe.md`](../specs/v290_spec-blacksmith-armor-recipe.md)
Plan: [`docs/plans/v290_2026-06-19-blacksmith-armor-recipe.md`](../plans/v290_2026-06-19-blacksmith-armor-recipe.md)

## What shipped

- Added `armor_reinforcement` / `Reinforce Armor` as a third blacksmith recipe.
- Server recipe validation now accepts the armor recipe and limits it to item templates with positive
  armor base stats in armor-bearing slots.
- Focused HTTP coverage proves the existing weapon recipe still rejects armor, the armor recipe
  rejects weapons, and the armor recipe upgrades armor stats through the existing item-level path.
- Extracted client recipe IDs, labels, eligibility text, item acceptance, and rejection copy into
  `blacksmith_recipes.gd`, shrinking `blacksmith_panel.gd` below the 600-line target.
- The blacksmith panel now shows `Reinforce Armor`, previews `Eligible: Armor pieces only`, enables
  armor pieces, and disables weapons for that recipe.
- Added client bot scenario `70_blacksmith_armor_recipe`, which selects the armor recipe, stages
  `cave_mail`, verifies the preview, upgrades once, and observes resource spend.

## Proof

Focused verification:

```bash
(cd server && go test ./internal/http -run Blacksmith -count=1)
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
make bot-client scenario=70_blacksmith_armor_recipe HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: deferred until the end of the selected autoloop queue.

## Manual visual command

```bash
make bot-visual scenario=70_blacksmith_armor_recipe
```

## Deferred

- Shared blacksmith recipe catalog, recipe unlocks, per-recipe costs, per-recipe success formulas,
  new upgrade resources, multi-resource recipes, durability, sockets, and item-level schema changes
  remain deferred.
- New blacksmith art, item icons, audio, external plugins, and production asset work remain deferred.
