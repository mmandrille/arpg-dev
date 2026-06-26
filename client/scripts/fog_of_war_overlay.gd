class_name FogOfWarOverlay
extends CanvasLayer

const HeroVisibilityFieldScript := preload("res://scripts/hero_visibility_field.gd")
const FogPresentationLoaderScript := preload("res://scripts/fog_presentation_loader.gd")
const FogLosShadowCacheScript := preload("res://scripts/fog_los_shadow_cache.gd")
const HeroLightSourceScript := preload("res://scripts/hero_light_source.gd")

const FALLBACK_WORLD_TO_SCREEN := 32.0
const ORGANIC_EDGE_MIN_PX := 5.0
const ORGANIC_EDGE_MAX_RADIUS_FRACTION := 0.10
const DARKNESS_FEATHER_MIN_PX := 8.0
const DARKNESS_FEATHER_MAX_RADIUS_FRACTION := 0.18
const SHADER_CODE := """
shader_type canvas_item;
render_mode blend_mix, unshaded;

uniform float perspective_mode = 0.0;
uniform vec2 center_px = vec2(0.0, 0.0);
uniform vec2 viewport_px = vec2(1.0, 1.0);
uniform vec2 hero_world_xz = vec2(0.0, 0.0);
uniform vec3 camera_origin = vec3(0.0, 0.0, 0.0);
uniform mat4 inv_projection;
uniform mat4 inv_view;
uniform float light_radius = 0.0;
uniform float light_radius_px = 0.0;
uniform float darkness_feather_px = 0.0;
uniform float edge_feather_world = 1.15;
uniform float falloff_power = 2.0;
uniform float darkness_alpha = 1.0;
uniform float organic_edge_px = 0.0;
uniform float organic_edge_amp_world = 0.0;
uniform float organic_edge_segments = 18.0;
uniform float organic_edge_seed = 41.0;
uniform float organic_edge_rotation = 0.0;
uniform vec4 darkness_color : source_color = vec4(0.0, 0.0, 0.0, 1.0);
uniform float sample_height_count = 1.0;
uniform float sample_height_1 = 1.25;
uniform float sample_height_2 = 2.5;
uniform float sample_height_3 = 3.75;
uniform float height_sample_max_ground_scale = 2.0;

const float ORGANIC_PI = 3.14159265359;
const float ORGANIC_TAU = 6.28318530718;
const float FAR_WORLD_DIST = 100000.0;

float hash1(float n) {
	return fract(sin(n) * 43758.5453123);
}

float smooth_noise(float x) {
	float i = floor(x);
	float f = fract(x);
	float u = f * f * (3.0 - 2.0 * f);
	return mix(hash1(i), hash1(i + 1.0), u);
}

float organic_edge_screen(vec2 delta_px) {
	if (organic_edge_px <= 0.0 || length(delta_px) <= 0.001) {
		return 0.0;
	}
	float angle = fract((atan(delta_px.y, delta_px.x) + ORGANIC_PI) / ORGANIC_TAU + organic_edge_rotation);
	float low = smooth_noise(angle * organic_edge_segments + organic_edge_seed);
	float high = smooth_noise(angle * organic_edge_segments * 2.17 + organic_edge_seed * 3.31);
	float combined = low * 0.72 + high * 0.28;
	return (combined - 0.5) * 2.0 * organic_edge_px;
}

float organic_edge_world(vec2 delta_xz) {
	if (organic_edge_amp_world <= 0.0 || length(delta_xz) <= 0.001) {
		return 0.0;
	}
	float angle = fract((atan(delta_xz.y, delta_xz.x) + ORGANIC_PI) / ORGANIC_TAU + organic_edge_rotation);
	float low = smooth_noise(angle * organic_edge_segments + organic_edge_seed);
	float high = smooth_noise(angle * organic_edge_segments * 2.17 + organic_edge_seed * 3.31);
	float combined = low * 0.72 + high * 0.28;
	return (combined - 0.5) * 2.0 * organic_edge_amp_world;
}

vec3 world_ray_direction(vec2 screen_px) {
	vec2 uv = screen_px / viewport_px;
	uv.y = 1.0 - uv.y;
	vec2 ndc = uv * 2.0 - 1.0;
	vec4 view = inv_projection * vec4(ndc, 1.0, 1.0);
	view.xyz /= view.w;
	vec3 dir_view = normalize(view.xyz);
	return normalize((inv_view * vec4(dir_view, 0.0)).xyz);
}

vec2 plane_xz_at_height(vec2 screen_px, float plane_y) {
	vec3 dir = world_ray_direction(screen_px);
	if (abs(dir.y) <= 0.0001) {
		return hero_world_xz + vec2(FAR_WORLD_DIST, 0.0);
	}
	float t = (plane_y - camera_origin.y) / dir.y;
	if (t < 0.0) {
		return hero_world_xz + vec2(FAR_WORLD_DIST, 0.0);
	}
	vec3 hit = camera_origin + dir * t;
	return hit.xz;
}

vec2 ground_xz_at_screen(vec2 screen_px) {
	return plane_xz_at_height(screen_px, 0.0);
}

float visibility_from_distance(float dist, float effective_radius, float feather) {
	float normalized = dist / max(0.001, effective_radius);
	float visibility = clamp(1.0 - pow(normalized, falloff_power), 0.0, 1.0);
	if (dist > effective_radius) {
		float feather_t = smoothstep(effective_radius, effective_radius + max(0.001, feather), dist);
		visibility = mix(visibility, 0.0, feather_t);
	}
	return visibility;
}

float visibility_at_world_xz(vec2 world_xz) {
	vec2 delta_xz = world_xz - hero_world_xz;
	float dist = length(delta_xz);
	float edge = organic_edge_world(delta_xz);
	float effective_radius = max(0.001, light_radius + edge * 0.45);
	return visibility_from_distance(dist, effective_radius, edge_feather_world);
}

float visibility_isometric(vec2 screen_px) {
	vec2 delta_px = screen_px - center_px;
	float d = length(delta_px);
	float edge = organic_edge_screen(delta_px);
	float effective_radius_px = max(1.0, light_radius_px + edge * 0.45);
	return visibility_from_distance(d, effective_radius_px, darkness_feather_px);
}

float visibility_perspective(vec2 screen_px) {
	// Canvas fog is transparent in perspective mode — OmniLight3D on the hero provides
	// the physical distance falloff. Shadow polygons (drawn above this rect) handle LOS.
	return 1.0;
}

void fragment() {
	vec2 screen_px = SCREEN_UV * viewport_px;
	float visibility = perspective_mode > 0.5
		? visibility_perspective(screen_px)
		: visibility_isometric(screen_px);
	float alpha = (1.0 - visibility) * darkness_alpha;
	COLOR = vec4(darkness_color.rgb, alpha);
}
"""

