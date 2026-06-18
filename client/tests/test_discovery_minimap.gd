# Unit tests for the session-local discovery minimap.
extends SceneTree

const DiscoveryMinimapScript := preload("res://scripts/discovery_minimap.gd")
const DiscoveryMinimapStateScript := preload("res://scripts/discovery_minimap_state.gd")

var _failures: int = 0


func _initialize() -> void:
	_test_state_exploration_accumulates()
	_test_state_is_scoped_by_level()
	_test_wall_and_objective_debug()
	_test_quest_path_inactive_without_active_objective()
	_test_points_of_interest_markers()
	_test_widget_defaults_toggle_size_and_opacity()
	_test_widget_cycles_fullscreen_mode()
	_test_widget_session_memory_scope()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_discovery_minimap")
	quit(0)


func _test_state_exploration_accumulates() -> void:
	var state: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
	var first := state.update(0, Vector3.ZERO, 4.0, [], {})
	var first_count := int(first.get("explored_count", 0))
	_assert_true("initial explored cells", first_count > 0)
	var second := state.update(0, Vector3(4.0, 0.0, 0.0), 4.0, [], {})
	_assert_true("moving accumulates cells", int(second.get("explored_count", 0)) > first_count)


func _test_state_is_scoped_by_level() -> void:
	var state: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
	var town := state.update(0, Vector3.ZERO, 4.0, [], {})
	var dungeon := state.update(-1, Vector3(30.0, 0.0, 30.0), 4.0, [], {})
	_assert_eq("level field", int(dungeon.get("level", 0)), -1)
	_assert_eq("new level fresh count", int(dungeon.get("explored_count", 0)), int(town.get("explored_count", 0)))
	var town_again := state.update(0, Vector3.ZERO, 4.0, [], {})
	_assert_eq("town retained count", int(town_again.get("explored_count", 0)), int(town.get("explored_count", 0)))


func _test_wall_and_objective_debug() -> void:
	var state: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
	var chest := Node3D.new()
	chest.position = Vector3(3.0, 0.0, -2.0)
	var mapped := state.update(
		0,
		Vector3.ZERO,
		5.0,
		[{"position": {"x": 1.0, "y": 0.0}, "size": {"x": 2.0, "y": 1.0}}],
		{"chest": {"elite_objective": true, "state": "closed", "node": chest}}
	)
	_assert_eq("known wall count", int(mapped.get("wall_count", 0)), 1)
	var objective: Dictionary = mapped.get("objective", {})
	_assert_true("objective has pin", bool(objective.get("has_pin", false)))
	_assert_eq("objective active", str(objective.get("status", "")), "active")
	var quest_path: Dictionary = mapped.get("quest_path", {})
	_assert_true("quest path active", bool(quest_path.get("has_marker", false)))
	_assert_eq("quest path starts at player", float(quest_path.get("start_x", 0.0)), 0.5)
	_assert_true("quest path points toward objective x", float(quest_path.get("end_x", 0.0)) > 0.5)
	_assert_true("quest path points toward objective y", float(quest_path.get("end_y", 1.0)) < 0.5)
	var marker_counts: Dictionary = mapped.get("marker_counts", {})
	_assert_eq("objective marker count", int(marker_counts.get("objective", 0)), 1)
	chest.free()


func _test_quest_path_inactive_without_active_objective() -> void:
	var state: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
	var completed_chest := Node3D.new()
	completed_chest.position = Vector3(3.0, 0.0, -2.0)
	var completed := state.update(
		0,
		Vector3.ZERO,
		5.0,
		[],
		{"chest": {"elite_objective": true, "state": "open", "node": completed_chest}}
	)
	var completed_path: Dictionary = completed.get("quest_path", {})
	_assert_true("completed quest path inactive", not bool(completed_path.get("has_marker", true)))
	var hidden := state.update(1, Vector3.ZERO, 5.0, [], {})
	var hidden_path: Dictionary = hidden.get("quest_path", {})
	_assert_true("hidden quest path inactive", not bool(hidden_path.get("has_marker", true)))
	completed_chest.free()


func _test_points_of_interest_markers() -> void:
	var state: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
	var stairs := Node3D.new()
	stairs.position = Vector3(2.0, 0.0, 0.0)
	var waypoint := Node3D.new()
	waypoint.position = Vector3(3.0, 0.0, 0.0)
	var service := Node3D.new()
	service.position = Vector3(4.0, 0.0, 0.0)
	var far_service := Node3D.new()
	far_service.position = Vector3(50.0, 0.0, 0.0)
	var mapped := state.update(
		0,
		Vector3.ZERO,
		6.0,
		[],
		{
			"stairs": {"type": "interactable", "interactable_def_id": "stairs_down", "node": stairs},
			"waypoint": {"type": "interactable", "interactable_def_id": "teleporter", "node": waypoint},
			"service": {"type": "interactable", "interactable_def_id": "town_vendor", "node": service},
			"far_service": {"type": "interactable", "interactable_def_id": "town_stash", "node": far_service},
		}
	)
	var counts: Dictionary = mapped.get("marker_counts", {})
	_assert_eq("poi marker total", int(mapped.get("marker_count", 0)), 3)
	_assert_eq("stairs marker", int(counts.get("stairs", 0)), 1)
	_assert_eq("waypoint marker", int(counts.get("waypoint", 0)), 1)
	_assert_eq("service marker", int(counts.get("service", 0)), 1)
	stairs.free()
	waypoint.free()
	service.free()
	far_service.free()


