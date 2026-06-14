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
	var delay_str := OS.get_environment("ARPG_BOT_STEP_DELAY")
	if delay_str != "":
		_runner.step_delay_s = maxf(0.0, float(delay_str))
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
		if str(action.get("type", "")) != "dispatch_intent":
			print("[bot-client] action %s scenario=%s" % [
				_format_action(action), _scenario_id
			])
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
	if _main != null and _main.has_method("bot_show_action_shadow"):
		_main.bot_show_action_shadow(action, state)
	var stype := str(action.get("_type", action.get("type", "")))
	match stype:
		"press_key":
			_do_press_key(str(action.get("keycode", "")))
		"click_entity":
			_do_click_entity(action, state)
		"click_loot_item":
			_do_click_loot_item(
				str(action.get("item_def_id", "")),
				state,
				int(action.get("occurrence", 0)),
				action.get("rolled", null),
			)
		"click_floor":
			_do_click_floor(float(action.get("x", 0.0)), float(action.get("z", 0.0)))
		"drag_bag_to_weapon_slot":
			_do_equip_from_bag(str(action.get("item_def_id", "")), state)
		"drag_bag_to_equipment_slot":
			_do_equip_from_bag_to_slot(str(action.get("item_def_id", "")), str(action.get("slot", "main_hand")), state)
		"drag_weapon_to_bag":
			_do_unequip_to_bag(state)
		"drag_equipment_to_bag":
			_do_unequip_slot_to_bag(str(action.get("slot", "main_hand")), state)
		"drag_bag_to_outside":
			_do_drop_from_bag(str(action.get("item_def_id", "")), state)
		"assign_hotbar_slot":
			_do_assign_hotbar_slot(
				int(action.get("slot_index", -1)),
				str(action.get("item_def_id", "")),
				int(action.get("bag_index", 0)),
				state,
			)
		"use_hotbar_slot":
			_do_use_hotbar_slot(int(action.get("slot_index", -1)))
		"double_click_bag_item":
			_do_double_click_bag_item(
				str(action.get("item_def_id", "")),
				int(action.get("bag_index", 0)),
				state,
			)
		"click_menu_button":
			_do_click_menu_button(str(action.get("button", "")))
		"enter_character_name":
			_do_enter_character_name(str(action.get("name", "")))
		"select_character":
			_do_select_character(int(action.get("index", 0)))
		"select_character_class":
			_do_select_character_class(str(action.get("class_id", "")))
		"select_window_size":
			_do_select_window_size(str(action.get("size", "")))
		"set_floating_combat_text":
			_do_set_floating_combat_text(bool(action.get("enabled", true)))
		"select_create_game_type":
			_do_select_create_game_type(str(action.get("session_type", "")))
		"click_stat_button":
			_do_click_stat_button(str(action.get("stat", "")))
		"click_skill_button":
			_do_click_skill_button(str(action.get("skill_id", "magic_bolt")))
		"use_skill_slot":
			_do_use_skill_slot(action, state)
		"dispatch_intent":
			if _main != null and _main.has_method("bot_dispatch_action"):
				_main.bot_dispatch_action(str(action.get("intent_type", "")), action.get("payload", {}))
		"click_shop_buy_offer":
			_do_click_shop_buy_offer(action)
		"click_shop_reroll":
			_do_click_shop_reroll()
		"click_shop_sell_item":
			_do_click_shop_sell_item(action)
		"click_waypoint_level":
			_do_click_waypoint_level(action)
		"drag_bag_to_stash":
			_do_drag_bag_to_stash(action)
		"drag_stash_to_bag":
			_do_drag_stash_to_bag(action)
		"click_stash_deposit_gold":
			_do_click_stash_deposit_gold(action)
		"click_stash_withdraw_gold":
			_do_click_stash_withdraw_gold(action)
		"click_bishop_respec":
			_do_click_bishop_respec()
		"click_blacksmith_upgrade":
			_do_click_blacksmith_upgrade(action)
		"set_stash_search":
			_do_set_stash_search(action)
		"select_stash_sort":
			_do_select_stash_sort(action)
		"set_multiplayer_search":
			_do_set_multiplayer_search(action)
		"select_multiplayer_sort":
			_do_select_multiplayer_sort(action)
		"set_market_publish_price":
			_do_set_market_publish_price(action)
		"click_market_publish_item":
			_do_click_market_publish_item(action)
		"click_market_purchase_listing":
			_do_click_market_purchase_listing(action)
		"click_market_view_offers":
			_do_click_market_view_offers(action)
		"click_market_accept_offer":
			_do_click_market_accept_offer(action)


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