var _camera: Camera3D
var _target: Node3D
var _rect: ColorRect
var _shadow_root: Node2D
var _material: ShaderMaterial
var _dungeon_active: bool = true
var _perspective_camera: bool = false
var _light_radius: float = 0.0
var _shadow_reach: float = 0.0
var _center_px := Vector2.ZERO
var _light_radius_px: float = 0.0
var _darkness_feather_px: float = 0.0
var _organic_edge_px: float = 0.0
var _wall_layout: Array = []
var _extra_occluder_layout: Array = []
var _shadow_gloom_polygons: Array = []
var _shadow_polygons: Array = []
var _shadow_debug: Array = []
var _occluder_count: int = 0
var _edge_rotation: float = 0.0
var _edge_rotation_active: bool = false
var _has_last_target_world: bool = false
var _last_target_world := Vector2.ZERO
var _shadow_gloom_color := Color(0.10, 0.11, 0.13, 0.42)
var _shadow_core_color := Color(0.0, 0.0, 0.0, 0.82)
var _shadow_gloom_scale := 1.035
var _point_light: OmniLight3D
var _character_visual: Node3D
var _shadow_cache: FogLosShadowCacheScript = FogLosShadowCacheScript.new()


func _ready() -> void:
	layer = 0
	FogPresentationLoaderScript.ensure_loaded()
	_load_shadow_colors()
	_ensure_rect()
	set_process(true)
	_sync_visibility()


