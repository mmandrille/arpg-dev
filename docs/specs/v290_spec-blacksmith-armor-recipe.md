# v290 Spec: Blacksmith Armor Recipe

Status: Implemented
Date: 2026-06-19
Codename: `blacksmith-armor-recipe`

## Purpose

Add a third blacksmith recipe that lets players reinforce armor pieces through the existing
server-authoritative upgrade flow. This slice expands the recipe selector from generic item upgrade
and weapon honing to an armor-focused option while preserving the current cost, resource, success,
pity, and item-level mechanics.

## Non-goals

- Do not add new upgrade resources, success formulas, cost tuning, per-recipe costs, recipe unlocks,
  item-level schema changes, durability, sockets, affix grammar, or multi-resource recipes.
- Do not add a shared blacksmith recipe catalog yet; this slice follows the existing hardcoded
  recipe-ID pattern and extracts client recipe presentation/eligibility into a focused helper.
- Do not change the underlying upgrade mutation algorithm beyond recipe eligibility.
- Do not add new item art, blacksmith art, icons, external plugins, audio, or new asset pipelines.

## Acceptance Criteria

- The server accepts `recipe_id=armor_reinforcement` on stash-item and inventory-item blacksmith
  upgrade requests.
- `armor_reinforcement` is eligible only for authored armor templates with a positive armor base stat
  in armor-bearing slots such as shields, helms, chest armor, gloves, belts, and boots.
- The armor recipe rejects weapons and unknown recipe ids with the existing error behavior.
- Successful armor reinforcement spends the existing configured gold/resource costs, increments item
  level through the existing upgrade path, and improves the armor-bearing rolled stats the same way
  other upgrade recipes do.
- The Godot blacksmith recipe selector shows `Reinforce Armor` as the third option and reports
  `Eligible: Armor pieces only` in the preview.
- Client-side recipe gating enables armor pieces such as `cave_mail` and disables weapons for
  `Reinforce Armor`.
- Client bot proof selects the armor recipe, stages an armor item, verifies the recipe-specific
  preview, upgrades once, and observes the upgraded armor item.

## Scope And Likely Files

- Server: `server/internal/http/account_stash.go`, `server/internal/http/blacksmith_recipe_test.go`.
- Client: `client/scripts/blacksmith_panel.gd`,
  `client/scripts/blacksmith_recipes.gd`, `client/tests/test_blacksmith_panel.gd`.
- Bot: `tools/bot/scenarios/client/70_blacksmith_armor_recipe.json`.
- Docs: v290 plan/as-built/lifecycle updates.

## Test And Bot Proof

Focused checks:

```bash
(cd server && go test ./internal/http -run Blacksmith -count=1)
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
make bot-client scenario=70_blacksmith_armor_recipe HEADLESS=1
make maintainability
```

Visual verification command for humans/agents:

```bash
make bot-visual scenario=70_blacksmith_armor_recipe
```

## Asset And Plugin Decision

- Adopt: existing blacksmith panel, recipe selector, upgrade preview, upgrade history, item icon
  presentation, vendor-lab loot, and inventory upgrade route.
- Borrow: existing weapon-honing test and client bot scenario structure.
- Reject: external assets/plugins, new model pipelines, new blacksmith art, new icons, or new audio.

## Outcome

- Implemented `armor_reinforcement` as a third blacksmith recipe, with server-authoritative
  eligibility and client-side preview gating for armor-bearing equipment.
- The blacksmith panel stayed under the file-size target by extracting recipe metadata and
  eligibility helpers into `blacksmith_recipes.gd`.
