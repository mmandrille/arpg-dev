# Unit tests for combat outcome punch presentation.
extends SceneTree

const CombatOutcomePunchScript := preload("res://scripts/combat_outcome_punch.gd")
const MainScript := preload("res://scripts/main.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_outcome_spawn_rules()
	_test_outcome_nodes()
	_test_special_outcome_integration()
	print("[gdtest] PASS: test_combat_outcome_punch (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_outcome_spawn_rules() -> void:
	for outcome in ["miss", "block", "immune", "crit"]:
		_assert_true("%s should spawn" % outcome, CombatOutcomePunchScript.should_spawn({"outcome": outcome}))
	_assert_true("critical flag should spawn", CombatOutcomePunchScript.should_spawn({"outcome": "hit", "critical": true}))
	_assert_true("normal hit should not spawn", not CombatOutcomePunchScript.should_spawn({"outcome": "hit"}))


func _test_outcome_nodes() -> void:
	var crit := CombatOutcomePunchScript.make_node({"outcome": "crit", "damage": 12, "damage_type": "fire"})
	_assert_eq("crit outcome meta", str(crit.get_meta("outcome")), "crit")
	_assert_eq("crit spark count", crit.get_child_count(), 9)
	crit.free()
	var block := CombatOutcomePunchScript.make_node({"outcome": "block"})
	_assert_eq("block ring exists", block.find_child("OutcomeRing", true, false) != null, true)
	block.free()


func _test_special_outcome_integration() -> void:
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
	main._camera.look_at_from_position(Vector3(4.0, 12.0, 14.0), monster.position, Vector3.UP)
	main._apply_delta({"events": [{
		"event_type": "attack_missed",
		"entity_id": "2001",
		"target_entity_id": "2001",
		"source_entity_id": "1001",
		"outcome": "block",
	}], "changes": []})
	_assert_eq("integrated outcome punch count", monster.find_children(CombatOutcomePunchScript.NODE_NAME, "", true, false).size(), 1)
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
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
