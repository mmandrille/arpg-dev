extends SceneTree

const FeedbackPresentationScript := preload("res://scripts/gameplay_feedback_presentation.gd")
const ReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const PresentationLoaderScript := preload("res://scripts/monster_death_presentation_loader.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	PresentationLoaderScript.reset_for_tests()
	_test_catalog_and_invalid_fallback()
	await _test_monster_death_event_flash_and_restore()
	await _test_early_reset_restores_baseline()
	await _test_player_and_companion_do_not_flash()
	await _test_other_monster_event_does_not_start_death_flash()
	if _fail_count > 0:
		print("[gdtest] FAIL: test_monster_death_presentation (%d failed / %d checks)" % [_fail_count, _pass_count + _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_monster_death_presentation (%d checks)" % _pass_count)
	quit(0)


func _test_catalog_and_invalid_fallback() -> void:
	var config := PresentationLoaderScript.rim_flash()
	_assert_true("catalog loads enabled death rim flash", bool(config.get("enabled", false)))
	_assert_true("catalog peak strength is positive", float(config.get("peak_strength", 0.0)) > 0.0)
	_assert_true("catalog tint is valid", Color.from_string(str(config.get("tint", "")), Color.TRANSPARENT) != Color.TRANSPARENT)
	config["peak_strength"] = 0.0
	_assert_approx("catalog accessor returns a defensive copy", float(PresentationLoaderScript.rim_flash().get("peak_strength", 0.0)), 0.85)
	var valid_raw := {
		"version": 0,
		"rim_flash": {
			"enabled": true,
			"tint": "#ffe4a3",
			"peak_strength": 0.7,
			"rise_seconds": 0.1,
			"hold_seconds": 0.0,
			"release_seconds": 0.3,
		},
	}
	_assert_true("valid config parses", not PresentationLoaderScript.parse_rim_flash(valid_raw).is_empty())
	var bad_tint: Dictionary = valid_raw.duplicate(true)
	(bad_tint["rim_flash"] as Dictionary)["tint"] = "warm"
	_assert_true("invalid tint disables the optional effect", PresentationLoaderScript.parse_rim_flash(bad_tint).is_empty())
	var bad_strength: Dictionary = valid_raw.duplicate(true)
	(bad_strength["rim_flash"] as Dictionary)["peak_strength"] = 2.0
	_assert_true("out-of-range strength disables the optional effect", PresentationLoaderScript.parse_rim_flash(bad_strength).is_empty())
	_assert_true("missing catalog disables the optional effect", PresentationLoaderScript.parse_rim_flash({}).is_empty())
	var bad_version: Dictionary = valid_raw.duplicate(true)
	bad_version["version"] = "0"
	_assert_true("wrong-typed catalog version disables the optional effect", PresentationLoaderScript.parse_rim_flash(bad_version).is_empty())


func _test_monster_death_event_flash_and_restore() -> void:
	var baseline := _baseline_material()
	var root := _visual_root(baseline)
	get_root().add_child(root)
	await process_frame
	var mesh := root.get_child(0) as MeshInstance3D
	var reaction = ReactionControllerScript.new(root, Color.WHITE)
	var entities := {
		"monster-1": {
			"type": "monster",
			"node": root,
			"reaction": reaction,
		}
	}
	var gate_allowed := FeedbackPresentationScript.entity_combat_impacts_allowed(entities, "monster-1")
	var selected_reaction = FeedbackPresentationScript.reaction_for_entity(entities, "player-1", null, "monster-1")
	_assert_true("monster death is allowed through the existing impact gate", gate_allowed)
	_assert_true("entity lookup resolves the target reaction controller", selected_reaction == reaction)
	FeedbackPresentationScript.play_entity_reaction(
		entities, "player-1", null, null, "monster-1",
		{"event_type": "monster_killed"}, "death", Callable(self, "_world_position")
	)
	var initial_frames := 0
	while float(reaction.get_debug_state().get("death_rim_flash_strength", 0.0)) <= 0.0 and initial_frames < 12:
		await process_frame
		initial_frames += 1
	var flash_config := PresentationLoaderScript.rim_flash()
	var flash_material := mesh.material_override as StandardMaterial3D
	_assert_true("monster death enters terminal reaction", bool(reaction.get_debug_state().get("terminal", false)))
	_assert_true("authoritative monster death starts the rim flash", bool(reaction.get_debug_state().get("death_rim_flash_active", false)))
	_assert_true("flash strength rises above zero while the reaction is active", float(reaction.get_debug_state().get("death_rim_flash_strength", 0.0)) > 0.0)
	_assert_true("flash strength follows the configured peak", float(reaction.get_debug_state().get("death_rim_flash_strength", 0.0)) <= float(PresentationLoaderScript.rim_flash().get("peak_strength", 0.0)))
	_assert_true("rim lighting is active during the flash", flash_material.rim_enabled and flash_material.rim > 0.0)
	var configured_tint := Color.from_string(str(flash_config.get("tint", "")), Color.TRANSPARENT)
	_assert_true("configured tint drives transient emission", flash_material.emission_enabled and flash_material.emission.is_equal_approx(configured_tint) and flash_material.emission_energy_multiplier > 0.0)
	_assert_true("source material remains unchanged", _material_matches(baseline, _baseline_values()))
	var release_frames := 0
	while bool(reaction.get_debug_state().get("death_rim_flash_active", true)) and release_frames < 120:
		await process_frame
		release_frames += 1
	_assert_true("flash becomes inactive after its release", not bool(reaction.get_debug_state().get("death_rim_flash_active", true)))
	_assert_true("flash completes within the bounded frame window", release_frames < 120)
	_assert_true("rim and emission values return to baseline", _material_matches(flash_material, _baseline_values()))
	reaction.dispose()
	root.free()
	await process_frame


func _test_early_reset_restores_baseline() -> void:
	var baseline := _baseline_material()
	var root := _visual_root(baseline)
	get_root().add_child(root)
	await process_frame
	var mesh := root.get_child(0) as MeshInstance3D
	var reaction = ReactionControllerScript.new(root, Color.WHITE)
	reaction.enter_death()
	var config := PresentationLoaderScript.rim_flash()
	_assert_true("explicit death reaction starts", reaction.play_death_rim_flash(config))
	var active_frames := 0
	while float(reaction.get_debug_state().get("death_rim_flash_strength", 0.0)) <= 0.0 and active_frames < 12:
		await process_frame
		active_frames += 1
	var flash_material := mesh.material_override as StandardMaterial3D
	_assert_true("flash is active before reset", bool(reaction.get_debug_state().get("death_rim_flash_active", false)))
	_assert_true("flash reaches configured positive strength before reset", float(reaction.get_debug_state().get("death_rim_flash_strength", 0.0)) > 0.0 and float(reaction.get_debug_state().get("death_rim_flash_strength", 0.0)) <= float(config.get("peak_strength", 0.0)))
	reaction.reset_terminal()
	_assert_true("reset cancels flash state", not bool(reaction.get_debug_state().get("death_rim_flash_active", true)))
	_assert_true("early reset restores all captured properties", _material_matches(mesh.material_override as StandardMaterial3D, _baseline_values()))
	reaction.dispose()
	root.free()
	await process_frame


func _test_player_and_companion_do_not_flash() -> void:
	await _assert_non_monster_has_no_flash("player")
	await _assert_non_monster_has_no_flash("companion")


func _test_other_monster_event_does_not_start_death_flash() -> void:
	var root := _visual_root(_baseline_material())
	get_root().add_child(root)
	await process_frame
	var reaction = ReactionControllerScript.new(root, Color.WHITE)
	var entities := {
		"monster-1": {
			"type": "monster",
			"node": root,
			"reaction": reaction,
		}
	}
	FeedbackPresentationScript.play_entity_reaction(
		entities, "player-1", null, null, "monster-1",
		{"event_type": "monster_damaged"}, "death", Callable(self, "_world_position")
	)
	await process_frame
	_assert_true("non-kill event keeps the ordinary terminal reaction", bool(reaction.get_debug_state().get("terminal", false)))
	_assert_true("non-kill event does not start the monster death flash", not bool(reaction.get_debug_state().get("death_rim_flash_active", false)))
	reaction.dispose()
	root.free()
	await process_frame


func _assert_non_monster_has_no_flash(entity_type: String) -> void:
	var baseline := _baseline_material()
	var root := _visual_root(baseline)
	get_root().add_child(root)
	await process_frame
	var mesh := root.get_child(0) as MeshInstance3D
	var reaction = ReactionControllerScript.new(root, Color.WHITE)
	var entities := {}
	var player_id := "not-the-target"
	var player_anchor: Node3D = null
	var player_reaction = null
	var entity_id := "entity-1"
	if entity_type == "player":
		player_id = entity_id
		player_anchor = root
		player_reaction = reaction
	else:
		entities[entity_id] = {"type": entity_type, "node": root, "reaction": reaction}
	FeedbackPresentationScript.play_entity_reaction(
		entities, player_id, player_anchor, player_reaction, entity_id,
		{"event_type": "player_killed" if entity_type == "player" else "monster_killed"},
		"death", Callable(self, "_world_position")
	)
	await process_frame
	var live_material := mesh.material_override as StandardMaterial3D
	_assert_true("%s death does not start monster rim flash" % entity_type, not bool(reaction.get_debug_state().get("death_rim_flash_active", false)))
	_assert_true("%s material remains at its original rim/emission values" % entity_type, _material_matches(live_material, _baseline_values()))
	reaction.dispose()
	root.free()
	await process_frame


func _baseline_material() -> StandardMaterial3D:
	var material := StandardMaterial3D.new()
	material.albedo_color = Color("#d4b38a")
	material.rim_enabled = false
	material.rim = 0.23
	material.rim_tint = 0.41
	material.emission_enabled = true
	material.emission = Color("#133344")
	material.emission_energy_multiplier = 0.19
	return material


func _baseline_values() -> Dictionary:
	return {
		"rim_enabled": false,
		"rim": 0.23,
		"rim_tint": 0.41,
		"emission_enabled": true,
		"emission": Color("#133344"),
		"emission_energy_multiplier": 0.19,
	}


func _visual_root(material: StandardMaterial3D) -> Node3D:
	var root := Node3D.new()
	var mesh := MeshInstance3D.new()
	var box := BoxMesh.new()
	box.material = material
	mesh.mesh = box
	mesh.material_override = material
	root.add_child(mesh)
	return root


func _material_matches(material: StandardMaterial3D, expected: Dictionary) -> bool:
	return (
		material != null
		and material.rim_enabled == bool(expected["rim_enabled"])
		and is_equal_approx(material.rim, float(expected["rim"]))
		and is_equal_approx(material.rim_tint, float(expected["rim_tint"]))
		and material.emission_enabled == bool(expected["emission_enabled"])
		and material.emission.is_equal_approx(expected["emission"])
		and is_equal_approx(material.emission_energy_multiplier, float(expected["emission_energy_multiplier"]))
	)


func _world_position(node: Node3D) -> Vector3:
	return node.global_position


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_approx(label: String, got: float, expected: float) -> void:
	_assert_true("%s (got %s, expected %s)" % [label, str(got), str(expected)], is_equal_approx(got, expected))
