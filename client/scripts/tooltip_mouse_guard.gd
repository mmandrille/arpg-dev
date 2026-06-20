class_name TooltipMouseGuard
extends RefCounted


static func ignore_mouse(root: Control) -> void:
	root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	for child in root.get_children():
		var control := child as Control
		if control != null:
			ignore_mouse(control)
