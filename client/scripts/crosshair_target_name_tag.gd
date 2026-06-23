## CrosshairTargetNameTag — compact screen-space label above a locked target.
class_name CrosshairTargetNameTag
extends Control

const WORLD_OFFSET := Vector3(0.0, 1.45, 0.0)
const PAD_H := 10.0
const PAD_V := 5.0
const FONT_SIZE := 14
const ACCENT_HEIGHT := 2.0

var _camera: Camera3D
var _target: Node3D
var _label: Label
var _background: ColorRect
var _accent: ColorRect


func attach_to(parent: Node) -> void:
	if parent == null:
		return
	if get_parent() == parent:
		return
	if get_parent() != null:
		get_parent().remove_child(self)
	parent.add_child(self)


func show_for(camera: Camera3D, target: Node3D, text: String) -> void:
	_camera = camera
	_target = target
	_ensure_built()
	_label.text = text.strip_edges()
	if _label.text == "":
		hide_tag()
		return
	_sync_layout()
	visible = true
	set_process(true)
	_update_position()


func hide_tag() -> void:
	visible = false
	set_process(false)
	_camera = null
	_target = null
	if _label != null:
		_label.text = ""


func get_label_text() -> String:
	if not visible or _label == null:
		return ""
	return _label.text


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	z_index = 99
	visible = false
	_ensure_built()


func _process(_delta: float) -> void:
	if not visible:
		return
	if _camera == null or not is_instance_valid(_target):
		hide_tag()
		return
	_update_position()


func _ensure_built() -> void:
	if _label != null:
		return
	_background = ColorRect.new()
	_background.color = Color(0.05, 0.045, 0.04, 0.82)
	_background.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_background)
	_accent = ColorRect.new()
	_accent.color = Color("#d4a017")
	_accent.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_accent)
	_label = Label.new()
	_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var settings := LabelSettings.new()
	settings.font_size = FONT_SIZE
	settings.font_color = Color("#f3e8c8")
	settings.outline_size = 2
	settings.outline_color = Color(0.03, 0.025, 0.02, 0.9)
	_label.label_settings = settings
	add_child(_label)


func _sync_layout() -> void:
	var font := _label.get_theme_font("font")
	var text_size := font.get_string_size(_label.text, HORIZONTAL_ALIGNMENT_LEFT, -1, FONT_SIZE)
	var width := maxf(72.0, text_size.x + PAD_H * 2.0)
	var height := maxf(22.0, text_size.y + PAD_V * 2.0)
	size = Vector2(width, height)
	_background.size = size
	_accent.position = Vector2(0.0, height - ACCENT_HEIGHT)
	_accent.size = Vector2(width, ACCENT_HEIGHT)
	_label.position = Vector2(PAD_H * 0.5, PAD_V * 0.25)
	_label.size = Vector2(width - PAD_H, height - PAD_V)


func _update_position() -> void:
	if _camera == null or not is_instance_valid(_target):
		return
	var screen := _camera.unproject_position(_target.global_position + WORLD_OFFSET)
	position = screen - size * 0.5
