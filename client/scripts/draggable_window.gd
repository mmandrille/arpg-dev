class_name DraggableWindow
extends PanelContainer

signal close_requested

const TITLEBAR_HEIGHT := 34.0
const CLOSE_SIZE := Vector2(30, 28)
static var layout_storage_path: String = "user://window_layout.cfg"
static var force_enable_persistence_for_tests: bool = false

var window_title: String = ""
var draggable: bool = true
var layout_key: String = ""
var _titlebar: PanelContainer
var _title_label: Label
var _close_button: Button
var _content_host: MarginContainer
var _dragging: bool = false
var _drag_offset: Vector2 = Vector2.ZERO


func _ready() -> void:
	if _titlebar == null:
		_build()
	get_viewport().size_changed.connect(clamp_to_viewport)


func configure(title: String, content_min_size: Vector2) -> void:
	window_title = title
	if _titlebar == null:
		_build()
	_title_label.text = title
	_content_host.custom_minimum_size = content_min_size


func set_content(content: Control) -> void:
	if _content_host == null:
		_build()
	for child in _content_host.get_children():
		_content_host.remove_child(child)
		child.queue_free()
	_content_host.add_child(content)


func close_button() -> Button:
	return _close_button


func set_layout_key(key: String) -> void:
	layout_key = key.strip_edges()
	_load_position()


func titlebar() -> Control:
	return _titlebar


func bot_drag_by(delta: Vector2) -> void:
	position += delta
	clamp_to_viewport()
	_save_position()


func get_debug_state() -> Dictionary:
	return {
		"title": window_title,
		"position": {"x": position.x, "y": position.y},
		"size": {"x": size.x, "y": size.y},
		"minimum_size": {"x": custom_minimum_size.x, "y": custom_minimum_size.y},
		"close_visible": _close_button != null and _close_button.visible,
		"draggable": draggable,
		"layout_key": layout_key,
		"persistence_enabled": _persistence_enabled(),
	}


func clamp_to_viewport() -> void:
	var viewport_size := get_viewport_rect().size if is_inside_tree() else Vector2(1280, 720)
	if viewport_size.x < 640.0 or viewport_size.y < 360.0:
		viewport_size = Vector2(1280, 720)
	var panel_size := custom_minimum_size
	if panel_size.x <= 0.0 or panel_size.y <= 0.0:
		panel_size = size
	var max_x := maxf(0.0, viewport_size.x - minf(panel_size.x, viewport_size.x))
	var max_y := maxf(0.0, viewport_size.y - TITLEBAR_HEIGHT)
	position = Vector2(clampf(position.x, 0.0, max_x), clampf(position.y, 0.0, max_y))


func _build() -> void:
	mouse_filter = Control.MOUSE_FILTER_STOP
	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 0)
	add_child(root)

	_titlebar = PanelContainer.new()
	_titlebar.custom_minimum_size = Vector2(0, TITLEBAR_HEIGHT)
	_titlebar.mouse_filter = Control.MOUSE_FILTER_STOP
	_titlebar.add_theme_stylebox_override("panel", _titlebar_style())
	_titlebar.gui_input.connect(_on_titlebar_gui_input)
	root.add_child(_titlebar)

	var title_row := HBoxContainer.new()
	title_row.add_theme_constant_override("separation", 8)
	_titlebar.add_child(title_row)

	_title_label = Label.new()
	_title_label.text = window_title
	_title_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	_title_label.add_theme_font_size_override("font_size", 18)
	_title_label.add_theme_color_override("font_color", Color("#f0dfbb"))
	title_row.add_child(_title_label)

	_close_button = Button.new()
	_close_button.text = "X"
	_close_button.tooltip_text = "Close"
	_close_button.focus_mode = Control.FOCUS_NONE
	_close_button.custom_minimum_size = CLOSE_SIZE
	_close_button.pressed.connect(func() -> void:
		close_requested.emit()
	)
	title_row.add_child(_close_button)

	_content_host = MarginContainer.new()
	_content_host.mouse_filter = Control.MOUSE_FILTER_PASS
	root.add_child(_content_host)


func _on_titlebar_gui_input(event: InputEvent) -> void:
	if not draggable:
		return
	if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT:
		_dragging = event.pressed
		_drag_offset = get_global_mouse_position() - global_position
		accept_event()
	elif event is InputEventMouseMotion and _dragging:
		global_position = get_global_mouse_position() - _drag_offset
		clamp_to_viewport()
		accept_event()
	if event is InputEventMouseButton and not event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		_save_position()


func _titlebar_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.025, 0.023, 0.02, 0.96)
	s.border_color = Color("#2f2718")
	s.border_width_bottom = 1
	s.content_margin_left = 10
	s.content_margin_right = 8
	s.content_margin_top = 3
	s.content_margin_bottom = 3
	return s


func _persistence_enabled() -> bool:
	if layout_key == "":
		return false
	if OS.get_environment("CLIENT_UNIT_ONLY") == "1" and not force_enable_persistence_for_tests:
		return false
	return true


func _load_position() -> void:
	if not _persistence_enabled():
		return
	var cfg := ConfigFile.new()
	if cfg.load(layout_storage_path) != OK:
		return
	if not cfg.has_section_key(layout_key, "x") or not cfg.has_section_key(layout_key, "y"):
		return
	position = Vector2(float(cfg.get_value(layout_key, "x", position.x)), float(cfg.get_value(layout_key, "y", position.y)))
	clamp_to_viewport()


func _save_position() -> void:
	if not _persistence_enabled():
		return
	var cfg := ConfigFile.new()
	cfg.load(layout_storage_path)
	cfg.set_value(layout_key, "x", position.x)
	cfg.set_value(layout_key, "y", position.y)
	var err := cfg.save(layout_storage_path)
	if err != OK:
		push_warning("could not save window layout: %s" % layout_storage_path)
