extends SceneTree

const LaneMarker := preload("res://scripts/boss_lane_marker.gd")
const BossVisualsContextScript := preload("res://scripts/boss_visuals_context.gd")
const BossVisualsControllerScript := preload("res://scripts/boss_visuals_controller.gd")

var failures := 0

func _initialize() -> void:
	call_deferred("_run")

func _run() -> void:
	var boss := Node3D.new()
	root.add_child(boss)
	boss.position = Vector3(2.0, 0.0, 2.0)
	var rec := {"node": boss}
	var lane := {
		"origin": {"x": 10.0, "y": 12.0},
		"forward": {"x": 1.0, "y": 0.0},
		"right": {"x": 0.0, "y": -1.0},
		"count": 3,
		"width": 2.0,
		"length": 6.0,
		"safe_color": "#42e5af",
		"danger_color": "#ff6838",
		"safe_intensity": 0.3,
		"safe_index": 1,
		"stage_index": 0,
		"intensity": 0.35,
		"danger_lanes": [0.0, 2.0],
	}
	LaneMarker.sync(rec, lane)
	var marker := boss.find_child("BossLaneMarker", false, false) as Node3D
	_check(marker != null, "lane marker exists")
	if marker != null:
		_check(marker.top_level, "lane marker uses world space")
		_check(marker.get_child_count() == 5, "lane strips and safe boundaries are visible")
		_check(marker.global_position.distance_to(Vector3(10.0, 0.0, 12.0)) < 0.001, "locked origin")
		var first_segment := marker.get_child(0)
		LaneMarker.sync(rec, lane)
		_check(marker.get_child(0) == first_segment, "unchanged lane reuses its meshes")
		var safe_segment := marker.find_child("SafeLane", false, false) as MeshInstance3D
		var overwritten_material := StandardMaterial3D.new()
		overwritten_material.albedo_color = Color("#ff6838")
		safe_segment.material_override = overwritten_material
		LaneMarker.sync(rec, lane)
		safe_segment = marker.find_child("SafeLane", false, false) as MeshInstance3D
		var expected_safe_color := Color("#42e5af")
		expected_safe_color.a = 0.3
		_check((safe_segment.material_override as StandardMaterial3D).albedo_color == expected_safe_color, "status tint cannot hide the safe corridor")
		boss.position = Vector3(4.0, 0.0, 4.0)
		_check(marker.global_position.distance_to(Vector3(10.0, 0.0, 12.0)) < 0.001, "boss movement leaves marker fixed")
		lane["stage_index"] = 1
		lane["intensity"] = 0.7
		LaneMarker.sync(rec, lane)
		_check(int(rec.get("telegraph_stage_index", -1)) == 1, "stage updates")
		_check(int(rec.get("safe_lane_index", -1)) == 1, "safe lane remains fixed")
		_check(marker.get_child_count() == 5, "stage retains all lanes and boundaries")
		var invalid_lane := lane.duplicate(true)
		invalid_lane["danger_lanes"] = "bad"
		LaneMarker.sync(rec, invalid_lane)
		_check(not bool(rec.get("has_boss_telegraph_marker", true)), "malformed lane clears marker state")
	LaneMarker.remove(rec)
	await process_frame
	_check(boss.find_child("BossLaneMarker", false, false) == null, "marker cleans up")
	_check(not rec.has("safe_lane_index"), "debug state cleans up")
	_check(not bool(rec.get("has_boss_telegraph_marker", true)), "marker presence clears")
	var context := BossVisualsContextScript.new()
	context.last_server_tick = 21
	var controller := BossVisualsControllerScript.new(context)
	rec["hp"] = 100
	rec["boss_phase"] = {
		"pattern_id": "shifting_bulwark", "phase_index": 1, "phase_kind": "telegraph",
		"started_tick": 20, "duration_ticks": 30, "lane": lane,
	}
	var legacy_marker := MeshInstance3D.new()
	legacy_marker.name = "BossTelegraphMarker"
	boss.add_child(legacy_marker)
	controller.sync_boss_telegraph_marker_from_record(rec)
	_check(boss.find_child("BossLaneMarker", false, false) != null, "snapshot restores warning marker")
	_check(boss.find_child("BossTelegraphMarker", false, false) == null, "lane removes previous shape marker")
	_check(int(rec.get("telegraph_stage_index", -1)) == 1, "snapshot restores stage")
	rec["hp"] = 0
	controller.sync_boss_telegraph_marker_from_record(rec)
	await process_frame
	_check(boss.find_child("BossLaneMarker", false, false) == null, "boss death removes marker")
	boss.queue_free()
	print("[gdtest] PASS: test_boss_lane_marker (%d failed)" % failures)
	quit(1 if failures > 0 else 0)

func _check(ok: bool, label: String) -> void:
	if not ok:
		failures += 1
		push_error("[gdtest] FAIL %s" % label)
