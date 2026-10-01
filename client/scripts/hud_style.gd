class_name HudStyle
extends RefCounted
## HUD colors, frames and slot styles (v517), all resolved from UiTheme tokens in
## shared/assets/ui_theme.v0.json (hud_* prefix). No literals live here: edit the catalog to restyle.
## Color getters are lower-case functions (hud_<name> tokens); frames are hud_* frame recipes.


static func bronze() -> Color:
	return UiTheme.color("hud_bronze")

static func bronze_bright() -> Color:
	return UiTheme.color("hud_bronze_bright")

static func gold() -> Color:
	return UiTheme.color("hud_gold")

static func text() -> Color:
	return UiTheme.color("hud_text")

static func text_muted() -> Color:
	return UiTheme.color("hud_text_muted")

static func panel_bg() -> Color:
	return UiTheme.color("hud_panel_bg")

static func slot_empty_bg() -> Color:
	return UiTheme.color("hud_slot_empty_bg")

static func slot_filled_bg() -> Color:
	return UiTheme.color("hud_slot_filled_bg")

static func slot_hover_bg() -> Color:
	return UiTheme.color("hud_slot_hover_bg")

static func slot_pressed_bg() -> Color:
	return UiTheme.color("hud_slot_pressed_bg")

static func mana() -> Color:
	return UiTheme.color("hud_mana")

static func hp_high() -> Color:
	return UiTheme.color("hud_hp_high")

static func hp_mid() -> Color:
	return UiTheme.color("hud_hp_mid")

static func hp_low() -> Color:
	return UiTheme.color("hud_hp_low")

static func hp_flash_damage() -> Color:
	return UiTheme.color("hud_hp_flash_damage")

static func hp_flash_heal() -> Color:
	return UiTheme.color("hud_hp_flash_heal")

static func mana_flash_restore() -> Color:
	return UiTheme.color("hud_mana_flash_restore")

static func globe_ring() -> Color:
	return UiTheme.color("hud_globe_ring")

static func globe_ring_edge() -> Color:
	return UiTheme.color("hud_globe_ring_edge")

static func globe_trough() -> Color:
	return UiTheme.color("hud_globe_trough")

static func globe_surface() -> Color:
	return UiTheme.color("hud_globe_surface")

static func globe_highlight() -> Color:
	return UiTheme.color("hud_globe_highlight")

static func globe_text() -> Color:
	return UiTheme.color("hud_globe_text")

static func globe_text_shadow() -> Color:
	return UiTheme.color("hud_globe_text_shadow")

static func xp_bg() -> Color:
	return UiTheme.color("hud_xp_bg")

static func xp_border() -> Color:
	return UiTheme.color("hud_xp_border")

static func minimap_bg() -> Color:
	return UiTheme.color("hud_minimap_bg")

static func minimap_border() -> Color:
	return UiTheme.color("hud_minimap_border")

static func boss_panel_bg() -> Color:
	return UiTheme.color("hud_boss_panel_bg")

static func boss_panel_border() -> Color:
	return UiTheme.color("hud_boss_panel_border")

static func boss_title() -> Color:
	return UiTheme.color("hud_boss_title")

static func boss_trough() -> Color:
	return UiTheme.color("hud_boss_trough")

static func boss_trough_border() -> Color:
	return UiTheme.color("hud_boss_trough_border")

static func boss_fill_high() -> Color:
	return UiTheme.color("hud_boss_fill_high")

static func boss_fill_mid() -> Color:
	return UiTheme.color("hud_boss_fill_mid")

static func boss_fill_low() -> Color:
	return UiTheme.color("hud_boss_fill_low")

static func boss_phase_trough() -> Color:
	return UiTheme.color("hud_boss_phase_trough")

static func boss_phase_telegraph() -> Color:
	return UiTheme.color("hud_boss_phase_telegraph")

static func boss_phase_active() -> Color:
	return UiTheme.color("hud_boss_phase_active")

static func boss_phase_recovery() -> Color:
	return UiTheme.color("hud_boss_phase_recovery")

static func boss_phase_other() -> Color:
	return UiTheme.color("hud_boss_phase_other")

static func boss_reward_bg() -> Color:
	return UiTheme.color("hud_boss_reward_bg")

static func boss_reward_border() -> Color:
	return UiTheme.color("hud_boss_reward_border")

static func boss_portrait_bg() -> Color:
	return UiTheme.color("hud_boss_portrait_bg")

static func clear() -> Color:
	return UiTheme.color("hud_clear")

static func shadow() -> Color:
	return UiTheme.color("hud_shadow")

static func slot_empty_border() -> Color:
	return UiTheme.color("hud_slot_empty_border")

static func minimap_shadow() -> Color:
	return UiTheme.color("hud_minimap_shadow")

static func boss_shadow() -> Color:
	return UiTheme.color("hud_boss_shadow")

static func boss_gloss() -> Color:
	return UiTheme.color("hud_boss_gloss")

static func boss_portrait_inner() -> Color:
	return UiTheme.color("hud_boss_portrait_inner")

static func boss_text() -> Color:
	return UiTheme.color("hud_boss_text")

static func boss_phase_text() -> Color:
	return UiTheme.color("hud_boss_phase_text")

static func boss_reward_hint() -> Color:
	return UiTheme.color("hud_boss_reward_hint")

static func hp_color(ratio: float) -> Color:
	if ratio > 0.6:
		return hp_high()
	if ratio > 0.3:
		return hp_mid()
	return hp_low()


static func globe_ring_width() -> float:
	return float(UiTheme.spacing("hud_globe_ring_width"))


static func _with_margin(style: StyleBoxFlat, horizontal: float, vertical: float) -> StyleBoxFlat:
	style.content_margin_left = horizontal
	style.content_margin_right = horizontal
	style.content_margin_top = vertical
	style.content_margin_bottom = vertical
	return style


static func _with_shadow(style: StyleBoxFlat, color_value: Color, size_token: String) -> StyleBoxFlat:
	style.shadow_color = color_value
	style.shadow_size = UiTheme.spacing(size_token)
	return style


## Shared HUD panel frame (hotbar, skill/character slots, name strip).
static func frame_panel(bg_alpha: float = 0.88, margin_h: float = 6.0, margin_v: float = 5.0) -> StyleBoxFlat:
	var s := UiTheme.frame("hud_panel")
	s.bg_color.a = bg_alpha
	_with_margin(s, margin_h, margin_v)
	return _with_shadow(s, shadow(), "hud_shadow_size")


## Slot states: "empty", "filled", "hover", "pressed", "disabled".
static func slot(kind: String) -> StyleBoxFlat:
	match kind:
		"filled":
			return UiTheme.frame("hud_slot_filled")
		"hover", "disabled":
			return UiTheme.frame("hud_slot_empty", kind)
		"pressed": # the schema's state vocabulary calls the pressed look "selected"
			return UiTheme.frame("hud_slot_empty", "selected")
		_:
			return UiTheme.frame("hud_slot_empty")


static func xp_background() -> StyleBoxFlat:
	return UiTheme.frame("hud_xp_bg")


static func xp_fill() -> StyleBoxFlat:
	return UiTheme.frame("hud_xp_fill")


static func minimap_frame(opacity: float) -> StyleBoxFlat:
	var s := UiTheme.frame("hud_minimap")
	s.bg_color.a = opacity
	return _with_shadow(s, minimap_shadow(), "hud_shadow_size")
