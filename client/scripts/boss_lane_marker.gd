class_name BossLaneMarker
extends RefCounted

const MARKER_NAME := "BossLaneMarker"

static func sync(rec: Dictionary, lane: Dictionary) -> void:
	var boss := rec.get("node", null) as Node3D
	if boss == null or not is_instance_valid(boss):
		return
	var count := int(lane.get("count", 0))
	var width := float(lane.get("width", 0.0))
	var length := float(lane.get("length", 0.0))
	var safe_index := int(lane.get("safe_index", -1))
	if count < 3 or count > 9 or width <= 0.0 or width > 10.0 or length <= 0.0 or length > 30.0 or safe_index < 0 or safe_index >= count:
		remove(rec)
		return
	var origin_value = lane.get("origin", {})
	var forward_value = lane.get("forward", {})
	var right_value = lane.get("right", {})
	var danger_value = lane.get("danger_lanes", [])
	if not origin_value is Dictionary or not forward_value is Dictionary or not right_value is Dictionary or not danger_value is Array:
		remove(rec)
		return
	var origin: Dictionary = origin_value
	var forward: Dictionary = forward_value
	var right: Dictionary = right_value
	var dangers: Array[int] = []
	for raw_index in danger_value:
		if typeof(raw_index) != TYPE_INT and typeof(raw_index) != TYPE_FLOAT:
			remove(rec)
			return
		var danger_index := int(raw_index)
		if float(danger_index) != float(raw_index):
			remove(rec)
			return
		dangers.append(danger_index)
	var f := Vector3(float(forward.get("x", 0.0)), 0.0, float(forward.get("y", 0.0)))
	var r := Vector3(float(right.get("x", 0.0)), 0.0, float(right.get("y", 0.0)))
	if absf(f.length() - 1.0) > 0.01 or absf(r.length() - 1.0) > 0.01 or absf(f.dot(r)) > 0.01:
		remove(rec)
		return
	if dangers.size() != count - 1:
		remove(rec)
		return
	for index in range(count):
		if index != safe_index and not dangers.has(index):
			remove(rec)
			return
	var signature := [
		float(origin.get("x", 0.0)), float(origin.get("y", 0.0)),
		f, r, count, width, length, safe_index,
		str(lane.get("safe_color", "#42e5af")), str(lane.get("danger_color", "#ff6838")),
		float(lane.get("safe_intensity", 0.3)), int(lane.get("stage_index", -1)),
		float(lane.get("intensity", 0.35)), dangers.duplicate(),
	]
	var marker := boss.find_child(MARKER_NAME, false, false) as Node3D
	if marker != null and marker.get_child_count() == count + 2 and rec.get("boss_lane_marker_signature", []) == signature and _materials_intact(marker, lane, safe_index, count):
		return
	if marker == null:
		marker = Node3D.new()
		marker.name = MARKER_NAME
		boss.add_child(marker)
		marker.top_level = true
	marker.global_position = Vector3(float(origin.get("x", 0.0)), 0.0, float(origin.get("y", 0.0)))
	marker.basis = Basis(r, Vector3.UP, f)
	for child in marker.get_children():
		marker.remove_child(child)
		child.queue_free()
	for index in range(count):
		var segment := MeshInstance3D.new()
		segment.name = "SafeLane" if index == safe_index else "DangerLane%d" % index
		segment.position = Vector3((float(index) - float(count - 1) / 2.0) * width, 0.035, length / 2.0)
		var mesh := BoxMesh.new()
		mesh.size = Vector3(width, 0.035, length)
		segment.mesh = mesh
		var material := StandardMaterial3D.new()
		var color := Color(str(lane.get("safe_color", "#42e5af")) if index == safe_index else str(lane.get("danger_color", "#ff6838")))
		color.a = clampf(float(lane.get("safe_intensity", 0.3)) if index == safe_index else float(lane.get("intensity", 0.35)), 0.0, 1.0)
		material.albedo_color = color
		material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
		material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
		material.cull_mode = BaseMaterial3D.CULL_DISABLED
		material.render_priority = 1 if index == safe_index else 0
		segment.material_override = material
		marker.add_child(segment)
	var safe_center := (float(safe_index) - float(count - 1) / 2.0) * width
	for side in [-1.0, 1.0]:
		var boundary := MeshInstance3D.new()
		boundary.name = "SafeBoundaryLeft" if side < 0.0 else "SafeBoundaryRight"
		boundary.position = Vector3(safe_center + side * width / 2.0, 0.09, length / 2.0)
		var edge_mesh := BoxMesh.new()
		edge_mesh.size = Vector3(minf(0.1, width * 0.06), 0.04, length)
		boundary.mesh = edge_mesh
		var edge_material := StandardMaterial3D.new()
		edge_material.albedo_color = Color(str(lane.get("safe_color", "#42e5af")))
		edge_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
		boundary.material_override = edge_material
		marker.add_child(boundary)
	rec["has_boss_telegraph_marker"] = true
	rec["telegraph_marker_shape"] = "lanes"
	rec["telegraph_radius"] = length
	rec["telegraph_marker_width"] = width * count
	rec["telegraph_marker_color"] = str(lane.get("danger_color", "#ff6838"))
	rec["telegraph_stage_index"] = int(lane.get("stage_index", -1))
	rec["safe_lane_index"] = safe_index
	rec["telegraph_world_origin"] = origin.duplicate(true)
	rec["telegraph_world_forward"] = forward.duplicate(true)
	rec["telegraph_danger_lanes"] = dangers.duplicate()
	rec["boss_lane_marker_signature"] = signature

static func _materials_intact(marker: Node3D, lane: Dictionary, safe_index: int, count: int) -> bool:
	for index in range(count):
		var segment := marker.get_child(index) as MeshInstance3D
		if segment == null:
			return false
		var material := segment.material_override as StandardMaterial3D
		if material == null:
			return false
		var color := Color(str(lane.get("safe_color", "#42e5af")) if index == safe_index else str(lane.get("danger_color", "#ff6838")))
		color.a = clampf(float(lane.get("safe_intensity", 0.3)) if index == safe_index else float(lane.get("intensity", 0.35)), 0.0, 1.0)
		if material.albedo_color != color:
			return false
	return true

static func remove(rec: Dictionary) -> void:
	var boss := rec.get("node", null) as Node3D
	if boss != null and is_instance_valid(boss):
		var marker := boss.find_child(MARKER_NAME, false, false)
		if marker != null:
			boss.remove_child(marker)
			marker.queue_free()
	rec.erase("telegraph_stage_index")
	rec.erase("safe_lane_index")
	rec.erase("telegraph_world_origin")
	rec.erase("telegraph_world_forward")
	rec.erase("telegraph_danger_lanes")
	rec.erase("boss_lane_marker_signature")
	if str(rec.get("telegraph_marker_shape", "")) == "lanes":
		rec["boss_telegraph_active"] = false
		rec["has_boss_telegraph_marker"] = false
		rec["telegraph_marker_shape"] = ""
		rec["telegraph_marker_width"] = 0.0
		rec["telegraph_radius"] = 0.0
		rec["telegraph_marker_color"] = ""
