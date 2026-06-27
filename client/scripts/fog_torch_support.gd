## Fog compositor torch uniform wiring (extracted from fog_of_war_overlay.gd).
class_name FogTorchSupport
extends RefCounted

const MAX_TORCHES := 32


static func apply_shader_params(
	material: ShaderMaterial,
	positions: Array,
	hero_world: Vector2,
	light_radius: float,
	feather_world: float,
) -> void:
	if material == null:
		return
	var nearest := _nearest_positions(positions, hero_world, MAX_TORCHES)
	material.set_shader_parameter("torch_count", nearest.size())
	material.set_shader_parameter("torch_light_radius", light_radius)
	material.set_shader_parameter("torch_feather_world", feather_world)
	var packed: Array[Vector2] = []
	for i in range(MAX_TORCHES):
		if i < nearest.size():
			packed.append(nearest[i] as Vector2)
		else:
			packed.append(Vector2.ZERO)
	material.set_shader_parameter("torch_world_xz", packed)


static func _nearest_positions(positions: Array, hero_world: Vector2, limit: int) -> Array:
	if positions.size() <= limit:
		return positions.duplicate()
	var ranked: Array = []
	for raw in positions:
		if typeof(raw) != TYPE_VECTOR2:
			continue
		var pos := raw as Vector2
		ranked.append({"pos": pos, "dist": hero_world.distance_squared_to(pos)})
	ranked.sort_custom(func(a, b): return float(a.get("dist", 0.0)) < float(b.get("dist", 0.0)))
	var out: Array = []
	for i in mini(limit, ranked.size()):
		out.append(ranked[i].get("pos", Vector2.ZERO))

	return out