func bind(camera: Camera3D, target: Node3D, character_visual: Node3D = null) -> void:
	_camera = camera
	_target = target
	_character_visual = character_visual
	_reset_motion_tracking()
	_reparent_point_light()
	_update_shader()
	_setup_point_light()


func set_active(active: bool) -> void:
	_dungeon_active = active
	_sync_visibility()
	_sync_point_light()


func set_perspective_camera(perspective: bool) -> void:
	if _perspective_camera != perspective:
		_shadow_cache.hard_invalidate()
	_perspective_camera = perspective
	_update_shader()
	_sync_point_light()
	if _character_visual != null:
		var mode := GeometryInstance3D.SHADOW_CASTING_SETTING_ON if not perspective else GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
		for node in _character_visual.find_children("*", "GeometryInstance3D", true, false):
			(node as GeometryInstance3D).cast_shadow = mode


func set_progression(progression: Dictionary) -> void:
	var derived: Dictionary = progression.get("derived_stats", {})
	set_light_radius(float(derived.get("light_radius", 0.0)))


func set_light_radius(radius: float) -> void:
	var next_radius := maxf(0.0, radius)
	if not is_equal_approx(_light_radius, next_radius):
		_shadow_cache.hard_invalidate()
	_light_radius = next_radius
	_shadow_reach = _light_radius * FogPresentationLoaderScript.shadow_reach_multiplier()
	_sync_visibility()
	_update_shader()
	_sync_point_light()


func set_wall_layout(walls: Array) -> void:
	_wall_layout = HeroVisibilityFieldScript.normalize_wall_layout(walls)
	_shadow_cache.hard_invalidate()
	_update_shader()


func set_occluder_layout(occluders: Array) -> void:
	_extra_occluder_layout = HeroVisibilityFieldScript.normalize_occluder_layout(occluders)
	_shadow_cache.hard_invalidate()
	_update_shader()


func set_performance_throttle(enabled: bool) -> void:
	var cache_cfg: Dictionary = FogPresentationLoaderScript.shadow_cache()
	var frames := int(cache_cfg.get("performance_min_rebuild_interval_frames", 3)) if enabled else 0
	_shadow_cache.set_performance_throttle_frames(frames)


func should_suppress_ambient() -> bool:
	return visible and _dungeon_active and _light_radius > 0.0


func ambient_suppression_params() -> Dictionary:
	if _perspective_camera:
		return FogPresentationLoaderScript.perspective_ambient_suppression()
	return FogPresentationLoaderScript.ambient_suppression()


