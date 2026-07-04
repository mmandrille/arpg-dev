extends SceneTree

const VisualCaptureScript := preload("res://scripts/showme/visual_capture.gd")
const ShowmeItemIconsCaptureScript := preload("res://scripts/showme/showme_item_icons_capture.gd")


func _initialize() -> void:
	if VisualCaptureScript == null:
		push_error("visual_capture preload failed")
		quit(1)
		return
	if ShowmeItemIconsCaptureScript == null:
		push_error("showme_item_icons_capture preload failed")
		quit(1)
		return
	print("[gdtest] PASS: showme capture scripts preload")
	quit(0)
