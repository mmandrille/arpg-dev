# Unit tests: combat hit/death events must not spawn shard markers.
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const GameplayFeedbackPresentationScript := preload("res://scripts/gameplay_feedback_presentation.gd")
const CombatEventPresentationScript := preload("res://scripts/combat_event_presentation.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")
const CombatFeelPresentationLoaderScript := preload("res://scripts/combat_feel_presentation_loader.gd")
const CombatOutcomePunchScript := preload("res://scripts/combat_outcome_punch.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	CombatFeelConfigScript.reset_for_tests()
	_test_damage_event_does_not_spawn_impact_sparks()
	_test_enemy_impact_feedback_toggle_blocks_hit_confirmation()
	_test_death_reaction_does_not_spawn_death_flourish()
	CombatEventPresentationScript.clear_session()
	CombatFeelConfigScript.reset_for_tests()
	print("[gdtest] PASS: test_impact_sparks (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_damage_event_does_not_spawn_impact_sparks() -> void:
	CombatFeelConfigScript.reset_for_tests()
	CombatFeelPresentationLoaderScript.ensure_loaded()
	CombatFeelPresentationLoaderScript.set_enemy_impact_feedback_enabled_for_tests(true)
	var main = MainScript.new()
	main.player_id = "1001"
	main.player_anchor = null
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	main.client_settings = ClientSettingsScript.new()
	main.damage_numbers_layer = CanvasLayer.new()
	main._camera = Camera3D.new()
	var monster := Node3D.new()
	monster.position = Vector3(4.0, 0.0, 4.0)
	main.entities_root.add_child(monster)
	main.entities["2001"] = {"node": monster, "type": "monster", "hp": 5, "controller": null}
	root.add_child(main.entities_root)
	root.add_child(main.walls_root)
	root.add_child(main.damage_numbers_layer)
	root.add_child(main._camera)
	GameplayFeedbackPresentationScript.bind_session(main, main.entities)
	main._camera.look_at_from_position(Vector3(4.0, 12.0, 14.0), monster.position, Vector3.UP)
	main._apply_delta({"events": [{
		"event_type": "monster_damaged",
		"entity_id": "2001",
		"target_entity_id": "2001",
		"source_entity_id": "1001",
		"outcome": "hit",
		"damage_type": "fire",
		"damage": 4
	}], "changes": []})
	_assert_eq("hit does not spawn impact sparks", monster.find_children("ImpactSparks", "", true, false).size(), 0)
	_assert_eq("hit spawns outcome punch", monster.find_children(CombatOutcomePunchScript.NODE_NAME, "", true, false).size(), 1)
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_enemy_impact_feedback_toggle_blocks_hit_confirmation() -> void:
	CombatFeelConfigScript.reset_for_tests()
	CombatFeelPresentationLoaderScript.ensure_loaded()
	CombatFeelPresentationLoaderScript.set_enemy_impact_feedback_enabled_for_tests(false)
	var main = MainScript.new()
	main.player_id = "1001"
	main.player_anchor = null
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	main.client_settings = ClientSettingsScript.new()
	main.damage_numbers_layer = CanvasLayer.new()
	main._camera = Camera3D.new()
	var monster := Node3D.new()
	monster.position = Vector3(4.0, 0.0, 4.0)
	main.entities_root.add_child(monster)
	main.entities["2001"] = {"node": monster, "type": "monster", "hp": 5, "controller": null}
	root.add_child(main.entities_root)
	root.add_child(main.walls_root)
	root.add_child(main.damage_numbers_layer)
	root.add_child(main._camera)
	GameplayFeedbackPresentationScript.bind_session(main, main.entities)
	main._camera.look_at_from_position(Vector3(4.0, 12.0, 14.0), monster.position, Vector3.UP)
	main._apply_delta({"events": [{
		"event_type": "monster_damaged",
		"entity_id": "2001",
		"target_entity_id": "2001",
		"source_entity_id": "1001",
		"outcome": "hit",
		"damage": 4
	}], "changes": []})
	_assert_eq("disabled impact feedback blocks outcome punch", monster.find_children(CombatOutcomePunchScript.NODE_NAME, "", true, false).size(), 0)
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()
	CombatFeelPresentationLoaderScript.set_enemy_impact_feedback_enabled_for_tests(true)


func _test_death_reaction_does_not_spawn_death_flourish() -> void:
	var monster := Node3D.new()
	root.add_child(monster)
	var reaction = ModelReactionControllerScript.new(monster, Color("#553322"))
	reaction.enter_death(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	_assert_eq("death reaction does not spawn flourish", monster.find_children("DeathFlourish", "", true, false).size(), 0)
	monster.queue_free()


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
