## Estimates hero-attached light height from the visible character mesh bounds.
class_name HeroLightSource
extends RefCounted


static func estimate_world_height(character_visual: Node3D, anchor: Node3D, cfg: Dictionary) -> float:
	var fraction := clampf(float(cfg.get("height_fraction", 0.55)), 0.0, 1.0)
	var min_height := maxf(0.0, float(cfg.get("min_height", cfg.get("height_offset", 0.75))))
	var anchor_y := _node_world_y(anchor)
	if character_visual != null and is_instance_valid(character_visual):
		var bounds := mesh_bounds_in_node_space(character_visual)
		if bounds.size.y > 0.001:
			var mesh_height := _mesh_height_world_y(character_visual, bounds, fraction)
			return maxf(anchor_y + min_height, mesh_height)

	if anchor != null and is_instance_valid(anchor):
		return anchor_y + min_height

	return min_height


static func _mesh_height_world_y(character_visual: Node3D, bounds: AABB, fraction: float) -> float:
	var local_point := bounds.position + Vector3(0.0, bounds.size.y * fraction, 0.0)
	if character_visual.is_inside_tree():
		return (character_visual.global_transform * local_point).y

	return _node_world_y(character_visual) + local_point.y * character_visual.scale.y


static func _node_world_y(node: Node3D) -> float:
	if node == null or not is_instance_valid(node):
		return 0.0
	if node.is_inside_tree():
		return node.global_position.y

	return node.position.y


static func local_light_position(character_visual: Node3D, cfg: Dictionary) -> Vector3:
	var fraction := clampf(float(cfg.get("height_fraction", 0.55)), 0.0, 1.0)
	var min_height := maxf(0.0, float(cfg.get("min_height", cfg.get("height_offset", 0.75))))
	if character_visual == null or not is_instance_valid(character_visual):
		return Vector3(0.0, min_height, 0.0)
	var bounds := mesh_bounds_in_node_space(character_visual)
	if bounds.size.y <= 0.001:
		return Vector3(0.0, min_height, 0.0)
	var center := bounds.position + bounds.size * 0.5
	var height := maxf(min_height, bounds.position.y + bounds.size.y * fraction)

	return Vector3(center.x, height, center.z)


static func mesh_bounds_in_node_space(root: Node3D) -> AABB:
	var found := false
	var bounds := AABB()
	for mesh in _mesh_instances(root):
		var mi := mesh as MeshInstance3D
		var local := mi.get_aabb()
		var rel := _relative_transform(root, mi)
		var mesh_bounds := AABB(rel * local.position, Vector3.ZERO)
		for i in range(8):
			mesh_bounds = mesh_bounds.expand(rel * local.get_endpoint(i))
		if not found:
			bounds = mesh_bounds
			found = true
		else:
			bounds = bounds.merge(mesh_bounds)
	if not found:
		return AABB(Vector3(-0.5, 0.0, -0.5), Vector3(1.0, 1.0, 1.0))

	return bounds


static func _relative_transform(root: Node3D, node: Node3D) -> Transform3D:
	if root.is_inside_tree() and node.is_inside_tree():
		return root.global_transform.affine_inverse() * node.global_transform

	return _local_transform_to_ancestor(root, node)


static func _local_transform_to_ancestor(ancestor: Node3D, node: Node3D) -> Transform3D:
	if node == ancestor:
		return Transform3D.IDENTITY
	var parent := node.get_parent()
	if parent == ancestor:
		return node.transform
	if parent is Node3D:
		return _local_transform_to_ancestor(ancestor, parent as Node3D) * node.transform

	return node.transform


static func _mesh_instances(node: Node) -> Array:
	var out: Array = []
	if node is MeshInstance3D:
		out.append(node)
	for child in node.get_children():
		out.append_array(_mesh_instances(child))

	return out