func get_debug_state() -> Dictionary:
	var organic_cfg: Dictionary = FogPresentationLoaderScript.organic_edge()
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()

	return {
		"enabled": visible,
		"active": _dungeon_active,
		"hero_centered_falloff": not _perspective_camera,
		"world_space_visibility": _perspective_camera,
		"falloff_mode": "point_light" if _perspective_camera else "screen_hero",
		"perspective_camera": _perspective_camera,
		"perspective_sample_heights": _perspective_sample_heights(),
		"height_sample_max_ground_scale": FogPresentationLoaderScript.height_sample_max_ground_scale(),
		"light_radius": _light_radius,
		"gloom_radius": _shadow_reach,
		"shadow_reach": _shadow_reach,
		"falloff_power": FogPresentationLoaderScript.falloff_power(),
		"edge_feather_world": FogPresentationLoaderScript.edge_feather_world(),
		"shadow_reach_multiplier": FogPresentationLoaderScript.shadow_reach_multiplier(),
		"light_radius_px": _light_radius_px,
		"gloom_radius_px": _projected_radius(_shadow_reach),
		"organic_edge_enabled": _organic_edge_enabled(),
		"organic_edge_px": _organic_edge_px,
		"organic_edge_world_amplitude": _organic_edge_world_amplitude(),
		"darkness_feather_px": _darkness_feather_px,
		"darkness_feather_world": FogPresentationLoaderScript.edge_feather_world(),
		"organic_edge_segments": int(organic_cfg.get("segments", 18.0)),
		"organic_edge_rotation": _edge_rotation,
		"organic_edge_rotation_active": _edge_rotation_active,
		"organic_edge_rotation_cycles_per_second": float(organic_cfg.get("rotation_cycles_per_second", 0.12)),
		"darkness_alpha": FogPresentationLoaderScript.darkness_alpha(),
		"shadow_gloom_alpha": _shadow_gloom_color.a,
		"shadow_core_alpha": _shadow_core_color.a,
		"wall_count": _wall_layout.size(),
		"extra_occluder_count": _extra_occluder_layout.size(),
		"occluder_count": _occluder_count,
		"shadow_count": _shadow_debug.size(),
		"shadow_start_offset": float(shadow_cfg.get("start_offset", 0.16)),
		"shadow_wall_height": float(shadow_cfg.get("wall_height", 1.0)),
		"shadow_polygons": _shadow_debug.duplicate(true),
		"center": {"x": _center_px.x, "y": _center_px.y},
		"hero_world": {"x": _target_world_position().x, "y": _target_world_position().y},
		"hero_light_height": _hero_light_world_height(),
		"shadow_cache_valid": _shadow_cache.is_valid(),
		"shadow_cache_hits": _shadow_cache.cache_hits,
		"shadow_rebuild_count": _shadow_cache.rebuild_count,
		"shadow_cache_last_rebuild_reason": _shadow_cache.last_rebuild_reason,
	}


func _process(delta: float) -> void:
	_shadow_cache.tick_frame()
	_update_motion_phase(delta)
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
	_material.set_shader_parameter("darkness_color", Color(0.0, 0.0, 0.0, FogPresentationLoaderScript.darkness_alpha()))
	_rect.material = _material
	add_child(_rect)
	_shadow_root = Node2D.new()
	_shadow_root.name = "FogLOSShadows"
	add_child(_shadow_root)


func _load_shadow_colors() -> void:
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()
	_shadow_gloom_color = _color_from_hex(str(shadow_cfg.get("gloom_color", "#1a1c21")), float(shadow_cfg.get("gloom_alpha", 0.42)))
	_shadow_core_color = _color_from_hex(str(shadow_cfg.get("core_color", "#000000")), float(shadow_cfg.get("core_alpha", 0.82)))
	_shadow_gloom_scale = float(shadow_cfg.get("gloom_scale", 1.035))


func _sync_visibility() -> void:
	visible = _dungeon_active and _light_radius > 0.0


