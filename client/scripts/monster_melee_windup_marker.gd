class_name MonsterMeleeWindupMarker
extends RefCounted

const TICK_DURATION_S := 0.1

static func sync_from_event(ev: Dictionary, entities: Dictionary) -> void:
	var source_id := str(ev.get("source_entity_id", ""))
	if source_id == "" or not entities.has(source_id):
		return
	var rec: Dictionary = entities[source_id]
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var total_ticks := int(ev.get("total_ticks", 0))
	var attack_style := str(ev.get("attack_style", "melee"))
	_ensure_marker(rec, node, total_ticks, attack_style)
	var ctrl = rec.get("controller", null)
	if ctrl != null:
		ctrl.play_one_shot("attack")


static func clear_for_record(rec: Dictionary) -> void:
	var marker = rec.get("melee_windup_marker", null)
	if marker != null and is_instance_valid(marker):
		marker.queue_free()
	rec["melee_windup_marker"] = null
	rec["has_melee_windup_marker"] = false


static func _ensure_marker(rec: Dictionary, node: Node3D, total_ticks: int, attack_style: String = "melee") -> void:
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
	var duration := maxf(TICK_DURATION_S, float(total_ticks) * TICK_DURATION_S)
	var tween := node.create_tween()
	tween.tween_property(material, "albedo_color:a", 0.0, duration)
	tween.tween_callback(Callable(MonsterMeleeWindupMarker, "_on_marker_fade_done").bind(rec, marker))


static func _on_marker_fade_done(rec: Dictionary, marker: MeshInstance3D) -> void:
	if marker != null and is_instance_valid(marker):
		marker.queue_free()
	if rec.get("melee_windup_marker", null) == marker:
		rec["melee_windup_marker"] = null
		rec["has_melee_windup_marker"] = false
