# Unit tests for _apply_delta and _apply_snapshot state mutations (v53).
#
# Covers the highest-risk client code flagged in the v53 review: these
# functions had zero unit coverage despite being the primary authoritative-
# boundary enforcement in the client.
#
# Scope: pure state-mutation ops that do not require a scene tree —
# gold_update, inventory_add/remove/update, equipped_update, hotbar_update,
# stash_gold_update, stash_item_add/remove, and snapshot field assignment.
# Entity spawn/update ops are excluded because they require scene-tree nodes.
#
# Run via: godot --headless --path client --script res://tests/test_delta_apply.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	_test_gold_update()
	_test_inventory_add()
	_test_inventory_remove()
	_test_inventory_update()
	_test_equipped_update()
	_test_weapon_set_update()
	_test_weapon_set_inventory_panel_hides_inactive_set_item()
	_test_hotbar_update()
	_test_stash_gold_update()
	_test_stash_item_add()
	_test_stash_item_remove()
	_test_snapshot_gold_and_inventory()
	_test_snapshot_equipped()
	_test_snapshot_weapon_sets()
	_test_snapshot_stash()
	_test_malformed_delta_does_not_crash()
	_test_malformed_envelope_payloads_do_not_crash()

	if _fail == 0:
		print("[gdtest] PASS: test_delta_apply (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: %d/%d assertions failed" % [_fail, _pass + _fail])
		quit(1)


# --- _apply_delta tests -------------------------------------------------------

func _test_gold_update() -> void:
	var m := _new_main()
	m._apply_delta({"changes": [{"op": "gold_update", "gold": 42}]})
	_assert_eq("gold_update sets gold", m.gold, 42)


func _test_inventory_add() -> void:
	var m := _new_main()
	var item := {"item_instance_id": "ii_1", "item_def_id": "rusty_sword"}
	m._apply_delta({"changes": [{"op": "inventory_add", "item": item}]})
	_assert_eq("inventory_add appends item", m.inventory.size(), 1)
	_assert_eq("inventory_add preserves item_def_id",
		str(m.inventory[0].get("item_def_id", "")), "rusty_sword")


func _test_inventory_remove() -> void:
	var m := _new_main()
	m.inventory = [{"item_instance_id": "ii_1"}, {"item_instance_id": "ii_2"}]
	m._apply_delta({"changes": [{"op": "inventory_remove", "item_instance_id": "ii_1"}]})
	_assert_eq("inventory_remove shrinks inventory", m.inventory.size(), 1)
	_assert_eq("inventory_remove keeps other item",
		str(m.inventory[0].get("item_instance_id", "")), "ii_2")


func _test_inventory_update() -> void:
	var m := _new_main()
	m.inventory = [{"item_instance_id": "ii_1", "hp": 10}]
	m._apply_delta({"changes": [{"op": "inventory_update", "item": {"item_instance_id": "ii_1", "hp": 5}}]})
	_assert_eq("inventory_update patches hp",
		int(m.inventory[0].get("hp", -1)), 5)


func _test_equipped_update() -> void:
	var m := _new_main()
	m._apply_delta({"changes": [
		{"op": "equipped_update", "slot": "main_hand", "item_instance_id": "ii_9"}
	]})
	_assert_eq("equipped_update sets slot",
		str(m.equipped.get("main_hand", "")), "ii_9")


func _test_weapon_set_update() -> void:
	var m := _new_main()
	m._apply_delta({"changes": [
		{"op": "weapon_set_update", "active_weapon_set": 1, "weapon_sets": [
			{"index": 0, "main_hand": "ii_1", "off_hand": null},
			{"index": 1, "main_hand": "ii_2", "off_hand": null},
		]}
	]})
	_assert_eq("weapon_set_update active", m.active_weapon_set, 1)
	_assert_eq("weapon_set_update set count", m.weapon_sets.size(), 2)
	_assert_eq("weapon_set_update set 2 main",
		str((m.weapon_sets[1] as Dictionary).get("main_hand", "")), "ii_2")