func _test_widget_defaults_toggle_size_and_opacity() -> void:
	var minimap: DiscoveryMinimap = DiscoveryMinimapScript.new()
	get_root().add_child(minimap)
	minimap._build()
	var initial := minimap.get_debug_state()
	_assert_true("default hidden", not bool(initial.get("visible", true)))
	_assert_true("default toggle false", not bool(initial.get("toggle_visible", true)))
	_assert_eq("map width doubled", int(initial.get("map_size_x", 0)), 208)
	_assert_eq("map height doubled", int(initial.get("map_size_y", 0)), 208)
	var opacity := float(initial.get("panel_opacity", 1.0))
	_assert_true("panel opacity transparent", opacity < 0.8 and opacity > 0.45)
	minimap.set_panel_opacity(0.35)
	_assert_true("panel opacity setting applies", absf(float(minimap.get_debug_state().get("panel_opacity", 1.0)) - 0.35) <= 0.001)
	minimap.set_panel_opacity(1.3)
	_assert_true("panel opacity clamps high", absf(float(minimap.get_debug_state().get("panel_opacity", 0.0)) - 1.0) <= 0.001)
	minimap.set_state({
		"level": 0,
		"player_x": 0.0,
		"player_y": 0.0,
		"map_world_radius": 16.0,
		"explored_cells": [{"x": 0, "y": 0}],
		"explored_count": 1,
		"walls": [],
		"wall_count": 0,
		"markers": [{"kind": "service", "label": "vendor", "x": 0.55, "y": 0.5}],
		"marker_count": 1,
		"marker_counts": {"service": 1},
		"objective": {"has_pin": true, "status": "active", "pin_x": 0.6, "pin_y": 0.4},
		"quest_path": {"has_marker": true, "start_x": 0.5, "start_y": 0.5, "end_x": 0.6, "end_y": 0.4, "direction_x": 0.707, "direction_y": -0.707, "angle_radians": -0.785},
	})
	minimap.toggle()
	var visible_state := minimap.get_debug_state()
	_assert_true("toggle visible", bool(visible_state.get("visible", false)))
	_assert_eq("debug explored count", int(visible_state.get("explored_count", 0)), 1)
	_assert_eq("debug marker count", int(visible_state.get("marker_count", 0)), 1)
	_assert_eq("debug service marker count", int(visible_state.get("service_marker_count", 0)), 1)
	_assert_true("debug pin", bool(visible_state.get("has_pin", false)))
	_assert_true("debug quest path", bool(visible_state.get("has_quest_path", false)))
	_assert_eq("debug quest path end x", float(visible_state.get("quest_path_end_x", 0.0)), 0.6)
	minimap.free()


func _test_widget_cycles_fullscreen_mode() -> void:
	var minimap: DiscoveryMinimap = DiscoveryMinimapScript.new()
	get_root().add_child(minimap)
	minimap._build()
	_assert_eq("initial mode", str(minimap.get_debug_state().get("display_mode", "")), "hidden")
	minimap.cycle_display_mode()
	var compact := minimap.get_debug_state()
	_assert_eq("compact mode", str(compact.get("display_mode", "")), "compact")
	_assert_true("compact visible", bool(compact.get("visible", false)))
	_assert_eq("compact width", int(compact.get("map_size_x", 0)), 208)
	_assert_true("compact not fullscreen", not bool(compact.get("full_screen", true)))
	minimap.cycle_display_mode()
	var fullscreen := minimap.get_debug_state()
	_assert_eq("fullscreen mode", str(fullscreen.get("display_mode", "")), "fullscreen")
	_assert_true("fullscreen visible", bool(fullscreen.get("visible", false)))
	_assert_true("fullscreen flag", bool(fullscreen.get("full_screen", false)))
	_assert_true("fullscreen larger", int(fullscreen.get("map_size_x", 0)) > int(compact.get("map_size_x", 0)))
	minimap.cycle_display_mode()
	var hidden := minimap.get_debug_state()
	_assert_eq("cycle hides", str(hidden.get("display_mode", "")), "hidden")
	_assert_true("cycle hidden", not bool(hidden.get("visible", true)))
	minimap.free()


func _test_widget_session_memory_scope() -> void:
	var minimap: DiscoveryMinimap = DiscoveryMinimapScript.new()
	get_root().add_child(minimap)
	minimap._build()
	minimap.sync_session("sess_a")
	minimap.sync(0, Vector3.ZERO, 4.0, [], {})
	var first := minimap.get_debug_state()
	_assert_eq("session key set", str(first.get("session_key", "")), "sess_a")
	var first_count := int(first.get("explored_count", 0))
	_assert_true("session explored", first_count > 0)
	minimap.sync_session("sess_a")
	minimap.cycle_display_mode()
	minimap.sync(0, Vector3(4.0, 0.0, 0.0), 4.0, [], {})
	var retained := minimap.get_debug_state()
	_assert_true("same session retains and grows", int(retained.get("explored_count", 0)) > first_count)
	minimap.sync_session("sess_b")
	var reset := minimap.get_debug_state()
	_assert_eq("new session key", str(reset.get("session_key", "")), "sess_b")
	_assert_eq("new session resets explored", int(reset.get("explored_count", -1)), 0)
	minimap.free()


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
