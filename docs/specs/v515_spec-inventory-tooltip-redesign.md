# v515 — Inventory, Slot, Rarity Border, and Item Tooltip Redesign

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Batch baseline:** `425b9ae4` (v510 closed)
- **Dependencies:** v514 `ui-theme-foundation` (`UiTheme` singleton + `shared/assets/ui_theme.v0.json`, specced in a sibling worktree, not yet available). Sibling overlap: v516 (character screen / paper doll) and v517.
- **Architecture:** ADR-0001 D2 (server owns items; display only), ADR-0018 D9 (real-renderer screenshot gate), v508 (color-safe rarity cues, must be preserved).

## Purpose

Inventory slots, rarity borders, and the item tooltip were built incrementally (v449-v456 growth, v508 cues) and read as a stack of special cases: slot borders are a lightened rarity background, the tooltip repeats rarity as a text line, and the title/meta/stat/requirement/comparison/footer blocks have weak hierarchy. `client/scripts/inventory_panel.gd` is 1,626 lines (grandfathered at 1,623) and mixes window/layout, slot rendering, drag-and-drop, and roughly 450 lines of pure tooltip-content logic. Redesign the slot look, rarity border treatment, and tooltip layout/hierarchy, and split the tooltip-content logic out of the panel.

## Scope

1. **Slot redesign.** One coherent slot visual for bag, equipment, and paper-doll slots: rarity border (width/inset), inner bevel/background, hover, selected-by-drag, invalid-requirement, two-handed-blocked, and empty states. The hotbar badge and the v508 rarity-cue badge keep their positions and legibility.
2. **Rarity borders.** Color is supplementary: the v508 shape+letter cue stays the primary non-hue signal. Border weight may also vary by rarity (a third redundant cue, not a replacement). Border colors come from one source; today they are duplicated in `InventoryPanelStyles.ITEM_RARITY_BACKGROUNDS`, `InventoryPanel._rarity_color`, and `ClientConstants.LOOT_LABEL_RARITY_COLORS`. Consolidate to one rarity-visual lookup (see D3).
3. **Tooltip hierarchy.** Rework `ItemTooltipPanel` / the line model so a reader sees, in order: name (rarity-colored) with rarity word and item-level; slot/type and weapon metadata; base stats; rolled/affix stats; set and unique-effect blocks; requirements (met/unmet); equipped comparison deltas; hotbar assignment; price footer. Section separators and spacing are consistent; the redundant `Rarity: X` line is folded into the header (the full rarity word must remain present in tooltip text, per v508 AC3).
4. **Real extraction (ratchet).** Move tooltip content building out of `inventory_panel.gd` into a new independently importable module (D1). Lower the `inventory_panel.gd` baseline in the same slice.
5. **Theme adoption (blocked on v514).** Colors, borders, spacing, and font sizes for slots/tooltip are read from `UiTheme` instead of hard-coded constants, with a safe fallback only if v514 explicitly provides one.

## Non-goals

- No protocol, schema, server, replay, rules, golden, item-mutation, drag/drop routing, equip, transfer, or price changes. Slot kinds, drag payloads, `intent_requested` payloads, and `InventoryTransferRouter` behavior are untouched (its logic stays; only its call sites are unaffected).
- No new rarities, rarity cue shapes, or changes to `shared/assets/rarity_cues.v0.json` content (v508 owns it).
- No paper-doll layout/preview redesign (v516), no skill/shop/stash/market panel redesign beyond what falls out of the shared `ItemTooltipPanel`/slot style (they must keep working and visually not regress).
- No new fonts, third-party UI themes, or assets. Adopt: existing in-repo `StyleBoxFlat`, `ItemIconDrawer`, `RarityCuePresenter`, `showme`/`regen-screenshots` harness. Borrow: v508 catalog pattern, v514 `UiTheme`. Reject: external UI/theme plugins, runtime downloads.

## Design decisions (spec-level)

