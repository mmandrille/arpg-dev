class_name SkillTreeStyles
extends RefCounted
## Theme-backed presentation helpers for skill nodes, connectors, and tooltips.

static func state_for(rank: int, spendable: bool) -> String:
	if rank > 0:
		return "learned"
	if spendable:
		return "available"
	return "locked"


static func panel_frame() -> StyleBoxFlat:
	return UiTheme.frame("skill_panel")


static func canvas_frame() -> StyleBoxFlat:
	return UiTheme.frame("skill_tree_surface")


static func tooltip_frame() -> StyleBoxFlat:
	return UiTheme.frame("skill_tooltip")


static func status_frame(state: String) -> StyleBoxFlat:
	return UiTheme.frame("skill_status_%s" % _normalized_state(state))


static func node_frame(state: String, selected: bool, hovered: bool) -> StyleBoxFlat:
	var normalized := _normalized_state(state)
	var frame_state := "hover" if hovered else "selected" if selected else UiTheme.FRAME_BASE_STATE
	var style := UiTheme.frame("skill_node_%s" % normalized, frame_state)
	var width := UiTheme.spacing("skill_tree_node_%s_border" % normalized)
	if selected:
		width += UiTheme.spacing("skill_tree_node_selected_extra_border")
	if hovered:
		width += UiTheme.spacing("skill_tree_node_hover_extra_border")
	if selected and hovered:
		style.border_color = UiTheme.color("skill_node_focus_border")
	elif selected:
		style.border_color = UiTheme.color("skill_node_selected_border")
	elif hovered:
		style.border_color = UiTheme.color("skill_node_hover_border")
	style.border_width_left = width
	style.border_width_top = width
	style.border_width_right = width
	style.border_width_bottom = width
	return style


static func state_text_color(state: String) -> Color:
	return UiTheme.color("skill_node_%s_text" % _normalized_state(state))


static func rank_text_color(state: String) -> Color:
	return UiTheme.color("skill_rank_%s" % _normalized_state(state))


static func icon_modulate(state: String, selected: bool) -> Color:
	if _normalized_state(state) != "locked":
		return Color.WHITE
	return UiTheme.color("skill_icon_locked_selected_modulate" if selected else "skill_icon_locked_modulate")


static func connector_color(met: bool) -> Color:
	return UiTheme.color("skill_connector_met" if met else "skill_connector_unmet")


static func connector_width() -> float:
	return float(UiTheme.spacing("skill_tree_connector_width"))


static func points_color(available: bool) -> Color:
	return UiTheme.color("skill_points_available" if available else "skill_points_empty")


static func _normalized_state(state: String) -> String:
	return state if state in ["available", "learned", "locked"] else "locked"
