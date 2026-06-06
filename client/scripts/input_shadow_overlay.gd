class_name InputShadowOverlay
extends CanvasLayer

# Ghost cursor overlay for windowed `make bot-client` runs — visualizes bot
# actions (clicks, keys, inventory drags) without faking replay events.

const CURSOR_COLOR := Color(0.35, 0.92, 1.0, 0.82)
const CURSOR_OUTLINE := Color(0.05, 0.08, 0.12, 0.95)
const KEY_BG := Color(0.08, 0.1, 0.14, 0.9)
const KEY_BORDER := Color(0.35, 0.92, 1.0, 0.95)
const KEY_TEXT := Color(0.92, 0.96, 1.0, 1.0)

var _camera: Camera3D
var _cursor: GhostCursor
var _click_pulse: ClickPulse
var _drag_trail: Line2D
var _keys: HBoxContainer
var _active: bool = false
var _cursor_pos := Vector2(160.0, 160.0)
var _motion_tween: Tween
var _keys_tween: Tween


class GhostCursor:
	extends Control

	func _draw() -> void:
		var pts := PackedVector2Array([
			Vector2(2, 2),
			Vector2(18, 12),
			Vector2(11, 12),
			Vector2(11, 24),
			Vector2(2, 24),
		])
		draw_colored_polygon(pts, CURSOR_COLOR)
		draw_polyline(pts + PackedVector2Array([pts[0]]), CURSOR_OUTLINE, 1.5, true)


class ClickPulse:
	extends Control

	func _draw() -> void:
		var radius := 18.0 * scale.x
		draw_arc(
			Vector2.ZERO,
			radius,
			0.0,
			TAU,
			32,
			Color(CURSOR_COLOR.r, CURSOR_COLOR.g, CURSOR_COLOR.b, 0.55 * modulate.a),
			2.0
		)


func _ready() -> void:
	layer = 20
	visible = false

	_drag_trail = Line2D.new()
	_drag_trail.width = 2.0
	_drag_trail.default_color = Color(CURSOR_COLOR.r, CURSOR_COLOR.g, CURSOR_COLOR.b, 0.45)
	_drag_trail.visible = false
	_drag_trail.z_index = 5
	add_child(_drag_trail)

	_click_pulse = ClickPulse.new()
	_click_pulse.z_index = 8
	_click_pulse.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_click_pulse.visible = false
	add_child(_click_pulse)

	_cursor = GhostCursor.new()
	_cursor.custom_minimum_size = Vector2(24, 28)
	_cursor.z_index = 10
	_cursor.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_cursor)
	_cursor.position = _cursor_pos - Vector2(2, 2)

	_keys = HBoxContainer.new()
	_keys.add_theme_constant_override("separation", 6)
	_keys.position = Vector2(16, 16)
	_keys.z_index = 12
	_keys.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_keys)


func bind_camera(camera: Camera3D) -> void:
	_camera = camera


func set_active(active: bool) -> void:
	_active = active
	visible = active
	Input.set_mouse_mode(Input.MOUSE_MODE_HIDDEN if active else Input.MOUSE_MODE_VISIBLE)
	if not active:
		_clear_keys()
		_drag_trail.visible = false


func pulse_world_target(world_pos: Vector3, keys: PackedStringArray = PackedStringArray(["LMB"])) -> void:
	if not _active or _camera == null:
		return
	pulse_screen_target(_camera.unproject_position(world_pos + Vector3(0.0, 0.55, 0.0)), keys)


func pulse_screen_target(screen_pos: Vector2, keys: PackedStringArray = PackedStringArray(["LMB"])) -> void:
	if not _active:
		return
	_move_cursor_to(screen_pos)
	_show_keys(keys)
	_play_click_pulse(screen_pos)


func show_keys(keys: PackedStringArray, duration: float = 0.95) -> void:
	if not _active:
		return
	_show_keys(keys, duration)


func show_drag(screen_from: Vector2, screen_to: Vector2, keys: PackedStringArray = PackedStringArray(["drag"])) -> void:
	if not _active:
		return
	_drag_trail.points = PackedVector2Array([screen_from, screen_to])
	_drag_trail.visible = true
	_show_keys(keys)
	_kill_motion_tween()
	_cursor_pos = screen_from
	_cursor.position = _cursor_pos - Vector2(2, 2)
	_motion_tween = create_tween()
	_motion_tween.tween_method(_set_cursor_pos, screen_from, screen_to, 0.42).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_IN_OUT)
	_motion_tween.tween_callback(func() -> void:
		_play_click_pulse(screen_to)
		_drag_trail.visible = false)


func _move_cursor_to(screen_pos: Vector2) -> void:
	_kill_motion_tween()
	_motion_tween = create_tween()
	_motion_tween.tween_method(_set_cursor_pos, _cursor_pos, screen_pos, 0.22).set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)


func _set_cursor_pos(pos: Vector2) -> void:
	_cursor_pos = pos
	_cursor.position = _cursor_pos - Vector2(2, 2)


func _show_keys(keys: PackedStringArray, duration: float = 0.95) -> void:
	_clear_keys()
	if keys.is_empty():
		return
	for key_name in keys:
		_keys.add_child(_make_key_badge(str(key_name)))
	if _keys_tween != null and is_instance_valid(_keys_tween):
		_keys_tween.kill()
	_keys.modulate.a = 1.0
	_keys_tween = create_tween()
	_keys_tween.tween_interval(duration)
	_keys_tween.tween_property(_keys, "modulate:a", 0.0, 0.3)
	_keys_tween.tween_callback(_clear_keys)


func _make_key_badge(text: String) -> PanelContainer:
	var panel := PanelContainer.new()
	panel.add_theme_stylebox_override("panel", _key_style())
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", KEY_TEXT)
	label.add_theme_font_size_override("font_size", 12)
	panel.add_child(label)
	return panel


func _key_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = KEY_BG
	style.border_color = KEY_BORDER
	style.border_width_left = 1
	style.border_width_top = 1
	style.border_width_right = 1
	style.border_width_bottom = 1
	style.corner_radius_top_left = 4
	style.corner_radius_top_right = 4
	style.corner_radius_bottom_left = 4
	style.corner_radius_bottom_right = 4
	style.content_margin_left = 8
	style.content_margin_top = 4
	style.content_margin_right = 8
	style.content_margin_bottom = 4
	return style


func _play_click_pulse(screen_pos: Vector2) -> void:
	_click_pulse.position = screen_pos
	_click_pulse.visible = true
	_click_pulse.modulate.a = 1.0
	_click_pulse.scale = Vector2.ONE * 0.35
	_click_pulse.queue_redraw()
	var tween := create_tween()
	tween.tween_property(_click_pulse, "scale", Vector2.ONE * 1.35, 0.28).set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	tween.parallel().tween_property(_click_pulse, "modulate:a", 0.0, 0.28)
	tween.tween_callback(func() -> void:
		if _click_pulse != null:
			_click_pulse.visible = false)


func _clear_keys() -> void:
	if _keys == null:
		return
	for child in _keys.get_children():
		child.queue_free()


func _kill_motion_tween() -> void:
	if _motion_tween != null and is_instance_valid(_motion_tween):
		_motion_tween.kill()
		_motion_tween = null
