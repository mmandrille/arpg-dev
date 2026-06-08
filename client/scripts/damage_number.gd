extends Label
class_name DamageNumber

const LIFETIME := 0.85
const RISE_PIXELS := 42.0
const SIDE_PIXELS := 14.0
const WORLD_OFFSET := Vector3(0.0, 1.7, 0.0)

var _camera: Camera3D
var _target: Node3D
var _world_position := Vector3.ZERO
var _age := 0.0
var _side_offset := 0.0
var combat_text: String = ""
var combat_variant: String = "normal"


func setup(camera: Camera3D, target: Node3D, world_position: Vector3, amount = null, color: Color = Color.WHITE, side: float = 1.0, prefix: String = "", variant: String = "normal", text_override: String = "") -> void:
	_camera = camera
	_target = target
	_world_position = world_position
	_side_offset = SIDE_PIXELS * side
	combat_variant = variant
	combat_text = text_override if text_override != "" else prefix + str(int(amount))
	text = combat_text
	z_index = 100
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	custom_minimum_size = Vector2(76, 30)
	size = custom_minimum_size
	rotation = deg_to_rad(4.0 * side) if variant == "crit" else 0.0

	var settings := LabelSettings.new()
	settings.font_size = 26 if variant == "crit" else 20 if variant in ["miss", "block"] else 22
	settings.font_color = color
	settings.outline_size = 4
	settings.outline_color = Color(0.08, 0.06, 0.04, 0.85)
	label_settings = settings
	_update_position()


func _process(delta: float) -> void:
	_age += delta
	if _age >= LIFETIME:
		queue_free()
		return

	var t := _age / LIFETIME
	modulate.a = 1.0 - smoothstep(0.62, 1.0, t)
	var start_scale := 1.28 if combat_variant == "crit" else 1.08 if combat_variant in ["miss", "block"] else 1.18
	scale = Vector2.ONE * (start_scale - 0.18 * t)
	_update_position()


func _update_position() -> void:
	if _camera == null:
		return

	var anchor := _world_position
	if is_instance_valid(_target):
		anchor = _target.global_position

	var screen := _camera.unproject_position(anchor + WORLD_OFFSET)
	var t := _age / LIFETIME
	position = screen + Vector2(_side_offset, -RISE_PIXELS * t) - size * 0.5
