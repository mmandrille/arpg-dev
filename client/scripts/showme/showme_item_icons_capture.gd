class_name ShowmeItemIconsCapture
extends RefCounted

const ItemIconsCatalogScript := preload("res://scripts/item_icons_catalog.gd")


static func setup(capture: SceneTree) -> void:
	var panel = ItemIconsCatalogScript.new()
	capture.get_root().add_child(panel)
	await capture.process_frame
	panel.ensure_display_visible()