- **D1 Extraction target.** New `client/scripts/inventory_tooltip_content.gd` (`class_name InventoryTooltipContent extends RefCounted`) with static functions plus a small typed `InventoryTooltipContext` (RefCounted: `item_rules`, `item_templates`, `item_presentations`, `equipped`, `hotbar`, `hotbar_capacity`, `character_class`, `viewed_weapon_set`, resource/shop price inputs). It owns what is now `_tooltip`, `_tooltip_lines`, `_metadata/_set_*/_compact_metadata_lines`, `_requirement_lines*`, `_comparison_*`, `_equip_preview_*`, `_detail_lines`, `_hotbar_*_tooltip` helpers, `_base/_random_stat_lines`, `_format_*`, `_display_stat`, and `_rarity_color`. It must not import `inventory_panel.gd` and must not use `globals()`-style laundering; the panel builds the context from its own fields and calls the module. Price helpers (`_item_gold_value`, `_item_buy_price`, `_generated_buy_price`, `_town_vendor_*`) move in a second extraction (`inventory_item_pricing.gd`) if the first does not reach the target. Expected result: `inventory_panel.gd` ≤ ~1,100 lines; baseline lowered to actual size (still grandfathered; further slot/drag extraction is a documented follow-up, not required here).
- **D2 Slot styling ownership.** `inventory_panel_styles.gd` stays the single slot/panel style factory; `InventorySlotButton` drawing stays in the panel unless the plan finds a clean seam. v516 also consumes these styles; this slice owns the slot/border API and v516 must consume it rather than fork it.
- **D3 One rarity visual lookup.** Rarity name -> {text color, border color, border width, background tint} lives in `UiTheme` (v514) once available; until then, one helper in `inventory_panel_styles.gd` replaces the three duplicates. Existing color values are retained unless the real-renderer contrast review (below) requires adjustment; any adjustment is recorded with measured ratios.
- **D4 Data-driven.** Presentation values are UI theme data (v514 owns `shared/assets/ui_theme.v0.json`); no gameplay tuning is added. If v515 needs new keys it adds them in a schema-backed edit coordinated with v514, not a hard-coded constant.

## Observable acceptance criteria

1. Bag, equipment, and paper-doll slots render the new design in all states (empty, item per rarity x5, hover, drag-over, invalid requirement, two-handed blocked, with hotbar badge, with v508 cue). Icon, cue, hotbar badge, and warning remain legible at actual slot sizes (`EQUIPMENT_SLOT_SIZE` 96x58 and bag cells).
2. Each rarity is distinguishable in a grayscale capture without color (shape+letter cue plus border treatment); border colors for dark/light slot backgrounds meet a recorded contrast ratio (>= 3:1 border vs adjacent background, measured and recorded; adjustments allowed only with evidence).
3. The tooltip shows the hierarchy in Scope 3 for: a common weapon, a magic armor, a rare with affixes, a unique with effects, a set piece (membership + bonus lines), a consumable/potion, an unmet-requirement item, an item with equipped comparison deltas, and a hotbar-assigned item. The rarity word, item level, price, requirement met/unmet states, and comparison deltas are all still present; no information shown today is dropped.
4. The tooltip text exposed through `slot.tooltip_text` and debug accessors (`debug_gold_value_text`, `debug_item_level_text`, `debug_requirement_texts`, `debug_border_width`, `debug_first_main_line_color`, `debug_main_line_font_sizes`) keeps working or is updated together with every test and bot scenario that reads them.
5. Shop, stash, market, blacksmith, and wallet tooltips that reuse `ItemTooltipPanel` still render without errors and without layout overflow at the standard tooltip width.
6. The extracted module(s) are importable and unit-testable without importing `inventory_panel.gd` (a focused Godot test preloads the module directly and covers lines/requirements/comparison/set/hotbar for several items). `make maintainability` passes, `inventory_panel.gd`'s baseline in `.maintainability/file-size-baseline.tsv` is lowered to its new size, and no new `helpers=globals()` site is added.
7. No server, protocol, shared-rules, golden, or replay file changes; item mutation intents sent from the inventory are unchanged (existing client scenarios prove it).
8. Real-renderer UI captures (before/after, same fixtures, grayscale pass, recorded renderer/quality tier) are attached to the as-built note.

