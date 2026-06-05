extends Control
class_name MonsterHealthBar

const SIZE := Vector2(76.0, 18.0)
const WORLD_OFFSET := Vector3(0.0, 2.75, 0.0)

var _camera: Camera3D
var _target: Node3D
var _max_hp := 1
var _hp := 1
var _background: ColorRect
var _fill: ColorRect
var _label: Label


func setup(camera: Camera3D, target: Node3D, hp: int, max_hp: int) -> void:
	_camera = camera
	_target = target
	_max_hp = max(1, max_hp)
	_build()
	update_hp(hp, max_hp)


func update_hp(hp: int, max_hp: int) -> void:
	_hp = clampi(hp, 0, max(1, max_hp))
	_max_hp = max(1, max_hp)
	_label.text = "%d/%d" % [_hp, _max_hp]
	var ratio := float(_hp) / float(_max_hp)
	_fill.size.x = maxf(0.0, SIZE.x * ratio)
	modulate.a = clampf(0.25 + 0.75 * ratio, 0.25, 1.0)
	_update_position()


func _process(_delta: float) -> void:
	if not is_instance_valid(_target):
		queue_free()
		return
	_update_position()


func _build() -> void:
	if _background != null:
		return

	size = SIZE
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	z_index = 90

	_background = ColorRect.new()
	_background.color = Color(0.08, 0.02, 0.02, 0.72)
	_background.size = SIZE
	add_child(_background)

	_fill = ColorRect.new()
	_fill.color = Color(0.85, 0.03, 0.03, 0.88)
	_fill.size = SIZE
	add_child(_fill)

	_label = Label.new()
	_label.size = SIZE
	_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var settings := LabelSettings.new()
	settings.font_size = 11
	settings.font_color = Color(1.0, 0.95, 0.88)
	settings.outline_size = 2
	settings.outline_color = Color(0.05, 0.02, 0.02, 0.8)
	_label.label_settings = settings
	add_child(_label)


func _update_position() -> void:
	if _camera == null or not is_instance_valid(_target):
		return
	var screen := _camera.unproject_position(_target.global_position + WORLD_OFFSET)
	position = screen - SIZE * 0.5
