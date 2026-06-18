class_name FogOfWarOverlay
extends CanvasLayer

const GLOOM_MULTIPLIER := 1.25
const FALLBACK_WORLD_TO_SCREEN := 32.0
const GLOOM_ALPHA := 0.52
const DARKNESS_ALPHA := 1.0
const SHADOW_EDGE_EPSILON := 0.08
const SHADOW_START_OFFSET := 0.16
const SHADOW_WALL_HEIGHT := 1.0
const ORGANIC_EDGE_WORLD_AMPLITUDE := 0.65
const ORGANIC_EDGE_MIN_PX := 5.0
const ORGANIC_EDGE_MAX_GLOOM_FRACTION := 0.10
const ORGANIC_EDGE_SEGMENTS := 18.0
const ORGANIC_EDGE_SEED := 41.0
const SHADER_CODE := """
shader_type canvas_item;
render_mode blend_mix, unshaded;

uniform vec2 center_px = vec2(0.0, 0.0);
uniform vec2 viewport_px = vec2(1.0, 1.0);
uniform float light_radius_px = 0.0;
uniform float gloom_radius_px = 0.0;
uniform float organic_edge_px = 0.0;
uniform float organic_edge_segments = 18.0;
uniform float organic_edge_seed = 41.0;
uniform vec4 gloom_color : source_color = vec4(0.22, 0.24, 0.27, 0.52);
uniform vec4 darkness_color : source_color = vec4(0.0, 0.0, 0.0, 1.0);

const float ORGANIC_PI = 3.14159265359;
const float ORGANIC_TAU = 6.28318530718;

float hash1(float n) {
	return fract(sin(n) * 43758.5453123);
}

float smooth_noise(float x) {
	float i = floor(x);
	float f = fract(x);
	float u = f * f * (3.0 - 2.0 * f);
	return mix(hash1(i), hash1(i + 1.0), u);
}

float organic_edge(vec2 delta) {
	if (organic_edge_px <= 0.0 || length(delta) <= 0.001) {
		return 0.0;
	}
	float angle = (atan(delta.y, delta.x) + ORGANIC_PI) / ORGANIC_TAU;
	float low = smooth_noise(angle * organic_edge_segments + organic_edge_seed);
	float high = smooth_noise(angle * organic_edge_segments * 2.17 + organic_edge_seed * 3.31);
	float combined = low * 0.72 + high * 0.28;
	return (combined - 0.5) * 2.0 * organic_edge_px;
}

void fragment() {
	vec2 pos = SCREEN_UV * viewport_px;
	vec2 delta = pos - center_px;
	float d = length(delta);
	float edge = organic_edge(delta);
	float visual_light_radius = max(0.0, light_radius_px + edge * 0.45);
	float visual_gloom_radius = max(visual_light_radius + 1.0, gloom_radius_px + edge);
	if (d <= visual_light_radius) {
		COLOR = vec4(0.0, 0.0, 0.0, 0.0);
	} else if (d <= visual_gloom_radius) {
		float t = smoothstep(visual_light_radius, visual_gloom_radius, d);
		COLOR = mix(gloom_color, darkness_color, t);
	} else {
		COLOR = darkness_color;
	}
}
"""

var _camera: Camera3D
var _target: Node3D
var _rect: ColorRect
var _shadow_root: Node2D
var _material: ShaderMaterial
var _active: bool = true
var _light_radius: float = 0.0
var _gloom_radius: float = 0.0
var _light_radius_px: float = 0.0
var _gloom_radius_px: float = 0.0
var _organic_edge_px: float = 0.0
var _center_px := Vector2.ZERO
var _wall_layout: Array = []
var _extra_occluder_layout: Array = []
var _shadow_polygons: Array = []
var _shadow_debug: Array = []
var _occluder_count: int = 0


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


func set_wall_layout(walls: Array) -> void:
	_wall_layout = []
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		var normalized := _normalized_occluder(wall as Dictionary)
		if not normalized.is_empty():
			_wall_layout.append(normalized)
	_update_shader()


func set_occluder_layout(occluders: Array) -> void:
	_extra_occluder_layout = []
	for occluder in occluders:
		if typeof(occluder) != TYPE_DICTIONARY:
			continue
		var normalized := _normalized_occluder(occluder as Dictionary)
		if not normalized.is_empty():
			_extra_occluder_layout.append(normalized)
	_update_shader()


