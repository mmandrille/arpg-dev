class_name InventoryPanelStyles
extends RefCounted
## Inventory-family panel styles. All values come from UiTheme (shared/assets/ui_theme.v0.json).
## Rarity text/border colors and border weights use the `inventory_*` theme tokens (v515).
## Color is supplementary: the v508 shape+letter cue stays the primary non-hue signal and
## border weight is a third redundant cue.

const RARITIES := ["common", "magic", "rare", "unique", "set"]
const HOVER_BORDER_LIGHTEN := 0.18
const MIN_INVALID_BORDER_WIDTH := 2

const PAPER_DOLL_CARD_TINT := 0.35
const PAPER_DOLL_CARD_ACCENT_DARKEN := 0.7
const PAPER_DOLL_CARD_BORDER_DARKEN := 0.45
const PAPER_DOLL_CONNECTOR_ALPHA := 0.32


static func panel_style() -> StyleBoxFlat:
	return UiTheme.frame("panel")


static func slot_style(hover: bool) -> StyleBoxFlat:
	return UiTheme.frame("slot", "hover" if hover else "")


## Unknown or empty rarity resolves to common, never to a rarer tier.
static func _rarity_key(rarity: String) -> String:
	var key := rarity.to_lower()
	return key if key in RARITIES else "common"


static func rarity_color(rarity: String) -> Color:
	return UiTheme.color("inventory_rarity_" + _rarity_key(rarity))


static func rarity_border_width(rarity: String) -> int:
	return UiTheme.spacing("inventory_border_" + _rarity_key(rarity))


static func rarity_border_color(rarity: String, hover: bool) -> Color:
	var key := _rarity_key(rarity)
	var base := UiTheme.color("inventory_border_common") if key == "common" else rarity_color(key)
	return base.lightened(HOVER_BORDER_LIGHTEN) if hover else base


static func item_slot_style(rarity: String, hover: bool, invalid_requirements: bool = false) -> StyleBoxFlat:
	var s := UiTheme.item_slot_frame(rarity, hover, invalid_requirements)
	var width := rarity_border_width(rarity)
	if invalid_requirements:
		width = maxi(width, MIN_INVALID_BORDER_WIDTH)
	else:
		s.border_color = rarity_border_color(rarity, hover)
	s.border_width_left = width
	s.border_width_top = width
	s.border_width_right = width
	s.border_width_bottom = width
	return s


static func empty_slot_style(hover: bool) -> StyleBoxFlat:
	return UiTheme.frame("empty_slot", "hover" if hover else "")


static func blocked_slot_style(hover: bool) -> StyleBoxFlat:
	return UiTheme.frame("blocked_slot", "hover" if hover else "")


static func paper_doll_style() -> StyleBoxFlat:
	return UiTheme.frame("paper_doll")


## Class-tinted card behind the paper-doll body (v516).
static func paper_doll_card_style(accent: Color) -> StyleBoxFlat:
	var s := UiTheme.frame("character_paper_doll_card")
	s.bg_color = UiTheme.color("paper_doll_bg").lerp(accent.darkened(PAPER_DOLL_CARD_ACCENT_DARKEN), PAPER_DOLL_CARD_TINT)
	s.border_color = accent.darkened(PAPER_DOLL_CARD_BORDER_DARKEN)
	return s


static func paper_doll_neutral_accent() -> Color:
	return UiTheme.color("character_paper_doll_accent_neutral")


static func paper_doll_connector_color(accent: Color) -> Color:
	return Color(accent.r, accent.g, accent.b, PAPER_DOLL_CONNECTOR_ALPHA)