func _do_click_menu_button(button: String) -> void:
	if _main != null and _main.has_method("bot_click_menu_button"):
		_main.bot_click_menu_button(button)


func _do_enter_character_name(name: String) -> void:
	if _main != null and _main.has_method("bot_enter_character_name"):
		_main.bot_enter_character_name(name)


func _do_select_character(index: int) -> void:
	if _main != null and _main.has_method("bot_select_character"):
		_main.bot_select_character(index)


func _do_select_character_class(class_id: String) -> void:
	if _main != null and _main.has_method("bot_select_character_class"):
		_main.bot_select_character_class(class_id)


func _do_select_window_size(size: String) -> void:
	if _main != null and _main.has_method("bot_select_window_size"):
		_main.bot_select_window_size(size)


func _do_set_floating_combat_text(enabled: bool) -> void:
	if _main != null and _main.has_method("bot_set_floating_combat_text"):
		_main.bot_set_floating_combat_text(enabled)


func _do_select_create_game_type(session_type: String) -> void:
	if _main != null and _main.has_method("bot_select_create_game_type"):
		_main.bot_select_create_game_type(session_type)


func _do_click_stat_button(stat: String) -> void:
	if _main != null and _main.has_method("bot_click_stat_button"):
		_main.bot_click_stat_button(stat)


func _do_click_skill_button(skill_id: String) -> void:
	if _main != null and _main.has_method("bot_click_skill_button"):
		_main.bot_click_skill_button(skill_id)


func _do_use_skill_slot(action: Dictionary, state: Dictionary) -> void:
	if _main == null:
		return
	var skill_id := str(action.get("skill_id", "magic_bolt"))
	if action.has("direction") and typeof(action.get("direction")) == TYPE_DICTIONARY:
		if _main.has_method("bot_cast_skill_direction"):
			_main.bot_cast_skill_direction(skill_id, action.get("direction", {}))
		return
	var target_id := str(action.get("target_id", ""))
	if target_id == "":
		var selector := {
			"entity_type": str(action.get("entity_type", "monster")),
		}
		for key in ["monster_def_id", "rarity", "state"]:
			if action.has(key):
				selector[key] = action[key]
		var ids := _select_entity_ids(state, selector)
		if not ids.is_empty():
			var entity_index := int(action.get("entity_index", 0))
			if entity_index >= 0 and entity_index < ids.size():
				target_id = str(ids[entity_index])
	if _main.has_method("bot_use_skill_bar"):
		_main.bot_use_skill_bar(skill_id, target_id, bool(action.get("force_direct", false)))


