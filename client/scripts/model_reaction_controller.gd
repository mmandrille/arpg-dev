extends RefCounted
class_name ModelReactionController

const HIT_LEAN_RADIANS := 0.22
const HIT_STOP_SECONDS := 0.04
const HIT_LEAN_SECONDS := 0.05
const HIT_RESTORE_SECONDS := 0.12
const HIT_FLASH_BLEND := 0.55
const DEATH_LEAN_RADIANS := 1.35
const DEATH_SECONDS := 0.18
const HIT_DARKEN := 0.45
const DEATH_DARKEN := 0.28
const DEATH_FLOURISH_COLOR := Color("#d9c089")
const HIGHLIGHT_EMISSION := Color("#d4a017")
const HIGHLIGHT_EMISSION_ENERGY := 0.35
const UNRESOLVED_SOURCE := Vector3(1000000000000.0, 1000000000000.0, 1000000000000.0)

var _root: Node3D
var _base_rotation := Vector3.ZERO
var _base_mesh_colors: Dictionary = {}
var _terminal: bool = false
var _highlighted: bool = false
var _last_reaction: String = ""
var _current_tint := Color.WHITE
var _base_tint := Color.WHITE
var _impact_feedback_count: int = 0
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
	for key in _base_mesh_colors.keys():
		var rec: Dictionary = _base_mesh_colors[key]
		rec["color"] = color
		_base_mesh_colors[key] = rec
	if not _terminal:
		_current_tint = color
		_apply_color_scale(1.0)


func set_highlight(on: bool) -> void:
	if _terminal or _highlighted == on:
		return
	_highlighted = on
	_sync_highlight_emission()


func play_hit(source_position: Vector3 = UNRESOLVED_SOURCE, fallback_direction: Vector3 = Vector3.BACK) -> void:
	if _terminal or _root == null:
		return
	_last_reaction = "hit"
	_impact_feedback_count += 1
	_kill_tween(_hit_tween)
	var direction := _reaction_direction(source_position, fallback_direction)
	var target_rotation := _base_rotation + Vector3(direction.z * HIT_LEAN_RADIANS, 0.0, -direction.x * HIT_LEAN_RADIANS)
	_apply_impact_flash()
	_hit_tween = _root.create_tween()
	_hit_tween.tween_interval(HIT_STOP_SECONDS)
	_hit_tween.tween_callback(_apply_color_scale.bind(HIT_DARKEN))
	_hit_tween.tween_property(_root, "rotation", target_rotation, HIT_LEAN_SECONDS)
	_hit_tween.parallel().tween_method(_apply_color_scale, HIT_DARKEN, 1.0, HIT_LEAN_SECONDS + HIT_RESTORE_SECONDS)
	_hit_tween.tween_property(_root, "rotation", _base_rotation, HIT_RESTORE_SECONDS)
	_hit_tween.finished.connect(_on_hit_finished)


func enter_death(source_position: Vector3 = UNRESOLVED_SOURCE, fallback_direction: Vector3 = Vector3.BACK) -> void:
	if _root == null:
		return
	_terminal = true
	_highlighted = false
	_last_reaction = "death"
	_impact_feedback_count += 1
	_kill_tween(_hit_tween)
	_kill_tween(_death_tween)
	var direction := _reaction_direction(source_position, fallback_direction)
	var target_rotation := _base_rotation + Vector3(direction.z * DEATH_LEAN_RADIANS, 0.0, -direction.x * DEATH_LEAN_RADIANS)
	_apply_impact_flash()
	_add_death_flourish(direction)
	_death_tween = _root.create_tween()
	_death_tween.tween_interval(HIT_STOP_SECONDS)
	_death_tween.tween_callback(_apply_color_scale.bind(DEATH_DARKEN))
	_death_tween.tween_property(_root, "rotation", target_rotation, DEATH_SECONDS)


func reset_terminal() -> void:
	_terminal = false
	_highlighted = false
	_last_reaction = ""
	_kill_tween(_hit_tween)
	_kill_tween(_death_tween)
	if _root != null:
		_root.rotation = _base_rotation
		_clear_death_flourish()
	_apply_color_scale(1.0)


func is_terminal() -> bool:
	return _terminal


func dispose() -> void:
	_highlighted = false
	_kill_tween(_hit_tween)
	_kill_tween(_death_tween)
	_hit_tween = null
	_death_tween = null
	_root = null
	_base_mesh_colors.clear()


