# BotController: client-side bot for CI scenario execution.
# Runs inside main.tscn when ARPG_BOT_CLIENT=1. Mounts as a child of the main
# node after normal auth/session/WS setup, loads a scenario from
# ARPG_BOT_SCENARIO, drives the input handlers through get_bot_state() /
# bot_dispatch_action() on main.gd, and exits with a clear sentinel + exit code.
#
# Headless input note: Godot headless mode does not expose a real display server,
# so synthetic mouse position (get_viewport().get_mouse_position()) is
# unreliable for targeting. click_entity and click_floor therefore use the
# documented direct fallback (bot_dispatch_action on main.gd, which routes
# through the same client.send("action_intent" / "move_to_intent") as
# _try_action_at_mouse()). press_key pushes a real InputEventKey through
# get_viewport().push_input() so it flows through _unhandled_input() exactly
# as a human keystroke would. Inventory intents route through
# bot_dispatch_inventory_intent() which calls _on_inventory_intent_requested().
class_name BotController
extends Node

const BotScenarioRunnerScript := preload("res://scripts/bot_scenario_runner.gd")

var _runner: BotScenarioRunner
var _scenario_id: String = ""
var _main = null  # main.gd node (parent)


func _ready() -> void:
	_main = get_parent()
	var scenario_path := OS.get_environment("ARPG_BOT_SCENARIO")
	if scenario_path == "":
		_fail_startup("ARPG_BOT_SCENARIO is not set")
		return

	if not FileAccess.file_exists(scenario_path):
		_fail_startup("scenario file not found: %s" % scenario_path)
		return

	var text := FileAccess.get_file_as_string(scenario_path)
	var data = JSON.parse_string(text)
	if typeof(data) != TYPE_DICTIONARY:
		_fail_startup("scenario JSON is not an object: %s" % scenario_path)
		return

	var err := BotScenarioRunnerScript.validate_scenario(data as Dictionary)
	if err != "":
		_fail_startup("scenario validation failed: %s -- %s" % [err, scenario_path])
		return

	_scenario_id = str((data as Dictionary).get("id", "unknown"))
	_runner = BotScenarioRunnerScript.new()
	_runner.bind_controller(self)
	if not _runner.load_scenario(data as Dictionary):
		_fail_startup("scenario has no client_steps: %s" % scenario_path)
		return

	print("[bot-client] RUN %s" % _scenario_id)


func _process(delta: float) -> void:
	if _runner == null:
		return

	var state := _get_state()
	var done := _runner.tick(delta, state)

	var action: Dictionary = _runner.pending_action
	if not action.is_empty():
		_execute_action(action, state)

	if done:
		if _runner.passed():
			print("[bot-client] PASS %s" % _scenario_id)
			get_tree().quit(0)
		else:
			printerr("[bot-client] FAIL %s -- %s" % [_scenario_id, _runner.failure_message()])
			get_tree().quit(1)


func _get_state() -> Dictionary:
	if _main == null or not _main.has_method("get_bot_state"):
		return {}
	return _main.get_bot_state()


func _execute_action(action: Dictionary, state: Dictionary) -> void:
	var stype := str(action.get("_type", action.get("type", "")))
	match stype:
		"press_key":
			_do_press_key(str(action.get("keycode", "")))
		"click_entity":
			_do_click_entity(str(action.get("entity_type", "")), state)
		"click_floor":
			_do_click_floor(float(action.get("x", 0.0)), float(action.get("z", 0.0)))
		"drag_bag_to_weapon_slot":
			_do_equip_from_bag(str(action.get("item_def_id", "")), state)
		"drag_weapon_to_bag":
			_do_unequip_to_bag(state)
		"drag_bag_to_outside":
			_do_drop_from_bag(str(action.get("item_def_id", "")), state)


func _do_press_key(keycode_str: String) -> void:
	var kc := _parse_keycode(keycode_str)
	if kc == KEY_NONE:
		printerr("[bot-client] unknown keycode: %s" % keycode_str)
		return
	var event := InputEventKey.new()
	event.keycode = kc
	event.pressed = true
	event.echo = false
	get_viewport().push_input(event)


# Headless fallback: dispatches action_intent directly via main.gd which routes
# through the same client.send() as _try_action_at_mouse(). Ray-pick via
# direct_space_state is unreliable in headless because get_mouse_position()
# returns (0,0) and Input.warp_mouse() has no effect without a real display server.
func _do_click_entity(entity_type: String, state: Dictionary) -> void:
	if _main == null:
		return
	var ids_key := "%s_ids" % entity_type
	var ids: Array = state.get(ids_key, [])
	if ids.is_empty():
		printerr("[bot-client] click_entity: no %s entity found" % entity_type)
		return
	var target_id := str(ids[0])
	if _main.has_method("bot_dispatch_action"):
		_main.bot_dispatch_action("action_intent", {"target_id": target_id})


func _do_click_floor(world_x: float, world_z: float) -> void:
	if _main == null:
		return
	if _main.has_method("bot_dispatch_action"):
		_main.bot_dispatch_action("move_to_intent", {"position": {"x": world_x, "y": world_z}})


func _do_equip_from_bag(item_def_id: String, state: Dictionary) -> void:
	if _main == null:
		return
	var item_id := _find_bag_item_id(item_def_id, state)
	if item_id == "":
		printerr("[bot-client] drag_bag_to_weapon_slot: item_def_id=%s not in bag" % item_def_id)
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("equip_intent", {"item_instance_id": item_id, "slot": "weapon"})


func _do_unequip_to_bag(state: Dictionary) -> void:
	if _main == null:
		return
	var eq: Dictionary = state.get("equipped", {})
	var item_id = eq.get("weapon", null)
	if item_id == null or str(item_id) == "" or str(item_id) == "null":
		printerr("[bot-client] drag_weapon_to_bag: nothing equipped in weapon slot")
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("unequip_intent", {"item_instance_id": str(item_id), "slot": "weapon"})


func _do_drop_from_bag(item_def_id: String, state: Dictionary) -> void:
	if _main == null:
		return
	var item_id := _find_bag_item_id(item_def_id, state)
	if item_id == "":
		printerr("[bot-client] drag_bag_to_outside: item_def_id=%s not in bag" % item_def_id)
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("drop_intent", {"item_instance_id": item_id})


func _find_bag_item_id(item_def_id: String, state: Dictionary) -> String:
	var inv: Array = state.get("inventory", [])
	var eq: Dictionary = state.get("equipped", {})
	var equipped_weapon = eq.get("weapon", null)
	for item in inv:
		if str(item.get("item_def_id", "")) == item_def_id:
			var iid := str(item.get("item_instance_id", ""))
			if str(equipped_weapon) != iid:
				return iid
	return ""


func _parse_keycode(name: String) -> Key:
	match name:
		"KEY_I": return KEY_I
		"KEY_E": return KEY_E
		"KEY_W": return KEY_W
		"KEY_A": return KEY_A
		"KEY_S": return KEY_S
		"KEY_D": return KEY_D
		"KEY_ESCAPE": return KEY_ESCAPE
	return KEY_NONE


func _fail_startup(msg: String) -> void:
	printerr("[bot-client] FAIL startup -- %s" % msg)
	get_tree().quit(1)
