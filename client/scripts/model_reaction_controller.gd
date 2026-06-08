extends RefCounted
class_name ModelReactionController

const HIT_LEAN_RADIANS := 0.22
const HIT_LEAN_SECONDS := 0.05
const HIT_RESTORE_SECONDS := 0.12
const DEATH_LEAN_RADIANS := 1.35
const DEATH_SECONDS := 0.18
const HIT_DARKEN := 0.45
const DEATH_DARKEN := 0.28
const UNRESOLVED_SOURCE := Vector3(1000000000000.0, 1000000000000.0, 1000000000000.0)

var _root: Node3D
var _base_rotation := Vector3.ZERO
var _base_mesh_colors: Dictionary = {}
var _terminal: bool = false
var _last_reaction: String = ""
var _current_tint := Color.WHITE
var _base_tint := Color.WHITE
var _hit_tween: Tween
var _death_tween: Tween


func _init(root: Node3D, base_tint: Color = Color.WHITE) -> void:
	_root = root
	_base_tint = base_tint
	_current_tint = base_tint
	if _root != null:
		_base_rotation = _root.rotation
		_capture_meshes(_root)


func set_base_tint(color: Color) -> void:
	_base_tint = color
	if not _terminal:
		_current_tint = color
		_apply_color_scale(1.0)


func play_hit(source_position: Vector3 = UNRESOLVED_SOURCE, fallback_direction: Vector3 = Vector3.BACK) -> void:
	if _terminal or _root == null:
		return
	_last_reaction = "hit"
	_kill_tween(_hit_tween)
	var direction := _reaction_direction(source_position, fallback_direction)
	var target_rotation := _base_rotation + Vector3(direction.z * HIT_LEAN_RADIANS, 0.0, -direction.x * HIT_LEAN_RADIANS)
	_apply_color_scale(HIT_DARKEN)
	_hit_tween = _root.create_tween()
	_hit_tween.tween_property(_root, "rotation", target_rotation, HIT_LEAN_SECONDS)
	_hit_tween.parallel().tween_method(_apply_color_scale, HIT_DARKEN, 1.0, HIT_LEAN_SECONDS + HIT_RESTORE_SECONDS)
	_hit_tween.tween_property(_root, "rotation", _base_rotation, HIT_RESTORE_SECONDS)
	_hit_tween.finished.connect(_on_hit_finished)


func enter_death(source_position: Vector3 = UNRESOLVED_SOURCE, fallback_direction: Vector3 = Vector3.BACK) -> void:
	if _root == null:
		return
	_terminal = true
	_last_reaction = "death"
	_kill_tween(_hit_tween)
	_kill_tween(_death_tween)
	var direction := _reaction_direction(source_position, fallback_direction)
	var target_rotation := _base_rotation + Vector3(direction.z * DEATH_LEAN_RADIANS, 0.0, -direction.x * DEATH_LEAN_RADIANS)
	_apply_color_scale(DEATH_DARKEN)
	_death_tween = _root.create_tween()
	_death_tween.tween_property(_root, "rotation", target_rotation, DEATH_SECONDS)


func get_debug_state() -> Dictionary:
	return {
		"terminal": _terminal,
		"last_reaction": _last_reaction,
		"base_tint": _base_tint.to_html(false),
		"current_tint": _current_tint.to_html(false),
		"base_rotation": _vec_debug(_base_rotation),
		"current_rotation": _vec_debug(_root.rotation if _root != null else Vector3.ZERO),
		"mesh_count": _base_mesh_colors.size(),
	}


func _capture_meshes(node: Node) -> void:
	if node is MeshInstance3D:
		var mesh_node := node as MeshInstance3D
		var mat := _material_for(mesh_node)
		mesh_node.material_override = mat
		_base_mesh_colors[mesh_node.get_instance_id()] = {
			"node": mesh_node,
			"color": mat.albedo_color,
		}
	for child in node.get_children():
		_capture_meshes(child)


func _material_for(mesh_node: MeshInstance3D) -> StandardMaterial3D:
	var source = mesh_node.material_override
	if source == null and mesh_node.mesh != null:
		source = mesh_node.mesh.surface_get_material(0)
	var mat: StandardMaterial3D
	if source is StandardMaterial3D:
		mat = (source as StandardMaterial3D).duplicate() as StandardMaterial3D
	else:
		mat = StandardMaterial3D.new()
		mat.albedo_color = _base_tint
	return mat


func _reaction_direction(source_position: Vector3, fallback_direction: Vector3) -> Vector3:
	var direction := fallback_direction
	if source_position != UNRESOLVED_SOURCE and _root != null:
		direction = _root.global_position - source_position
	direction.y = 0.0
	if direction.length() <= 0.001:
		direction = fallback_direction
	direction.y = 0.0
	if direction.length() <= 0.001:
		direction = Vector3.BACK
	return direction.normalized()


func _apply_color_scale(scale: float) -> void:
	for key in _base_mesh_colors.keys():
		var rec: Dictionary = _base_mesh_colors[key]
		var mesh_node := rec.get("node", null) as MeshInstance3D
		if mesh_node == null or not is_instance_valid(mesh_node):
			continue
		var mat := mesh_node.material_override as StandardMaterial3D
		if mat == null:
			continue
		var base: Color = rec.get("color", _base_tint)
		mat.albedo_color = Color(base.r * scale, base.g * scale, base.b * scale, base.a)
	_current_tint = Color(_base_tint.r * scale, _base_tint.g * scale, _base_tint.b * scale, _base_tint.a)


func _on_hit_finished() -> void:
	if _terminal:
		return
	if _root != null:
		_root.rotation = _base_rotation
	_apply_color_scale(1.0)


func _kill_tween(tween: Tween) -> void:
	if tween != null and tween.is_valid():
		tween.kill()


func _vec_debug(v: Vector3) -> Dictionary:
	return {"x": v.x, "y": v.y, "z": v.z}