func _do_click_shop_buy_offer(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_shop_buy_offer"):
		_main.bot_click_shop_buy_offer(
			str(action.get("offer_id", "")),
			str(action.get("offer_kind", "")),
			int(action.get("offer_index", 0))
		)


func _do_click_shop_reroll() -> void:
	if _main != null and _main.has_method("bot_click_shop_reroll"):
		_main.bot_click_shop_reroll()


func _do_click_shop_sell_item(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_shop_sell_item"):
		_main.bot_click_shop_sell_item(
			str(action.get("item_def_id", "")),
			action.get("rolled", null),
			int(action.get("bag_index", 0))
		)


func _do_click_waypoint_level(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_waypoint_level"):
		_main.bot_click_waypoint_level(int(action.get("target_level", 0)))


func _do_drag_bag_to_stash(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_drag_bag_to_stash"):
		_main.bot_drag_bag_to_stash(
			str(action.get("item_def_id", "")),
			action.get("rolled", null),
			int(action.get("bag_index", 0))
		)


func _do_drag_stash_to_bag(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_drag_stash_to_bag"):
		_main.bot_drag_stash_to_bag(
			str(action.get("stash_item_id", "")),
			str(action.get("item_def_id", "")),
			action.get("rolled", null),
			int(action.get("stash_index", 0))
		)


func _do_click_stash_deposit_gold(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_stash_deposit_gold"):
		_main.bot_click_stash_deposit_gold(int(action.get("amount", 1)))


func _do_click_stash_withdraw_gold(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_stash_withdraw_gold"):
		_main.bot_click_stash_withdraw_gold(int(action.get("amount", 1)))


func _do_click_bishop_respec() -> void:
	if _main != null and _main.has_method("bot_click_bishop_respec"):
		_main.bot_click_bishop_respec()


func _do_click_blacksmith_upgrade(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_blacksmith_upgrade"):
		_main.bot_click_blacksmith_upgrade(
			str(action.get("stash_item_id", "")),
			str(action.get("item_def_id", "")),
			int(action.get("stash_index", 0))
		)


func _do_set_stash_search(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_set_stash_search"):
		_main.bot_set_stash_search(str(action.get("text", "")))


func _do_select_stash_sort(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_select_stash_sort"):
		_main.bot_select_stash_sort(str(action.get("mode", "acquired")))


func _do_set_multiplayer_search(action: Dictionary) -> void:
	if _main == null or not _main.has_method("bot_set_multiplayer_search"):
		return
	var env_key := str(action.get("text_env", ""))
	_main.bot_set_multiplayer_search(OS.get_environment(env_key) if env_key != "" else str(action.get("text", "")))


func _do_select_multiplayer_sort(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_select_multiplayer_sort"):
		_main.bot_select_multiplayer_sort(str(action.get("mode", "recent")))


func _do_set_market_publish_price(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_set_market_publish_price"):
		_main.bot_set_market_publish_price(int(action.get("price_gold", 1)))


func _do_click_market_publish_item(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_market_publish_item"):
		_main.bot_click_market_publish_item(
			str(action.get("stash_item_id", "")),
			str(action.get("item_def_id", "")),
			action.get("rolled", null),
			int(action.get("stash_index", 0))
		)


func _do_click_market_purchase_listing(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_market_purchase_listing"):
		_main.bot_click_market_purchase_listing(
			str(action.get("listing_id", "")),
			str(action.get("item_def_id", "")),
			int(action.get("price_gold", -1)),
			int(action.get("listing_index", 0))
		)


func _do_click_market_view_offers(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_market_view_offers"):
		_main.bot_click_market_view_offers(
			str(action.get("listing_id", "")),
			str(action.get("item_def_id", "")),
			int(action.get("price_gold", -1)),
			int(action.get("listing_index", 0))
		)


func _do_click_market_accept_offer(action: Dictionary) -> void:
	if _main != null and _main.has_method("bot_click_market_accept_offer"):
		_main.bot_click_market_accept_offer(str(action.get("offer_id", "")), int(action.get("offer_index", 0)))


# Headless fallback: dispatches action_intent directly via main.gd which routes
# through the same client.send() as _try_action_at_mouse(). Ray-pick via
# direct_space_state is unreliable in headless because get_mouse_position()
# returns (0,0) and Input.warp_mouse() has no effect without a real display server.
func consume_pending_event_at(index: int) -> void:
	if _main != null and _main.has_method("bot_consume_pending_event_at"):
		_main.bot_consume_pending_event_at(index)


func _do_click_entity(action: Dictionary, state: Dictionary) -> void:
	if _main == null:
		return
	var entity_type := str(action.get("entity_type", ""))
	var entity_index := int(action.get("entity_index", 0))
	var ids_key := "%s_ids" % entity_type
	var ids: Array = _select_entity_ids(state, action)
	if ids.is_empty() and not _has_entity_filter(action):
		ids = state.get(ids_key, [])
	if ids.is_empty():
		printerr("[bot-client] click_entity: no %s entity found" % entity_type)
		return
	if entity_index < 0 or entity_index >= ids.size():
		printerr("[bot-client] click_entity: index %d out of range for %s" % [entity_index, entity_type])
		return
	var target_id := str(ids[entity_index])
	if _main.has_method("bot_click_entity_id"):
		_main.bot_click_entity_id(target_id)
	elif _main.has_method("bot_dispatch_action"):
		_main.bot_dispatch_action("action_intent", {"target_id": target_id})


func _has_entity_filter(selector: Dictionary) -> bool:
	for key in ["monster_def_id", "interactable_def_id", "item_def_id", "rarity", "state"]:
		if selector.has(key) and str(selector.get(key, "")) != "":
			return true
	for key in ["is_boss", "elite_objective"]:
		if selector.has(key):
			return true
	return false


func _select_entity_ids(state: Dictionary, selector: Dictionary) -> Array:
	var out: Array = []
	var entity_type := str(selector.get("entity_type", ""))
	for row in state.get("entities_debug", []):
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if entity_type != "" and str(rec.get("type", "")) != entity_type:
			continue
		var ok := true
		for key in ["monster_def_id", "interactable_def_id", "item_def_id", "rarity", "state"]:
			if selector.has(key) and str(selector.get(key, "")) != "" and str(rec.get(key, "")) != str(selector[key]):
				ok = false
				break
		if selector.has("is_boss") and bool(rec.get("is_boss", false)) != bool(selector.get("is_boss", false)):
			ok = false
		if selector.has("elite_objective") and bool(rec.get("elite_objective", false)) != bool(selector.get("elite_objective", false)):
			ok = false
		if ok:
			out.append(str(rec.get("id", "")))
	return out


func _do_click_loot_item(item_def_id: String, state: Dictionary, occurrence: int = 0, rolled: Variant = null) -> void:
	if _main == null:
		return
	var loot: Array = state.get("loot", [])
	var seen := 0
	for row in loot:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if rolled != null and (str(rec.get("item_template_id", "")) != "") != bool(rolled):
			continue
		if seen == occurrence:
			var target_id := str(rec.get("id", ""))
			if _main.has_method("bot_click_entity_id"):
				_main.bot_click_entity_id(target_id)
			elif _main.has_method("bot_dispatch_action"):
				_main.bot_dispatch_action("action_intent", {"target_id": target_id})
			return
		seen += 1
	printerr("[bot-client] click_loot_item: item_def_id=%s rolled=%s occurrence=%d not found" % [item_def_id, str(rolled), occurrence])


func _do_click_floor(world_x: float, world_z: float) -> void:
	if _main == null:
		return
	if _main.has_method("bot_dispatch_action"):
		_main.bot_dispatch_action("move_to_intent", {"position": {"x": world_x, "y": world_z}})


func _do_equip_from_bag(item_def_id: String, state: Dictionary) -> void:
	_do_equip_from_bag_to_slot(item_def_id, "main_hand", state)


func _do_equip_from_bag_to_slot(item_def_id: String, slot: String, state: Dictionary) -> void:
	if _main == null:
		return
	var item_id := _find_bag_item_id(item_def_id, state)
	if item_id == "":
		printerr("[bot-client] drag_bag_to_weapon_slot: item_def_id=%s not in bag" % item_def_id)
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("equip_intent", {"item_instance_id": item_id, "slot": slot})


func _do_unequip_to_bag(state: Dictionary) -> void:
	_do_unequip_slot_to_bag("main_hand", state)


func _do_unequip_slot_to_bag(slot: String, state: Dictionary) -> void:
	if _main == null:
		return
	var eq: Dictionary = state.get("equipped", {})
	var item_id = eq.get(slot, null)
	if item_id == null or str(item_id) == "" or str(item_id) == "null":
		printerr("[bot-client] drag_equipment_to_bag: nothing equipped in %s slot" % slot)
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("unequip_intent", {"item_instance_id": str(item_id), "slot": slot})


func _do_assign_hotbar_slot(slot_index: int, item_def_id: String, bag_index: int, state: Dictionary) -> void:
	if _main == null:
		return
	var item_id := _find_bag_item_id(item_def_id, state, bag_index)
	if item_id == "":
		printerr("[bot-client] assign_hotbar_slot: item_def_id=%s bag_index=%d not in bag" % [item_def_id, bag_index])
		return
	if _main.has_method("bot_assign_consumable_hotbar"):
		_main.bot_assign_consumable_hotbar(slot_index, item_id)


func _do_use_hotbar_slot(slot_index: int) -> void:
	if _main == null:
		return
	if _main.has_method("bot_use_consumable_hotbar"):
		_main.bot_use_consumable_hotbar(slot_index)


func _do_double_click_bag_item(item_def_id: String, bag_index: int, state: Dictionary) -> void:
	if _main == null:
		return
	var item_id := _find_bag_item_id(item_def_id, state, bag_index)
	if item_id == "":
		printerr("[bot-client] double_click_bag_item: item_def_id=%s bag_index=%d not in bag" % [item_def_id, bag_index])
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("use_intent", {"item_instance_id": item_id})


func _do_drop_from_bag(item_def_id: String, state: Dictionary) -> void:
	if _main == null:
		return
	var item_id := _find_bag_item_id(item_def_id, state)
	if item_id == "":
		printerr("[bot-client] drag_bag_to_outside: item_def_id=%s not in bag" % item_def_id)
		return
	if _main.has_method("bot_dispatch_inventory_intent"):
		_main.bot_dispatch_inventory_intent("drop_intent", {"item_instance_id": item_id})


func _find_bag_item_id(item_def_id: String, state: Dictionary, bag_index: int = 0) -> String:
	var inv: Array = state.get("inventory", [])
	var eq: Dictionary = state.get("equipped", {})
	var equipped_weapon = eq.get("main_hand", null)
	var matches: Array = []
	for item in inv:
		if str(item.get("item_def_id", "")) == item_def_id:
			var iid := str(item.get("item_instance_id", ""))
			if str(equipped_weapon) != iid:
				matches.append(iid)
	if matches.is_empty():
		var bar: Dictionary = state.get("consumable_bar", {})
		var assigned: Array = bar.get("assigned_slots", [])
		for slot in assigned:
			if typeof(slot) != TYPE_DICTIONARY:
				continue
			var slot_item := slot as Dictionary
			if str(slot_item.get("item_def_id", "")) == item_def_id:
				var iid := str(slot_item.get("item_instance_id", ""))
				if iid != "":
					matches.append(iid)
	if bag_index < 0 or bag_index >= matches.size():
		return ""
	return str(matches[bag_index])


func _parse_keycode(name: String) -> Key:
	match name:
		"KEY_0": return KEY_0
		"KEY_1": return KEY_1
		"KEY_2": return KEY_2
		"KEY_3": return KEY_3
		"KEY_4": return KEY_4
		"KEY_5": return KEY_5
		"KEY_6": return KEY_6
		"KEY_7": return KEY_7
		"KEY_8": return KEY_8
		"KEY_9": return KEY_9
		"KEY_F1": return KEY_F1
		"KEY_F2": return KEY_F2
		"KEY_F3": return KEY_F3
		"KEY_F4": return KEY_F4
		"KEY_F5": return KEY_F5
		"KEY_F6": return KEY_F6
		"KEY_F7": return KEY_F7
		"KEY_F8": return KEY_F8
		"KEY_I": return KEY_I
		"KEY_K": return KEY_K
		"KEY_P": return KEY_P
		"KEY_Q": return KEY_Q
		"KEY_E": return KEY_E
		"KEY_W": return KEY_W
		"KEY_A": return KEY_A
		"KEY_C": return KEY_C
		"KEY_S": return KEY_S
		"KEY_D": return KEY_D
		"KEY_ESCAPE": return KEY_ESCAPE
	return KEY_NONE


func _format_action(action: Dictionary) -> String:
	var stype := str(action.get("_type", action.get("type", "")))
	match stype:
		"click_entity":
			return "click_entity type=%s index=%s" % [
				str(action.get("entity_type", "")), str(action.get("entity_index", 0))
			]
		"click_loot_item":
			return "click_loot_item item=%s rolled=%s occurrence=%s" % [
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("occurrence", 0))
			]
		"click_floor":
			return "click_floor x=%s z=%s" % [str(action.get("x", "")), str(action.get("z", ""))]
		"press_key":
			return "press_key %s" % str(action.get("keycode", ""))
		"click_menu_button":
			return "click_menu_button %s" % str(action.get("button", ""))
		"enter_character_name":
			return "enter_character_name %s" % str(action.get("name", ""))
		"select_character":
			return "select_character index=%s" % str(action.get("index", 0))
		"select_window_size":
			return "select_window_size %s" % str(action.get("size", ""))
		"select_create_game_type":
			return "select_create_game_type %s" % str(action.get("session_type", ""))
		"click_stat_button":
			return "click_stat_button %s" % str(action.get("stat", ""))
		"click_skill_button":
			return "click_skill_button %s" % str(action.get("skill_id", "magic_bolt"))
		"use_skill_slot":
			return "use_skill_slot skill=%s target=%s monster=%s force=%s direction=%s" % [
				str(action.get("skill_id", "magic_bolt")),
				str(action.get("target_id", "")),
				str(action.get("monster_def_id", "")),
				str(action.get("force_direct", false)),
				str(action.get("direction", {})),
			]
		"click_shop_buy_offer":
			return "click_shop_buy offer_id=%s kind=%s index=%s" % [
				str(action.get("offer_id", "")),
				str(action.get("offer_kind", "")),
				str(action.get("offer_index", 0)),
			]
		"click_shop_reroll":
			return "click_shop_reroll"
		"click_shop_sell_item":
			return "click_shop_sell item=%s rolled=%s bag_index=%s" % [
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("bag_index", 0)),
			]
		"click_waypoint_level":
			return "click_waypoint_level target=%s" % str(action.get("target_level", ""))
		"drag_bag_to_stash":
			return "drag_bag_to_stash item=%s rolled=%s bag_index=%s" % [
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("bag_index", 0)),
			]
		"drag_stash_to_bag":
			return "drag_stash_to_bag stash_item=%s item=%s rolled=%s stash_index=%s" % [
				str(action.get("stash_item_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("stash_index", 0)),
			]
		"click_stash_deposit_gold":
			return "click_stash_deposit_gold amount=%s" % str(action.get("amount", 1))
		"click_stash_withdraw_gold":
			return "click_stash_withdraw_gold amount=%s" % str(action.get("amount", 1))
		"click_bishop_respec":
			return "click_bishop_respec"
		"click_blacksmith_upgrade":
			return "click_blacksmith_upgrade stash_item=%s item=%s stash_index=%s" % [
				str(action.get("stash_item_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("stash_index", 0)),
			]
		"set_stash_search":
			return "set_stash_search text=%s" % str(action.get("text", ""))
		"select_stash_sort":
			return "select_stash_sort mode=%s" % str(action.get("mode", "acquired"))
		"set_market_publish_price":
			return "set_market_publish_price price=%s" % str(action.get("price_gold", 1))
		"click_market_publish_item":
			return "click_market_publish_item stash_item=%s item=%s rolled=%s stash_index=%s" % [
				str(action.get("stash_item_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("stash_index", 0)),
			]
		"click_market_purchase_listing":
			return "click_market_purchase_listing listing=%s item=%s price=%s index=%s" % [
				str(action.get("listing_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("price_gold", "")),
				str(action.get("listing_index", 0)),
			]
		"click_market_view_offers":
			return "click_market_view_offers listing=%s item=%s price=%s index=%s" % [
				str(action.get("listing_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("price_gold", "")),
				str(action.get("listing_index", 0)),
			]
		"click_market_accept_offer":
			return "click_market_accept_offer offer=%s index=%s" % [str(action.get("offer_id", "")), str(action.get("offer_index", 0))]
		"assign_hotbar_slot":
			return "assign_hotbar slot=%s item=%s bag_index=%s" % [
				str(action.get("slot_index", "")),
				str(action.get("item_def_id", "")),
				str(action.get("bag_index", "")),
			]
		"double_click_bag_item":
			return "double_click_bag item=%s bag_index=%s" % [
				str(action.get("item_def_id", "")), str(action.get("bag_index", ""))
			]
		_:
			return stype


func _fail_startup(msg: String) -> void:
	printerr("[bot-client] FAIL startup -- %s" % msg)
	get_tree().quit(1)
