## Fog compositor torch uniform wiring (extracted from fog_of_war_overlay.gd).
class_name FogTorchSupport
extends RefCounted

const MAX_TORCHES := 8


static func apply_shader_params(material: ShaderMaterial, positions: Array, light_radius: float) -> void:
	if material == null:
		return
	var count := mini(positions.size(), MAX_TORCHES)
	material.set_shader_parameter("torch_count", count)
	material.set_shader_parameter("torch_light_radius", light_radius)
	var packed: Array[Vector2] = []
	for i in range(MAX_TORCHES):
		if i < count:
			packed.append(positions[i] as Vector2)
		else:
			packed.append(Vector2.ZERO)
	material.set_shader_parameter("torch_world_xz", packed)
