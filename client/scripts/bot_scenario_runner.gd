# BotScenarioRunner: frame-tick step executor for the Godot client bot.
# Runs one step at a time from a loaded client scenario, tracks elapsed time
# per step, and delegates to BotController for state access and action dispatch.
class_name BotScenarioRunner
extends RefCounted

const STEP_TYPES_WAIT := [
	"wait_ws_open", "wait_entity", "wait_event", "wait_inventory_item",
	"wait_loot_item", "wait_player_near", "assert_entity_removed",
]
const STEP_TYPES_ASSERT := [
	"assert_panel_visible", "assert_equipped",
	"assert_unequipped", "assert_inventory_missing",
]
const STEP_TYPES_ACTION := [
	"press_key", "click_entity", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_outside",
]
const ALL_STEP_TYPES: Array = [
	"wait_ws_open", "wait_entity", "wait_event", "assert_entity_removed",
	"assert_panel_visible", "wait_inventory_item", "assert_equipped",
	"assert_unequipped", "assert_inventory_missing", "wait_loot_item",
	"wait_player_near", "press_key", "click_entity", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_outside",
]

var scenario: Dictionary = {}
var step_delay_s: float = 0.0  # pause after each completed step (visual mode)
var _steps: Array = []
var _step_index: int = 0
var _step_elapsed: float = 0.0
var _post_step_wait: float = 0.0  # countdown after a step completes
var _done: bool = false
var _passed: bool = false
var _failure_msg: String = ""
var _controller = null  # BotController reference

# Filled by tick() on first action call; consumed by controller's _process.
var pending_action: Dictionary = {}


func load_scenario(data: Dictionary) -> bool:
	scenario = data
	_steps = data.get("client_steps", [])
	_step_index = 0
	_step_elapsed = 0.0
	_done = false
	_passed = false
	_failure_msg = ""
	pending_action = {}
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

	var timeout_s := float(step.get("timeout_s", 0.0))
	if timeout_s > 0.0 and _step_elapsed > timeout_s:
		_fail("timeout after %.1fs at step %d (%s) scenario=%s" % [
			timeout_s, _step_index, stype, str(scenario.get("id", "?"))
		])
		return true

	if stype in STEP_TYPES_WAIT:
		if _eval_wait(step, stype, state):
			_advance()
			_check_complete()
	elif stype in STEP_TYPES_ASSERT:
		if not _eval_assert(step, stype, state):
			return true  # _eval_assert already called _fail
		_advance()
		_check_complete()
	elif stype in STEP_TYPES_ACTION:
		_queue_action(step, stype)
		_advance()
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
		"wait_entity":
			var etype := str(step.get("entity_type", ""))
			var eids: Array = state.get("%s_ids" % etype, state.get("entities_by_type", {}).get(etype, []))
			return eids.size() > 0
		"wait_event":
			var evtype := str(step.get("event_type", ""))
			var pending: Array = state.get("pending_events", [])
			for ev in pending:
				if str(ev.get("event_type", "")) == evtype:
					return true
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
	return true


func _queue_action(step: Dictionary, stype: String) -> void:
	pending_action = step.duplicate()
	pending_action["_type"] = stype


func _advance() -> void:
	_step_index += 1
	_step_elapsed = 0.0
	if step_delay_s > 0.0:
		_post_step_wait = step_delay_s


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
	if stype == "wait_player_near":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	if stype in ["drag_bag_to_weapon_slot", "drag_bag_to_outside"]:
		if str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires item_def_id" % [index, stype]
	if stype == "press_key":
		if str(step.get("keycode", "")) == "":
			return "client_steps[%d] (%s) requires keycode" % [index, stype]
	if stype == "click_entity":
		if str(step.get("entity_type", "")) == "":
			return "client_steps[%d] (%s) requires entity_type" % [index, stype]
	if stype == "click_floor":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	return ""
