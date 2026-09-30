# v485: remote co-op players render the hero of their authoritative entity
# `character_class`, and a later snapshot/delta with a different class swaps it.
# Run via: godot --headless --path client --script res://tests/test_remote_player_class.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const RemotePlayerClassSyncScript := preload("res://scripts/remote_player_class_sync.gd")
const BotRemotePlayerAssertionsScript := preload("res://scripts/bot_remote_player_assertions.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	var classes := _two_kit_hero_classes()
	if classes.size() < 2:
		_fail("class presentations define two distinct kit heroes", str(classes))
	else:
		_test_snapshot_and_delta_drive_remote_hero(classes[0], classes[1])
	_test_class_changed_only_for_player_class_updates()
	_test_bot_remote_player_class_expectations()

	if _fail_count > 0:
		print("[gdtest] FAIL: test_remote_player_class (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_remote_player_class (%d assertions)" % _pass_count)
	quit(0)


## Two classes that resolve to different, non-fallback character assets.
func _two_kit_hero_classes() -> Array:
	ClassPresentationsLoaderScript.ensure_loaded()
	var ids: Array = ClassPresentationsLoaderScript._classes.keys()
	ids.sort()
	var picked: Array = []
	var assets: Dictionary = {}
	for id in ids:
		var asset_id := str(ClassPresentationsLoaderScript.resolve(str(id)).get("asset_id", ""))
		if asset_id == ClassPresentationsLoaderScript.FALLBACK_ASSET_ID or assets.has(asset_id):
			continue
		assets[asset_id] = true
		picked.append(str(id))
		if picked.size() == 2:
			break
	return picked


func _test_snapshot_and_delta_drive_remote_hero(first_class: String, second_class: String) -> void:
	var main = _make_main()
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 1.0, "y": 1.0}, "hp": 10, "max_hp": 10, "character_class": second_class},
			{"id": "1002", "type": "player", "position": {"x": 2.0, "y": 2.0}, "hp": 10, "max_hp": 10, "character_class": first_class},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	var rec: Dictionary = main.entities.get("1002", {})
	var node := rec.get("node", null) as Node3D
	_assert_true("remote player record exists", node != null)
	if node == null:
		_free_main(main)
		return
	_assert_eq("snapshot records remote class", str(rec.get("character_class", "")), first_class)
	_assert_eq("snapshot builds remote hero model", _model_scene(node), _class_scene(first_class))
	var controller_before = rec.get("controller", null)

	main._apply_delta({"events": [], "changes": [
		{"op": "entity_update", "entity": {"id": "1002", "type": "player", "position": {"x": 3.0, "y": 2.0}, "hp": 10, "max_hp": 10}},
	]})
	_assert_true("delta without class keeps the controller", rec.get("controller", null) == controller_before)
	_assert_eq("delta without class keeps the model", _model_scene(node), _class_scene(first_class))

	main._apply_delta({"events": [], "changes": [
		{"op": "entity_update", "entity": {"id": "1002", "type": "player", "position": {"x": 3.0, "y": 2.0}, "character_class": second_class}},
	]})
	_assert_true("class swap keeps the entity node", main.entities["1002"].get("node", null) == node)
	_assert_eq("class swap records new class", str(rec.get("character_class", "")), second_class)
	_assert_eq("class swap rebuilds hero model", _model_scene(node), _class_scene(second_class))
	_assert_eq("class swap leaves one model root", _model_root_count(node), 1)
	_assert_eq("class swap updates visual class id", str(node.get("class_id")), second_class)
	_assert_true("class swap rebinds animation controller", rec.get("controller", null) != null and rec.get("controller", null) != controller_before)
	_assert_true("class swap rebuilds reaction controller", rec.get("reaction", null) != null)
	_assert_eq("bot state remote ids", RemotePlayerClassSyncScript.remote_player_ids(main.entities), ["1002"])
	_assert_eq("bot state rendered classes", RemotePlayerClassSyncScript.rendered_classes(main.entities), {"1002": second_class})
	_free_main(main)


func _test_class_changed_only_for_player_class_updates() -> void:
	var player_rec := {"type": "player", "character_class": "a"}
	_assert_true("same class is not a change", not RemotePlayerClassSyncScript.class_changed(player_rec, {"character_class": "a"}))
	_assert_true("missing class is not a change", not RemotePlayerClassSyncScript.class_changed(player_rec, {}))
	_assert_true("different class is a change", RemotePlayerClassSyncScript.class_changed(player_rec, {"character_class": "b"}))
	_assert_true("companion class is ignored", not RemotePlayerClassSyncScript.class_changed({"type": "companion", "character_class": "a"}, {"character_class": "b"}))


func _test_bot_remote_player_class_expectations() -> void:
	var state := {"remote_player_ids": ["1002"], "remote_player_classes": {"1002": "a"}, "local_player_rendered_class": "b"}
	_assert_true("bot matches rendered remote class", BotRemotePlayerAssertionsScript.matches({"at_least": 1, "rendered_classes": ["a"], "local_rendered_class": "b"}, state))
	_assert_true("bot rejects missing remote class", not BotRemotePlayerAssertionsScript.matches({"at_least": 1, "rendered_classes": ["b"]}, state))
	_assert_true("bot rejects wrong local class", not BotRemotePlayerAssertionsScript.matches({"at_least": 1, "local_rendered_class": "a"}, state))
	_assert_true("bot still requires a count", not BotRemotePlayerAssertionsScript.matches({"rendered_classes": ["a"]}, state))


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	get_root().add_child(main.player_anchor)
	get_root().add_child(main.entities_root)
	get_root().add_child(main.walls_root)
	return main


func _free_main(main) -> void:
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _model_scene(node: Node3D) -> String:
	var model := node.find_child("ModelRoot", false, false)
	return model.scene_file_path if model != null else ""


func _model_root_count(node: Node3D) -> int:
	var count := 0
	for child in node.get_children():
		if child.name == "ModelRoot":
			count += 1
	return count


func _class_scene(class_id: String) -> String:
	return ClassPresentationsLoaderScript.packed_scene_for_class(class_id).resource_path


func _assert_true(label: String, cond: bool) -> void:
	if cond:
		_pass_count += 1
	else:
		_fail(label, "expected true")


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail(label, "got=%s want=%s" % [str(got), str(want)])


func _fail(label: String, detail: String) -> void:
	_fail_count += 1
	print("[gdtest] FAIL %s: %s" % [label, detail])
