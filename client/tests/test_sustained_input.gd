# Unit tests for sustained left-click hold state (v27).
# Run via: godot --headless --path client --script res://tests/test_sustained_input.gd
extends SceneTree

const SustainedClickInputScript := preload("res://scripts/sustained_click_input.gd")
const CombatInputBufferScript := preload("res://scripts/combat_input_buffer.gd")
const CombatReachScript := preload("res://scripts/combat_reach.gd")
const CombatStickyTargetScript := preload("res://scripts/combat_sticky_target.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_begin_monster_hold()
	_test_begin_floor_hold()
	_test_begin_directional_attack_hold()
	_test_begin_oneshot_no_hold()
	_test_should_stop_missing_target()
	_test_should_stop_dead_monster()
	_test_can_repeat_move_epsilon()
	_test_clear_resets()
	_test_attack_buffer_queue_replace_and_expire()
	_test_attack_buffer_clear_guards()
	_test_sticky_target_replacement_and_clear_guards()
	_test_combat_reach_uses_equipped_weapon()
	_test_combat_reach_attack_approach_point()

	print("[gdtest] PASS: test_sustained_input (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _test_begin_monster_hold() -> void:
	var hold := SustainedClickInputScript.new()
	_assert_true("monster hold starts", hold.begin_from_pick({"kind": "monster", "target_id": "1002"}))
	_assert_true("monster hold active", hold.active)
	_assert_eq("monster hold mode", hold.mode, "attack")
	_assert_eq("monster sticky target", hold.target_id, "1002")


func _test_begin_floor_hold() -> void:
	var hold := SustainedClickInputScript.new()
	var ground := Vector3(3.0, 0.0, 4.0)
	_assert_true("floor hold starts", hold.begin_from_pick({"kind": "floor", "ground": ground}))
	_assert_true("floor hold active", hold.active)
	_assert_eq("floor hold mode", hold.mode, "move")
	_assert_eq("floor last ground x", hold.last_ground.x, 3.0)
	_assert_eq("floor last ground y", hold.last_ground.y, 4.0)


func _test_begin_directional_attack_hold() -> void:
	var hold := SustainedClickInputScript.new()
	_assert_true("directional attack hold starts", hold.begin_directional_attack())
	_assert_true("directional attack hold active", hold.active)
	_assert_eq("directional attack hold mode", hold.mode, "directional_attack")
	_assert_eq("directional attack target empty", hold.target_id, "")


func _test_begin_oneshot_no_hold() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	_assert_false("loot click does not hold", hold.begin_from_pick({"kind": "oneshot", "target_id": "2001"}))
	_assert_false("oneshot not active", hold.active)


func _test_should_stop_missing_target() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	_assert_true("missing target stops", hold.should_stop(10, {}))


func _test_should_stop_dead_monster() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	var entities := {
		"1002": {"type": "monster", "hp": 0},
	}
	_assert_true("dead monster stops", hold.should_stop(10, entities))


func _test_can_repeat_move_epsilon() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "floor", "ground": Vector3(0.0, 0.0, 0.0)})
	_assert_false("small move below epsilon", hold.can_repeat_move(Vector3(0.1, 0.0, 0.1)))
	_assert_true("large move above epsilon", hold.can_repeat_move(Vector3(0.5, 0.0, 0.0)))


func _test_clear_resets() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	hold.clear()
	_assert_false("clear deactivates", hold.active)
	_assert_eq("clear mode", hold.mode, "")
	_assert_eq("clear target", hold.target_id, "")


func _test_attack_buffer_queue_replace_and_expire() -> void:
	var buffer := CombatInputBufferScript.new()
	buffer.queue_attack("1002", 0.30)
	_assert_true("attack buffer active", buffer.active())
	_assert_eq("attack buffer target", buffer.target_id, "1002")
	buffer.queue_attack("1003", 0.30)
	_assert_eq("attack buffer replaces target", buffer.target_id, "1003")
	buffer.tick(0.10)
	_assert_true("attack buffer remains before expiry", buffer.active())
	buffer.tick(0.25)
	_assert_false("attack buffer expires", buffer.active())
	_assert_eq("attack buffer target cleared", buffer.target_id, "")


func _test_attack_buffer_clear_guards() -> void:
	var buffer := CombatInputBufferScript.new()
	var entities := {
		"1002": {"type": "monster", "hp": 3},
		"1003": {"type": "monster", "hp": 0},
		"loot": {"type": "loot", "hp": 1},
	}
	buffer.queue_attack("1002")
	_assert_false("valid monster does not clear", buffer.should_clear(10, entities))
	_assert_true("dead player clears buffer", buffer.should_clear(0, entities))
	buffer.queue_attack("missing")
	_assert_true("missing target clears buffer", buffer.should_clear(10, entities))
	buffer.queue_attack("1003")
	_assert_true("dead monster clears buffer", buffer.should_clear(10, entities))
	buffer.queue_attack("loot")
	_assert_true("non-monster target clears buffer", buffer.should_clear(10, entities))


func _test_sticky_target_replacement_and_clear_guards() -> void:
	var sticky := CombatStickyTargetScript.new()
	var entities := {
		"1002": {"type": "monster", "hp": 3},
		"1003": {"type": "monster", "hp": 0},
		"loot": {"type": "loot", "hp": 1},
	}
	sticky.set_target("1002")
	_assert_true("sticky target active", sticky.active())
	_assert_false("valid sticky target does not clear", sticky.should_clear(10, entities))
	sticky.set_target("1003")
	_assert_eq("sticky target replaces", sticky.target_id, "1003")
	_assert_true("dead sticky target clears", sticky.should_clear(10, entities))
	sticky.set_target("missing")
	_assert_true("missing sticky target clears", sticky.should_clear(10, entities))
	sticky.set_target("loot")
	_assert_true("non-monster sticky target clears", sticky.should_clear(10, entities))
	sticky.set_target("1002")
	_assert_true("dead player clears sticky target", sticky.should_clear(0, entities))


func _test_combat_reach_uses_equipped_weapon() -> void:
	var player := Node3D.new()
	player.position = Vector3(2.0, 0.0, 5.0)
	var monster := Node3D.new()
	monster.position = Vector3(13.0, 0.0, 5.0)
	var entities := {
		"1002": {"type": "monster", "hp": 10, "node": monster},
	}
	var inventory := [{
		"item_instance_id": "1004",
		"item_def_id": "training_bow",
		"slot": "main_hand",
		"equipped": true,
	}]
	var equipped := {"main_hand": "1004"}
	_assert_true("equipped bow reach covers control lab target", CombatReachScript.target_in_local_attack_range(player, entities, inventory, equipped, "1002"))
	player.free()
	monster.free()


func _test_combat_reach_attack_approach_point() -> void:
	var player := Node3D.new()
	player.position = Vector3(2.0, 0.0, 5.0)
	var monster := Node3D.new()
	monster.position = Vector3(13.0, 0.0, 5.0)
	var entities := {
		"1002": {"type": "monster", "hp": 10, "node": monster},
	}
	var point := CombatReachScript.attack_approach_point(player, entities, [], {}, "1002")
	_assert_true("approach point stops left of target", point.x < monster.position.x)
	_assert_true("approach point stays on target lane", absf(point.z - monster.position.z) < 0.001)
	_assert_true("approach point is inside unarmed reach", point.distance_to(monster.position) < ClientConstants.LOCAL_UNARMED_REACH + ClientConstants.LOCAL_MONSTER_RADIUS)
	player.free()
	monster.free()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
