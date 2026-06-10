# BotScenarioRunner: frame-tick step executor for the Godot client bot.
# Runs one step at a time from a loaded client scenario, tracks elapsed time
# per step, and delegates to BotController for state access and action dispatch.
class_name BotScenarioRunner
extends RefCounted

const STEP_TYPES_WAIT := [
	"wait_ws_open", "wait_entity", "wait_event", "wait_inventory_item",
	"wait_inventory_count", "wait_loot_item", "wait_loot_count", "wait_hotbar_assigned",
	"wait_hotbar_capacity",
	"wait_player_near", "assert_entity_removed",
	"click_entity_until_event", "wait_main_menu", "wait_character_panel",
	"wait_multiplayer_panel", "wait_settings_panel", "wait_pause_menu", "wait_character_progression",
	"wait_skill_progression", "wait_skill_bar",
	"wait_damage_number", "wait_no_damage_number", "wait_entity_reaction",
	"wait_wall_layout", "wait_shop_panel", "wait_stash_panel",
	"wait_remote_player_count",
]
const STEP_TYPES_ASSERT := [
	"assert_panel_visible", "assert_waypoint_panel_visible", "assert_equipped",
	"assert_unequipped", "assert_inventory_missing", "assert_inventory_count",
	"assert_loot_presentation", "assert_inventory_presentation",
	"assert_hotbar_assigned", "assert_player_hp", "assert_main_menu_visible",
	"assert_main_menu_actions", "assert_character_panel", "assert_create_game_type",
	"assert_current_session",
	"assert_character_panel_visible", "assert_settings_panel_visible",
	"assert_pause_menu_visible", "assert_session_changed",
	"assert_multiplayer_panel_visible", "assert_multiplayer_session_rows",
	"assert_player_position_unchanged", "assert_character_stats_panel_visible",
	"assert_character_info_panel_visible", "assert_character_info",
	"assert_character_progression", "assert_stat_button_enabled", "assert_xp_bar",
	"assert_skills_panel_visible", "assert_skill_progression",
	"assert_skill_button_enabled", "assert_skill_bar",
	"assert_hotbar_capacity", "assert_hotbar_slot_disabled",
	"assert_inventory_capacity", "assert_bag_grid", "assert_paper_doll_layout",
	"assert_inventory_panel_details",
	"assert_floating_combat_text_enabled", "assert_damage_number", "assert_no_damage_number",
	"assert_entity_reaction",
	"assert_wall_layout", "assert_shop_panel_visible", "assert_shop_offer_count",
	"assert_shop_buy_button", "assert_shop_sell_rows", "assert_shop_offer_details",
	"assert_shop_sell_details", "assert_stash_panel_visible", "assert_stash_item_count",
	"assert_stash_gold",
	"assert_remote_player_count",
]
const STEP_TYPES_ACTION := [
	"press_key", "click_entity", "click_loot_item", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_equipment_slot",
	"drag_equipment_to_bag", "drag_bag_to_outside", "assign_hotbar_slot",
	"use_hotbar_slot", "double_click_bag_item", "click_menu_button",
	"enter_character_name", "select_character", "select_window_size",
	"set_floating_combat_text", "select_create_game_type",
	"remember_session", "remember_player_position", "click_stat_button",
	"click_skill_button", "use_skill_slot", "click_shop_buy_offer", "click_shop_sell_item",
	"drag_bag_to_stash", "drag_stash_to_bag", "click_stash_deposit_gold",
	"click_stash_withdraw_gold", "click_waypoint_level",
]
const WAIT_LOG_INTERVAL_S := 2.0

const ALL_STEP_TYPES: Array = [
	"wait_ws_open", "wait_entity", "wait_event", "assert_entity_removed",
	"assert_panel_visible", "assert_waypoint_panel_visible", "wait_inventory_item", "wait_inventory_count",
	"assert_equipped", "assert_unequipped", "assert_inventory_missing",
	"assert_inventory_count", "wait_loot_item", "wait_loot_count",
	"wait_player_near", "press_key", "click_entity", "click_loot_item", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_equipment_slot",
	"drag_equipment_to_bag", "drag_bag_to_outside", "assert_loot_presentation",
	"assert_inventory_presentation", "click_entity_until_event", "assign_hotbar_slot",
	"use_hotbar_slot", "assert_hotbar_assigned", "wait_hotbar_assigned",
	"assert_hotbar_capacity", "wait_hotbar_capacity",
	"assert_hotbar_slot_disabled", "assert_inventory_capacity", "assert_bag_grid",
	"assert_paper_doll_layout", "assert_inventory_panel_details",
	"assert_player_hp", "double_click_bag_item", "wait_main_menu",
	"wait_character_panel", "wait_settings_panel", "wait_pause_menu",
	"assert_main_menu_visible", "assert_main_menu_actions",
	"assert_character_panel_visible", "assert_character_panel",
	"assert_create_game_type", "assert_current_session",
	"wait_multiplayer_panel", "assert_multiplayer_panel_visible",
	"assert_multiplayer_session_rows", "assert_settings_panel_visible", "assert_pause_menu_visible",
	"click_menu_button", "enter_character_name", "select_character",
	"select_window_size", "remember_session", "assert_session_changed",
	"select_create_game_type",
	"remember_player_position", "assert_player_position_unchanged",
	"assert_character_stats_panel_visible", "assert_character_info_panel_visible", "assert_character_info",
	"wait_character_progression",
	"assert_character_progression", "click_stat_button",
	"assert_stat_button_enabled", "assert_xp_bar",
	"wait_skill_progression", "assert_skills_panel_visible",
	"assert_skill_progression", "assert_skill_button_enabled", "click_skill_button",
	"wait_skill_bar", "assert_skill_bar", "use_skill_slot",
	"set_floating_combat_text", "assert_floating_combat_text_enabled",
	"wait_damage_number", "wait_no_damage_number", "assert_damage_number", "assert_no_damage_number",
	"wait_entity_reaction", "assert_entity_reaction",
	"wait_wall_layout", "assert_wall_layout",
	"wait_shop_panel", "assert_shop_panel_visible", "assert_shop_offer_count",
	"assert_shop_buy_button", "assert_shop_sell_rows", "assert_shop_offer_details",
	"assert_shop_sell_details", "click_shop_buy_offer", "click_shop_sell_item",
	"wait_stash_panel", "assert_stash_panel_visible", "assert_stash_item_count",
	"assert_stash_gold", "drag_bag_to_stash", "drag_stash_to_bag",
	"click_stash_deposit_gold", "click_stash_withdraw_gold",
	"click_waypoint_level",
	"wait_remote_player_count", "assert_remote_player_count",
]

var scenario: Dictionary = {}
var step_delay_s: float = 0.0  # pause after each completed step (visual mode)
var _steps: Array = []
var _step_index: int = 0
var _step_elapsed: float = 0.0
var _last_retry_at: float = -999.0
var _last_wait_log_at: float = 0.0
var _step_begin_logged: bool = false
var _post_step_wait: float = 0.0  # countdown after a step completes
var _done: bool = false
var _passed: bool = false
var _failure_msg: String = ""
var _controller = null  # BotController reference
var _memory: Dictionary = {}

# Filled by tick() on first action call; consumed by controller's _process.
var pending_action: Dictionary = {}


func load_scenario(data: Dictionary) -> bool:
	scenario = data
	_steps = data.get("client_steps", [])
	_step_index = 0
	_step_elapsed = 0.0
	_last_retry_at = -999.0
	_last_wait_log_at = 0.0
	_step_begin_logged = false
	_done = false
	_passed = false
	_failure_msg = ""
	pending_action = {}
	_memory = {}
	return _steps.size() > 0


func bind_controller(ctrl) -> void:
	_controller = ctrl


func is_done() -> bool:
	return _done


func passed() -> bool:
	return _passed


func failure_message() -> String:
	return _failure_msg


# Called once per _process frame. Returns true when the scenario has ended.
# pending_action is set here for action steps; the caller (BotController) reads
# it after tick() returns, before the NEXT tick clears it.
func tick(delta: float, state: Dictionary) -> bool:
	pending_action = {}  # clear action from previous frame
	if _done:
		return true
	if _post_step_wait > 0.0:
		_post_step_wait -= delta
		return false
	if _step_index >= _steps.size():
		_pass()
		return true

	var step: Dictionary = _steps[_step_index]
	var stype := str(step.get("type", ""))
	_step_elapsed += delta

	if not _step_begin_logged:
		_log_step_begin(step, stype)
		_step_begin_logged = true

	var timeout_s := float(step.get("timeout_s", 0.0))
	if timeout_s > 0.0 and _step_elapsed > timeout_s:
		_fail("timeout after %.1fs at step %d (%s) scenario=%s" % [
			timeout_s, _step_index, stype, str(scenario.get("id", "?"))
		])
		return true

	if stype in STEP_TYPES_WAIT:
		if _step_elapsed - _last_wait_log_at >= WAIT_LOG_INTERVAL_S:
			_log_wait_progress(step, stype, state)
			_last_wait_log_at = _step_elapsed
		if _eval_wait(step, stype, state):
			_advance(stype)
			_check_complete()
	elif stype in STEP_TYPES_ASSERT:
		if not _eval_assert(step, stype, state):
			return true  # _eval_assert already called _fail
		_advance(stype)
		_check_complete()
	elif stype in STEP_TYPES_ACTION:
		_queue_action(step, stype, state)
		_advance(stype)
		_check_complete()
	else:
		_fail("unknown step type '%s' at step %d scenario=%s" % [
			stype, _step_index, str(scenario.get("id", "?"))
		])
		return true

	return _done