func get_debug_state() -> Dictionary:
	return {
		"terminal": _terminal,
		"highlighted": _highlighted,
		"last_reaction": _last_reaction,
		"base_tint": _base_tint.to_html(false),
		"current_tint": _current_tint.to_html(false),
		"impact_feedback_count": _impact_feedback_count,
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
		var raw_node = rec.get("node", null)
		if raw_node == null or not is_instance_valid(raw_node):
			continue
		var mesh_node := raw_node as MeshInstance3D
		if mesh_node == null:
			continue
		var mat := mesh_node.material_override as StandardMaterial3D
		if mat == null:
			continue
		var base: Color = rec.get("color", _base_tint)
		mat.albedo_color = Color(base.r * scale, base.g * scale, base.b * scale, base.a)
	_current_tint = Color(_base_tint.r * scale, _base_tint.g * scale, _base_tint.b * scale, _base_tint.a)
	_sync_highlight_emission()


func _sync_highlight_emission() -> void:
	for key in _base_mesh_colors.keys():
		var rec: Dictionary = _base_mesh_colors[key]
		var raw_node = rec.get("node", null)
		if raw_node == null or not is_instance_valid(raw_node):
			continue
		var mesh_node := raw_node as MeshInstance3D
		if mesh_node == null:
			continue
		var mat := mesh_node.material_override as StandardMaterial3D
		if mat == null:
			continue
		if _highlighted and not _terminal:
			mat.emission_enabled = true
			mat.emission = HIGHLIGHT_EMISSION
			mat.emission_energy_multiplier = HIGHLIGHT_EMISSION_ENERGY
		else:
			mat.emission_enabled = false
			mat.emission = Color.BLACK
			mat.emission_energy_multiplier = 0.0


func _apply_impact_flash() -> void:
	for key in _base_mesh_colors.keys():
		var rec: Dictionary = _base_mesh_colors[key]
		var raw_node = rec.get("node", null)
		if raw_node == null or not is_instance_valid(raw_node):
			continue
		var mesh_node := raw_node as MeshInstance3D
		var mat := mesh_node.material_override as StandardMaterial3D
		if mesh_node == null or mat == null:
			continue
		var base: Color = rec.get("color", _base_tint)
		mat.albedo_color = base.lerp(Color.WHITE, HIT_FLASH_BLEND)
	_current_tint = _base_tint.lerp(Color.WHITE, HIT_FLASH_BLEND)


func _add_death_flourish(direction: Vector3) -> void:
	_clear_death_flourish()
	var root := Node3D.new()
	root.name = "DeathFlourish"
	root.position = Vector3(0.0, 0.18, 0.0)
	root.look_at_from_position(root.position, root.position + direction, Vector3.UP)
	for i in range(6):
		var shard := MeshInstance3D.new()
		shard.name = "DeathShard_%d" % i
		var mesh := BoxMesh.new()
		mesh.size = Vector3(0.04, 0.04, 0.34 + float(i % 2) * 0.08)
		shard.mesh = mesh
		var angle := (TAU / 6.0) * float(i)
		shard.position = Vector3(cos(angle) * 0.24, 0.16 + float(i % 3) * 0.045, sin(angle) * 0.24)
		shard.rotation = Vector3(0.4, angle, 0.7)
		shard.material_override = _death_flourish_material(0.68 - float(i) * 0.05)
		root.add_child(shard)
	_root.add_child(root)


func _clear_death_flourish() -> void:
	if _root == null:
		return
	for child in _root.find_children("DeathFlourish", "", false, false):
		_root.remove_child(child)
		child.queue_free()


func _death_flourish_material(alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(DEATH_FLOURISH_COLOR.r, DEATH_FLOURISH_COLOR.g, DEATH_FLOURISH_COLOR.b, alpha)
	mat.emission_enabled = true
	mat.emission = DEATH_FLOURISH_COLOR
	mat.emission_energy_multiplier = 0.55
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	return mat


func _on_hit_finished() -> void:
	if _terminal:
		return
	if _root != null:
		_root.rotation = _base_rotation
	_apply_color_scale(1.0)
	_sync_highlight_emission()


func _kill_tween(tween: Tween) -> void:
	if tween != null and tween.is_valid():
		tween.kill()


func _vec_debug(v: Vector3) -> Dictionary:
	return {"x": v.x, "y": v.y, "z": v.z}
