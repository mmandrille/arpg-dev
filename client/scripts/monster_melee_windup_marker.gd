class_name MonsterMeleeWindupMarker
extends RefCounted

const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")
const TICK_DURATION_S := 0.1

static var _active_markers: Array = []


static func sync_from_event(ev: Dictionary, entities: Dictionary, hero_pos: Vector3 = Vector3.ZERO) -> void:
	var source_id := str(ev.get("source_entity_id", ""))
	if source_id == "" or not entities.has(source_id):
		return
	var rec: Dictionary = entities[source_id]
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	_enforce_cap(hero_pos)
	var total_ticks := int(ev.get("total_ticks", 0))
	var attack_style := str(ev.get("attack_style", "melee"))
	_ensure_marker(rec, node, total_ticks, attack_style, hero_pos)


static func clear_for_record(rec: Dictionary) -> void:
	var marker = rec.get("melee_windup_marker", null)
	if marker != null and is_instance_valid(marker):
		_remove_active(marker)
		marker.queue_free()
	rec["melee_windup_marker"] = null
	rec["has_melee_windup_marker"] = false


static func _enforce_cap(hero_pos: Vector3) -> void:
	var max_count := MainConfigLoaderScript.windup_marker_max_concurrent()
	while _active_markers.size() >= max_count:
		var farthest_idx := 0
		var farthest_dist := -1.0
		for i in range(_active_markers.size()):
			var entry: Dictionary = _active_markers[i]
			var marker_node := entry.get("marker", null) as Node3D
			if marker_node == null or not is_instance_valid(marker_node):
				continue
			var dist_sq := Vector2(marker_node.global_position.x - hero_pos.x, marker_node.global_position.z - hero_pos.z).length_squared()
			if dist_sq > farthest_dist:
				farthest_dist = dist_sq
				farthest_idx = i
		var drop: Dictionary = _active_markers[farthest_idx]
		_active_markers.remove_at(farthest_idx)
		var rec: Dictionary = drop.get("rec", {})
		clear_for_record(rec)


static func _ensure_marker(rec: Dictionary, node: Node3D, total_ticks: int, attack_style: String = "melee", hero_pos: Vector3 = Vector3.ZERO) -> void:
	clear_for_record(rec)
	var marker := MeshInstance3D.new()
	marker.name = "MeleeWindupMarker"
	var mesh := CylinderMesh.new()
	var radius := 0.72 if attack_style != "pounce" else 0.95
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = 0.04
	marker.mesh = mesh
	var material := StandardMaterial3D.new()
	material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	material.albedo_color = Color("#ff8a5c", 0.55) if attack_style != "pounce" else Color("#ffd166", 0.62)
	marker.material_override = material
	marker.position = Vector3(0.0, 0.05, 0.0)
	node.add_child(marker)
	rec["melee_windup_marker"] = marker
	rec["has_melee_windup_marker"] = true
	_active_markers.append({"marker": marker, "rec": rec})
	var duration := maxf(TICK_DURATION_S, float(total_ticks) * TICK_DURATION_S)
	var tween := node.create_tween()
	tween.tween_property(material, "albedo_color:a", 0.0, duration)
	tween.tween_callback(_on_marker_fade_done.bind(rec, marker))


static func _on_marker_fade_done(rec: Dictionary, marker: MeshInstance3D) -> void:
	if marker != null and is_instance_valid(marker):
		_remove_active(marker)
		marker.queue_free()
	if rec.get("melee_windup_marker", null) == marker:
		rec["melee_windup_marker"] = null
		rec["has_melee_windup_marker"] = false


static func _remove_active(marker: MeshInstance3D) -> void:
	for i in range(_active_markers.size()):
		var entry: Dictionary = _active_markers[i]
		if entry.get("marker", null) == marker:
			_active_markers.remove_at(i)
			return


static func reset_for_tests() -> void:
	_active_markers.clear()
