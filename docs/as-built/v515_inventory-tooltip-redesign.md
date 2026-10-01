# v515 — Inventory, Slot, Rarity Border, and Item Tooltip Redesign (as built)

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)

- **Base:** `425b9ae4` + v514 UiTheme (transferred by the coordinator). Presentation only; no protocol, server, rules, golden, replay or item-mutation change.
- **Spec / plan:** [spec](../specs/v515_spec-inventory-tooltip-redesign.md), [plan](../plans/v515_2026-10-01-inventory-tooltip-redesign.md)

## What shipped

- **Real extraction.** `client/scripts/inventory_tooltip_content.gd` (`InventoryTooltipContent`, static, nested typed `Context` carrying only `hotbar`) holds tooltip lines/text, requirement, comparison, set, hotbar and rarity-color logic. `client/scripts/inventory_item_pricing.gd` (`InventoryItemPricing`) holds the display-only price helpers. Neither imports `inventory_panel.gd`; no `globals()` helper injection (ratchet reports 0). `inventory_panel.gd` 1,626 -> 1,119 lines; baseline lowered 1623 -> 1119. Four dead helpers were deleted.
- **Slots/rarity borders.** `InventoryPanelStyles` stays a UiTheme facade and adds `rarity_color`, `rarity_border_width`, `rarity_border_color`, and a theme-backed `item_slot_style`. Theme additions in `shared/assets/ui_theme.v0.json` (additive, `inventory_*` prefix): colors `inventory_rarity_{common,magic,rare,unique,set}` and `inventory_border_common`; spacing `inventory_slot_radius` (2) and `inventory_border_{rarity}` = common 1, magic 2, rare 2, unique 3, set 3; `slot`/`empty_slot`/`blocked_slot` frames gain the 2px radius. Unknown rarity resolves to common. Slot backgrounds are unchanged; v508 cues are untouched. The invalid-requirement border stays red (min 2px).
- **Tooltip hierarchy** (`item_tooltip_panel.gd`, shared by inventory/shop/stash/market/blacksmith): larger name, rarity-colored `Rarity:` line (text kept), rarity-weighted header rule, separator rules before Requirements / Class affinity, rarity-bordered preview with the v508 cue, border from the single lookup. Tooltip line text and Label names used by debug accessors are unchanged.
- **Validator.** `tools/validate_ui_theme.py` also requires the `inventory_rarity_*`/`inventory_border_*` tokens for every gameplay rarity (+ test).
- **Loot label colors.** `ClientConstants.LOOT_LABEL_RARITY_COLORS` is a const and cannot call UiTheme, so it stays; a test asserts it equals the theme rarity colors.

## Evidence

- Contrast (border vs slot background, WCAG ratio, theme colors): common 3.47 normal / 3.14 hover; magic 6.94 / 5.38; rare 6.32 / 4.82; unique 6.38 / 4.98; set 7.27 / 5.46. All >= 3:1; asserted in `test_inventory_tooltip_content.gd`.
- Tests: `test_inventory_tooltip_content` (59), `test_ui_theme` (74), `test_inventory_panel` (17), `test_shop_panel` (206), `test_stash_panel` (121), `test_look_and_feel_polish`, `test_rarity_cues` (54), `test_character_stats_panel` (54), `test_market_item_comparison`, `test_shop_tooltip_stability`, `test_blacksmith_panel` (43), `test_inventory_transfer_router` (46) all pass. `validate_shared.py` 2244 checks OK; `pytest tools/test_validate_ui_theme.py tools/test_validate_rarity_cues.py` 9 passed; `make maintainability` passes.
- Client bot (headless) 1/1 each: inventory_open_close, inventory_equip_unequip, inventory_lab_drop_item, client_inventory_paper_doll, v508_rarity_cue_preset_camera, town_vendor_shop_panel, account_stash_panel, market_board_ui, blacksmith_upgrade_ui.
- Real-renderer captures (`render_focus.py`, windowed Godot 4.7.2, Metal Forward+, Apple M4 Pro): [`assets/v515/`](assets/v515/) `inventory|corpse-inventory|shop` `-before.png` / `-after.png` and `-gray.png` variants. "Before" images are the v514 captures, which v514 verified byte-identical to base `425b9ae4`. In grayscale the five rarities remain distinguishable by cue shape+letter and by border weight; the tooltip rarity word stays legible as text.

## Limits

- Shop, stash, market, blacksmith and wallet panels keep private slot-style and tooltip-line builders (including their own `Rarity:` lines); their slot borders are unchanged and only the shared `ItemTooltipPanel` look changed for them. Consolidating these is follow-up work.
- `inventory_panel.gd` is still grandfathered (1,119 > 600); slot/drag/paper-doll handling remains in it (v516 territory).
- No frame-time or draw-call measurement was taken (UI-only, no per-frame work added). No crowded/small-slot or hover/blocked/invalid capture matrix beyond the three focuses above.
- Theme `fonts` roles remain unused by this slice; tooltip font sizes are still local constants.
