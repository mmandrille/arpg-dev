class_name ShowmeItemIconCapture
extends RefCounted

const ItemFamilyIconPreviewScript := preload("res://scripts/item_family_icon_preview.gd")


static func setup(capture: SceneTree, family_id: String) -> void:
	var root := Control.new()
	root.name = "VisualFeedbackItemIcon"
	root.set_anchors_preset(Control.PRESET_FULL_RECT)
	capture.get_root().add_child(root)

	var panel := Panel.new()
	panel.custom_minimum_size = Vector2(280, 320)
	panel.position = Vector2(180, 80)
	panel.add_theme_stylebox_override("panel", _panel_style())
	root.add_child(panel)

	var layout := VBoxContainer.new()
	layout.position = Vector2(24, 24)
	layout.add_theme_constant_override("separation", 12)
	panel.add_child(layout)

	ItemRulesLoader.ensure_loaded()
	var families: Dictionary = ItemRulesLoader.item_presentation_families
	var effective_id := family_id
	if effective_id == "" or not families.has(effective_id):
		var ids: Array = families.keys()
		ids.sort()
		effective_id = str(ids[0]) if not ids.is_empty() else ""
	var family: Dictionary = families.get(effective_id, {})

	var frame := Panel.new()
	frame.custom_minimum_size = Vector2(128, 128)
	frame.add_theme_stylebox_override("panel", _cell_style())
	layout.add_child(frame)

	var icon := ItemFamilyIconPreviewScript.new()
	icon.position = Vector2(0, 0)
	icon.size = Vector2(128, 128)
	icon.configure(family.get("icon", {}))
	frame.add_child(icon)

	var title := Label.new()
	title.text = effective_id
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 22)
	title.add_theme_color_override("font_color", Color("#f0dfbb"))
	layout.add_child(title)

	var meta := Label.new()
	var shape := str(family.get("icon", {}).get("shape", ""))
	var label := str(family.get("icon", {}).get("label", ""))
	meta.text = "%s · %s" % [shape, label]
	meta.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	meta.add_theme_font_size_override("font_size", 14)
	meta.add_theme_color_override("font_color", Color("#9a9080"))
	layout.add_child(meta)

	await capture.process_frame


static func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#1a1b1d")
	style.border_color = Color("#4a4034")
	style.set_border_width_all(2)
	style.set_corner_radius_all(4)
	return style


static func _cell_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#151617")
	style.border_color = Color("#3a342c")
	style.set_border_width_all(1)
	style.set_corner_radius_all(4)
	return style
