# v134 Spec: Unique Inspection UI

Status: Complete
Date: 2026-06-13
Codename: `unique-inspection-ui`

## Purpose

Show readable unique-effect descriptions at the bottom of item tooltips for items that carry unique
effect ids. Players should be able to inspect a unique item in inventory, stash, unique chest, or
market contexts and understand what the effect does without knowing the raw catalog id.

## Non-goals

- No new unique items, unique effects, drop rules, or balance tuning.
- No protocol or server payload schema changes.
- No new item art, tooltip layout rewrite, or replacement of the shared tooltip panel.
- No persistence, replay, combat, or bot behavior changes.

## Acceptance Criteria

- The Godot client loads `shared/rules/unique_effects.v0.json` through the shared item rule loader.
- Tooltip builders can resolve each item `effect_ids` entry to its effect display name and summary.
- Inventory tooltips append unique-effect text after the existing stat, requirement, and comparison
  sections so the effect description appears at the bottom of the tooltip.
- Stash, unique chest, and market item tooltips include the same readable unique-effect text for
  items with `effect_ids`.
- Items without unique effects keep their current tooltip content.
- Client unit tests cover a unique item with `everburning_wound` and prove the readable summary is
  present at the bottom.

## Likely Files

- `client/scripts/item_rules_loader.gd`
- `client/scripts/unique_effect_tooltip.gd`
- `client/scripts/inventory_panel.gd`
- `client/scripts/stash_panel.gd`
- `client/scripts/market_panel.gd`
- `client/tests/test_shop_panel.gd`
- `client/tests/test_stash_panel.gd`
- `PROGRESS.md`
- `docs/as-built/v134_unique-inspection-ui.md`

## Test And Bot Proof

- `make client-unit`
- `make maintainability`
- `make ci`

No dedicated visual bot scenario is required for this small tooltip text change. Manual visual
verification can use any inventory/stash context that contains a unique item with `effect_ids`.

## Open Questions And Risks

- Server summaries may still include raw `Effect: <id>` metadata. This slice adds the readable
  bottom description without changing server summary payloads.
