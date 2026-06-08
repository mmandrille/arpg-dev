# Unit tests for local/remote co-op snapshot handling (v33).
# Run via: godot --headless --path client --script res://tests/test_coop_client.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_local_and_remote_players_apply_from_snapshot()
	_test_remote_player_delta_and_remove()

	print("[gdtest] PASS: test_coop_client (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	return main


func _test_local_and_remote_players_apply_from_snapshot() -> void:
	var main = _make_main()
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [
			{"player_id": "1001", "role": "host", "connected": true},
			{"player_id": "1002", "role": "guest", "connected": true},
		],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 9, "max_hp": 10, "character_id": "char_host"},
			{"id": "1002", "type": "player", "position": {"x": 4.0, "y": 5.0}, "hp": 8, "max_hp": 10, "character_id": "char_guest"},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	_assert_eq("local player id from snapshot", main.player_id, "1001")
	_assert_vec3("local player anchor position", main.player_anchor.position, Vector3(2.0, 0.0, 3.0))
	_assert_true("remote player entity stored", main.entities.has("1002"))
	_assert_eq("remote entity type", str(main.entities["1002"].get("type", "")), "player")
	_assert_eq("remote character metadata", str(main.entities["1002"].get("character_id", "")), "char_guest")
	_assert_eq("party count", main.party.size(), 2)
	main.free()


func _test_remote_player_delta_and_remove() -> void:
	var main = _make_main()
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 10, "max_hp": 10},
			{"id": "1002", "type": "player", "position": {"x": 4.0, "y": 5.0}, "hp": 10, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	main._apply_delta({
		"events": [
			{"event_type": "player_damaged", "entity_id": "1002", "target_entity_id": "1002", "source_entity_id": "1003", "damage": 2},
		],
		"changes": [
			{"op": "entity_update", "entity": {"id": "1002", "type": "player", "position": {"x": 6.0, "y": 7.0}, "hp": 8, "max_hp": 10}},
		],
	})
	_assert_vec3("remote player authoritative position", (main.entities["1002"]["node"] as Node3D).position, Vector3(6.0, 0.0, 7.0))
	_assert_eq("remote hp updated", int(main.entities["1002"].get("hp", 0)), 8)
	_assert_vec3("local prediction untouched by remote delta", main.predicted_pos, Vector3(2.0, 0.0, 3.0))
	main._apply_delta({"events": [], "changes": [{"op": "entity_remove", "entity_id": "1002"}]})
	_assert_true("remote player removed", not main.entities.has("1002"))
	main.free()


func _assert_true(name: String, cond: bool) -> void:
	if cond:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s" % name)


func _assert_eq(name: String, got, want) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s got=%s want=%s" % [name, str(got), str(want)])


func _assert_vec3(name: String, got: Vector3, want: Vector3) -> void:
	if got.distance_to(want) <= 0.0001:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s got=%s want=%s" % [name, str(got), str(want)])
