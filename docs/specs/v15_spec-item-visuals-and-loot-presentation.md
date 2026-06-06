# v15 Spec — Item visuals and loot presentation

## 1. Intent

Make current items readable in the Godot client without changing authoritative gameplay.
Ground loot and inventory slots should use shared presentation metadata instead of generic
colored cubes and text initials, while equipped weapons keep the ADR-0006 manifest-backed
mount path.

## 2. Baseline

Latest completed slice: v14 `godot-client-bot`.

Existing state:

- `rusty_sword` and `training_bow` have equipment GLB assets and `item_visuals` mappings.
- Non-equipped items (`training_badge`, `quest_leaf`, `red_potion`) have item rules but no
  distinct presentation metadata.
- Ground loot is rendered as a single primitive cube colored by category.
- Inventory slots render text initials from item names.
- Server protocol and replay already carry `item_def_id` on loot/inventory entities.

## 3. Non-goals

- No server gameplay changes, new intents, protocol schema bump, item stats, stack splitting,
  stash, vendors, crafting, consumable use, or character-scoped persistence.
- No production art pipeline, remote asset patcher, texture budget, or external asset import.
- No plugin adoption for inventory logic.

## 4. Shared presentation contract

Add `shared/assets/item_presentations.v0.json` plus schema. It is presentation-only data:

- keyed by `item_def_id`
- validates that every key resolves to `shared/rules/items.v0.json`
- contains no gameplay stats
- provides inventory icon shape/color/accent/label and ground loot shape/color/accent/scale

The data may be consumed by any client renderer. The Go server must not read it.

## 5. Godot rendering requirements

- Inventory panel renders each item with a shaped icon from item presentation metadata.
- Tooltips keep using item rules for names/stats.
- Ground loot renders using item presentation metadata and is visually distinguishable across:
  `rusty_sword`, `training_bow`, `training_badge`, `quest_leaf`, and `red_potion`.
- Missing presentation metadata falls back to the existing category-colored behavior.
- Equipped weapon GLB mounting remains unchanged.

## 6. Validation and tests

- `make validate-shared` validates the new schema/instance and cross-checks item keys.
- `make validate-assets` stays green.
- Godot item visual test checks every current item has presentation metadata.
- Client bot can assert presentation metadata is visible through `get_bot_state()` for inventory
  and loot items.
- `make ci` remains the final gate.

