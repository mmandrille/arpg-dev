class_name BossBarFrame
extends RefCounted
## Look of the boss health bar widget (v517): panel/reward frames, trough and fill colors, gloss and
## portrait frame. Style only; layout constants and behavior stay in boss_health_bar.gd.
## Colors and frames come from UiTheme hud_* tokens.


static func panel_style() -> StyleBoxFlat:
	var s := UiTheme.frame("hud_boss_panel")
	s.shadow_color = HudStyle.boss_shadow()
	s.shadow_size = UiTheme.spacing("hud_boss_shadow_size")
	return s


static func reward_style() -> StyleBoxFlat:
	var s := UiTheme.frame("hud_boss_reward")
	s.shadow_color = HudStyle.boss_shadow()
	s.shadow_size = UiTheme.spacing("hud_boss_shadow_size")
	return s


static func fill_color(ratio: float) -> Color:
	if ratio > 0.6:
		return HudStyle.boss_fill_high()
	if ratio > 0.3:
		return HudStyle.boss_fill_mid()
	return HudStyle.boss_fill_low()


static func phase_color(kind: String) -> Color:
	match kind:
		"telegraph":
			return HudStyle.boss_phase_telegraph()
		"active":
			return HudStyle.boss_phase_active()
		"recovery":
			return HudStyle.boss_phase_recovery()
		_:
			return HudStyle.boss_phase_other()


## Thin highlight along the top edge of a bar fill; follows the fill width through anchors.
static func add_gloss(fill: ColorRect, height: float) -> void:
	var gloss := ColorRect.new()
	gloss.mouse_filter = Control.MOUSE_FILTER_IGNORE
	gloss.color = HudStyle.boss_gloss()
	gloss.set_anchors_preset(Control.PRESET_TOP_WIDE)
	gloss.offset_bottom = height
	fill.add_child(gloss)


## Trough: dark inset with a 1px frame drawn by a child panel so the ColorRect contract is unchanged.
static func add_trough_frame(trough: ColorRect) -> void:
	var frame := Panel.new()
	frame.mouse_filter = Control.MOUSE_FILTER_IGNORE
	frame.add_theme_stylebox_override("panel", UiTheme.frame("hud_boss_trough"))
	frame.set_anchors_preset(Control.PRESET_FULL_RECT)
	trough.add_child(frame)


static func draw_portrait_frame(portrait: Control, rect: Rect2) -> void:
	portrait.draw_rect(rect, HudStyle.boss_portrait_bg(), true)
	portrait.draw_rect(rect.grow(-1.0), HudStyle.boss_panel_border(), false, 2.0)
	portrait.draw_rect(rect.grow(-4.0), HudStyle.boss_portrait_inner(), false, 1.0)
