class_name TextInputFocusGuard
extends RefCounted


static func has_text_input_focus(viewport: Viewport) -> bool:
	if viewport == null:
		return false
	var owner := viewport.gui_get_focus_owner()
	return owner is LineEdit or owner is TextEdit
