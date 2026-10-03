# v532 — Item tooltip without icon preview

- **Status:** Complete (combined `make ci` passed 2026-10-03, 20m36s)
- **Date:** 2026-10-02
- **Codename:** `tooltip-no-icon`
- **Baseline:** `dab60eb5` (`main`, v531 closed)
- **Sequence:** Standalone run of an accepted six-slice set: v532 → v533 `camera-zoom-out` → v535 `wider-entrances` → v534 `light-radius-buff` → v536 `solid-dungeon-props` → v537 `door-wall-visual-proof`. This slice has no dependencies and no overlap with the others.
- **ADRs:** [ADR-0001](../adr/0001-technology-stack.md) D2 (client is display only; nothing here touches authority). [v515 spec](v515_spec-inventory-tooltip-redesign.md) is the prior owner of the tooltip layout.

## Purpose

`ItemTooltipPanel` draws a 96×96 `ItemPreview` beside the item name using the legacy 2D `ItemIconDrawer`. It is stale next to the item art the slots already show, and the player already sees the item in the bag, stash, shop, market or wallet slot they are hovering. Remove the preview so the tooltip is text only and the text column uses the full tooltip width.

## Scope

1. Delete the `ItemPreview` inner class and its instantiation in `ItemTooltipPanel.setup`.
2. Drop the now-dead constants and imports in `client/scripts/item_tooltip_panel.gd`: `PREVIEW_SIZE`, `PREVIEW_GAP`, `MAIN_STAT_WIDTH`, `ICON_FONT_SIZE`, `ItemIconDrawerScript`, and `RarityCuePresenterScript` if nothing else in the file uses it.
3. Main stat lines use `CONTENT_WIDTH` (360) in all cases. Remove the `top_row` `HBoxContainer` if it holds only `main_stats`; otherwise keep the layout minimal. An empty-item tooltip must render the same as before.
4. Remove the `item_presentations` and `fallback_label` parameters of `ItemTooltipPanel.setup` if the plan confirms the caller edits are mechanical (about ten call sites in `inventory_panel.gd`, `stash_panel.gd`, `shop_panel.gd`, `market_panel.gd`, `material_wallet_panel.gd`, and tests); otherwise keep the signature, rename them with a leading underscore, and record the cleanup as a follow-up. Do not leave live-looking dead parameters undocumented.

## Non-goals

- No change to inventory, bag, stash, shop, market, wallet, paper-doll, hotbar, or consumable-bar slot icons, to `item_icon_drawer.gd` itself (other panels use it), or to `item_family_icon_preview.gd`.
- No change to tooltip text content, ordering, colors, rarity border, requirement, comparison, price, or level footer lines.
- No protocol, server, shared rules, golden, or asset manifest change. No new gameplay tuning.
- No model-thumbnail or 3D preview replacement. The decision is removal, not substitution.

## Observable acceptance criteria

1. An item tooltip built through `ItemTooltipPanel.setup` for a common, magic, rare, unique and set item contains no `ItemPreview` node and no `Control` that draws an icon. The first label is the item name at the same font size as before.
2. The text column is `CONTENT_WIDTH` wide with no reserved empty strip on the right; long lines wrap at the full width, so tooltip height does not grow versus the previous layout for the same item (it may shrink).
3. Rarity remains readable without hue: the `Rarity: X` text line, rarity name color, and rarity-weighted border and rule are unchanged. The slot-shape cue (v508) the preview used to draw is intentionally not reproduced in the tooltip; the slot the player hovers still shows it. Record this in the as-built.
4. Every existing tooltip consumer still builds: inventory, equipped comparison, stash, vendor, mystery offer, market listing and staged offer, material wallet. `debug_gold_value_text`, `debug_item_level_text`, `debug_requirement_texts`, `debug_border_width`, `debug_first_main_line_color`, `debug_main_line_font_sizes` and `border_width_for_rarity` keep their contracts.
5. `item_tooltip_panel.gd` ends smaller than its current 391 lines, with no unused constant or import, and `make maintainability` stays green.
6. A real-renderer capture of one tooltip per rarity shows the text-only layout with no gap, clipped text, or overlapping lines. Compare against a capture from `dab60eb5`.

## Surfaces

| Area | Files |
|---|---|
| Client | `client/scripts/item_tooltip_panel.gd`; call sites in `inventory_panel.gd`, `stash_panel.gd`, `shop_panel.gd`, `market_panel.gd`, `material_wallet_panel.gd` only if the signature is pruned |
| Client tests | `test_inventory_tooltip_content.gd`, `test_shop_panel.gd`, `test_shop_tooltip_stability.gd`, `test_stash_panel.gd`, `test_character_stats_panel.gd`, `test_look_and_feel_polish.gd`; add a focused assertion that no preview node exists and the text column spans `CONTENT_WIDTH` |
| Docs | `docs/as-built/v532_tooltip-no-icon.md`, lifecycle row, `docs/CODEMAP.md` only if a file-role description changes |

## Adopt / borrow / reject

- **Adopt:** the existing `ItemTooltipPanel`, `InventoryPanelStyles` rarity helpers and `showme` / `make regen-screenshots` harness.
- **Borrow:** the v515 tooltip hierarchy, unchanged.
- **Reject:** a replacement 3D or model-thumbnail preview (duplicates the slot), external UI plugins, and any new asset.

## Focused verification (no `make ci`)

- Godot headless runs of the tests listed above through their `res://tests/...` scripts (as wired in `scripts/client_smoke.sh`), or `make client-unit` if a single run is cheaper than the list.
- A new or extended headless assertion for criteria 1 and 2.
- `make maintainability` and `git diff --check`.
- A tooltip screenshot per rarity before and after, using the existing rarity fixtures (`showme_rarity_ui_fixtures.gd`) or a `showme` suite the plan names. Godot shutdown leak warnings are reported, not hidden.
- `make validate-shared` is not required: no shared file changes.

## Integration risks

- **Hidden consumers of the preview geometry.** Tests or the tooltip-stability code may assume the old width or a top row; the plan greps for `top_row`, `MAIN_STAT_WIDTH` and the preview before editing.
- **Positional `setup` arguments.** Pruning the parameters shifts later positional arguments (`price`, `affordable`, `affinity_lines`) at every call site. Prune only with a full call-site sweep, otherwise keep the signature.
- **Lost non-hue rarity cue in the tooltip** (criterion 3) is a conscious tradeoff to confirm at spec review.

## Open questions

1. Prune the `setup` signature now (scope item 4) or keep it and note the cleanup? Default: prune if the plan counts at most 12 mechanical edits, else keep with underscored names.
