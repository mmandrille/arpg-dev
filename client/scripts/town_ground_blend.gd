## Town-only ground blend. The shader samples the existing grass texture and uses world-space
## distance to the data-backed plaza/path capsules, so the color transition does not move with UVs.
class_name TownGroundBlend
extends RefCounted

const ShaderResource := preload("res://shaders/town_ground_blend.gdshader")
const TownLoader := preload("res://scripts/town_presentation_loader.gd")
const GroundDetail := preload("res://scripts/town_ground_detail.gd")

const MAX_PATHS := 8


static func material(base: StandardMaterial3D, dressing: Dictionary) -> ShaderMaterial:
	var terrain: Dictionary = dressing.get("terrain", {})
	var shader_mat := ShaderMaterial.new()
	shader_mat.shader = ShaderResource
	shader_mat.set_shader_parameter("ground_texture", base.albedo_texture)
	shader_mat.set_shader_parameter("ground_uv_scale", Vector2(base.uv1_scale.x, base.uv1_scale.y))
	shader_mat.set_shader_parameter("roughness_value", base.roughness)
	shader_mat.set_shader_parameter("soil_color", Color(str(terrain.get("soil_color", "#706d4c"))))
	shader_mat.set_shader_parameter("blend_width_m", float(terrain.get("blend_width_m", 4.0)))
	shader_mat.set_shader_parameter("noise_scale_per_m", float(terrain.get("noise_scale_per_m", 0.45)))
	shader_mat.set_shader_parameter("noise_amplitude_m", float(terrain.get("noise_amplitude_m", 0.45)))
	var center := TownLoader.center()
	shader_mat.set_shader_parameter("plaza_center", center)
	var margin := float(terrain.get("footprint_margin_m", 0.0))
	shader_mat.set_shader_parameter("plaza_radius", float((dressing.get("plaza", {}) as Dictionary).get("radius_m", 0.0)) + margin)
	var frame := {"center": center, "gate": TownLoader.gate_position()}
	var capsules := GroundDetail.region_capsules(dressing, frame)
	var paths := PackedVector4Array()
	var widths := PackedFloat32Array()
	# The first capsule is the plaza disc, already sent through plaza_radius.
	for index in range(1, mini(capsules.size(), MAX_PATHS + 1)):
		var capsule: Dictionary = capsules[index]
		var a: Vector2 = capsule["a"]
		var b: Vector2 = capsule["b"]
		paths.append(Vector4(a.x, a.y, b.x, b.y))
		widths.append(float(capsule["half"]) + margin)
	while paths.size() < MAX_PATHS:
		paths.append(Vector4.ZERO)
		widths.append(0.0)
	shader_mat.set_shader_parameter("path_count", mini(capsules.size() - 1, MAX_PATHS))
	shader_mat.set_shader_parameter("paths", paths)
	shader_mat.set_shader_parameter("path_half_widths", widths)
	return shader_mat
