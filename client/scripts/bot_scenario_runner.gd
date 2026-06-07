# BotScenarioRunner: frame-tick step executor for the Godot client bot.
# Runs one step at a time from a loaded client scenario, tracks elapsed time
# per step, and delegates to BotController for state access and action dispatch.
class_name BotScenarioRunner
extends RefCounted

const STEP_TYPES_WAIT := [
	"wait_ws_open", "wait_entity", "wait_event", "wait_inventory_item",
	"wait_inventory_count", "wait_loot_item", "wait_loot_count",
	"wait_player_near", "assert_entity_removed",
	"click_entity_until_event", "wait_main_menu", "wait_character_panel",
	"wait_settings_panel", "wait_pause_menu",
]
const STEP_TYPES_ASSERT := [
	"assert_panel_visible", "assert_waypoint_panel_visible", "assert_equipped",
	"assert_unequipped", "assert_inventory_missing", "assert_inventory_count",
	"assert_loot_presentation", "assert_inventory_presentation",
	"assert_hotbar_assigned", "assert_player_hp", "assert_main_menu_visible",
	"assert_character_panel_visible", "assert_settings_panel_visible",
	"assert_pause_menu_visible", "assert_session_changed",
	"assert_player_position_unchanged",
]
const STEP_TYPES_ACTION := [
	"press_key", "click_entity", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_outside",
	"assign_hotbar_slot", "double_click_bag_item", "click_menu_button",
	"enter_character_name", "select_character", "select_window_size",
	"remember_session", "remember_player_position",
]
const WAIT_LOG_INTERVAL_S := 2.0

