class_name FogOfWarOverlay
extends CanvasLayer

const GLOOM_MULTIPLIER := 1.25
const FALLBACK_WORLD_TO_SCREEN := 32.0
const SHADER_CODE := """
shader_type canvas_item;
render_mode blend_mix, unshaded;

uniform vec2 center_px = vec2(0.0, 0.0);
uniform vec2 viewport_px = vec2(1.0, 1.0);
uniform float light_radius_px = 0.0;
uniform float gloom_radius_px = 0.0;
uniform vec4 gloom_color : source_color = vec4(0.22, 0.24, 0.27, 0.52);
uniform vec4 darkness_color : source_color = vec4(0.0, 0.0, 0.0, 0.90);

void fragment() {
	vec2 pos = SCREEN_UV * viewport_px;
	float d = distance(pos, center_px);
	if (d <= light_radius_px) {
		COLOR = vec4(0.0, 0.0, 0.0, 0.0);
	} else if (d <= gloom_radius_px) {
		float t = smoothstep(light_radius_px, gloom_radius_px, d);
		COLOR = mix(gloom_color, darkness_color, t);
	} else {
		COLOR = darkness_color;
	}
}
"""

var _camera: Camera3D
var _target: Node3D
var _rect: ColorRect
var _material: ShaderMaterial
var _active: bool = true
var _light_radius: float = 0.0
var _gloom_radius: float = 0.0
var _light_radius_px: float = 0.0
var _gloom_radius_px: float = 0.0
var _center_px := Vector2.ZERO


func _ready() -> void:
	layer = 0
	_ensure_rect()
	set_process(true)
	_sync_visibility()


func bind(camera: Camera3D, target: Node3D) -> void:
	_camera = camera
	_target = target
	_update_shader()


func set_active(active: bool) -> void:
	_active = active
	_sync_visibility()


func set_progression(progression: Dictionary) -> void:
	var derived: Dictionary = progression.get("derived_stats", {})
	set_light_radius(float(derived.get("light_radius", 0.0)))


func set_light_radius(radius: float) -> void:
	_light_radius = maxf(0.0, radius)
	_gloom_radius = _light_radius * GLOOM_MULTIPLIER
	_sync_visibility()
	_update_shader()


func get_debug_state() -> Dictionary:
	return {
		"enabled": visible,
		"active": _active,
		"light_radius": _light_radius,
		"gloom_radius": _gloom_radius,
		"light_radius_px": _light_radius_px,
		"gloom_radius_px": _gloom_radius_px,
		"center": {"x": _center_px.x, "y": _center_px.y},
	}


func _process(_delta: float) -> void:
	_update_shader()


func _ensure_rect() -> void:
	if _rect != null:
		return
	_rect = ColorRect.new()
	_rect.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_rect.set_anchors_preset(Control.PRESET_FULL_RECT)
	var shader := Shader.new()
	shader.code = SHADER_CODE
	_material = ShaderMaterial.new()
	_material.shader = shader
	_rect.material = _material
	add_child(_rect)


func _sync_visibility() -> void:
	visible = _active and _light_radius > 0.0


func _update_shader() -> void:
	if _material == null:
		return
	var viewport_size := get_viewport().get_visible_rect().size if get_viewport() != null else Vector2(1.0, 1.0)
	_center_px = _project_target()
	_light_radius_px = _projected_radius(_light_radius)
	_gloom_radius_px = _projected_radius(_gloom_radius)
	_material.set_shader_parameter("viewport_px", viewport_size)
	_material.set_shader_parameter("center_px", _center_px)
	_material.set_shader_parameter("light_radius_px", _light_radius_px)
	_material.set_shader_parameter("gloom_radius_px", _gloom_radius_px)


func _project_target() -> Vector2:
	if _camera != null and _target != null:
		return _camera.unproject_position(_target.global_position)
	return get_viewport().get_visible_rect().size * 0.5 if get_viewport() != null else Vector2.ZERO


func _projected_radius(world_radius: float) -> float:
	if world_radius <= 0.0:
		return 0.0
	if _camera != null and _target != null:
		var center := _camera.unproject_position(_target.global_position)
		var edge := _camera.unproject_position(_target.global_position + Vector3(world_radius, 0.0, 0.0))
		return maxf(1.0, center.distance_to(edge))
	return world_radius * FALLBACK_WORLD_TO_SCREEN
