class_name ClassBodyTint
extends RefCounted

const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const DEFAULT_SKIN := Color("#c99666")


static func apply_to_model(root: Node, class_id: String, strength_scale: float = 1.0) -> void:
	if root == null or class_id == "":
		return
	var cfg := ClassPresentationsLoaderScript.body_tint_for_class(class_id)
	if cfg.is_empty():
		return
	var tint_color := Color(str(cfg.get("color", "#ffffff")))
	var strength := clampf(float(cfg.get("strength", 0.0)) * strength_scale, 0.0, 1.0)
	if strength <= 0.0:
		return
	_apply_recursive(root, tint_color, strength)


static func representative_color(class_id: String) -> Color:
	var cfg := ClassPresentationsLoaderScript.body_tint_for_class(class_id)
	if cfg.is_empty():
		return DEFAULT_SKIN
	var tint_color := Color(str(cfg.get("color", "#ffffff")))
	var strength := clampf(float(cfg.get("strength", 0.0)), 0.0, 1.0)
	return DEFAULT_SKIN.lerp(tint_color, strength)


static func _apply_recursive(node: Node, tint_color: Color, strength: float) -> void:
	if node is MeshInstance3D:
		var mesh_node := node as MeshInstance3D
		var mat := _duplicate_surface_material(mesh_node)
		mat.albedo_color = mat.albedo_color.lerp(tint_color, strength)
		mesh_node.material_override = mat
	for child in node.get_children():
		_apply_recursive(child, tint_color, strength)


static func _duplicate_surface_material(mesh_node: MeshInstance3D) -> StandardMaterial3D:
	var source = mesh_node.material_override
	if source == null and mesh_node.mesh != null and mesh_node.mesh.get_surface_count() > 0:
		source = mesh_node.mesh.surface_get_material(0)
	if source is StandardMaterial3D:
		return (source as StandardMaterial3D).duplicate() as StandardMaterial3D
	var mat := StandardMaterial3D.new()
	mat.albedo_color = DEFAULT_SKIN
	return mat
