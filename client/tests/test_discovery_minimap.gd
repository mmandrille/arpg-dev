# Unit tests for the session-local discovery minimap.
extends SceneTree

const DiscoveryMinimapScript := preload("res://scripts/discovery_minimap.gd")
const DiscoveryMinimapStateScript := preload("res://scripts/discovery_minimap_state.gd")

var _failures: int = 0


func _initialize() -> void:
	_test_state_exploration_accumulates()
	_test_state_is_scoped_by_level()
	_test_wall_and_objective_debug()
	_test_widget_defaults_toggle_size_and_opacity()
	_test_widget_cycles_fullscreen_mode()
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
	chest.free()


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
	minimap.set_state({
		"level": 0,
		"player_x": 0.0,
		"player_y": 0.0,
		"map_world_radius": 16.0,
		"explored_cells": [{"x": 0, "y": 0}],
		"explored_count": 1,
		"walls": [],
		"wall_count": 0,
		"objective": {"has_pin": true, "status": "active", "pin_x": 0.6, "pin_y": 0.4},
	})
	minimap.toggle()
	var visible_state := minimap.get_debug_state()
	_assert_true("toggle visible", bool(visible_state.get("visible", false)))
	_assert_eq("debug explored count", int(visible_state.get("explored_count", 0)), 1)
	_assert_true("debug pin", bool(visible_state.get("has_pin", false)))
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


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
