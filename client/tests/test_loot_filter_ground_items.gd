# Headless tests for loot filter ground item presentation.
# Run via: godot --headless --path client --script res://tests/test_loot_filter_ground_items.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	_test_loot_filter_hides_ground_items_and_pick_colliders()
	if _fail == 0:
		print("[gdtest] PASS: test_loot_filter_ground_items (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_loot_filter_ground_items (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _test_loot_filter_hides_ground_items_and_pick_colliders() -> void:
	var main = _make_main()
	main._upsert_entity({
		"id": "loot_common",
		"type": "loot",
		"item_def_id": "rusty_sword",
		"display_name": "Rusty Sword",
		"rarity": "common",
		"position": {"x": 2.0, "y": 0.0},
	})
	main._upsert_entity({
		"id": "loot_rare",
		"type": "loot",
		"item_def_id": "long_sword",
		"display_name": "Long Sword",
		"rarity": "rare",
		"position": {"x": 3.0, "y": 0.0},
	})
	main._loot_filter.cycle() # Magic+
	main._loot_filter.cycle() # Rare+
	main._refresh_loot_label_visibility()
	_check(not _loot_visible(main, "loot_common"), "Rare+ hides common ground loot")
	_check(not _loot_pickable(main, "loot_common"), "Rare+ disables common loot pick collider")
	_check(_loot_visible(main, "loot_rare"), "Rare+ keeps rare ground loot visible")
	_check(_loot_pickable(main, "loot_rare"), "Rare+ keeps rare loot pick collider enabled")
	main._loot_filter.cycle() # Unique
	main._loot_filter.cycle() # All
	main._refresh_loot_label_visibility()
	_check(_loot_visible(main, "loot_common"), "All restores common ground loot")
	_check(_loot_pickable(main, "loot_common"), "All restores common loot pick collider")
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.queue_free()


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	get_root().add_child(main.player_anchor)
	get_root().add_child(main.entities_root)
	get_root().add_child(main.walls_root)
	return main


func _loot_visible(main, id: String) -> bool:
	var node := main.entities.get(id, {}).get("node", null) as Node3D
	return node != null and node.visible


func _loot_pickable(main, id: String) -> bool:
	var node := main.entities.get(id, {}).get("node", null) as Node3D
	var body := node.find_child("PickBody", true, false) as CollisionObject3D if node != null else null
	return body != null and body.collision_layer != 0 and body.input_ray_pickable


func _check(cond: bool, msg: String) -> void:
	if cond:
		_pass += 1
	else:
		_fail += 1
		push_error("assertion failed: " + msg)
		print("  FAIL: ", msg)