func _test_weapon_set_inventory_panel_hides_inactive_set_item() -> void:
	var m := _new_main()
	m.inventory_panel = InventoryPanelScript.new()
	m.add_child(m.inventory_panel)
	m.inventory = [
		{"item_instance_id": "ii_1", "item_def_id": "long_sword", "slot": "main_hand", "equipped": true},
		{"item_instance_id": "ii_2", "item_def_id": "bow", "slot": "main_hand", "equipped": true},
	]
	m.equipped = {"main_hand": "ii_1"}
	m.active_weapon_set = 0
	m.weapon_sets = [
		{"index": 0, "main_hand": "ii_1", "off_hand": null},
		{"index": 1, "main_hand": "ii_2", "off_hand": null},
	]
	m._refresh_inventory_ui()
	_assert_eq("inactive set equipped item hidden from bag",
		int(m.inventory_panel.get_debug_state().get("visible_bag_count", -1)), 0)


func _test_hotbar_update() -> void:
	var m := _new_main()
	m._apply_delta({"changes": [
		{"op": "hotbar_update", "slot_index": 0, "item_instance_id": "ii_7"}
	]})
	_assert_eq("hotbar_update sets slot 0",
		str(m.hotbar[0].get("item_instance_id", "")), "ii_7")


func _test_stash_gold_update() -> void:
	var m := _new_main()
	m._apply_delta({"changes": [{"op": "stash_gold_update", "stash_gold": 100}]})
	_assert_eq("stash_gold_update sets stash_gold", m.stash_gold, 100)


func _test_stash_item_add() -> void:
	var m := _new_main()
	var item := {"stash_item_id": "si_1", "item_def_id": "red_potion"}
	m._apply_delta({"changes": [{"op": "stash_item_add", "item": item}]})
	_assert_eq("stash_item_add appends item", m.stash_items.size(), 1)
	_assert_eq("stash_item_add preserves def",
		str(m.stash_items[0].get("item_def_id", "")), "red_potion")


func _test_stash_item_remove() -> void:
	var m := _new_main()
	m.stash_items = [{"stash_item_id": "si_1"}, {"stash_item_id": "si_2"}]
	m._apply_delta({"changes": [{"op": "stash_item_remove", "stash_item_id": "si_1"}]})
	_assert_eq("stash_item_remove shrinks stash", m.stash_items.size(), 1)
	_assert_eq("stash_item_remove keeps other",
		str(m.stash_items[0].get("stash_item_id", "")), "si_2")


# --- _apply_snapshot tests ----------------------------------------------------

func _test_snapshot_gold_and_inventory() -> void:
	var m := _new_main()
	var snap := {
		"local_player_id": "p1",
		"gold": 77,
		"inventory": [{"item_instance_id": "ii_3", "item_def_id": "ring"}],
		"equipped": {},
		"hotbar": [],
		"inventory_rows": 3,
		"inventory_capacity": 15,
		"stash_items": [],
		"stash_gold": 0,
		"stash_capacity": 50,
		"hotbar_capacity": 2,
		"character_progression": {},
		"skill_progression": {},
		"skill_cooldowns": [],
		"entities": [],
		"events": [],
	}
	m._apply_snapshot(snap)
	_assert_eq("snapshot sets gold", m.gold, 77)
	_assert_eq("snapshot sets inventory size", m.inventory.size(), 1)
	_assert_eq("snapshot inventory item def",
		str(m.inventory[0].get("item_def_id", "")), "ring")


func _test_snapshot_equipped() -> void:
	var m := _new_main()
	var snap := _base_snapshot()
	snap["equipped"] = {"main_hand": "ii_5"}
	m._apply_snapshot(snap)
	_assert_eq("snapshot sets equipped slot",
		str(m.equipped.get("main_hand", "")), "ii_5")


