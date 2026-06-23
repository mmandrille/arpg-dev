# Unit tests for CrosshairTargetSystem lock gating, click-pick routing, and highlight wiring.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_crosshair_target_system.gd
extends SceneTree

const AimReticleOverlayScript := preload("res://scripts/aim_reticle_overlay.gd")
const CrosshairTargetContextScript := preload("res://scripts/crosshair_target_context.gd")
const CrosshairTargetNamesScript := preload("res://scripts/crosshair_target_names.gd")
const CrosshairTargetSystemScript := preload("res://scripts/crosshair_target_system.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const PickTargetHighlightScript := preload("res://scripts/pick_target_highlight.gd")

var _pass_count: int = 0
var _fail_count: int = 0
var _reach_ids: Array[String] = []
var _ray_pick_id: String = ""


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_locks_reachable_entity()
	await _test_skips_out_of_reach_entity()
	await _test_build_use_pick_routing()
	await _test_reticle_locked_state()
	await _test_reticle_canvas_layer_layout()
	await _test_highlight_toggles_with_lock()
	await _test_interactable_node_highlight()
	await _test_locked_target_shows_name_tag()
	await _test_isometric_mouse_pick_highlights_target()
	await _test_name_resolution_helpers()
	print("[gdtest] PASS: test_crosshair_target_system (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_locks_reachable_entity() -> void:
	_reach_ids = ["monster_a"]
	_ray_pick_id = "monster_a"
	var system = _make_system("monster_a")
	system.tick(get_root().get_viewport(), null)
	_assert_eq("reachable monster locks", system.locked_target_id(), "monster_a")
	system.clear()


func _test_skips_out_of_reach_entity() -> void:
	_reach_ids = []
	_ray_pick_id = "monster_far"
	var system = _make_system("monster_far")
	system.tick(get_root().get_viewport(), null)
	_assert_eq("out-of-reach target does not lock", system.locked_target_id(), "")
	system.clear()


func _test_build_use_pick_routing() -> void:
	_reach_ids = ["monster_a"]
	_ray_pick_id = "monster_a"
	var monster_system = _make_system("monster_a")
	monster_system.tick(get_root().get_viewport(), null)
	_assert_true("living monster use pick is empty", monster_system.build_use_pick().is_empty())
	monster_system.clear()

	_reach_ids = ["loot_a"]
	_ray_pick_id = "loot_a"
	var loot_system = _make_system("loot_a", "loot")
	loot_system.tick(get_root().get_viewport(), null)
	var loot_pick: Dictionary = loot_system.build_use_pick()
	_assert_eq("loot use pick kind", loot_pick.get("kind", ""), "oneshot")
	_assert_eq("loot use pick target", loot_pick.get("target_id", ""), "loot_a")
	loot_system.clear()

	_reach_ids = ["corpse_a"]
	_ray_pick_id = "corpse_a"
	var visual := MeshInstance3D.new()
	visual.mesh = BoxMesh.new()
	var entities := {
		"corpse_a": {
			"type": "monster",
			"hp": 0,
			"node": visual,
			"reaction": ModelReactionControllerScript.new(visual, Color.WHITE),
		},
	}
	var corpse_system = _make_system_with_entities("corpse_a", entities)
	corpse_system.tick(get_root().get_viewport(), null)
	var corpse_pick: Dictionary = corpse_system.build_use_pick()
	_assert_eq("dead monster use pick kind", corpse_pick.get("kind", ""), "oneshot")
	_assert_eq("dead monster use pick target", corpse_pick.get("target_id", ""), "corpse_a")
	corpse_system.clear()


func _test_reticle_locked_state() -> void:
	_reach_ids = ["monster_a"]
	_ray_pick_id = "monster_a"
	var reticle := AimReticleOverlayScript.new()
	get_root().add_child(reticle)
	await process_frame
	var settings := ClientSettingsScript.new("user://test_reticle_lock.json")
	settings.camera_mode = ClientSettingsScript.CAMERA_MODE_CHEST_VIEW
	var system = _make_system("monster_a", "monster", reticle)
	system.tick_runtime(get_root().get_viewport(), null, [], {}, settings, false)
	_assert_eq("target locks in chest view", system.locked_target_id(), "monster_a")
	_assert_true("reticle reports locked", reticle.is_locked())
	system.clear()
	_assert_true("reticle unlocks after clear", not reticle.is_locked())
	reticle.queue_free()


func _test_reticle_canvas_layer_layout() -> void:
	var layer := CanvasLayer.new()
	get_root().add_child(layer)
	var reticle := AimReticleOverlayScript.new()
	layer.add_child(reticle)
	await process_frame
	reticle.visible = true
	await process_frame
	_assert_true("reticle fills viewport width", reticle.size.x > 1.0)
	_assert_true("reticle fills viewport height", reticle.size.y > 1.0)
	layer.queue_free()


func _test_highlight_toggles_with_lock() -> void:
	_reach_ids = ["monster_a"]
	_ray_pick_id = "monster_a"
	var visual := MeshInstance3D.new()
	visual.mesh = BoxMesh.new()
	var reaction := ModelReactionControllerScript.new(visual, Color.WHITE)
	var entities := {
		"monster_a": {
			"type": "monster",
			"hp": 10,
			"node": visual,
			"reaction": reaction,
		},
	}
	var system = _make_system_with_entities("monster_a", entities)
	system.tick(get_root().get_viewport(), null)
	_assert_true("target highlights when locked", reaction.get_debug_state().get("highlighted", false) == true)
	system.clear()
	_assert_true("highlight clears when unlocked", reaction.get_debug_state().get("highlighted", false) == false)


func _test_interactable_node_highlight() -> void:
	_reach_ids = ["chest_a"]
	_ray_pick_id = "chest_a"
	var visual := MeshInstance3D.new()
	visual.mesh = BoxMesh.new()
	var entities := {
		"chest_a": {
			"type": "interactable",
			"hp": 1,
			"node": visual,
			"interactable_def_id": "treasure_chest",
		},
	}
	var system = _make_system_with_entities("chest_a", entities)
	system.tick(get_root().get_viewport(), null)
	_assert_true("interactable mesh has highlight material", visual.material_override is StandardMaterial3D)
	var mat := visual.material_override as StandardMaterial3D
	_assert_true("interactable highlight enables emission", mat.emission_enabled)
	system.clear()
	_assert_true("interactable highlight clears on unlock", not (visual.material_override as StandardMaterial3D).emission_enabled)


func _test_locked_target_shows_name_tag() -> void:
	var layer := CanvasLayer.new()
	get_root().add_child(layer)
	var visual := MeshInstance3D.new()
	visual.mesh = BoxMesh.new()
	get_root().add_child(visual)
	var entities := {
		"chest_a": {
			"type": "interactable",
			"node": visual,
			"interactable_def_id": "treasure_chest",
		},
	}
	_reach_ids = ["chest_a"]
	_ray_pick_id = "chest_a"
	var system = _make_system_with_entities("chest_a", entities, null, layer)
	system.tick(get_root().get_viewport(), null)
	await process_frame
	_assert_eq("locked interactable shows name tag", system.get_name_tag_text(), "Treasure Chest")
	system.clear()
	await process_frame
	_assert_eq("name tag hides on clear", system.get_name_tag_text(), "")
	layer.queue_free()
	visual.queue_free()


func _test_isometric_mouse_pick_highlights_target() -> void:
	var layer := CanvasLayer.new()
	get_root().add_child(layer)
	var visual := MeshInstance3D.new()
	visual.mesh = BoxMesh.new()
	get_root().add_child(visual)
	var entities := {
		"vendor_a": {
			"type": "interactable",
			"node": visual,
			"interactable_def_id": "town_vendor",
		},
	}
	_reach_ids = ["vendor_a"]
	_ray_pick_id = "vendor_a"
	var settings := ClientSettingsScript.new("user://test_iso_crosshair_target.json")
	settings.camera_mode = ClientSettingsScript.CAMERA_MODE_ISOMETRIC
	var system = _make_system_with_entities("vendor_a", entities, null, layer)
	system.tick_runtime(get_root().get_viewport(), null, [], {}, settings, false)
	await process_frame
	_assert_eq("isometric mouse pick locks vendor", system.locked_target_id(), "vendor_a")
	_assert_eq("isometric hover shows name tag", system.get_name_tag_text(), "Town Vendor")
	_assert_true("isometric mode skips center-ray pick", not CrosshairTargetSystemScript.uses_center_ray_pick(settings))
	layer.queue_free()
	visual.queue_free()


func _test_name_resolution_helpers() -> void:
	var interactable_name := CrosshairTargetNamesScript.display_name({
		"type": "interactable",
		"interactable_def_id": "town_vendor",
	})
	_assert_eq("interactable name from rules", interactable_name, "Town Vendor")
	var loot_name := CrosshairTargetNamesScript.display_name({
		"type": "loot",
		"display_name": "Iron Sword",
	})
	_assert_eq("loot uses display_name", loot_name, "Iron Sword")
	var monster_name := CrosshairTargetNamesScript.display_name({
		"type": "monster",
		"hp": 10,
		"monster_def_id": "cave_bat",
	})
	_assert_eq("monster titleizes def id", monster_name, "Cave Bat")


func _make_system(entity_id: String, entity_type: String = "monster", reticle = null):
	var visual := MeshInstance3D.new()
	visual.mesh = BoxMesh.new()
	var reaction := ModelReactionControllerScript.new(visual, Color.WHITE)
	var entities := {
		entity_id: {
			"type": entity_type,
			"hp": 10,
			"node": visual,
			"reaction": reaction,
		},
	}
	return _make_system_with_entities(entity_id, entities, reticle)


func _make_system_with_entities(entity_id: String, entities: Dictionary, reticle = null, name_parent: Node = null):
	var camera := Camera3D.new()
	var anchor := Node3D.new()
	if reticle == null:
		reticle = AimReticleOverlayScript.new()
	var tag_parent := name_parent
	if tag_parent == null:
		tag_parent = CanvasLayer.new()
		get_root().add_child(tag_parent)
	var ctx := CrosshairTargetContextScript.make(
		camera,
		anchor,
		entities,
		[],
		{},
		reticle,
		Callable(self, "_target_in_reach"),
		Callable(self, "_revive_disabled"),
		Callable(self, "_no_loot"),
		Callable(self, "_ground_origin"),
		Callable(self, "_ray_pick"),
		tag_parent,
	)
	var system := CrosshairTargetSystemScript.new()
	system.setup(ctx)
	return system


func _ray_pick(_viewport: Viewport, _world: World3D) -> String:
	return _ray_pick_id


func _target_in_reach(target_id: String) -> bool:
	return _reach_ids.has(target_id)


func _revive_disabled() -> bool:
	return false


func _no_loot(_ground: Vector3) -> String:
	return ""


func _ground_origin() -> Vector3:
	return Vector3.ZERO


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	print("[gdtest] FAIL: %s — got %s want %s" % [label, str(got), str(want)])


func _assert_true(label: String, cond: bool) -> void:
	if cond:
		_pass_count += 1
		return
	_fail_count += 1
	print("[gdtest] FAIL: %s" % label)