func _eval_wait(step: Dictionary, stype: String, state: Dictionary) -> bool:
	match stype:
		"wait_ws_open":
			return bool(state.get("ws_open", false))
		"wait_main_menu":
			return bool(state.get("main_menu_visible", false))
		"wait_character_panel":
			return bool(state.get("character_panel_visible", false))
		"wait_multiplayer_panel":
			return bool(state.get("multiplayer_panel_visible", false))
		"wait_settings_panel":
			return bool(state.get("settings_panel_visible", false))
		"wait_pause_menu":
			return bool(state.get("pause_menu_visible", false))
		"wait_character_progression":
			return _progression_matches(step, state)
		"wait_skill_progression":
			return _skill_progression_matches(step, state)
		"wait_skill_bar":
			return _skill_bar_matches(step, state)
		"wait_damage_number":
			return _damage_number_matches(step, state)
		"wait_no_damage_number":
			return (state.get("damage_numbers", []) as Array).is_empty()
		"wait_entity_reaction":
			return _presentation_matches(step, state)
		"wait_wall_layout":
			return _wall_layout_matches(step, state)
		"wait_shop_panel":
			if not bool(state.get("shop_panel_visible", false)):
				return false
			return _shop_offer_count_matches(step, state)
		"wait_stash_panel":
			if not bool(state.get("stash_panel_visible", false)):
				return false
			return _stash_item_count_matches(step, state)
		"wait_remote_player_count":
			return _remote_player_count_matches(step, state)
		"wait_entity":
			var etype := str(step.get("entity_type", ""))
			var eids: Array = state.get("%s_ids" % etype, state.get("entities_by_type", {}).get(etype, []))
			return eids.size() > 0
		"wait_event":
			var evtype := str(step.get("event_type", ""))
			var pending: Array = state.get("pending_events", [])
			for i in range(pending.size()):
				if str(pending[i].get("event_type", "")) == evtype and _event_matches(step, pending[i]):
					if _controller != null and _controller.has_method("consume_pending_event_at"):
						_controller.consume_pending_event_at(i)
					return true
			return false
		"wait_inventory_count":
			var def_id := str(step.get("item_def_id", ""))
			var want := int(step.get("equals", 0))
			return _inventory_count(state, def_id) == want
		"wait_loot_count":
			var min_count := int(step.get("min_count", 1))
			return (state.get("loot_ids", []) as Array).size() >= min_count
		"click_entity_until_event":
			var evtype := str(step.get("event_type", ""))
			var pending: Array = state.get("pending_events", [])
			for i in range(pending.size()):
				var ev = pending[i]
				if str(ev.get("event_type", "")) == evtype and _event_matches(step, ev):
					if bool(step.get("consume_event", false)) and _controller != null and _controller.has_method("consume_pending_event_at"):
						_controller.consume_pending_event_at(i)
					return true
			var retry_s := float(step.get("retry_s", 0.25))
			if _step_elapsed - _last_retry_at >= retry_s:
				_last_retry_at = _step_elapsed
				pending_action = {
					"type": "click_entity",
					"_type": "click_entity",
					"entity_type": str(step.get("entity_type", "")),
					"entity_index": int(step.get("entity_index", 0)),
				}
				for key in ["monster_def_id", "interactable_def_id", "item_def_id", "rarity", "state"]:
					if step.has(key):
						pending_action[key] = step[key]
			return false
		"wait_inventory_item":
			var def_id := str(step.get("item_def_id", ""))
			var inv: Array = state.get("inventory", [])
			if def_id == "":
				return inv.size() > 0
			for item in inv:
				if str(item.get("item_def_id", "")) == def_id:
					return true
			return false
		"wait_loot_item":
			return (state.get("loot_ids", []) as Array).size() > 0
		"wait_hotbar_assigned":
			return _hotbar_slot_matches(step, state)
		"wait_hotbar_capacity":
			return _hotbar_capacity_matches(step, state)
		"wait_player_near":
			var tx := float(step.get("x", 0.0))
			var tz := float(step.get("z", 0.0))
			var max_dist := float(step.get("distance", 2.5))
			var pp: Dictionary = state.get("player_pos", {})
			var px := float(pp.get("x", 0.0))
			var pz := float(pp.get("z", 0.0))
			var dist := sqrt((px - tx) * (px - tx) + (pz - tz) * (pz - tz))
			return dist <= max_dist
		"assert_entity_removed":
			# Treated as a wait step: server entity_remove may arrive in a
			# subsequent delta after the kill event. Times out via timeout_s.
			var etype := str(step.get("entity_type", ""))
			var eids: Array = state.get("%s_ids" % etype, [])
			return eids.is_empty()
	return false


