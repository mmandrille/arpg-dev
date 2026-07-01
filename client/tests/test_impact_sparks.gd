# Unit tests for combat impact spark presentation.
extends SceneTree

const ImpactSparksScript := preload("res://scripts/impact_sparks.gd")
const MainScript := preload("res://scripts/main.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const GameplayFeedbackPresentationScript := preload("res://scripts/gameplay_feedback_presentation.gd")
const CombatEventPresentationScript := preload("res://scripts/combat_event_presentation.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	CombatFeelConfigScript.reset_for_tests()
	_test_typed_hit_uses_damage_type_color()
	_test_special_outcomes_do_not_spawn()
	_test_damage_event_integration_respects_disabled_monster_impacts()
	_test_terminal_death_skips_monster_reaction_artifacts()
	CombatEventPresentationScript.clear_session()
	CombatFeelConfigScript.reset_for_tests()
	print("[gdtest] PASS: test_impact_sparks (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_typed_hit_uses_damage_type_color() -> void:
	var ev := {"event_type": "monster_damaged", "damage": 4, "damage_type": "fire", "outcome": "hit"}
	_assert_true("hit should spawn", ImpactSparksScript.should_spawn(ev))
	var node := ImpactSparksScript.make_node(ev, Color.WHITE)
	_assert_eq("spark child count", node.get_child_count(), 5)
	_assert_eq("spark damage type meta", str(node.get_meta("damage_type")), "fire")
	var spark := node.get_child(0) as MeshInstance3D
	var mat := spark.material_override as StandardMaterial3D
	_assert_true("spark uses fire color", mat != null and mat.albedo_color.r > mat.albedo_color.g and mat.albedo_color.g > mat.albedo_color.b)
	node.free()


func _test_special_outcomes_do_not_spawn() -> void:
	for outcome in ["miss", "block", "immune"]:
		_assert_true("%s should not spawn" % outcome, not ImpactSparksScript.should_spawn({"event_type": "attack_missed", "outcome": outcome}))


func _test_damage_event_integration_respects_disabled_monster_impacts() -> void:
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
	_assert_eq("disabled monster spark count", monster.find_children("ImpactSparks", "", true, false).size(), 0)
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_terminal_death_skips_monster_reaction_artifacts() -> void:
	var main = MainScript.new()
	main.entities_root = Node3D.new()
	var monster := Node3D.new()
	main.entities_root.add_child(monster)
	var reaction = ModelReactionControllerScript.new(monster, Color("#553322"))
	main.entities["2001"] = {
		"node": monster,
		"type": "monster",
		"hp": 0,
		"controller": null,
		"reaction": reaction,
	}
	root.add_child(main.entities_root)
	GameplayFeedbackPresentationScript.bind_session(main, main.entities)
	main._enter_entity_terminal_death("2001", main.entities["2001"])
	_assert_eq("terminal death skips death flourish", monster.find_children("DeathFlourish", "", true, false).size(), 0)
	main.entities_root.queue_free()
	main.free()


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
