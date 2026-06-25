## PickTargetHighlight — emissive outline for interactables/loot under crosshair lock.
class_name PickTargetHighlight
extends RefCounted

const HIGHLIGHT_EMISSION := Color("#d4a017")
const HIGHLIGHT_EMISSION_ENERGY := 0.4

static var _mesh_states: Dictionary = {}


static func set_highlight(root: Node3D, on: bool) -> void:
	if root == null:
		return
	if on:
		_apply_node(root)
	else:
		_clear_node(root)


static func clear_all() -> void:
	while not _mesh_states.is_empty():
		var key = _mesh_states.keys()[0]
		var rec: Dictionary = _mesh_states[key]
		var mesh_node := rec.get("node", null) as MeshInstance3D
		if mesh_node != null and is_instance_valid(mesh_node):
			_restore_mesh(mesh_node)
		else:
			_mesh_states.erase(key)


static func _apply_node(node: Node) -> void:
	if node is MeshInstance3D:
		_apply_mesh(node as MeshInstance3D)
	for child in node.get_children():
		_apply_node(child)


static func _clear_node(node: Node) -> void:
	if node is MeshInstance3D:
		_restore_mesh(node as MeshInstance3D)
	for child in node.get_children():
		_clear_node(child)


static func _apply_mesh(mesh_node: MeshInstance3D) -> void:
	var key := mesh_node.get_instance_id()
	if not _mesh_states.has(key):
		var mat := _material_for(mesh_node)
		mesh_node.material_override = mat
		_mesh_states[key] = {
			"node": mesh_node,
			"emission_enabled": mat.emission_enabled,
			"emission": mat.emission,
			"energy": mat.emission_energy_multiplier,
		}
	var mat := mesh_node.material_override as StandardMaterial3D
	if mat == null:
		return
	mat.emission_enabled = true
	mat.emission = HIGHLIGHT_EMISSION
	mat.emission_energy_multiplier = HIGHLIGHT_EMISSION_ENERGY


static func _restore_mesh(mesh_node: MeshInstance3D) -> void:
	var key := mesh_node.get_instance_id()
	if not _mesh_states.has(key):
		return
	var rec: Dictionary = _mesh_states[key]
	_mesh_states.erase(key)
	if not is_instance_valid(mesh_node):
		return
	var mat := mesh_node.material_override as StandardMaterial3D
	if mat == null:
		return
	mat.emission_enabled = bool(rec.get("emission_enabled", false))
	mat.emission = rec.get("emission", Color.BLACK)
	mat.emission_energy_multiplier = float(rec.get("energy", 0.0))


static func _material_for(mesh_node: MeshInstance3D) -> StandardMaterial3D:
	var source = mesh_node.material_override
	if source == null and mesh_node.mesh != null:
		source = mesh_node.mesh.surface_get_material(0)
	if source is StandardMaterial3D:
		return (source as StandardMaterial3D).duplicate() as StandardMaterial3D
	return StandardMaterial3D.new()