## Likely surfaces

| Area | Paths |
|------|-------|
| Client | `client/scripts/inventory_panel.gd` (shrink), `inventory_panel_styles.gd`, `item_tooltip_panel.gd`, `item_tooltip_stat_sections.gd`, new `inventory_tooltip_content.gd` (+ optional `inventory_item_pricing.gd`), `rarity_cue_presenter.gd` (only if cue placement changes), `inventory_transfer_router.gd` (no behavior change; touch only if a constant moves) |
| Consumers to regression-check | `shop_panel.gd`, `stash_panel.gd`, `market_panel.gd`, `material_wallet_panel.gd`, `blacksmith_*`, `character_stats_panel.gd`, `tests/test_look_and_feel_polish.gd`, `test_shop_panel.gd`, `test_shop_tooltip_stability.gd`, `test_stash_panel.gd` |
| Tests | new `client/tests/test_inventory_tooltip_content.gd`; extend `test_inventory_panel.gd`, `test_rarity_cues.gd`; register new tests in `scripts/client_smoke.sh` |
| Shared (via v514) | `shared/assets/ui_theme.v0.json` + schema (owned by v514; v515 adds keys only by coordination) |
| Proof | `skills/showme` `inventory`, `shop`, `corpse-inventory` focuses; `make regen-screenshots` (register an inventory/tooltip suite if none exists; check `make regen-screenshots-list` first); `docs/as-built/v515_inventory-tooltip-redesign.md` |
| Bot | client scenarios `02_inventory_open_close`, `03_inventory_equip_unequip`, `04_inventory_lab_drop_item`, `13_inventory_paper_doll`, `v508_rarity_cue_preset_camera`, plus existing stash/vendor/market client scenarios as flow-preservation checks |

## Verification (focused only)

- `make validate-shared` (only if v514 keys are added); `make client-unit`; `make maintainability`; `make test-py` subset only if a validator is touched.
- Client bot: `make bot-client SCENARIO=<name> HEADLESS=1` for the five inventory scenarios above and the affected stash/vendor/market ones.
- `skills/showme` inventory/shop/corpse-inventory focuses and `make regen-screenshots` suite for before/after; headless bot passes are not readability proof.
- No per-slice `make ci`; the coordinator runs the combined gate.

## Integration risks

- **v514 not available:** theme-driven styling is blocked; extraction and tooltip hierarchy against current constants can proceed first, with a mechanical theme-swap task after v514 lands.
- **v516 shares `inventory_panel_styles.gd` and the paper-doll slot code in `inventory_panel.gd`** (`PAPER_DOLL_SLOT_POSITIONS`, `_debug_paper_doll_slots`, `_equipment_slots`). v516 will also shrink/edit `inventory_panel.gd`; line moves conflict heavily. Integrate v515's extraction first or agree ownership of the paper-doll block.
- **v517** (unknown scope): coordinator to confirm whether it touches tooltips/panels or shared theme JSON.
- Many consumers reuse `ItemTooltipPanel`; layout changes risk shop/stash/market tooltip overflow (AC5).
- Tests/bot scenarios reading tooltip text or debug accessors may be string-sensitive (AC4).

## Open questions

1. Should border weight vary by rarity (extra redundant cue) or stay uniform and rely on the v508 badge? (Spec assumes weight varies modestly; confirm.)
2. Is rarity restyling in shop/stash/market slots in scope when they share style helpers, or must their visuals stay byte-identical until a later slice? (Spec assumes shared `ItemTooltipPanel` changes apply everywhere; slot styles for other panels change only if they already call `InventoryPanelStyles`.)
3. Target size for `inventory_panel.gd`: is ~1,100 lines (grandfathered, lowered) acceptable, or must this slice also split slot/drag handling to reach a lower figure?
4. Does v514 expose a rarity-visual section, or does v515 add it (D3/D4 ownership)?