const ALL_STEP_TYPES: Array = [
	"wait_ws_open", "wait_entity", "wait_event", "assert_entity_removed",
	"assert_panel_visible", "assert_waypoint_panel_visible", "wait_inventory_item", "wait_inventory_count",
	"assert_equipped", "assert_unequipped", "assert_inventory_missing",
	"assert_inventory_count", "wait_loot_item", "wait_loot_count",
	"wait_player_near", "press_key", "click_entity", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_outside",
	"assert_loot_presentation", "assert_inventory_presentation",
	"click_entity_until_event", "assign_hotbar_slot", "assert_hotbar_assigned",
	"assert_player_hp", "double_click_bag_item", "wait_main_menu",
	"wait_character_panel", "wait_settings_panel", "wait_pause_menu",
	"assert_main_menu_visible", "assert_character_panel_visible",
	"assert_settings_panel_visible", "assert_pause_menu_visible",
	"click_menu_button", "enter_character_name", "select_character",
	"select_window_size", "remember_session", "assert_session_changed",
	"remember_player_position", "assert_player_position_unchanged",
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
		"wait_settings_panel":
			return bool(state.get("settings_panel_visible", false))
		"wait_pause_menu":
			return bool(state.get("pause_menu_visible", false))
		"wait_entity":
			var etype := str(step.get("entity_type", ""))
			var eids: Array = state.get("%s_ids" % etype, state.get("entities_by_type", {}).get(etype, []))
			return eids.size() > 0
		"wait_event":
			var evtype := str(step.get("event_type", ""))
			var pending: Array = state.get("pending_events", [])
			for i in range(pending.size()):
				if str(pending[i].get("event_type", "")) == evtype:
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
			for ev in pending:
				if str(ev.get("event_type", "")) == evtype:
					return true
			var retry_s := float(step.get("retry_s", 0.25))
			if _step_elapsed - _last_retry_at >= retry_s:
				_last_retry_at = _step_elapsed
				pending_action = {
					"type": "click_entity",
					"_type": "click_entity",
					"entity_type": str(step.get("entity_type", "")),
				}
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
		"assert_character_panel_visible":
			return _assert_bool_state("assert_character_panel_visible", "character_panel_visible", step, state)
		"assert_settings_panel_visible":
			return _assert_bool_state("assert_settings_panel_visible", "settings_panel_visible", step, state)
		"assert_pause_menu_visible":
			return _assert_bool_state("assert_pause_menu_visible", "pause_menu_visible", step, state)
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
			var slot := str(step.get("slot", "weapon"))
			var eq: Dictionary = state.get("equipped", {})
			var val = eq.get(slot, null)
			if val == null or str(val) == "":
				_fail("assert_equipped failed: slot=%s equipped=%s step=%d scenario=%s" % [
					slot, str(eq), _step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
		"assert_unequipped":
			var slot := str(step.get("slot", "weapon"))
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
			var slot_index := int(step.get("slot_index", -1))
			var want_def := str(step.get("item_def_id", ""))
			var bar: Dictionary = state.get("consumable_bar", {})
			var assigned: Array = bar.get("assigned_slots", [])
			if slot_index < 0 or slot_index >= assigned.size():
				_fail("assert_hotbar_assigned failed: invalid slot_index=%d step=%d scenario=%s" % [
					slot_index, _step_index, str(scenario.get("id", "?"))
				])
				return false
			var slot_val = assigned[slot_index]
			if slot_val == null or typeof(slot_val) != TYPE_DICTIONARY:
				_fail("assert_hotbar_assigned failed: slot %d empty step=%d scenario=%s" % [
					slot_index, _step_index, str(scenario.get("id", "?"))
				])
				return false
			if str((slot_val as Dictionary).get("item_def_id", "")) != want_def:
				_fail("assert_hotbar_assigned failed: slot %d has %s want %s step=%d scenario=%s" % [
					slot_index, str((slot_val as Dictionary).get("item_def_id", "")), want_def,
					_step_index, str(scenario.get("id", "?"))
				])
				return false
			return true
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


func _inventory_count(state: Dictionary, item_def_id: String) -> int:
	var inv: Array = state.get("inventory", [])
	if item_def_id == "":
		return inv.size()
	var count := 0
	for item in inv:
		if str(item.get("item_def_id", "")) == item_def_id:
			count += 1
	return count


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
	if want != got:
		_fail("%s failed: want=%s got=%s step=%d scenario=%s" % [
			label, want, got, _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


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
		"assert_inventory_missing", "double_click_bag_item", "assign_hotbar_slot":
			return "item_def_id=%s" % str(step.get("item_def_id", ""))
		"wait_loot_count":
			return "min_count=%s" % str(step.get("min_count", ""))
		"assert_player_hp":
			return "hp=%s" % str(step.get("equals", ""))
		"click_menu_button":
			return "button=%s" % str(step.get("button", ""))
		"enter_character_name":
			return "name=%s" % str(step.get("name", ""))
		"select_character":
			return "index=%s" % str(step.get("index", 0))
		"select_window_size":
			return "size=%s" % str(step.get("size", ""))
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
	if stype == "wait_loot_count":
		parts.append("loot_count=%d" % (state.get("loot_ids", []) as Array).size())
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
	if stype == "wait_player_near":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	if stype in ["drag_bag_to_weapon_slot", "drag_bag_to_outside", "assign_hotbar_slot", "double_click_bag_item"]:
		if str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires item_def_id" % [index, stype]
	if stype == "assign_hotbar_slot":
		if not step.has("slot_index"):
			return "client_steps[%d] (%s) requires slot_index" % [index, stype]
	if stype in ["wait_inventory_count", "assert_inventory_count"]:
		if not step.has("equals"):
			return "client_steps[%d] (%s) requires equals" % [index, stype]
	if stype == "wait_loot_count":
		if not step.has("min_count"):
			return "client_steps[%d] (%s) requires min_count" % [index, stype]
	if stype == "assert_hotbar_assigned":
		if not step.has("slot_index") or str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires slot_index and item_def_id" % [index, stype]
	if stype == "assert_player_hp":
		if not step.has("equals"):
			return "client_steps[%d] (%s) requires equals" % [index, stype]
	if stype == "press_key":
		if str(step.get("keycode", "")) == "":
			return "client_steps[%d] (%s) requires keycode" % [index, stype]
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
	if stype == "click_floor":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	return ""
