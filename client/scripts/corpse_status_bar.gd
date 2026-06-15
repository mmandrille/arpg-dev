extends Control
class_name CorpseStatusBar

const SIZE := Vector2(150.0, 24.0)
const WORLD_OFFSET := Vector3(0.0, 2.15, 0.0)

var _camera: Camera3D
var _target: Node3D
var _label: Label


func setup(camera: Camera3D, target: Node3D, text: String) -> void:
	_camera = camera
	_target = target
	_build()
	_label.text = text
	_update_position()


func _process(_delta: float) -> void:
	if not is_instance_valid(_target):
		queue_free()
		return
	_update_position()


func _build() -> void:
	if _label != null:
		return
	size = SIZE
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	z_index = 95

	var background := ColorRect.new()
	background.color = Color(0.03, 0.08, 0.025, 0.86)
	background.size = SIZE
	background.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(background)

	var accent := ColorRect.new()
	accent.color = Color(0.42, 0.88, 0.30, 0.96)
	accent.position = Vector2(0.0, SIZE.y - 4.0)
	accent.size = Vector2(SIZE.x, 4.0)
	accent.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(accent)

	_label = Label.new()
	_label.size = SIZE
	_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var settings := LabelSettings.new()
	settings.font_size = 13
	settings.font_color = Color("#e8f5cf")
	settings.outline_size = 2
	settings.outline_color = Color(0.02, 0.04, 0.01, 0.92)
	_label.label_settings = settings
	add_child(_label)


func _update_position() -> void:
	if _camera == null or not is_instance_valid(_target):
		return
	var screen := _camera.unproject_position(_target.global_position + WORLD_OFFSET)
	position = screen - SIZE * 0.5