func _update_shader() -> void:
	if _material == null:
		return
	var viewport_size := get_viewport().get_visible_rect().size if get_viewport() != null else Vector2(1.0, 1.0)
	_center_px = _project_target()
	_light_radius_px = _projected_radius(_light_radius)
	_organic_edge_px = _organic_edge_pixel_radius()
	_darkness_feather_px = _darkness_feather_pixel_radius()
	_material.set_shader_parameter("viewport_px", viewport_size)
	_material.set_shader_parameter("perspective_mode", 1.0 if _perspective_camera else 0.0)
	_material.set_shader_parameter("center_px", _center_px)
	_material.set_shader_parameter("hero_world_xz", _target_world_position())
	_material.set_shader_parameter("light_radius", _light_radius)
	_material.set_shader_parameter("light_radius_px", _light_radius_px)
	_material.set_shader_parameter("darkness_feather_px", _darkness_feather_px)
	_material.set_shader_parameter("edge_feather_world", FogPresentationLoaderScript.edge_feather_world())
	_material.set_shader_parameter("falloff_power", FogPresentationLoaderScript.falloff_power())
	_material.set_shader_parameter("darkness_alpha", FogPresentationLoaderScript.darkness_alpha())
	_material.set_shader_parameter("organic_edge_px", _organic_edge_px)
	_material.set_shader_parameter("organic_edge_amp_world", _organic_edge_world_amplitude())
	if _camera != null and _perspective_camera:
		_material.set_shader_parameter("camera_origin", _camera.global_position)
		_material.set_shader_parameter("inv_projection", _camera.get_camera_projection().inverse())
		_material.set_shader_parameter("inv_view", _camera.get_global_transform().affine_inverse())
	var organic_cfg: Dictionary = FogPresentationLoaderScript.organic_edge()
	_material.set_shader_parameter("organic_edge_segments", float(organic_cfg.get("segments", 18.0)))
	_material.set_shader_parameter("organic_edge_seed", float(organic_cfg.get("seed", 41.0)))
	_material.set_shader_parameter("organic_edge_rotation", _edge_rotation)
	_apply_perspective_shader_params()
	_update_shadows(viewport_size)
	_sync_point_light()


func _apply_perspective_shader_params() -> void:
	var heights := _perspective_sample_heights()
	var count := heights.size()
	_material.set_shader_parameter("sample_height_count", float(count))
	_material.set_shader_parameter("sample_height_1", float(heights[1]) if count > 1 else 0.0)
	_material.set_shader_parameter("sample_height_2", float(heights[2]) if count > 2 else 0.0)
	_material.set_shader_parameter("sample_height_3", float(heights[3]) if count > 3 else 0.0)
	_material.set_shader_parameter(
		"height_sample_max_ground_scale",
		FogPresentationLoaderScript.height_sample_max_ground_scale(),
	)


func _perspective_sample_heights() -> Array:
	var cfg: Dictionary = FogPresentationLoaderScript.perspective()
	var raw: Array = cfg.get("sample_heights", [0.0])
	var heights: Array = []
	for value in raw:
		heights.append(maxf(0.0, float(value)))
	if heights.is_empty():
		heights.append(0.0)
	heights.sort()
	var deduped: Array = []
	for height in heights:
		if deduped.is_empty() or absf(float(deduped[-1]) - float(height)) > 0.001:
			deduped.append(height)
	if deduped.size() > 4:
		deduped = deduped.slice(0, 4)

	return deduped


func _update_motion_phase(delta: float) -> void:
	var organic_cfg: Dictionary = FogPresentationLoaderScript.organic_edge()
	var move_epsilon := float(organic_cfg.get("rotation_move_epsilon", 0.006))
	var rotation_speed := float(organic_cfg.get("rotation_cycles_per_second", 0.12))
	var current := _target_world_position()
	if not visible:
		_last_target_world = current
		_has_last_target_world = true
		_edge_rotation_active = false
		return
	if not _has_last_target_world:
		_last_target_world = current
		_has_last_target_world = true
		_edge_rotation_active = false
		return
	var distance := current.distance_to(_last_target_world)
	_last_target_world = current
	_edge_rotation_active = distance > move_epsilon
	if _edge_rotation_active:
		_edge_rotation = fposmod(_edge_rotation + maxf(0.0, delta) * rotation_speed, 1.0)


func _reset_motion_tracking() -> void:
	_has_last_target_world = false
	_last_target_world = Vector2.ZERO
	_edge_rotation_active = false


func _hero_light_world_height() -> float:
	return HeroLightSourceScript.estimate_world_height(_character_visual, _target, FogPresentationLoaderScript.point_light())


func _project_target() -> Vector2:
	if _camera != null and _target != null:
		var pos := _target.global_position
		return _camera.unproject_position(Vector3(pos.x, _hero_light_world_height(), pos.z))

	return get_viewport().get_visible_rect().size * 0.5 if get_viewport() != null else Vector2.ZERO


