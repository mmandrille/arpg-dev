# v514 — UI Theme Foundation (as built)

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)

- **Base:** `425b9ae4`. Presentation only; no protocol, server, rules, golden, bot or ADR-0014 change.
- **Spec / plan:** [spec](../specs/v514_spec-ui-theme-foundation.md), [plan](../plans/v514_2026-10-01-ui-theme-foundation.md)

## What shipped

- `shared/assets/ui_theme.v0.json` + `.schema.json`: `colors`, `rarity_slot_backgrounds`, `spacing`, `fonts` (roles; `family` is `default`), `frames` (recipes referencing color/spacing tokens, optional `states`), `state_modifiers` (item-slot tint factors and invalid-border tokens). Colors are `#rrggbb`, `#rrggbbaa` or `[r,g,b,a]` floats (floats keep the 0.92 panel alpha exact).
- `client/scripts/ui_theme.gd` (`class_name UiTheme`, static, `ensure_loaded()`/`invalidate()`, `load_from_path()` for tests). Unknown tokens log one `push_error` and return magenta / 0 / `{}`-equivalents; nothing returns null.
- `InventoryPanelStyles` and `CharacterPanelStyles` are now one-line delegates with unchanged signatures (91 -> 27 and 21 -> 7 lines).
- `tools/validate_ui_theme.py` (registered in `validate_shared.py`): token references resolve, rarity keys equal `item_templates.v0.json`, rarity backgrounds distinct. `tools/test_validate_ui_theme.py` covers green and broken catalogs.
- `client/tests/test_ui_theme.gd` (74 checks, registered in `scripts/client_smoke.sh`), expectations derived from the catalog.

## API for v515-v517

```
UiTheme.color(token) / spacing(token) / rarity_background(rarity)
UiTheme.frame(name, state = "") -> StyleBoxFlat            # fresh instance
UiTheme.item_slot_frame(rarity, hover = false, dimmed = false, base = "slot")
UiTheme.font_size(role) / font_color(role) / apply_font(control, role)
UiTheme.has_token(kind, name)  # kind: color|spacing|font|frame|rarity
UiTheme.catalog() -> deep-ish copy; UiTheme.invalidate(); UiTheme.load_from_path(abs_path)
```

Extension rules: add catalog entries under your own prefix (`inventory_*`, `character_*`, `hud_*`), one entry per line; schema additions must be additive and optional; do not rename or remove tokens without the coordinator. Frames currently support bg, border, border width, radius, margin and `hover`/`disabled`/`invalid`/`selected` color-state overrides (only `hover` is used today). Shadows/gradients are not implemented. `fonts` entries (`title`, `body`, `caption`, `value`) are reserved placeholders with new values (no current call site migrated); v515-v517 should tune them when they consume them.

## Evidence

- Legacy-vs-migrated style dump (41 style variants: every function x hover x 8 rarity spellings x invalid): identical on bg/border colors, border widths, radii and margins.
- Real-renderer captures (`render_focus.py`, windowed Godot) before/after for `inventory`, `character-menu`, `corpse-inventory`, `rarity-cues`, `shop`, `blacksmith`, `skills`, `hud`: all eight PNG pairs are **byte-identical** (`docs/as-built/assets/v514/*-before.png` / `*-after.png`; `rarity-cues-*.json` differ only by frame timing ms).
- Note: the first capture run persists panel positions into Godot user data, so a first-ever "before" differs from later runs. The final "before" images were retaken against the legacy styles files after that state existed.
- Unit/validator: `validate_shared.py` OK (2243 checks); `pytest tools/test_validate_ui_theme.py tools/test_validate_rarity_cues.py` 8 passed; `test_ui_theme`, `test_inventory_panel`, `test_shop_panel`, `test_rarity_cues`, `test_character_stats_panel` pass; `make maintainability` passes.
- Inventory-rebuild timing was not separately measured (styles are plain dictionary lookups plus one `StyleBoxFlat`); no frame-time change appears in the rarity-cues capture JSON.

## Limits / not done

- Only the two `*_styles.gd` files migrated. Private duplicates remain in `shop_panel.gd`, `stash_panel.gd`, `market_panel.gd`, `skills_panel.gd`, `consumable_bar.gd`, bars and tooltips (left to v515-v517).
- No Godot `Theme` resource. No font assets.
