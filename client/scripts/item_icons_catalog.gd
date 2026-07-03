class_name ItemIconsCatalog
extends Control

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const ItemFamilyIconPreviewScript := preload("res://scripts/item_family_icon_preview.gd")

const ICON_SIZE := Vector2(72, 72)
const CELL_SIZE := Vector2(108, 118)
const GRID_COLUMNS := 6


func _init() -> void:
	_build()


func ensure_display_visible() -> void:
	visible = true
	show()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	var panel := DraggableWindowScript.new()
	panel.custom_minimum_size = Vector2(700, 640)
	panel.position = Vector2(130, 20)
	panel.configure("Item Icon Families", Vector2(668, 592))
	panel.set_layout_key("item_icons_catalog")
	panel.mouse_filter = Control.MOUSE_FILTER_STOP
	panel.add_theme_stylebox_override("panel", _panel_style())
	add_child(panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	root.custom_minimum_size = Vector2(668, 592)
	panel.set_content(root)

	var subtitle := Label.new()
	subtitle.text = "Shared presentation families from item_presentations.v0.json"
	subtitle.add_theme_font_size_override("font_size", 14)
	subtitle.add_theme_color_override("font_color", Color("#b9ad97"))
	root.add_child(subtitle)

	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(668, 548)
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	root.add_child(scroll)

	var grid := GridContainer.new()
	grid.columns = GRID_COLUMNS
	grid.add_theme_constant_override("h_separation", 10)
	grid.add_theme_constant_override("v_separation", 10)
	scroll.add_child(grid)

	ItemRulesLoader.ensure_loaded()
	var family_ids: Array = ItemRulesLoader.item_presentation_families.keys()
	family_ids.sort()
	for family_id in family_ids:
		var family: Dictionary = ItemRulesLoader.item_presentation_families.get(family_id, {})
		grid.add_child(_make_family_cell(str(family_id), family))


func _make_family_cell(family_id: String, family: Dictionary) -> Control:
	var cell := VBoxContainer.new()
	cell.custom_minimum_size = CELL_SIZE
	cell.add_theme_constant_override("separation", 4)

	var frame := Panel.new()
	frame.custom_minimum_size = ICON_SIZE
	frame.add_theme_stylebox_override("panel", _cell_style())
	cell.add_child(frame)

	var icon := ItemFamilyIconPreviewScript.new()
	icon.position = Vector2(0, 0)
	icon.size = ICON_SIZE
	icon.configure(family.get("icon", {}))
	frame.add_child(icon)

	var name_label := Label.new()
	name_label.text = family_id
	name_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	name_label.add_theme_font_size_override("font_size", 13)
	name_label.add_theme_color_override("font_color", Color("#f0dfbb"))
	cell.add_child(name_label)

	var shape_label := Label.new()
	var shape := str(family.get("icon", {}).get("shape", ""))
	var label := str(family.get("icon", {}).get("label", ""))
	shape_label.text = "%s · %s" % [shape, label] if label != "" else shape
	shape_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	shape_label.add_theme_font_size_override("font_size", 11)
	shape_label.add_theme_color_override("font_color", Color("#9a9080"))
	cell.add_child(shape_label)

	return cell


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#1a1b1d")
	style.border_color = Color("#4a4034")
	style.set_border_width_all(2)
	style.set_corner_radius_all(4)
	return style


func _cell_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#151617")
	style.border_color = Color("#3a342c")
	style.set_border_width_all(1)
	style.set_corner_radius_all(4)
	return style