func _projected_radius(world_radius: float) -> float:
	if world_radius <= 0.0:
		return 0.0
	if _camera != null and _target != null:
		var ground := Vector3(_target.global_position.x, 0.0, _target.global_position.z)
		var center := _camera.unproject_position(ground)
		var edge := _camera.unproject_position(ground + Vector3(world_radius, 0.0, 0.0))
		return maxf(1.0, center.distance_to(edge))

	return world_radius * FALLBACK_WORLD_TO_SCREEN


func _organic_edge_enabled() -> bool:
	if not visible or _light_radius <= 0.0:
		return false
	if _perspective_camera:
		return _organic_edge_world_amplitude() > 0.0

	return _organic_edge_px > 0.0


func ground_xz_at_screen(screen_pos: Vector2) -> Vector2:
	return plane_xz_at_screen(screen_pos, 0.0)


func plane_xz_at_screen(screen_pos: Vector2, plane_y: float) -> Vector2:
	if _camera == null:
		return _target_world_position()
	var ray_origin := _camera.project_ray_origin(screen_pos)
	var ray_dir := _camera.project_ray_normal(screen_pos)
	if absf(ray_dir.y) <= 0.0001:
		return _target_world_position() + Vector2(1e6, 0.0)
	var t := (plane_y - ray_origin.y) / ray_dir.y
	if t < 0.0:
		return _target_world_position() + Vector2(1e6, 0.0)
	var hit := ray_origin + ray_dir * t

	return Vector2(hit.x, hit.z)


func _organic_edge_pixel_radius() -> float:
	var amplitude := _organic_edge_world_amplitude()
	if _light_radius <= 0.0 or _light_radius_px <= 0.0 or amplitude <= 0.0:
		return 0.0
	var projected := _projected_radius(amplitude)
	var max_edge := _light_radius_px * ORGANIC_EDGE_MAX_RADIUS_FRACTION
	return clampf(projected, ORGANIC_EDGE_MIN_PX, max_edge)


func _darkness_feather_pixel_radius() -> float:
	if _light_radius <= 0.0 or _light_radius_px <= 0.0:
		return 0.0
	var projected := _projected_radius(FogPresentationLoaderScript.edge_feather_world())
	var max_feather := _light_radius_px * DARKNESS_FEATHER_MAX_RADIUS_FRACTION
	return clampf(projected, DARKNESS_FEATHER_MIN_PX, max_feather)


func _organic_edge_world_amplitude() -> float:
	if not visible or _light_radius <= 0.0:
		return 0.0
	var organic_cfg: Dictionary = FogPresentationLoaderScript.organic_edge()
	var enabled := bool(organic_cfg.get("enabled_perspective", false)) if _perspective_camera else bool(organic_cfg.get("enabled_isometric", true))
	if not enabled:
		return 0.0

	return float(organic_cfg.get("world_amplitude", 0.65))


func _update_shadows(viewport_size: Vector2) -> void:
	_shadow_debug = []
	_occluder_count = 0
	if _shadow_root == null:
		return
	if _perspective_camera:
		_hide_shadow_polygons()
		return
	var occluders := HeroVisibilityFieldScript.combined_occluders(_wall_layout, _extra_occluder_layout)
	if not visible or _shadow_reach <= 0.0 or occluders.is_empty():
		_hide_shadow_polygons()
		return
	var hero_world := _target_world_position()
	var built := _shadow_cache.resolve(
		_camera,
		hero_world,
		_shadow_reach,
		viewport_size,
		occluders,
		FogPresentationLoaderScript.shadow(),
		_center_px,
	)
	_occluder_count = int(built.get("occluder_count", 0))
	_shadow_debug = built.get("debug", [])
	_sync_shadow_polygons(built.get("polygons", []))