func _eval_assert(step: Dictionary, stype: String, state: Dictionary) -> bool:
	match stype:
		"assert_panel_visible":
			var want := bool(step.get("visible", true))
			var got := bool(state.get("inventory_panel_visible", false))
			if want != got:
				_fail("assert_panel_visible failed: want=%s got=%s step=%d scenario=%s" % [
					want, got, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_main_menu_visible":
			return _assert_bool_state("assert_main_menu_visible", "main_menu_visible", step, state)
		"assert_main_menu_actions":
			return _assert_main_menu_actions(step, state)
		"assert_character_panel_visible":
			return _assert_bool_state("assert_character_panel_visible", "character_panel_visible", step, state)
		"assert_character_panel":
			return _assert_character_panel(step, state)
		"assert_create_game_type":
			return _assert_create_game_type(step, state)
		"assert_current_session":
			return _assert_current_session(step, state)
		"assert_multiplayer_panel_visible":
			return _assert_bool_state("assert_multiplayer_panel_visible", "multiplayer_panel_visible", step, state)
		"assert_multiplayer_session_rows":
			return _assert_multiplayer_session_rows(step, state)
		"assert_settings_panel_visible":
			return _assert_bool_state("assert_settings_panel_visible", "settings_panel_visible", step, state)
		"assert_pause_menu_visible":
			return _assert_bool_state("assert_pause_menu_visible", "pause_menu_visible", step, state)
		"assert_character_stats_panel_visible":
			return _assert_bool_state("assert_character_stats_panel_visible", "character_stats_panel_visible", step, state)
		"assert_character_info_panel_visible":
			return _assert_bool_state("assert_character_info_panel_visible", "character_info_panel_visible", step, state)
		"assert_character_info":
			return _assert_character_info(step, state)
		"assert_character_progression":
			return _assert_character_progression(step, state)
		"assert_stat_button_enabled":
			return _assert_stat_button_enabled(step, state)
		"assert_xp_bar":
			return _assert_xp_bar(step, state)
		"assert_skills_panel_visible":
			return _assert_bool_state("assert_skills_panel_visible", "skills_panel_visible", step, state)
		"assert_skill_progression":
			return _assert_skill_progression(step, state)
		"assert_skill_button_enabled":
			return _assert_skill_button_enabled(step, state)
		"assert_skill_bar":
			if not _skill_bar_matches(step, state):
				_fail("assert_skill_bar failed: want=%s got=%s step=%d scenario=%s" % [
					str(_skill_bar_expectation(step)), str(state.get("skill_bar", {})), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_floating_combat_text_enabled":
			return _assert_bool_value("assert_floating_combat_text_enabled", step, bool(state.get("floating_combat_text_enabled", false)), bool(step.get("enabled", true)))
		"assert_damage_number":
			if not _damage_number_matches(step, state):
				_fail("assert_damage_number failed: want=%s damage_numbers=%s step=%d scenario=%s" % [
					str(step), str(state.get("damage_numbers", [])), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_no_damage_number":
			var numbers: Array = state.get("damage_numbers", [])
			if not numbers.is_empty():
				_fail("assert_no_damage_number failed: damage_numbers=%s step=%d scenario=%s" % [
					str(numbers), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_entity_reaction":
			if not _presentation_matches(step, state):
				_fail("assert_entity_reaction failed: want=%s local=%s entities=%s step=%d scenario=%s" % [
					str(step), str(state.get("local_player_presentation", {})),
					str(state.get("entities_presentation_debug", [])), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_wall_layout":
			if not _wall_layout_matches(step, state):
				_fail("assert_wall_layout failed: want=%s wall_count=%d generated=%d non_perimeter=%d level=%d walls=%s step=%d scenario=%s" % [
					str(step),
					int(state.get("wall_count", 0)),
					int(state.get("generated_wall_count", 0)),
					int(state.get("non_perimeter_wall_count", 0)),
					int(state.get("current_level", 0)),
					str(state.get("walls", [])),
					_step_index,
					str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_shop_panel_visible":
			return _assert_bool_state("assert_shop_panel_visible", "shop_panel_visible", step, state)
		"assert_shop_offer_count":
			return _assert_shop_offer_count(step, state)
		"assert_shop_buy_button":
			return _assert_shop_buy_button(step, state)
		"assert_shop_sell_rows":
			return _assert_shop_sell_rows(step, state)
		"assert_shop_offer_details":
			return _assert_shop_offer_details(step, state)
		"assert_shop_sell_details":
			return _assert_shop_sell_details(step, state)
		"assert_stash_panel_visible":
			return _assert_bool_state("assert_stash_panel_visible", "stash_panel_visible", step, state)
		"assert_stash_item_count":
			return _assert_stash_item_count(step, state)
		"assert_stash_gold":
			return _assert_stash_gold(step, state)
		"assert_remote_player_count":
			if not _remote_player_count_matches(step, state):
				_fail("assert_remote_player_count failed: want=%s remote_player_ids=%s step=%d scenario=%s" % [
					str(step), str(state.get("remote_player_ids", [])), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_session_changed":
			var remembered_session := str(_memory.get("session_id", ""))
			var current_session := str(state.get("current_session_id", ""))
			if current_session == "" or remembered_session == "" or current_session == remembered_session:
				_fail("assert_session_changed failed: remembered=%s current=%s step=%d scenario=%s" % [
					remembered_session, current_session, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_player_position_unchanged":
			var remembered_pos: Dictionary = _memory.get("player_pos", {})
			var current_pos: Dictionary = state.get("player_pos", {})
			var tolerance := float(step.get("tolerance", 0.01))
			var dx := float(current_pos.get("x", 0.0)) - float(remembered_pos.get("x", 0.0))
			var dz := float(current_pos.get("z", 0.0)) - float(remembered_pos.get("z", 0.0))
			if sqrt(dx * dx + dz * dz) > tolerance:
				_fail("assert_player_position_unchanged failed: remembered=%s current=%s step=%d scenario=%s" % [
					str(remembered_pos), str(current_pos), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_waypoint_panel_visible":
			var want := bool(step.get("visible", true))
			var got := bool(state.get("waypoint_panel_visible", false))
			if want != got:
				_fail("assert_waypoint_panel_visible failed: want=%s got=%s step=%d scenario=%s" % [
					want, got, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_equipped":
			var slot := str(step.get("slot", "main_hand"))
			var eq: Dictionary = state.get("equipped", {})
			var val = eq.get(slot, null)
			if val == null or str(val) == "":
				_fail("assert_equipped failed: slot=%s equipped=%s step=%d scenario=%s" % [
					slot, str(eq), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_unequipped":
			var slot := str(step.get("slot", "main_hand"))
			var eq: Dictionary = state.get("equipped", {})
			var val = eq.get(slot, null)
			if val != null and str(val) != "" and str(val) != "null":
				_fail("assert_unequipped failed: slot=%s still has %s step=%d scenario=%s" % [
					slot, str(val), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_missing":
			var def_id := str(step.get("item_def_id", ""))
			var inv: Array = state.get("inventory", [])
			for item in inv:
				if str(item.get("item_def_id", "")) == def_id:
					_fail("assert_inventory_missing failed: %s still in inventory step=%d scenario=%s" % [
						def_id, _step_index, str(scenario.get("id", "?"))
					])
					return false
			return true
		"assert_loot_presentation":
			var def_id := str(step.get("item_def_id", ""))
			var presentations: Dictionary = state.get("loot_presentations", {})
			if not bool(presentations.get(def_id, false)):
				_fail("assert_loot_presentation failed: %s missing from loot presentation state step=%d scenario=%s" % [
					def_id, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_presentation":
			var def_id := str(step.get("item_def_id", ""))
			var panel: Dictionary = state.get("inventory_panel", {})
			var presentations: Dictionary = panel.get("item_presentations", {})
			if not bool(presentations.get(def_id, false)):
				_fail("assert_inventory_presentation failed: %s missing from inventory presentation state step=%d scenario=%s" % [
					def_id, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_hotbar_assigned":
			if not _hotbar_slot_matches(step, state):
				_fail("assert_hotbar_assigned failed: slot=%d item_def_id=%s step=%d scenario=%s" % [
					int(step.get("slot_index", -1)), str(step.get("item_def_id", "")),
					_step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_hotbar_capacity":
			if not _hotbar_capacity_matches(step, state):
				var bar: Dictionary = state.get("consumable_bar", {})
				_fail("assert_hotbar_capacity failed: got=%d want=%d step=%d scenario=%s" % [
					int(bar.get("hotbar_capacity", -1)), int(step.get("equals", -1)),
					_step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_hotbar_slot_disabled":
			var disabled_slot_index := int(step.get("slot_index", -1))
			var bar: Dictionary = state.get("consumable_bar", {})
			var cap := int(bar.get("hotbar_capacity", 2))
			if disabled_slot_index < cap:
				_fail("assert_hotbar_slot_disabled failed: slot=%d capacity=%d step=%d scenario=%s" % [
					disabled_slot_index, cap, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_capacity":
			return _assert_inventory_capacity(step, state)
		"assert_bag_grid":
			return _assert_bag_grid(step, state)
		"assert_paper_doll_layout":
			return _assert_paper_doll_layout(step, state)
		"assert_inventory_panel_details":
			return _assert_inventory_panel_details(step, state)
		"assert_player_hp":
			var want_hp := int(step.get("equals", -1))
			var got_hp := int(state.get("player_hp", -1))
			if got_hp != want_hp:
				_fail("assert_player_hp failed: want=%d got=%d step=%d scenario=%s" % [
					want_hp, got_hp, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_count":
			var def_id := str(step.get("item_def_id", ""))
			var want := int(step.get("equals", 0))
			var got := _inventory_count(state, def_id)
			if got != want:
				_fail("assert_inventory_count failed: %s want=%d got=%d step=%d scenario=%s" % [
					def_id, want, got, _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
	return true


func _assert_character_progression(step: Dictionary, state: Dictionary) -> bool:
	if _progression_matches(step, state):
		return true
	var progression: Dictionary = state.get("character_progression", {})
	_fail("assert_character_progression failed: want=%s got=%s step=%d scenario=%s" % [
		str(_progression_expectation(step)), str(progression), _step_index, str(scenario.get("id", "?"))
	])
	return false


func _assert_character_info(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("character_info_panel", {})
	for key in ["name", "area"]:
		if step.has(key) and str(panel.get(key, "")) != str(step.get(key, "")):
			_fail("assert_character_info failed: %s want=%s got=%s panel=%s step=%d scenario=%s" % [
				key, str(step.get(key, "")), str(panel.get(key, "")), str(panel),
				_step_index, str(scenario.get("id", "?"))
			])
			return false
	if step.has("level") and int(panel.get("level", -999999)) != int(step.get("level", 0)):
		_fail("assert_character_info failed: level want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("level", 0)), int(panel.get("level", -999999)), str(panel),
			_step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _progression_matches(step: Dictionary, state: Dictionary) -> bool:
	var progression: Dictionary = state.get("character_progression", {})
	if progression.is_empty():
		return false
	for key in ["level", "experience", "unspent_stat_points", "gold", "deepest_dungeon_depth"]:
		if step.has(key) and int(progression.get(key, -999999)) != int(step.get(key, 0)):
			return false
	var base: Dictionary = progression.get("base_stats", {})
	for stat in ["str", "dex", "vit", "magic"]:
		if step.has(stat) and int(base.get(stat, -999999)) != int(step.get(stat, 0)):
			return false
	if step.has("derived_stats"):
		var expected: Dictionary = step.get("derived_stats", {})
		var derived: Dictionary = progression.get("derived_stats", {})
		for key in expected.keys():
			if not _float_close(float(derived.get(key, -999999.0)), float(expected[key]), float(step.get("tolerance", 0.001))):
				return false
	if step.has("player_max_hp") and int(state.get("player_max_hp", -999999)) != int(step.get("player_max_hp", 0)):
		return false
	if step.has("stat_breakdowns") and not _stat_breakdowns_match(step.get("stat_breakdowns", []), progression):
		return false
	return true


func _progression_expectation(step: Dictionary) -> Dictionary:
	var out := {}
	for key in ["level", "experience", "unspent_stat_points", "gold", "deepest_dungeon_depth", "str", "dex", "vit", "magic", "derived_stats", "player_max_hp", "stat_breakdowns"]:
		if step.has(key):
			out[key] = step[key]
	return out


func _assert_skill_progression(step: Dictionary, state: Dictionary) -> bool:
	if _skill_progression_matches(step, state):
		return true
	var progression: Dictionary = state.get("skill_progression", {})
	_fail("assert_skill_progression failed: want=%s got=%s step=%d scenario=%s" % [
		str(_skill_progression_expectation(step)), str(progression), _step_index, str(scenario.get("id", "?"))
	])
	return false


func _skill_progression_matches(step: Dictionary, state: Dictionary) -> bool:
	var progression: Dictionary = state.get("skill_progression", {})
	if progression.is_empty():
		return false
	if step.has("unspent_skill_points") and int(progression.get("unspent_skill_points", -999999)) != int(step.get("unspent_skill_points", 0)):
		return false
	var skill_id := str(step.get("skill_id", "magic_bolt"))
	var row := _skill_progression_row(progression, skill_id)
	for key in ["rank", "max_rank"]:
		if step.has(key) and int(row.get(key, -999999)) != int(step.get(key, 0)):
			return false
	if step.has("can_spend") and bool(row.get("can_spend", false)) != bool(step.get("can_spend", false)):
		return false
	return true


func _skill_progression_row(progression: Dictionary, skill_id: String) -> Dictionary:
	var rows: Array = progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return row as Dictionary
	return {}


func _skill_progression_expectation(step: Dictionary) -> Dictionary:
	var out := {}
	for key in ["unspent_skill_points", "skill_id", "rank", "max_rank", "can_spend"]:
		if step.has(key):
			out[key] = step[key]
	return out


func _assert_skill_button_enabled(step: Dictionary, state: Dictionary) -> bool:
	var want := bool(step.get("enabled", true))
	var panel: Dictionary = state.get("skills_panel", {})
	var got := bool(panel.get("spend_button_enabled", false))
	if want != got:
		_fail("assert_skill_button_enabled failed: skill=%s want=%s got=%s panel=%s step=%d scenario=%s" % [
			str(step.get("skill_id", "magic_bolt")), want, got, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _skill_bar_matches(step: Dictionary, state: Dictionary) -> bool:
	var bar: Dictionary = state.get("skill_bar", {})
	if bar.is_empty():
		return false
	if step.has("skill_id") and str(bar.get("skill_id", "")) != str(step.get("skill_id", "")):
		return false
	for key in ["rank", "max_rank", "remaining_ticks", "total_ticks"]:
		if step.has(key) and int(bar.get(key, -999999)) != int(step.get(key, 0)):
			return false
	if step.has("enabled") and bool(bar.get("enabled", false)) != bool(step.get("enabled", false)):
		return false
	if step.has("disabled") and bool(bar.get("disabled", false)) != bool(step.get("disabled", false)):
		return false
	if step.has("remaining_ticks_min") and int(bar.get("remaining_ticks", -999999)) < int(step.get("remaining_ticks_min", 0)):
		return false
	if step.has("remaining_ticks_max") and int(bar.get("remaining_ticks", 999999)) > int(step.get("remaining_ticks_max", 0)):
		return false
	if step.has("cooldown_fraction_min") and float(bar.get("cooldown_fraction", -1.0)) < float(step.get("cooldown_fraction_min", 0.0)):
		return false
	if step.has("cooldown_fraction_max") and float(bar.get("cooldown_fraction", 2.0)) > float(step.get("cooldown_fraction_max", 1.0)):
		return false
	return true


func _skill_bar_expectation(step: Dictionary) -> Dictionary:
	var out := {}
	for key in ["skill_id", "rank", "max_rank", "enabled", "disabled", "remaining_ticks", "remaining_ticks_min", "remaining_ticks_max", "total_ticks", "cooldown_fraction_min", "cooldown_fraction_max"]:
		if step.has(key):
			out[key] = step[key]
	return out


func _stat_breakdowns_match(expected_rows, progression: Dictionary) -> bool:
	if typeof(expected_rows) != TYPE_ARRAY:
		return false
	var by_key := {}
	var actual_rows: Array = progression.get("stat_breakdowns", [])
	for row in actual_rows:
		if typeof(row) == TYPE_DICTIONARY:
			by_key[str((row as Dictionary).get("key", ""))] = row
	for expected in expected_rows:
		if typeof(expected) != TYPE_DICTIONARY:
			return false
		var want := expected as Dictionary
		var key := str(want.get("key", ""))
		if not by_key.has(key):
			return false
		var got: Dictionary = by_key[key]
		var tolerance := float(want.get("tolerance", 0.001))
		if want.has("value") and not _float_close(float(got.get("value", -999999.0)), float(want.get("value", 0.0)), tolerance):
			return false
		if want.has("min_value") and float(got.get("value", -999999.0)) < float(want.get("min_value", 0.0)):
			return false
		if want.has("uncapped_value") and not _float_close(float(got.get("uncapped_value", -999999.0)), float(want.get("uncapped_value", 0.0)), tolerance):
			return false
		if want.has("min_uncapped_value") and float(got.get("uncapped_value", -999999.0)) < float(want.get("min_uncapped_value", 0.0)):
			return false
		if want.has("cap"):
			var got_cap = got.get("cap", null)
			if got_cap == null or not _float_close(float(got_cap), float(want.get("cap", 0.0)), tolerance):
				return false
		if want.has("source_kinds"):
			var source_kinds: Array = []
			for source in got.get("sources", []):
				if typeof(source) == TYPE_DICTIONARY:
					source_kinds.append(str((source as Dictionary).get("kind", "")))
			for kind in want.get("source_kinds", []):
				if not source_kinds.has(str(kind)):
					return false
	return true


func _event_matches(step: Dictionary, event) -> bool:
	if typeof(event) != TYPE_DICTIONARY:
		return false
	var ev := event as Dictionary
	for key in ["outcome", "source_entity_id", "target_entity_id", "shop_id", "offer_id", "item_instance_id", "skill_id"]:
		if step.has(key) and str(ev.get(key, "")) != str(step.get(key, "")):
			return false
	for key in ["damage", "raw_damage", "mitigated_damage", "price", "total_gold", "level", "from_level", "to_level", "rank", "mana", "remaining_ticks", "total_ticks", "amount", "unspent_skill_points"]:
		if step.has(key) and int(ev.get(key, -999999)) != int(step.get(key, 0)):
			return false
	if step.has("min_damage") and int(ev.get("damage", -999999)) < int(step.get("min_damage", 0)):
		return false
	for key in ["blocked", "critical"]:
		if step.has(key) and bool(ev.get(key, false)) != bool(step.get(key, false)):
			return false
	return true


func _damage_number_matches(step: Dictionary, state: Dictionary) -> bool:
	var numbers: Array = state.get("damage_numbers", [])
	for number in numbers:
		if typeof(number) != TYPE_DICTIONARY:
			continue
		var rec := number as Dictionary
		if step.has("text") and str(rec.get("text", "")) != str(step.get("text", "")):
			continue
		if step.has("variant") and str(rec.get("variant", "")) != str(step.get("variant", "")):
			continue
		return true
	return false


func _presentation_matches(step: Dictionary, state: Dictionary) -> bool:
	var rows: Array = []
	if bool(step.get("local_player", false)):
		rows.append(state.get("local_player_presentation", {}))
	else:
		rows = state.get("entities_presentation_debug", [])
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if not _presentation_row_matches(step, rec):
			continue
		return true
	return false


func _presentation_row_matches(step: Dictionary, rec: Dictionary) -> bool:
	if step.has("entity_type") and str(rec.get("type", "")) != str(step.get("entity_type", "")):
		return false
	for key in ["id", "monster_def_id", "character_id", "visual_model", "base_tint"]:
		if step.has(key) and str(rec.get(key, "")) != str(step.get(key, "")):
			return false
	if step.has("hp") and int(rec.get("hp", -999999)) != int(step.get("hp", 0)):
		return false
	if step.has("has_bow_marker") and bool(rec.get("has_bow_marker", false)) != bool(step.get("has_bow_marker", false)):
		return false
	var reaction: Dictionary = rec.get("reaction", {})
	if step.has("reaction") and str(reaction.get("last_reaction", "")) != str(step.get("reaction", "")):
		return false
	if step.has("terminal") and bool(reaction.get("terminal", false)) != bool(step.get("terminal", false)):
		return false
	if step.has("current_tint") and str(reaction.get("current_tint", "")) != str(step.get("current_tint", "")):
		return false
	return true


func _wall_layout_matches(step: Dictionary, state: Dictionary) -> bool:
	if step.has("current_level") and int(state.get("current_level", -999999)) != int(step.get("current_level", 0)):
		return false
	var wall_count := int(state.get("wall_count", 0))
	var generated_count := int(state.get("generated_wall_count", 0))
	var non_perimeter_count := int(state.get("non_perimeter_wall_count", 0))
	if step.has("equals") and wall_count != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and wall_count < int(step.get("at_least", 0)):
		return false
	if step.has("generated_at_least") and generated_count < int(step.get("generated_at_least", 0)):
		return false
	if step.has("non_perimeter_at_least") and non_perimeter_count < int(step.get("non_perimeter_at_least", 0)):
		return false
	return true


func _assert_stat_button_enabled(step: Dictionary, state: Dictionary) -> bool:
	var stat := str(step.get("stat", ""))
	var want := bool(step.get("enabled", true))
	var panel: Dictionary = state.get("character_stats_panel", {})
	var buttons: Dictionary = panel.get("stat_buttons", {})
	var button: Dictionary = buttons.get(stat, {})
	var got := bool(button.get("enabled", false))
	if want != got:
		_fail("assert_stat_button_enabled failed: stat=%s want=%s got=%s panel=%s step=%d scenario=%s" % [
			stat, want, got, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_xp_bar(step: Dictionary, state: Dictionary) -> bool:
	var bar: Dictionary = state.get("consumable_bar", {})
	var xp_bar: Dictionary = bar.get("xp_bar", {})
	for key in ["level", "experience"]:
		if step.has(key) and int(xp_bar.get(key, -999999)) != int(step.get(key, 0)):
			_fail("assert_xp_bar failed: %s want=%s got=%s xp_bar=%s step=%d scenario=%s" % [
				key, str(step.get(key, "")), str(xp_bar.get(key, "")), str(xp_bar),
				_step_index, str(scenario.get("id", "?"))
			])
			return false
	if step.has("progress_min") and float(xp_bar.get("progress", -1.0)) < float(step.get("progress_min", 0.0)):
		_fail("assert_xp_bar failed: progress below min xp_bar=%s step=%d scenario=%s" % [
			str(xp_bar), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("progress_max") and float(xp_bar.get("progress", 2.0)) > float(step.get("progress_max", 1.0)):
		_fail("assert_xp_bar failed: progress above max xp_bar=%s step=%d scenario=%s" % [
			str(xp_bar), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_inventory_capacity(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("inventory_panel", {})
	var got_rows := int(panel.get("inventory_rows", state.get("inventory_rows", -1)))
	var got_capacity := int(panel.get("inventory_capacity", state.get("inventory_capacity", -1)))
	if step.has("rows") and got_rows != int(step.get("rows", -1)):
		_fail("assert_inventory_capacity failed: rows want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("rows", -1)), got_rows, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("capacity") and got_capacity != int(step.get("capacity", -1)):
		_fail("assert_inventory_capacity failed: capacity want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("capacity", -1)), got_capacity, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_bag_grid(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("inventory_panel", {})
	var got_columns := int(panel.get("bag_columns", -1))
	var got_slots := int(panel.get("available_slot_count", -1))
	if step.has("columns") and got_columns != int(step.get("columns", -1)):
		_fail("assert_bag_grid failed: columns want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("columns", -1)), got_columns, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("available_slot_count") and got_slots != int(step.get("available_slot_count", -1)):
		_fail("assert_bag_grid failed: slots want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("available_slot_count", -1)), got_slots, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_paper_doll_layout(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("inventory_panel", {})
	var preview: Dictionary = panel.get("paper_doll_preview", {})
	if bool(step.get("preview", true)) and (not bool(preview.get("exists", false)) or not bool(preview.get("visible", false))):
		_fail("assert_paper_doll_layout failed: preview missing preview=%s step=%d scenario=%s" % [
			str(preview), _step_index, str(scenario.get("id", "?"))
		])
		return false
	var slots: Dictionary = panel.get("paper_doll_slots", {})
	var expected_slots: Array = step.get("slots", [])
	for slot in expected_slots:
		var slot_id := str(slot)
		var rec: Dictionary = slots.get(slot_id, {})
		if not bool(rec.get("exists", false)):
			_fail("assert_paper_doll_layout failed: missing slot=%s slots=%s step=%d scenario=%s" % [
				slot_id, str(slots.keys()), _step_index, str(scenario.get("id", "?"))
			])
			return false
	return true


func _assert_inventory_panel_details(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("inventory_panel", {})
	if bool(step.get("visible", false)) and not bool(state.get("inventory_panel_visible", false)):
		_fail("assert_inventory_panel_details failed: panel hidden step=%d scenario=%s" % [
			_step_index, str(scenario.get("id", "?"))
		])
		return false
	var requirement_rows := int(panel.get("requirement_row_count", 0))
	var preview_rows := int(panel.get("equip_preview_row_count", 0))
	if bool(step.get("requires_requirement_status", false)) and requirement_rows <= 0:
		_fail("assert_inventory_panel_details failed: missing requirement rows panel=%s step=%d scenario=%s" % [
			str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if bool(step.get("requires_equip_preview", false)) and preview_rows <= 0:
		_fail("assert_inventory_panel_details failed: missing equip preview rows panel=%s step=%d scenario=%s" % [
			str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("requirement_rows_at_least") and requirement_rows < int(step.get("requirement_rows_at_least", 0)):
		_fail("assert_inventory_panel_details failed: requirement rows want>=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("requirement_rows_at_least", 0)), requirement_rows, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("equip_preview_rows_at_least") and preview_rows < int(step.get("equip_preview_rows_at_least", 0)):
		_fail("assert_inventory_panel_details failed: preview rows want>=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("equip_preview_rows_at_least", 0)), preview_rows, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_shop_offer_count(step: Dictionary, state: Dictionary) -> bool:
	if _shop_offer_count_matches(step, state):
		return true
	var panel: Dictionary = state.get("shop_panel", {})
	_fail("assert_shop_offer_count failed: want=%s panel=%s step=%d scenario=%s" % [
		str(step), str(panel), _step_index, str(scenario.get("id", "?"))
	])
	return false


func _assert_shop_buy_button(step: Dictionary, state: Dictionary) -> bool:
	var offer_id := str(step.get("offer_id", ""))
	var panel: Dictionary = state.get("shop_panel", {})
	var buttons: Dictionary = panel.get("buy_buttons", {})
	var button: Dictionary = buttons.get(offer_id, {})
	if button.is_empty():
		_fail("assert_shop_buy_button failed: missing offer_id=%s buttons=%s step=%d scenario=%s" % [
			offer_id, str(buttons), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("enabled") and bool(button.get("enabled", false)) != bool(step.get("enabled", true)):
		_fail("assert_shop_buy_button failed: offer_id=%s enabled want=%s got=%s step=%d scenario=%s" % [
			offer_id, str(step.get("enabled", true)), str(button.get("enabled", false)),
			_step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_shop_sell_rows(step: Dictionary, state: Dictionary) -> bool:
	var rows := _matching_shop_sell_rows(step, state)
	if step.has("equals") and rows.size() != int(step.get("equals", 0)):
		_fail("assert_shop_sell_rows failed: equals want=%d got=%d rows=%s step=%d scenario=%s" % [
			int(step.get("equals", 0)), rows.size(), str(rows), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("at_least") and rows.size() < int(step.get("at_least", 0)):
		_fail("assert_shop_sell_rows failed: at_least want=%d got=%d rows=%s step=%d scenario=%s" % [
			int(step.get("at_least", 0)), rows.size(), str(rows), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_shop_offer_details(step: Dictionary, state: Dictionary) -> bool:
	return _assert_shop_detail_rows("assert_shop_offer_details", step, _matching_shop_offer_rows(step, state), "buy_price")


func _assert_shop_sell_details(step: Dictionary, state: Dictionary) -> bool:
	return _assert_shop_detail_rows("assert_shop_sell_details", step, _matching_shop_sell_rows(step, state), "sell_price")


func _assert_stash_item_count(step: Dictionary, state: Dictionary) -> bool:
	if _stash_item_count_matches(step, state):
		return true
	var panel: Dictionary = state.get("stash_panel", {})
	_fail("assert_stash_item_count failed: want=%s panel=%s step=%d scenario=%s" % [
		str(step), str(panel), _step_index, str(scenario.get("id", "?"))
	])
	return false


func _assert_stash_gold(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("stash_panel", {})
	var got := int(panel.get("stash_gold", state.get("stash_gold", 0)))
	if step.has("equals") and got != int(step.get("equals", 0)):
		_fail("assert_stash_gold failed: equals want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("equals", 0)), got, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("at_least") and got < int(step.get("at_least", 0)):
		_fail("assert_stash_gold failed: at_least want=%d got=%d panel=%s step=%d scenario=%s" % [
			int(step.get("at_least", 0)), got, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_shop_detail_rows(label: String, step: Dictionary, rows: Array, price_key: String) -> bool:
	if step.has("equals") and rows.size() != int(step.get("equals", 0)):
		_fail("%s failed: equals want=%d got=%d rows=%s step=%d scenario=%s" % [
			label, int(step.get("equals", 0)), rows.size(), str(rows), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("at_least") and rows.size() < int(step.get("at_least", 0)):
		_fail("%s failed: at_least want=%d got=%d rows=%s step=%d scenario=%s" % [
			label, int(step.get("at_least", 0)), rows.size(), str(rows), _step_index, str(scenario.get("id", "?"))
		])
		return false
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			_fail("%s failed: non-dictionary row=%s step=%d scenario=%s" % [
				label, str(row), _step_index, str(scenario.get("id", "?"))
			])
			return false
		var rec := row as Dictionary
		if bool(step.get("requires_price", false)) and int(rec.get(price_key, 0)) <= 0:
			_fail("%s failed: missing positive %s row=%s step=%d scenario=%s" % [
				label, price_key, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_summary", false)) and _row_summary_lines(rec).is_empty():
			_fail("%s failed: missing summary row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_comparison", false)) and int(rec.get("comparison_count", 0)) <= 0:
			_fail("%s failed: missing comparison row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_requirement_status", false)) and int(rec.get("requirement_count", 0)) <= 0:
			_fail("%s failed: missing requirement status row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_equip_preview", false)) and int(rec.get("equip_preview_count", 0)) <= 0:
			_fail("%s failed: missing equip preview row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_slot", false)) and str(rec.get("slot", "")) == "":
			_fail("%s failed: missing slot row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_category", false)) and str(rec.get("category", "")) == "":
			_fail("%s failed: missing category row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_concealed", false)) and not bool(rec.get("concealed", false)):
			_fail("%s failed: missing concealed flag row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("requires_mystery_label", false)) and str(rec.get("mystery_label", "")) == "":
			_fail("%s failed: missing mystery label row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if bool(step.get("forbids_item_identity", false)) and _row_has_item_identity(rec):
			_fail("%s failed: item identity leaked row=%s step=%d scenario=%s" % [
				label, str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
		if step.has("summary_contains") and not _row_summary_contains(rec, step.get("summary_contains", "")):
			_fail("%s failed: summary_contains=%s row=%s step=%d scenario=%s" % [
				label, str(step.get("summary_contains", "")), str(rec), _step_index, str(scenario.get("id", "?"))
			])
			return false
	return true


func _shop_offer_count_matches(step: Dictionary, state: Dictionary) -> bool:
	if not step.has("equals") and not step.has("at_least"):
		return true
	var panel: Dictionary = state.get("shop_panel", {})
	var key := "offer_count"
	match str(step.get("offer_kind", "")):
		"fixed":
			key = "fixed_offer_count"
		"generated":
			key = "generated_offer_count"
		"mystery":
			key = "mystery_offer_count"
		"buyback":
			key = "buyback_offer_count"
	var got := int(panel.get(key, 0))
	if step.has("equals") and got != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and got < int(step.get("at_least", 0)):
		return false
	return true


func _stash_item_count_matches(step: Dictionary, state: Dictionary) -> bool:
	if not step.has("equals") and not step.has("at_least"):
		return true
	var rows := _matching_stash_rows(step, state)
	if step.has("equals") and rows.size() != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and rows.size() < int(step.get("at_least", 0)):
		return false
	return true


func _matching_stash_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("stash_panel", {})
	var rows: Array = panel.get("stash_rows", state.get("stash_items", []))
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("stash_item_id") and str(rec.get("stash_item_id", "")) != str(step.get("stash_item_id", "")):
			continue
		if step.has("item_def_id") and str(rec.get("item_def_id", "")) != str(step.get("item_def_id", "")):
			continue
		if step.has("item_template_id") and str(rec.get("item_template_id", "")) != str(step.get("item_template_id", "")):
			continue
		if step.has("rolled") and (str(rec.get("item_template_id", "")) != "") != bool(step.get("rolled", false)):
			continue
		out.append(rec)
	return out


func _matching_shop_offer_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("shop_panel", {})
	var rows: Array = panel.get("offer_rows", [])
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("offer_kind") and str(rec.get("kind", "")) != str(step.get("offer_kind", "")):
			continue
		if step.has("offer_id") and str(rec.get("offer_id", "")) != str(step.get("offer_id", "")):
			continue
		if step.has("item_def_id") and str(rec.get("item_def_id", "")) != str(step.get("item_def_id", "")):
			continue
		if step.has("item_template_id") and str(rec.get("item_template_id", "")) != str(step.get("item_template_id", "")):
			continue
		if step.has("concealed") and bool(rec.get("concealed", false)) != bool(step.get("concealed", false)):
			continue
		if step.has("mystery_label") and str(rec.get("mystery_label", "")) != str(step.get("mystery_label", "")):
			continue
		var row_source_min := int(rec.get("source_depth_min", rec.get("source_depth", 0)))
		var row_source_max := int(rec.get("source_depth_max", rec.get("source_depth", 0)))
		if step.has("source_depth_min") and row_source_min < int(step.get("source_depth_min", 0)):
			continue
		if step.has("source_depth_max") and row_source_max > int(step.get("source_depth_max", 0)):
			continue
		if step.has("rolled") and (str(rec.get("item_template_id", "")) != "") != bool(step.get("rolled", false)):
			continue
		out.append(rec)
	return out


func _matching_shop_sell_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("shop_panel", {})
	var rows: Array = panel.get("sell_rows", [])
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("item_def_id") and str(rec.get("item_def_id", "")) != str(step.get("item_def_id", "")):
			continue
		if step.has("item_template_id") and str(rec.get("item_template_id", "")) != str(step.get("item_template_id", "")):
			continue
		if step.has("rolled") and (str(rec.get("item_template_id", "")) != "") != bool(step.get("rolled", false)):
			continue
		out.append(rec)
	return out


func _row_has_item_identity(row: Dictionary) -> bool:
	if int(row.get("identity_field_count", 0)) > 0:
		return true
	for key in ["item_def_id", "item_template_id", "rarity"]:
		if str(row.get(key, "")) != "":
			return true
	return false


func _row_summary_lines(row: Dictionary) -> Array:
	var summary = row.get("summary_lines", [])
	if typeof(summary) != TYPE_ARRAY:
		return []
	return summary as Array


func _row_summary_contains(row: Dictionary, needle_value: Variant) -> bool:
	var needles: Array = []
	if typeof(needle_value) == TYPE_ARRAY:
		needles = needle_value as Array
	else:
		needles.append(str(needle_value))
	var lines := _row_summary_lines(row)
	for needle in needles:
		var text := str(needle)
		if text == "":
			continue
		var matched := false
		for line in lines:
			if str(line).findn(text) >= 0:
				matched = true
				break
		if not matched:
			return false
	return true


func _float_close(got: float, want: float, tolerance: float) -> bool:
	return absf(got - want) <= tolerance


func _inventory_count(state: Dictionary, item_def_id: String) -> int:
	var inv: Array = state.get("inventory", [])
	if item_def_id == "":
		return inv.size()
	var count := 0
	for item in inv:
		if str(item.get("item_def_id", "")) == item_def_id:
			count += 1
	return count


func _hotbar_slot_matches(step: Dictionary, state: Dictionary) -> bool:
	var slot_index := int(step.get("slot_index", -1))
	var want_def := str(step.get("item_def_id", ""))
	var bar: Dictionary = state.get("consumable_bar", {})
	var assigned: Array = bar.get("assigned_slots", [])
	if slot_index < 0 or slot_index >= assigned.size():
		return false
	var slot_val = assigned[slot_index]
	if slot_val == null or typeof(slot_val) != TYPE_DICTIONARY:
		return false
	return str((slot_val as Dictionary).get("item_def_id", "")) == want_def


func _hotbar_capacity_matches(step: Dictionary, state: Dictionary) -> bool:
	var bar: Dictionary = state.get("consumable_bar", {})
	var got := int(bar.get("hotbar_capacity", -1))
	if step.has("equals"):
		return got == int(step.get("equals", -1))
	if step.has("at_least"):
		return got >= int(step.get("at_least", -1))
	return false


func _queue_action(step: Dictionary, stype: String, state: Dictionary) -> void:
	if stype == "remember_session":
		_memory["session_id"] = str(state.get("current_session_id", ""))
		return
	if stype == "remember_player_position":
		_memory["player_pos"] = (state.get("player_pos", {}) as Dictionary).duplicate(true)
		return
	pending_action = step.duplicate()
	pending_action["_type"] = stype


func _assert_bool_state(label: String, key: String, step: Dictionary, state: Dictionary) -> bool:
	var want := bool(step.get("visible", true))
	var got := bool(state.get(key, false))
	return _assert_bool_value(label, step, got, want)


func _assert_bool_value(label: String, step: Dictionary, got: bool, want: bool) -> bool:
	if want != got:
		_fail("%s failed: want=%s got=%s step=%d scenario=%s" % [
			label, want, got, _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_main_menu_actions(step: Dictionary, state: Dictionary) -> bool:
	var labels: Array = state.get("main_menu_button_labels", [])
	if step.has("labels"):
		var expected_labels: Array = step.get("labels", [])
		if labels.size() != expected_labels.size():
			_fail("assert_main_menu_actions labels failed: want=%s got=%s step=%d scenario=%s" % [
				str(expected_labels), str(labels), _step_index, str(scenario.get("id", "?"))
			])
			return false
		for i in range(expected_labels.size()):
			if str(labels[i]) != str(expected_labels[i]):
				_fail("assert_main_menu_actions label[%d] failed: want=%s got=%s labels=%s step=%d scenario=%s" % [
					i, str(expected_labels[i]), str(labels[i]), str(labels), _step_index, str(scenario.get("id", "?"))
				])
				return false
	if step.has("actions"):
		var actions: Array = state.get("main_menu_actions", [])
		for action in step.get("actions", []):
			if not actions.has(str(action)):
				_fail("assert_main_menu_actions action missing: want=%s actions=%s step=%d scenario=%s" % [
					str(action), str(actions), _step_index, str(scenario.get("id", "?"))
				])
				return false
	return true


func _assert_character_panel(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("character_panel", {})
	if step.has("visible"):
		var visible := bool(panel.get("visible", state.get("character_panel_visible", false)))
		if visible != bool(step.get("visible", true)):
			_fail("assert_character_panel visible failed: want=%s got=%s step=%d scenario=%s" % [
				str(step.get("visible", true)), str(visible), _step_index, str(scenario.get("id", "?"))
			])
			return false
	if step.has("mode") and str(panel.get("mode", state.get("character_panel_mode", ""))) != str(step.get("mode", "")):
		_fail("assert_character_panel mode failed: want=%s got=%s panel=%s step=%d scenario=%s" % [
			str(step.get("mode", "")), str(panel.get("mode", "")), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("title") and str(panel.get("title", state.get("character_panel_title", ""))) != str(step.get("title", "")):
		_fail("assert_character_panel title failed: want=%s got=%s panel=%s step=%d scenario=%s" % [
			str(step.get("title", "")), str(panel.get("title", "")), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	var characters: Array = panel.get("characters", state.get("known_characters", []))
	if step.has("character_count") and characters.size() != int(step.get("character_count", 0)):
		_fail("assert_character_panel character_count failed: want=%d got=%d step=%d scenario=%s" % [
			int(step.get("character_count", 0)), characters.size(), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("min_character_count") and characters.size() < int(step.get("min_character_count", 0)):
		_fail("assert_character_panel min_character_count failed: want=%d got=%d step=%d scenario=%s" % [
			int(step.get("min_character_count", 0)), characters.size(), _step_index, str(scenario.get("id", "?"))
		])
		return false
	for key in ["name_field_visible", "create_button_visible", "empty_visible"]:
		if step.has(key) and bool(panel.get(key, false)) != bool(step.get(key, false)):
			_fail("assert_character_panel %s failed: want=%s got=%s panel=%s step=%d scenario=%s" % [
				key, str(step.get(key, false)), str(panel.get(key, false)), str(panel), _step_index, str(scenario.get("id", "?"))
			])
			return false
	return true


func _assert_create_game_type(step: Dictionary, state: Dictionary) -> bool:
	var want := str(step.get("session_type", ""))
	var got := str(state.get("create_game_session_type", ""))
	if got != want:
		_fail("assert_create_game_type failed: want=%s got=%s step=%d scenario=%s" % [
			want, got, _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_current_session(step: Dictionary, state: Dictionary) -> bool:
	var session_id := str(state.get("current_session_id", ""))
	var expected_session_id := _expected_session_id(step)
	if step.has("exists") and bool(step.get("exists", true)) != (session_id != ""):
		_fail("assert_current_session exists failed: want=%s got_session=%s step=%d scenario=%s" % [
			str(step.get("exists", true)), session_id, _step_index, str(scenario.get("id", "?"))
		])
		return false
	if expected_session_id != "" and session_id != expected_session_id:
		_fail("assert_current_session session_id failed: want=%s got=%s step=%d scenario=%s" % [
			expected_session_id, session_id, _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("mode") and str(state.get("current_session_mode", "")) != str(step.get("mode", "")):
		_fail("assert_current_session mode failed: want=%s got=%s step=%d scenario=%s" % [
			str(step.get("mode", "")), str(state.get("current_session_mode", "")), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("listed") and bool(state.get("current_session_listed", false)) != bool(step.get("listed", false)):
		_fail("assert_current_session listed failed: want=%s got=%s step=%d scenario=%s" % [
			str(step.get("listed", false)), str(state.get("current_session_listed", false)), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _assert_multiplayer_session_rows(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("multiplayer_panel", {})
	var sessions: Array = panel.get("sessions", [])
	var expected_session_id := _expected_session_id(step)
	if step.has("equals") and sessions.size() != int(step.get("equals", 0)):
		_fail("assert_multiplayer_session_rows equals failed: count=%d equals=%d rows=%s step=%d scenario=%s" % [
			sessions.size(), int(step.get("equals", 0)), str(sessions), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("min_count") and sessions.size() < int(step.get("min_count", 0)):
		_fail("assert_multiplayer_session_rows failed: count=%d min_count=%d rows=%s step=%d scenario=%s" % [
			sessions.size(), int(step.get("min_count", 0)), str(sessions), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("selected") and bool(step.get("selected", false)) != (str(panel.get("selected_session_id", "")) != ""):
		_fail("assert_multiplayer_session_rows selected failed: selected_id=%s want_selected=%s step=%d scenario=%s" % [
			str(panel.get("selected_session_id", "")), str(step.get("selected", false)), _step_index, str(scenario.get("id", "?"))
		])
		return false
	var rows_to_check: Array = sessions
	if expected_session_id != "":
		rows_to_check = []
		for row in sessions:
			if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("session_id", "")) == expected_session_id:
				rows_to_check.append(row)
		if rows_to_check.is_empty():
			_fail("assert_multiplayer_session_rows session_id missing: want=%s rows=%s step=%d scenario=%s" % [
				expected_session_id, str(sessions), _step_index, str(scenario.get("id", "?"))
			])
			return false
	if step.has("listed"):
		for row in rows_to_check:
			if typeof(row) == TYPE_DICTIONARY and bool((row as Dictionary).get("listed", false)) != bool(step.get("listed", true)):
				_fail("assert_multiplayer_session_rows listed failed: row=%s step=%d scenario=%s" % [
					str(row), _step_index, str(scenario.get("id", "?"))
				])
				return false
	if step.has("mode"):
		for row in rows_to_check:
			if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("mode", "")) != str(step.get("mode", "")):
				_fail("assert_multiplayer_session_rows mode failed: row=%s want=%s step=%d scenario=%s" % [
					str(row), str(step.get("mode", "")), _step_index, str(scenario.get("id", "?"))
				])
				return false
	if step.has("member_count_min"):
		for row in rows_to_check:
			if typeof(row) == TYPE_DICTIONARY and int((row as Dictionary).get("member_count", 0)) < int(step.get("member_count_min", 0)):
				_fail("assert_multiplayer_session_rows member_count_min failed: row=%s step=%d scenario=%s" % [
					str(row), _step_index, str(scenario.get("id", "?"))
				])
				return false
	if step.has("connected_count_min"):
		for row in rows_to_check:
			if typeof(row) == TYPE_DICTIONARY and int((row as Dictionary).get("connected_count", 0)) < int(step.get("connected_count_min", 0)):
				_fail("assert_multiplayer_session_rows connected_count_min failed: row=%s step=%d scenario=%s" % [
					str(row), _step_index, str(scenario.get("id", "?"))
				])
				return false
	return true


func _expected_session_id(step: Dictionary) -> String:
	if step.has("session_id"):
		return str(step.get("session_id", ""))
	var env_key := str(step.get("session_id_env", ""))
	if env_key != "":
		return OS.get_environment(env_key)
	return ""


func _remote_player_count_matches(step: Dictionary, state: Dictionary) -> bool:
	var remote_ids: Array = state.get("remote_player_ids", [])
	if step.has("equals") and remote_ids.size() != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and remote_ids.size() < int(step.get("at_least", 0)):
		return false
	return step.has("equals") or step.has("at_least")


func _advance(completed_type: String = "") -> void:
	print("[bot-client] step done idx=%d type=%s elapsed=%.2fs scenario=%s" % [
		_step_index, completed_type, _step_elapsed, str(scenario.get("id", "?"))
	])
	_step_index += 1
	_step_elapsed = 0.0
	_last_retry_at = -999.0
	_last_wait_log_at = 0.0
	_step_begin_logged = false
	if step_delay_s > 0.0:
		_post_step_wait = step_delay_s


func _log_step_begin(step: Dictionary, stype: String) -> void:
	var timeout_s := float(step.get("timeout_s", 0.0))
	var detail := _step_detail(step, stype)
	print("[bot-client] step begin idx=%d/%d type=%s timeout=%.1fs %s scenario=%s" % [
		_step_index, _steps.size(), stype, timeout_s, detail, str(scenario.get("id", "?"))
	])


func _step_detail(step: Dictionary, stype: String) -> String:
	match stype:
		"wait_entity", "click_entity", "click_entity_until_event", "assert_entity_removed":
			return "entity_type=%s" % str(step.get("entity_type", ""))
		"wait_event", "click_entity_until_event":
			return "event_type=%s entity_type=%s" % [
				str(step.get("event_type", "")), str(step.get("entity_type", ""))
			]
		"wait_inventory_item", "wait_inventory_count", "assert_inventory_count", \
		"assert_inventory_missing", "double_click_bag_item", "assign_hotbar_slot", \
		"click_loot_item", "wait_hotbar_assigned":
			return "item_def_id=%s" % str(step.get("item_def_id", ""))
		"wait_loot_count":
			return "min_count=%s" % str(step.get("min_count", ""))
		"assert_player_hp":
			return "hp=%s" % str(step.get("equals", ""))
		"wait_character_progression", "assert_character_progression":
			return "progression=%s" % str(_progression_expectation(step))
		"assert_stat_button_enabled", "click_stat_button":
			return "stat=%s" % str(step.get("stat", ""))
		"wait_skill_progression", "assert_skill_progression":
			return "skill_progression=%s" % str(_skill_progression_expectation(step))
		"assert_skill_button_enabled", "click_skill_button":
			return "skill_id=%s" % str(step.get("skill_id", "magic_bolt"))
		"wait_skill_bar", "assert_skill_bar", "use_skill_slot":
			return "skill_bar=%s" % str(_skill_bar_expectation(step))
		"assert_xp_bar":
			return "xp_bar=%s" % str(step)
		"click_menu_button":
			return "button=%s" % str(step.get("button", ""))
		"assert_main_menu_actions":
			return "main_menu=%s" % str(step)
		"assert_character_panel":
			return "character_panel=%s" % str(step)
		"assert_create_game_type", "select_create_game_type":
			return "session_type=%s" % str(step.get("session_type", ""))
		"assert_current_session":
			return "session=%s" % str(step)
		"enter_character_name":
			return "name=%s" % str(step.get("name", ""))
		"select_character":
			return "index=%s" % str(step.get("index", 0))
		"select_window_size":
			return "size=%s" % str(step.get("size", ""))
		"set_floating_combat_text", "assert_floating_combat_text_enabled":
			return "enabled=%s" % str(step.get("enabled", true))
		"wait_damage_number", "wait_no_damage_number", "assert_damage_number", "assert_no_damage_number":
			return "damage_number=%s" % str(step)
		"wait_entity_reaction", "assert_entity_reaction":
			return "presentation=%s" % str(step)
		"wait_wall_layout", "assert_wall_layout":
			return "wall_layout=%s" % str(step)
		"wait_shop_panel", "assert_shop_offer_count", "assert_shop_buy_button", "assert_shop_sell_rows", \
		"assert_shop_offer_details", "assert_shop_sell_details", \
		"click_shop_buy_offer", "click_shop_sell_item":
			return "shop=%s" % str(step)
		"click_waypoint_level":
			return "target_level=%s" % str(step.get("target_level", ""))
		"wait_stash_panel", "assert_stash_item_count", "assert_stash_gold", "assert_stash_panel_visible", \
		"drag_bag_to_stash", "drag_stash_to_bag", \
		"click_stash_deposit_gold", "click_stash_withdraw_gold":
			return "stash=%s" % str(step)
		"press_key":
			return "key=%s" % str(step.get("keycode", ""))
		"click_entity":
			return "entity_type=%s index=%s" % [
				str(step.get("entity_type", "")), str(step.get("entity_index", 0))
			]
		_:
			return ""


func _log_wait_progress(step: Dictionary, stype: String, state: Dictionary) -> void:
	var parts: PackedStringArray = PackedStringArray([
		"waiting",
		stype,
		"elapsed=%.1fs" % _step_elapsed,
		"ws=%s" % ("open" if bool(state.get("ws_open", false)) else "closed"),
		"tick=%s" % str(state.get("last_tick", "?")),
		"hp=%s" % str(state.get("player_hp", "?")),
	])
	if stype in ["wait_entity", "assert_entity_removed", "click_entity_until_event"]:
		var etype := str(step.get("entity_type", ""))
		var eids: Array = state.get("%s_ids" % etype, [])
		parts.append("%s_count=%d" % [etype, eids.size()])
	if stype in ["wait_event", "click_entity_until_event"]:
		var pending: Array = state.get("pending_events", [])
		var event_names: PackedStringArray = PackedStringArray()
		for ev in pending:
			event_names.append(str(ev.get("event_type", "?")))
		parts.append("pending_events=[%s]" % ", ".join(event_names))
	if stype in ["wait_inventory_count", "wait_inventory_item", "assert_inventory_count"]:
		var def_id := str(step.get("item_def_id", ""))
		parts.append("inventory_%s=%d" % [def_id, _inventory_count(state, def_id)])
	if stype == "wait_character_progression":
		parts.append("progression=%s" % str(state.get("character_progression", {})))
	if stype == "wait_skill_progression":
		parts.append("skill_progression=%s" % str(state.get("skill_progression", {})))
	if stype == "wait_skill_bar":
		parts.append("skill_bar=%s" % str(state.get("skill_bar", {})))
	if stype in ["wait_damage_number", "wait_no_damage_number"]:
		parts.append("damage_numbers=%s" % str(state.get("damage_numbers", [])))
	if stype == "wait_entity_reaction":
		parts.append("local_presentation=%s" % str(state.get("local_player_presentation", {})))
		parts.append("entity_presentations=%s" % str(state.get("entities_presentation_debug", [])))
	if stype == "wait_loot_count":
		parts.append("loot_count=%d" % (state.get("loot_ids", []) as Array).size())
	if stype == "wait_wall_layout":
		parts.append("wall_count=%d generated=%d non_perimeter=%d level=%d" % [
			int(state.get("wall_count", 0)),
			int(state.get("generated_wall_count", 0)),
			int(state.get("non_perimeter_wall_count", 0)),
			int(state.get("current_level", 0)),
		])
	if stype == "wait_shop_panel":
		parts.append("shop_panel=%s" % str(state.get("shop_panel", {})))
	if stype == "wait_stash_panel":
		parts.append("stash_panel=%s" % str(state.get("stash_panel", {})))
	if stype == "click_entity_until_event" and _step_elapsed - _last_retry_at < float(step.get("retry_s", 0.25)):
		parts.append("next_attack_soon=true")
	print("[bot-client] %s scenario=%s step=%d" % [" ".join(parts), str(scenario.get("id", "?")), _step_index])


func _check_complete() -> void:
	if not _done and _step_index >= _steps.size():
		_pass()


func _pass() -> void:
	_done = true
	_passed = true


func _fail(msg: String) -> void:
	_done = true
	_passed = false
	_failure_msg = msg


# --- Static validation -------------------------------------------------------

static func validate_scenario(data: Dictionary) -> String:
	if str(data.get("runner", "")) != "godot_client":
		return "runner must be 'godot_client', got '%s'" % str(data.get("runner", ""))
	if str(data.get("id", "")) == "":
		return "id must be non-empty"
	if str(data.get("world_id", "")) == "":
		return "world_id must be non-empty"
	var steps = data.get("client_steps", null)
	if steps == null or typeof(steps) != TYPE_ARRAY or (steps as Array).size() == 0:
		return "client_steps must be a non-empty array"
	for i in range((steps as Array).size()):
		var step = (steps as Array)[i]
		if typeof(step) != TYPE_DICTIONARY:
			return "client_steps[%d] must be an object" % i
		var err := validate_step(step as Dictionary, i)
		if err != "":
			return err
	return ""


static func validate_step(step: Dictionary, index: int) -> String:
	var stype := str(step.get("type", ""))
	if stype == "":
		return "client_steps[%d].type is missing" % index
	if stype not in ALL_STEP_TYPES:
		return "client_steps[%d].type '%s' is unknown" % [index, stype]
	if stype in STEP_TYPES_WAIT and stype != "wait_loot_item":
		var timeout = step.get("timeout_s", null)
		if timeout == null or float(timeout) <= 0.0:
			return "client_steps[%d] (%s) requires a positive timeout_s" % [index, stype]
	if stype == "wait_entity" or stype == "assert_entity_removed":
		if str(step.get("entity_type", "")) == "":
			return "client_steps[%d] (%s) requires entity_type" % [index, stype]
	if stype == "wait_event":
		if str(step.get("event_type", "")) == "":
			return "client_steps[%d] (%s) requires event_type" % [index, stype]
	if stype == "click_entity_until_event":
		if str(step.get("entity_type", "")) == "" or str(step.get("event_type", "")) == "":
			return "client_steps[%d] (%s) requires entity_type and event_type" % [index, stype]
	if stype == "click_loot_item":
		if str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires item_def_id or rolled" % [index, stype]
	if stype == "wait_player_near":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	if stype in ["drag_bag_to_weapon_slot", "drag_bag_to_equipment_slot", "drag_bag_to_outside", "assign_hotbar_slot", "double_click_bag_item"]:
		if str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires item_def_id" % [index, stype]
	if stype in ["drag_bag_to_equipment_slot", "drag_equipment_to_bag"]:
		if str(step.get("slot", "")) == "":
			return "client_steps[%d] (%s) requires slot" % [index, stype]
	if stype == "assign_hotbar_slot":
		if not step.has("slot_index"):
			return "client_steps[%d] (%s) requires slot_index" % [index, stype]
	if stype == "use_hotbar_slot":
		if not step.has("slot_index"):
			return "client_steps[%d] (%s) requires slot_index" % [index, stype]
	if stype in ["wait_inventory_count", "assert_inventory_count"]:
		if not step.has("equals"):
			return "client_steps[%d] (%s) requires equals" % [index, stype]
	if stype == "wait_loot_count":
		if not step.has("min_count"):
			return "client_steps[%d] (%s) requires min_count" % [index, stype]
	if stype in ["assert_hotbar_assigned", "wait_hotbar_assigned"]:
		if not step.has("slot_index") or str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires slot_index and item_def_id" % [index, stype]
	if stype in ["assert_hotbar_capacity", "wait_hotbar_capacity", "assert_hotbar_slot_disabled"]:
		if not step.has("equals") and not step.has("at_least") and stype in ["assert_hotbar_capacity", "wait_hotbar_capacity"]:
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
		if not step.has("slot_index") and stype == "assert_hotbar_slot_disabled":
			return "client_steps[%d] (%s) requires slot_index" % [index, stype]
	if stype == "assert_inventory_capacity":
		if not step.has("rows") and not step.has("capacity"):
			return "client_steps[%d] (%s) requires rows or capacity" % [index, stype]
	if stype == "assert_bag_grid":
		if not step.has("columns") and not step.has("available_slot_count"):
			return "client_steps[%d] (%s) requires columns or available_slot_count" % [index, stype]
	if stype == "assert_paper_doll_layout":
		if not step.has("slots"):
			return "client_steps[%d] (%s) requires slots" % [index, stype]
	if stype == "assert_player_hp":
		if not step.has("equals"):
			return "client_steps[%d] (%s) requires equals" % [index, stype]
	if stype == "press_key":
		if str(step.get("keycode", "")) == "":
			return "client_steps[%d] (%s) requires keycode" % [index, stype]
	if stype in ["click_stat_button", "assert_stat_button_enabled"]:
		if str(step.get("stat", "")) == "":
			return "client_steps[%d] (%s) requires stat" % [index, stype]
	if stype == "assert_stat_button_enabled":
		if not step.has("enabled"):
			return "client_steps[%d] (%s) requires enabled" % [index, stype]
	if stype == "assert_character_info":
		if not step.has("name") and not step.has("level") and not step.has("area"):
			return "client_steps[%d] (%s) requires name, level, or area" % [index, stype]
	if stype in ["wait_character_progression", "assert_character_progression"]:
		var has_any := false
		for key in ["level", "experience", "unspent_stat_points", "gold", "deepest_dungeon_depth", "str", "dex", "vit", "magic", "derived_stats", "player_max_hp", "stat_breakdowns"]:
			if step.has(key):
				has_any = true
		if not has_any:
			return "client_steps[%d] (%s) requires at least one progression expectation" % [index, stype]
	if stype in ["wait_skill_progression", "assert_skill_progression"]:
		var has_skill_expectation := false
		for key in ["unspent_skill_points", "rank", "max_rank", "can_spend"]:
			if step.has(key):
				has_skill_expectation = true
		if not has_skill_expectation:
			return "client_steps[%d] (%s) requires at least one skill progression expectation" % [index, stype]
	if stype == "assert_skill_button_enabled":
		if not step.has("enabled"):
			return "client_steps[%d] (%s) requires enabled" % [index, stype]
	if stype in ["wait_skill_bar", "assert_skill_bar"]:
		var has_bar_expectation := false
		for key in ["rank", "max_rank", "enabled", "disabled", "remaining_ticks", "remaining_ticks_min", "remaining_ticks_max", "total_ticks", "cooldown_fraction_min", "cooldown_fraction_max"]:
			if step.has(key):
				has_bar_expectation = true
		if not has_bar_expectation:
			return "client_steps[%d] (%s) requires at least one skill bar expectation" % [index, stype]
	if stype in ["set_floating_combat_text", "assert_floating_combat_text_enabled"]:
		if not step.has("enabled"):
			return "client_steps[%d] (%s) requires enabled" % [index, stype]
	if stype == "select_create_game_type" or stype == "assert_create_game_type":
		var session_type := str(step.get("session_type", ""))
		if session_type != "coop" and session_type != "solo":
			return "client_steps[%d] (%s) requires session_type coop or solo" % [index, stype]
	if stype == "assert_main_menu_actions":
		if not step.has("labels") and not step.has("actions"):
			return "client_steps[%d] (%s) requires labels or actions" % [index, stype]
	if stype == "assert_character_panel":
		var has_panel_expectation := false
		for key in ["visible", "mode", "title", "character_count", "min_character_count", "name_field_visible", "create_button_visible", "empty_visible"]:
			if step.has(key):
				has_panel_expectation = true
		if not has_panel_expectation:
			return "client_steps[%d] (%s) requires a panel expectation" % [index, stype]
	if stype == "assert_current_session":
		if not step.has("exists") and not step.has("mode") and not step.has("listed") and not step.has("session_id") and not step.has("session_id_env"):
			return "client_steps[%d] (%s) requires exists, mode, listed, session_id, or session_id_env" % [index, stype]
	if stype in ["wait_damage_number", "assert_damage_number"]:
		if not step.has("text") and not step.has("variant"):
			return "client_steps[%d] (%s) requires text or variant" % [index, stype]
	if stype in ["wait_wall_layout", "assert_wall_layout"]:
		if not step.has("equals") and not step.has("at_least") and not step.has("generated_at_least") and not step.has("non_perimeter_at_least") and not step.has("current_level"):
			return "client_steps[%d] (%s) requires a wall count or current_level expectation" % [index, stype]
	if stype in ["wait_shop_panel", "assert_shop_offer_count"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype in ["wait_stash_panel", "assert_stash_item_count", "assert_stash_gold"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype == "assert_shop_buy_button":
		if str(step.get("offer_id", "")) == "":
			return "client_steps[%d] (%s) requires offer_id" % [index, stype]
	if stype in ["assert_shop_sell_rows", "assert_shop_offer_details", "assert_shop_sell_details"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype == "click_shop_buy_offer":
		if str(step.get("offer_id", "")) == "" and str(step.get("offer_kind", "")) == "":
			return "client_steps[%d] (%s) requires offer_id or offer_kind" % [index, stype]
	if stype == "drag_bag_to_stash":
		if str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires item_def_id or rolled" % [index, stype]
	if stype == "drag_stash_to_bag":
		if str(step.get("stash_item_id", "")) == "" and str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires stash_item_id, item_def_id, or rolled" % [index, stype]
	if stype in ["click_stash_deposit_gold", "click_stash_withdraw_gold"]:
		if int(step.get("amount", 0)) <= 0:
			return "client_steps[%d] (%s) requires positive amount" % [index, stype]
	if stype == "assert_multiplayer_session_rows":
		if not step.has("equals") and not step.has("min_count") and not step.has("selected") and not step.has("listed") and not step.has("session_id") and not step.has("session_id_env") and not step.has("mode") and not step.has("member_count_min") and not step.has("connected_count_min"):
			return "client_steps[%d] (%s) requires a row expectation" % [index, stype]
	if stype in ["wait_remote_player_count", "assert_remote_player_count"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype == "click_menu_button":
		if str(step.get("button", "")) == "":
			return "client_steps[%d] (%s) requires button" % [index, stype]
	if stype == "enter_character_name":
		if str(step.get("name", "")) == "":
			return "client_steps[%d] (%s) requires name" % [index, stype]
	if stype == "select_window_size":
		if str(step.get("size", "")) == "":
			return "client_steps[%d] (%s) requires size" % [index, stype]
	if stype == "click_entity":
		if str(step.get("entity_type", "")) == "":
			return "client_steps[%d] (%s) requires entity_type" % [index, stype]
	if stype == "click_waypoint_level":
		if not step.has("target_level"):
			return "client_steps[%d] (%s) requires target_level" % [index, stype]
	if stype == "click_floor":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	return ""
