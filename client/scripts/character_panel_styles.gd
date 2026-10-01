class_name CharacterPanelStyles
extends RefCounted
## Character-family panel styles. Colors, type scale and frame recipes come from UiTheme
## (shared/assets/ui_theme.v0.json, `character_*` entries). Only accent-derived shading
## factors (how far a class accent is darkened) live here: they are behavior, not tokens.

const CARD_BORDER_DARKEN := 0.55
const XP_FILL_DARKEN := 0.1
const CHIP_FILLED_DARKEN := 0.45
const ALLOCATE_DARKEN := 0.5
const ALLOCATE_HOVER_DARKEN := 0.3


static func panel_style() -> StyleBoxFlat:
	return UiTheme.frame("panel_character")


static func text() -> Color:
	return UiTheme.color("character_text")


static func text_dim() -> Color:
	return UiTheme.color("character_text_dim")


static func section() -> Color:
	return UiTheme.color("character_section")


static func neutral_accent() -> Color:
	return UiTheme.color("character_accent_neutral")


static func font_title() -> int:
	return UiTheme.font_size("character_title")


static func font_row() -> int:
	return UiTheme.font_size("character_row")


static func font_caption() -> int:
	return UiTheme.font_size("character_caption")


## Raised card behind the header and each stat group.
static func card_style(accent: Color = Color(0, 0, 0, 0)) -> StyleBoxFlat:
	var s := UiTheme.frame("character_card")
	if accent.a > 0.0:
		s.border_color = accent.darkened(CARD_BORDER_DARKEN)
	return s


static func xp_bar_background_style() -> StyleBoxFlat:
	return UiTheme.frame("character_xp_background")


static func xp_bar_fill_style(accent: Color) -> StyleBoxFlat:
	var s := UiTheme.frame("character_xp_fill")
	s.bg_color = accent.darkened(XP_FILL_DARKEN)
	return s


## Level chip and unspent-points badge.
static func chip_style(accent: Color, filled: bool) -> StyleBoxFlat:
	var s := UiTheme.frame("character_chip")
	s.border_color = accent
	if filled:
		s.bg_color = accent.darkened(CHIP_FILLED_DARKEN)
	return s


static func allocate_button_style(accent: Color, hover: bool) -> StyleBoxFlat:
	var s := UiTheme.frame("character_allocate")
	s.bg_color = accent.darkened(ALLOCATE_HOVER_DARKEN if hover else ALLOCATE_DARKEN)
	s.border_color = accent
	return s