func _sync_shadow_polygons(polygons: Array) -> void:
	while _shadow_gloom_polygons.size() < polygons.size():
		var gloom_node := Polygon2D.new()
		gloom_node.color = _shadow_gloom_color
		_shadow_root.add_child(gloom_node)
		_shadow_gloom_polygons.append(gloom_node)
	while _shadow_polygons.size() < polygons.size():
		var node := Polygon2D.new()
		node.color = _shadow_core_color
		_shadow_root.add_child(node)
		_shadow_polygons.append(node)
	for i in range(_shadow_gloom_polygons.size()):
		var gloom_node := _shadow_gloom_polygons[i] as Polygon2D
		if i < polygons.size():
			gloom_node.visible = true
			gloom_node.polygon = PackedVector2Array(HeroVisibilityFieldScript.expanded_polygon(polygons[i] as Array, _shadow_gloom_scale))
		else:
			gloom_node.visible = false
			gloom_node.polygon = PackedVector2Array()
	for i in range(_shadow_polygons.size()):
		var node := _shadow_polygons[i] as Polygon2D
		if i < polygons.size():
			node.visible = true
			node.polygon = PackedVector2Array(polygons[i])
		else:
			node.visible = false
			node.polygon = PackedVector2Array()


func _hide_shadow_polygons() -> void:
	for node in _shadow_gloom_polygons:
		var gloom_polygon := node as Polygon2D
		gloom_polygon.visible = false
		gloom_polygon.polygon = PackedVector2Array()
	for node in _shadow_polygons:
		var polygon := node as Polygon2D
		polygon.visible = false
		polygon.polygon = PackedVector2Array()


func _target_world_position() -> Vector2:
	if _target != null:
		return Vector2(_target.global_position.x, _target.global_position.z)

	return Vector2.ZERO


func _color_from_hex(hex: String, alpha: float) -> Color:
	var normalized := hex.strip_edges()
	if not normalized.begins_with("#"):
		normalized = "#" + normalized
	var color := Color(normalized)

	return Color(color.r, color.g, color.b, alpha)


func _reparent_point_light() -> void:
	if _point_light == null:
		return
	var parent := _character_visual if _character_visual != null else _target
	if parent == null or _point_light.get_parent() == parent:
		return
	if _point_light.get_parent() != null:
		_point_light.get_parent().remove_child(_point_light)
	parent.add_child(_point_light)


func _setup_point_light() -> void:
	if _point_light != null:
		return
	var parent := _character_visual if _character_visual != null else _target
	if parent == null:
		return
	var cfg: Dictionary = FogPresentationLoaderScript.point_light()
	_point_light = OmniLight3D.new()
	_point_light.name = "HeroPointLight"
	_point_light.light_energy = float(cfg.get("energy", 2.0))
	_point_light.omni_attenuation = float(cfg.get("attenuation", 1.5))
	_point_light.light_color = Color(str(cfg.get("color", "#ffffff")))
	_point_light.shadow_enabled = bool(cfg.get("shadow_enabled", false))
	_point_light.shadow_bias = float(cfg.get("shadow_bias", 0.08))
	_point_light.shadow_normal_bias = float(cfg.get("shadow_normal_bias", 1.2))
	_point_light.position = HeroLightSourceScript.local_light_position(_character_visual, cfg)
	_point_light.omni_range = 0.0
	_point_light.visible = false
	parent.add_child(_point_light)


func _sync_point_light() -> void:
	if _point_light == null:
		return
	var active := _dungeon_active and _perspective_camera and _light_radius > 0.0
	_point_light.visible = active
	if active:
		var cfg: Dictionary = FogPresentationLoaderScript.point_light()
		_point_light.omni_range = maxf(0.1, _light_radius * float(cfg.get("range_multiplier", 1.0)))
		_point_light.shadow_bias = float(cfg.get("shadow_bias", 0.08))
		_point_light.shadow_normal_bias = float(cfg.get("shadow_normal_bias", 1.2))
		_point_light.position = HeroLightSourceScript.local_light_position(_character_visual, cfg)