func get_debug_state() -> Dictionary:
	return {
		"enabled": visible,
		"active": _active,
		"light_radius": _light_radius,
		"gloom_radius": _gloom_radius,
		"light_radius_px": _light_radius_px,
		"gloom_radius_px": _gloom_radius_px,
		"organic_edge_enabled": _organic_edge_enabled(),
		"organic_edge_px": _organic_edge_px,
		"organic_edge_world_amplitude": ORGANIC_EDGE_WORLD_AMPLITUDE,
		"organic_edge_segments": int(ORGANIC_EDGE_SEGMENTS),
		"gloom_alpha": GLOOM_ALPHA,
		"darkness_alpha": DARKNESS_ALPHA,
		"wall_count": _wall_layout.size(),
		"extra_occluder_count": _extra_occluder_layout.size(),
		"occluder_count": _occluder_count,
		"shadow_count": _shadow_debug.size(),
		"shadow_start_offset": SHADOW_START_OFFSET,
		"shadow_wall_height": SHADOW_WALL_HEIGHT,
		"shadow_polygons": _shadow_debug.duplicate(true),
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
	_material.set_shader_parameter("gloom_color", Color(0.22, 0.24, 0.27, GLOOM_ALPHA))
	_material.set_shader_parameter("darkness_color", Color(0.0, 0.0, 0.0, DARKNESS_ALPHA))
	_material.set_shader_parameter("organic_edge_segments", ORGANIC_EDGE_SEGMENTS)
	_material.set_shader_parameter("organic_edge_seed", ORGANIC_EDGE_SEED)
	_rect.material = _material
	add_child(_rect)
	_shadow_root = Node2D.new()
	_shadow_root.name = "FogLOSShadows"
	add_child(_shadow_root)


func _sync_visibility() -> void:
	visible = _active and _light_radius > 0.0


func _update_shader() -> void:
	if _material == null:
		return
	var viewport_size := get_viewport().get_visible_rect().size if get_viewport() != null else Vector2(1.0, 1.0)
	_center_px = _project_target()
	_light_radius_px = _projected_radius(_light_radius)
	_gloom_radius_px = _projected_radius(_gloom_radius)
	_organic_edge_px = _organic_edge_pixel_radius()
	_material.set_shader_parameter("viewport_px", viewport_size)
	_material.set_shader_parameter("center_px", _center_px)
	_material.set_shader_parameter("light_radius_px", _light_radius_px)
	_material.set_shader_parameter("gloom_radius_px", _gloom_radius_px)
	_material.set_shader_parameter("organic_edge_px", _organic_edge_px)
	_update_shadows(viewport_size)


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


func _organic_edge_enabled() -> bool:
	return _active and _light_radius > 0.0 and _organic_edge_px > 0.0


func _organic_edge_pixel_radius() -> float:
	if _light_radius <= 0.0 or _gloom_radius_px <= 0.0:
		return 0.0
	var projected := _projected_radius(ORGANIC_EDGE_WORLD_AMPLITUDE)
	var max_edge := _gloom_radius_px * ORGANIC_EDGE_MAX_GLOOM_FRACTION
	return clampf(projected, ORGANIC_EDGE_MIN_PX, max_edge)


func _update_shadows(viewport_size: Vector2) -> void:
	_shadow_debug = []
	_occluder_count = 0
	if _shadow_root == null:
		return
	var occluders := _combined_occluders()
	if not visible or _gloom_radius <= 0.0 or occluders.is_empty():
		_hide_shadow_polygons()
		return
	var hero_world := _target_world_position()
	var polygons: Array = []
	for occluder in occluders:
		var poly := _shadow_polygon_for_wall(occluder as Dictionary, hero_world, viewport_size)
		if poly.size() < 4:
			continue
		_occluder_count += 1
		polygons.append(poly)
		_shadow_debug.append(_debug_shadow(poly))
	_sync_shadow_polygons(polygons)


func _sync_shadow_polygons(polygons: Array) -> void:
	while _shadow_polygons.size() < polygons.size():
		var node := Polygon2D.new()
		node.color = Color(0.0, 0.0, 0.0, DARKNESS_ALPHA)
		_shadow_root.add_child(node)
		_shadow_polygons.append(node)
	for i in range(_shadow_polygons.size()):
		var node := _shadow_polygons[i] as Polygon2D
		if i < polygons.size():
			node.visible = true
			node.polygon = PackedVector2Array(polygons[i])
		else:
			node.visible = false
			node.polygon = PackedVector2Array()


func _hide_shadow_polygons() -> void:
	for node in _shadow_polygons:
		var polygon := node as Polygon2D
		polygon.visible = false
		polygon.polygon = PackedVector2Array()


func _shadow_polygon_for_wall(wall: Dictionary, hero_world: Vector2, viewport_size: Vector2) -> Array:
	var center := Vector2(float(wall.get("x", 0.0)), float(wall.get("y", 0.0)))
	var size := Vector2(float(wall.get("w", 0.0)), float(wall.get("h", 0.0)))
	if size.x <= 0.0 or size.y <= 0.0:
		return []
	var wall_reach := size.length() * 0.5
	if hero_world.distance_to(center) > _gloom_radius + wall_reach:
		return []
	if _point_inside_wall(hero_world, center, size):
		return []
	var tangent := _tangent_corners(hero_world, _wall_corners(center, size))
	if tangent.size() < 2:
		return []
	var corner_a: Vector2 = tangent[0]
	var corner_b: Vector2 = tangent[1]
	var dir_a := (corner_a - hero_world).normalized()
	var dir_b := (corner_b - hero_world).normalized()
	if dir_a.length() <= 0.0 or dir_b.length() <= 0.0:
		return []
	var edge_offset := clampf(SHADOW_START_OFFSET, SHADOW_EDGE_EPSILON, minf(size.x, size.y) * 0.5)
	var extend := maxf(_gloom_radius * 4.0, hero_world.distance_to(center) + viewport_size.length() / FALLBACK_WORLD_TO_SCREEN + wall_reach)
	var start_a := corner_a + dir_a * edge_offset
	var start_b := corner_b + dir_b * edge_offset
	var end_a := hero_world + dir_a * extend
	var end_b := hero_world + dir_b * extend
	return [
		_project_world_point(start_a, SHADOW_WALL_HEIGHT),
		_project_world_point(end_a),
		_project_world_point(end_b),
		_project_world_point(start_b, SHADOW_WALL_HEIGHT),
	]


func _tangent_corners(hero_world: Vector2, corners: Array) -> Array:
	var entries: Array = []
	for corner in corners:
		var point := corner as Vector2
		entries.append({"angle": atan2(point.y - hero_world.y, point.x - hero_world.x), "corner": point})
	entries.sort_custom(func(a, b): return float(a.get("angle", 0.0)) < float(b.get("angle", 0.0)))
	var max_gap := -1.0
	var gap_index := 0
	for i in range(entries.size()):
		var next_i := (i + 1) % entries.size()
		var a := float(entries[i].get("angle", 0.0))
		var b := float(entries[next_i].get("angle", 0.0))
		var gap := b - a
		if next_i == 0:
			gap += TAU
		if gap > max_gap:
			max_gap = gap
			gap_index = i
	var first: Vector2 = entries[(gap_index + 1) % entries.size()].get("corner", Vector2.ZERO)
	var second: Vector2 = entries[gap_index].get("corner", Vector2.ZERO)
	if first.distance_to(second) <= 0.001:
		return []
	return [first, second]


func _wall_corners(center: Vector2, size: Vector2) -> Array:
	var half := size * 0.5
	return [
		Vector2(center.x - half.x, center.y - half.y),
		Vector2(center.x + half.x, center.y - half.y),
		Vector2(center.x + half.x, center.y + half.y),
		Vector2(center.x - half.x, center.y + half.y),
	]


func _point_inside_wall(point: Vector2, center: Vector2, size: Vector2) -> bool:
	var half := size * 0.5
	return point.x >= center.x - half.x and point.x <= center.x + half.x and point.y >= center.y - half.y and point.y <= center.y + half.y


func _project_world_point(world_point: Vector2, height: float = 0.0) -> Vector2:
	if _camera != null:
		return _camera.unproject_position(Vector3(world_point.x, height, world_point.y))
	return _center_px + (world_point - _target_world_position()) * FALLBACK_WORLD_TO_SCREEN


func _target_world_position() -> Vector2:
	if _target != null:
		return Vector2(_target.global_position.x, _target.global_position.z)
	return Vector2.ZERO


func _combined_occluders() -> Array:
	var occluders := _wall_layout.duplicate()
	for occluder in _extra_occluder_layout:
		occluders.append(occluder)
	return occluders


func _normalized_occluder(occluder: Dictionary) -> Dictionary:
	var pos: Dictionary = occluder.get("position", {})
	var size: Dictionary = occluder.get("size", {})
	var w := float(size.get("x", 0.0))
	var h := float(size.get("y", 0.0))
	if w <= 0.0 or h <= 0.0:
		return {}
	return {
		"x": float(pos.get("x", 0.0)),
		"y": float(pos.get("y", 0.0)),
		"w": w,
		"h": h,
	}


func _debug_shadow(points: Array) -> Dictionary:
	var min_p := points[0] as Vector2
	var max_p := points[0] as Vector2
	var serialized: Array = []
	for point in points:
		var p := point as Vector2
		min_p.x = minf(min_p.x, p.x)
		min_p.y = minf(min_p.y, p.y)
		max_p.x = maxf(max_p.x, p.x)
		max_p.y = maxf(max_p.y, p.y)
		serialized.append({"x": p.x, "y": p.y})
	return {
		"points": serialized,
		"bounds": {
			"min_x": min_p.x,
			"min_y": min_p.y,
			"max_x": max_p.x,
			"max_y": max_p.y,
		},
	}