func _test_snapshot_weapon_sets() -> void:
	var m := _new_main()
	var snap := _base_snapshot()
	snap["active_weapon_set"] = 1
	snap["weapon_sets"] = [
		{"index": 0, "main_hand": "ii_5", "off_hand": null},
		{"index": 1, "main_hand": "ii_6", "off_hand": null},
	]
	m._apply_snapshot(snap)
	_assert_eq("snapshot active weapon set", m.active_weapon_set, 1)
	_assert_eq("snapshot weapon set 1 main",
		str((m.weapon_sets[0] as Dictionary).get("main_hand", "")), "ii_5")


func _test_snapshot_stash() -> void:
	var m := _new_main()
	var snap := _base_snapshot()
	snap["stash_gold"] = 200
	snap["stash_items"] = [{"stash_item_id": "si_9", "item_def_id": "red_potion"}]
	m._apply_snapshot(snap)
	_assert_eq("snapshot sets stash_gold", m.stash_gold, 200)
	_assert_eq("snapshot sets stash_items size", m.stash_items.size(), 1)


# --- robustness ---------------------------------------------------------------

func _test_malformed_delta_does_not_crash() -> void:
	var m := _new_main()
	# Missing required fields — should not throw.
	m._apply_delta({})
	m._apply_delta({"changes": [{"op": "gold_update"}]})
	m._apply_delta({"changes": [{"op": "inventory_add"}]})
	m._apply_delta({"changes": [{"op": "unknown_op", "data": "x"}]})
	m._apply_delta({"changes": [{"op": "equipped_update", "item_instance_id": "ii_x"}]})
	_assert_eq("malformed equipped_update without slot does not crash", true, true)
	_assert_eq("malformed delta does not crash", true, true)


func _test_malformed_envelope_payloads_do_not_crash() -> void:
	var m := _new_main()
	m.pending_action_targets["msg_accepted"] = "1001"
	m.pending_action_targets["msg_missing"] = "1002"
	m.pending_interactable_action = {"entity_id": "1003"}
	m.pending_waypoint_travel = true

	m._handle_message({"type": "intent_accepted", "payload": {"accepted_message_id": "msg_accepted"}})
	m._handle_message({"type": "intent_accepted"})
	m._handle_message({"type": "intent_accepted", "payload": "not-a-dictionary"})
	m._handle_message({"type": "intent_rejected"})
	m._handle_message({"type": "intent_rejected", "payload": null})
	m._handle_message({"type": "error"})
	m._handle_message({"type": "error", "payload": []})
	m._handle_message({"type": "state_delta"})
	m._handle_message({"type": "state_delta", "payload": "not-a-dictionary"})

	_assert_eq("accepted payload erases matching pending action", m.pending_action_targets.has("msg_accepted"), false)
	_assert_eq("malformed accepted leaves unrelated pending action", m.pending_action_targets.has("msg_missing"), true)
	_assert_eq("malformed reject clears interactable action", m.pending_interactable_action.is_empty(), true)
	_assert_eq("malformed reject clears waypoint travel", m.pending_waypoint_travel, false)


# --- helpers ------------------------------------------------------------------

func _new_main() -> Object:
	var m := MainScript.new()
	# Inject item rules so methods that look up item defs find something.
	ItemRulesLoader.ensure_loaded()
	return m


func _base_snapshot() -> Dictionary:
	return {
		"local_player_id": "",
		"gold": 0,
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"inventory_rows": 3,
		"inventory_capacity": 15,
		"stash_items": [],
		"stash_gold": 0,
		"stash_capacity": 50,
		"hotbar_capacity": 2,
		"character_progression": {},
		"skill_progression": {},
		"skill_cooldowns": [],
		"entities": [],
		"events": [],
	}


func _assert_eq(label: String, got, expected) -> void:
	_pass += 1
	if got != expected:
		_fail += 1
		_pass -= 1
		printerr("[gdtest] FAIL %s: got %s, want %s" % [label, str(got), str(expected)])
