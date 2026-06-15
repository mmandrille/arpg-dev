# BotScenarioRunner: frame-tick step executor for the Godot client bot.
# Runs one step at a time from a loaded client scenario, tracks elapsed time
# per step, and delegates to BotController for state access and action dispatch.
class_name BotScenarioRunner
extends RefCounted

const BotStepCatalogScript := preload("res://scripts/bot_step_catalog.gd")
const BotWaitHandlersScript := preload("res://scripts/bot_wait_handlers.gd")
const BotAssertionHandlersScript := preload("res://scripts/bot_assertion_handlers.gd")
const BotActionHandlersScript := preload("res://scripts/bot_action_handlers.gd")

const STEP_TYPES_WAIT := BotStepCatalogScript.STEP_TYPES_WAIT
const STEP_TYPES_ASSERT := BotStepCatalogScript.STEP_TYPES_ASSERT
const STEP_TYPES_ACTION := BotStepCatalogScript.STEP_TYPES_ACTION
const WAIT_LOG_INTERVAL_S := BotStepCatalogScript.WAIT_LOG_INTERVAL_S
const ALL_STEP_TYPES := BotStepCatalogScript.ALL_STEP_TYPES

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
	return BotWaitHandlersScript.evaluate(self, step, stype, state)


func _wait_ticks(step: Dictionary, state: Dictionary) -> bool:
	var current_tick := int(state.get("last_tick", 0))
	if not _memory.has("wait_ticks_target"):
		_memory["wait_ticks_target"] = current_tick + int(step.get("ticks", 0))
		_memory["wait_ticks_last_pulse"] = -999.0
	var target_tick := int(_memory.get("wait_ticks_target", current_tick))
	if current_tick >= target_tick:
		_memory.erase("wait_ticks_target")
		_memory.erase("wait_ticks_last_pulse")
		return true
	var pulse_s := float(step.get("pulse_s", 0.05))
	if _step_elapsed - float(_memory.get("wait_ticks_last_pulse", -999.0)) >= pulse_s:
		_memory["wait_ticks_last_pulse"] = _step_elapsed
		pending_action = {
			"type": "dispatch_intent",
			"_type": "dispatch_intent",
			"intent_type": "move_intent",
			"payload": {"direction": {"x": 0, "y": 0}, "duration_ticks": 1},
		}
	return false


func _eval_assert(step: Dictionary, stype: String, state: Dictionary) -> bool:
	return BotAssertionHandlersScript.evaluate(self, step, stype, state)


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
	var skill_id := str(step.get("skill_id", "magic_bolt"))
	var skill_panel := _panel_skill_state(panel, skill_id)
	var got := bool(skill_panel.get("spend_button_enabled", false))
	if want != got:
		_fail("assert_skill_button_enabled failed: skill=%s want=%s got=%s panel=%s step=%d scenario=%s" % [
			skill_id, want, got, str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("requirements_met") and bool(skill_panel.get("requirements_met", false)) != bool(step.get("requirements_met", false)):
		_fail("assert_skill_button_enabled requirements failed: want=%s got=%s panel=%s step=%d scenario=%s" % [
			bool(step.get("requirements_met", false)), bool(skill_panel.get("requirements_met", false)), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("skill_name") and str(skill_panel.get("skill_name", "")) != str(step.get("skill_name", "")):
		_fail("assert_skill_button_enabled skill_name failed: want=%s got=%s step=%d scenario=%s" % [
			str(step.get("skill_name", "")), str(skill_panel.get("skill_name", "")), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("tooltip_contains") and not str(skill_panel.get("tooltip_body", "")).contains(str(step.get("tooltip_contains", ""))):
		_fail("assert_skill_button_enabled tooltip failed: want contains=%s got=%s step=%d scenario=%s" % [
			str(step.get("tooltip_contains", "")), str(skill_panel.get("tooltip_body", "")), _step_index, str(scenario.get("id", "?"))
		])
		return false
	return true


func _panel_skill_state(panel: Dictionary, skill_id: String) -> Dictionary:
	var skill_rows = panel.get("skills", {})
	if typeof(skill_rows) == TYPE_DICTIONARY:
		var rows: Dictionary = skill_rows
		if rows.has(skill_id) and typeof(rows[skill_id]) == TYPE_DICTIONARY:
			return rows[skill_id] as Dictionary
	return panel


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
	if step.has("slot_text") and str(bar.get("slot_text", "")) != str(step.get("slot_text", "")):
		return false
	if step.has("tooltip_contains") and not str(bar.get("tooltip_text", "")).contains(str(step.get("tooltip_contains", ""))):
		return false
	return true


func _skill_bar_expectation(step: Dictionary) -> Dictionary:
	var out := {}
	for key in ["skill_id", "rank", "max_rank", "enabled", "disabled", "remaining_ticks", "remaining_ticks_min", "remaining_ticks_max", "total_ticks", "cooldown_fraction_min", "cooldown_fraction_max", "slot_text", "tooltip_contains"]:
		if step.has(key):
			out[key] = step[key]
	return out


func _assert_boss_health_bar(step: Dictionary, state: Dictionary) -> bool:
	if _boss_health_bar_matches(step, state):
		return true
	_fail("assert_boss_health_bar failed: want=%s got=%s step=%d scenario=%s" % [
		str(_boss_health_bar_expectation(step)),
		str(state.get("boss_health_bar", {})),
		_step_index,
		str(scenario.get("id", "?"))
	])
	return false


func _boss_health_bar_matches(step: Dictionary, state: Dictionary) -> bool:
	var bar: Dictionary = state.get("boss_health_bar", {})
	if step.has("visible") and bool(bar.get("visible", false)) != bool(step.get("visible", true)):
		return false
	for key in ["boss_id", "boss_template_id", "title"]:
		if step.has(key) and str(bar.get(key, "")) != str(step.get(key, "")):
			return false
	for key in ["hp", "max_hp", "phase_index", "duration_ticks"]:
		if step.has(key) and int(bar.get(key, -999999)) != int(step.get(key, 0)):
			return false
	for key in ["phase_kind", "pattern_id"]:
		if step.has(key) and str(bar.get(key, "")) != str(step.get(key, "")):
			return false
	if step.has("hp_min") and int(bar.get("hp", -999999)) < int(step.get("hp_min", 0)):
		return false
	if step.has("hp_max") and int(bar.get("hp", 999999)) > int(step.get("hp_max", 0)):
		return false
	if step.has("remaining_ticks_min") and int(bar.get("remaining_ticks", -999999)) < int(step.get("remaining_ticks_min", 0)):
		return false
	if step.has("remaining_ticks_max") and int(bar.get("remaining_ticks", 999999)) > int(step.get("remaining_ticks_max", 0)):
		return false
	if step.has("ratio_min") and float(bar.get("ratio", -1.0)) < float(step.get("ratio_min", 0.0)):
		return false
	if step.has("ratio_max") and float(bar.get("ratio", 2.0)) > float(step.get("ratio_max", 1.0)):
		return false
	if step.has("phase_ratio_min") and float(bar.get("phase_ratio", -1.0)) < float(step.get("phase_ratio_min", 0.0)):
		return false
	if step.has("phase_ratio_max") and float(bar.get("phase_ratio", 2.0)) > float(step.get("phase_ratio_max", 1.0)):
		return false
	return true


func _boss_health_bar_expectation(step: Dictionary) -> Dictionary:
	var out := {}
	for key in ["visible", "boss_id", "boss_template_id", "title", "hp", "max_hp", "hp_min", "hp_max", "ratio_min", "ratio_max", "phase_kind", "pattern_id", "phase_index", "duration_ticks", "remaining_ticks_min", "remaining_ticks_max", "phase_ratio_min", "phase_ratio_max"]:
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
	for key in ["outcome", "source_entity_id", "target_entity_id", "shop_id", "offer_id", "item_instance_id", "skill_id", "damage_type"]:
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
	if step.has("has_burning_effect") and bool(rec.get("has_burning_effect", false)) != bool(step.get("has_burning_effect", false)):
		return false
	if step.has("has_elite_command_effect") and bool(rec.get("has_elite_command_effect", false)) != bool(step.get("has_elite_command_effect", false)):
		return false
	if step.has("has_elite_command_radius_preview") and bool(rec.get("has_elite_command_radius_preview", false)) != bool(step.get("has_elite_command_radius_preview", false)):
		return false
	if step.has("elite_command_radius_min") and float(rec.get("elite_command_radius_preview", -1.0)) < float(step.get("elite_command_radius_min", 0.0)):
		return false
	if step.has("elite_command_radius_max") and float(rec.get("elite_command_radius_preview", 999999.0)) > float(step.get("elite_command_radius_max", 0.0)):
		return false
	if step.has("monster_pack_leader") and bool(rec.get("monster_pack_leader", false)) != bool(step.get("monster_pack_leader", false)):
		return false
	if step.has("elite_objective") and bool(rec.get("elite_objective", false)) != bool(step.get("elite_objective", false)):
		return false
	if step.has("has_objective_marker") and bool(rec.get("has_objective_marker", false)) != bool(step.get("has_objective_marker", false)):
		return false
	if step.has("quest_reward") and bool(rec.get("quest_reward", false)) != bool(step.get("quest_reward", false)) or step.has("has_quest_marker") and bool(rec.get("has_quest_marker", false)) != bool(step.get("has_quest_marker", false)):
		return false
	if step.has("is_boss") and bool(rec.get("is_boss", false)) != bool(step.get("is_boss", false)):
		return false
	if step.has("boss_template_id") and str(rec.get("boss_template_id", "")) != str(step.get("boss_template_id", "")):
		return false
	if step.has("boss_telegraph_active") and bool(rec.get("boss_telegraph_active", false)) != bool(step.get("boss_telegraph_active", false)):
		return false
	if step.has("has_boss_telegraph_marker") and bool(rec.get("has_boss_telegraph_marker", false)) != bool(step.get("has_boss_telegraph_marker", false)):
		return false
	if step.has("telegraph_tint") and str(rec.get("telegraph_tint", "")) != str(step.get("telegraph_tint", "")):
		return false
	if step.has("telegraph_radius_min") and float(rec.get("telegraph_radius", -1.0)) < float(step.get("telegraph_radius_min", 0.0)):
		return false
	if step.has("telegraph_radius_max") and float(rec.get("telegraph_radius", 999999.0)) > float(step.get("telegraph_radius_max", 0.0)):
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


func _assert_shop_reroll_button(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("shop_panel", {})
	if step.has("visible") and bool(panel.get("reroll_visible", false)) != bool(step.get("visible", true)):
		_fail("assert_shop_reroll_button failed: visible want=%s panel=%s step=%d scenario=%s" % [
			str(step.get("visible", true)), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("enabled") and bool(panel.get("reroll_enabled", false)) != bool(step.get("enabled", true)):
		_fail("assert_shop_reroll_button failed: enabled want=%s panel=%s step=%d scenario=%s" % [
			str(step.get("enabled", true)), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("cost") and int(panel.get("reroll_cost", 0)) != int(step.get("cost", 0)):
		_fail("assert_shop_reroll_button failed: cost want=%d panel=%s step=%d scenario=%s" % [
			int(step.get("cost", 0)), str(panel), _step_index, str(scenario.get("id", "?"))
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


func _assert_stash_filter(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("stash_panel", {})
	if step.has("search_text") and str(panel.get("stash_search_text", "")) != str(step.get("search_text", "")):
		_fail("assert_stash_filter failed: search_text want=%s panel=%s step=%d scenario=%s" % [
			str(step.get("search_text", "")), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("sort_mode") and str(panel.get("stash_sort_mode", "")) != str(step.get("sort_mode", "")):
		_fail("assert_stash_filter failed: sort_mode want=%s panel=%s step=%d scenario=%s" % [
			str(step.get("sort_mode", "")), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("filtered_equals") and int(panel.get("filtered_stash_item_count", 0)) != int(step.get("filtered_equals", 0)):
		_fail("assert_stash_filter failed: filtered_equals want=%d panel=%s step=%d scenario=%s" % [
			int(step.get("filtered_equals", 0)), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("first_item_def_id"):
		var rows: Array = panel.get("stash_rows", [])
		if rows.is_empty() or str((rows[0] as Dictionary).get("item_def_id", "")) != str(step.get("first_item_def_id", "")):
			_fail("assert_stash_filter failed: first_item_def_id want=%s rows=%s step=%d scenario=%s" % [
				str(step.get("first_item_def_id", "")), str(rows), _step_index, str(scenario.get("id", "?"))
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
	if step.has("container_mode") and str((state.get("stash_panel", {}) as Dictionary).get("container_mode", "")) != str(step.get("container_mode", "")):
		return false
	var rows := _matching_stash_rows(step, state)
	if step.has("equals") and rows.size() != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and rows.size() < int(step.get("at_least", 0)):
		return false
	return true


func _market_listing_rows_match(step: Dictionary, state: Dictionary) -> bool:
	if not step.has("equals") and not step.has("at_least") and not step.has("price_gold") and not step.has("item_def_id") and not step.has("rolled"):
		return true
	var rows := _matching_market_listing_rows(step, state)
	if step.has("equals") and rows.size() != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and rows.size() < int(step.get("at_least", 0)):
		return false
	return true


func _assert_market_listing_rows(step: Dictionary, state: Dictionary) -> bool:
	if _market_listing_rows_match(step, state):
		return true
	_fail("assert_market_listing_rows failed: want=%s panel=%s step=%d scenario=%s" % [
		str(step), str(state.get("market_panel", {})), _step_index, str(scenario.get("id", "?"))
	])
	return false


func _market_offer_rows_match(step: Dictionary, state: Dictionary) -> bool:
	if not (step.has("offer_equals") or step.has("offer_at_least") or step.has("offer_item_def_id") or step.has("offer_status")):
		return true
	var rows := _matching_market_offer_rows(step, state)
	if step.has("offer_equals") and rows.size() != int(step.get("offer_equals", 0)):
		return false
	if step.has("offer_at_least") and rows.size() < int(step.get("offer_at_least", 0)):
		return false
	return true


func _assert_market_offer_rows(step: Dictionary, state: Dictionary) -> bool:
	if _market_offer_rows_match(step, state):
		return true
	_fail("assert_market_offer_rows failed: want=%s panel=%s step=%d scenario=%s" % [
		str(step), str(state.get("market_panel", {})), _step_index, str(scenario.get("id", "?"))
	])
	return false


func _bishop_panel_matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("bishop_panel", {})
	for key in ["price", "gold"]:
		if step.has(key) and int(panel.get(key, -1)) != int(step.get(key, 0)):
			return false
	for key in ["affordable", "respec_enabled", "visible", "debug_enabled"]:
		if step.has(key) and bool(panel.get(key, false)) != bool(step.get(key, false)):
			return false
	if step.has("service_id") and str(panel.get("service_id", "")) != str(step.get("service_id", "")):
		return false
	if step.has("status_contains") and not str(panel.get("status", "")).contains(str(step.get("status_contains", ""))):
		return false
	return true


func _blacksmith_panel_matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("blacksmith_panel", {})
	if step.has("visible") and bool(panel.get("visible", false)) != bool(step.get("visible", false)):
		return false
	if step.has("stash_gold_equals") and int(panel.get("stash_gold", 0)) != int(step.get("stash_gold_equals", 0)):
		return false
	if step.has("stash_gold_at_least") and int(panel.get("stash_gold", 0)) < int(step.get("stash_gold_at_least", 0)):
		return false
	if step.has("item_count") and int(panel.get("item_count", 0)) != int(step.get("item_count", 0)):
		return false
	if step.has("status_contains") and not str(panel.get("status", "")).contains(str(step.get("status_contains", ""))):
		return false
	var rows := _matching_blacksmith_rows(step, state)
	if step.has("item_def_id") or step.has("stash_item_id") or step.has("item_level") or step.has("upgrade_enabled"):
		if rows.is_empty():
			return false
	if step.has("item_level"):
		var want_level := int(step.get("item_level", 0))
		var found_level := false
		for row in rows:
			if _blacksmith_row_item_level(row as Dictionary) == want_level:
				found_level = true
				break
		if not found_level:
			return false
	if step.has("upgrade_enabled"):
		var want_enabled := bool(step.get("upgrade_enabled", false))
		var found_enabled := false
		for row in rows:
			if bool((row as Dictionary).get("upgrade_enabled", false)) == want_enabled:
				found_enabled = true
				break
		if not found_enabled:
			return false
	return true


func _blacksmith_row_item_level(row: Dictionary) -> int:
	if row.has("item_level"):
		return int(row.get("item_level", -1))
	var rolled: Variant = row.get("rolled_stats", {})
	if typeof(rolled) == TYPE_DICTIONARY:
		var payload := rolled as Dictionary
		if typeof(payload.get("stats", {})) == TYPE_DICTIONARY:
			return int((payload.get("stats", {}) as Dictionary).get("item_level", 0))
		return int(payload.get("item_level", 0))
	return -1


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
		if step.has("display_name") and str(rec.get("display_name", "")) != str(step.get("display_name", "")): continue
		if step.has("rolled") and (str(rec.get("item_template_id", "")) != "") != bool(step.get("rolled", false)):
			continue
		if step.has("summary_contains") and not _row_summary_contains(rec, step.get("summary_contains", "")): continue
		out.append(rec)
	return out


func _matching_blacksmith_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("blacksmith_panel", {})
	var rows: Array = []
	var staged: Variant = panel.get("staged_item", {})
	if typeof(staged) == TYPE_DICTIONARY and not (staged as Dictionary).is_empty():
		rows.append(staged)
	rows.append_array(panel.get("rows", []))
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("stash_item_id") and str(rec.get("stash_item_id", "")) != str(step.get("stash_item_id", "")):
			continue
		if step.has("item_def_id") and str(rec.get("item_def_id", "")) != str(step.get("item_def_id", "")):
			continue
		out.append(rec)
	return out


func _matching_market_listing_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("market_panel", {})
	var rows: Array = panel.get("listing_rows", [])
	if bool(step.get("seller_owned", false)):
		rows = panel.get("owned_listing_rows", [])
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("listing_id") and str(rec.get("listing_id", "")) != str(step.get("listing_id", "")):
			continue
		if step.has("item_def_id") and str(rec.get("item_def_id", "")) != str(step.get("item_def_id", "")):
			continue
		if step.has("rolled") and (str(rec.get("item_template_id", "")) != "") != bool(step.get("rolled", false)):
			continue
		if step.has("price_gold") and int(rec.get("price_gold", 0)) != int(step.get("price_gold", 0)):
			continue
		if step.has("seller_owned") and bool(step.get("seller_owned", false)) != (str(rec.get("seller_account_id", "")) == str(panel.get("account_id", ""))):
			continue
		out.append(rec)
	return out


func _matching_market_offer_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("market_panel", {})
	var rows: Array = panel.get("offer_rows", [])
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("offer_id") and str(rec.get("offer_id", "")) != str(step.get("offer_id", "")):
			continue
		if step.has("offer_status") and str(rec.get("status", "")) != str(step.get("offer_status", "")):
			continue
		if step.has("offer_item_def_id") and not (rec.get("item_def_ids", []) as Array).has(str(step.get("offer_item_def_id", ""))):
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
	BotActionHandlersScript.queue(self, step, stype, state)


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
	if _has_character_summary_expectation(step):
		var rows: Array = panel.get("character_rows", characters)
		var matched := false
		for row in rows:
			if typeof(row) != TYPE_DICTIONARY:
				continue
			if _character_row_matches_summary(row as Dictionary, step):
				matched = true
				break
		if not matched:
			_fail("assert_character_panel summary failed: want=%s rows=%s step=%d scenario=%s" % [
				str(_character_summary_expectation(step)), str(rows), _step_index, str(scenario.get("id", "?"))
			])
			return false
	if step.has("selected_class") and str(panel.get("selected_class", "")) != str(step.get("selected_class", "")):
		_fail("assert_character_panel selected_class failed: want=%s got=%s panel=%s step=%d scenario=%s" % [
			str(step.get("selected_class", "")), str(panel.get("selected_class", "")), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("class_picker_visible") and bool(panel.get("class_picker_visible", false)) != bool(step.get("class_picker_visible", false)):
		_fail("assert_character_panel class_picker_visible failed: want=%s got=%s panel=%s step=%d scenario=%s" % [
			str(step.get("class_picker_visible", false)), str(panel.get("class_picker_visible", false)), str(panel), _step_index, str(scenario.get("id", "?"))
		])
		return false
	if step.has("row_character_class"):
		var rows: Array = panel.get("character_rows", characters)
		var found_class := false
		for row in rows:
			if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("character_class", "")) == str(step.get("row_character_class", "")):
				found_class = true
				break
		if not found_class:
			_fail("assert_character_panel row_character_class failed: want=%s rows=%s step=%d scenario=%s" % [
				str(step.get("row_character_class", "")), str(rows), _step_index, str(scenario.get("id", "?"))
			])
			return false
	return true


func _has_character_summary_expectation(step: Dictionary) -> bool:
	for key in ["character_id", "status", "level", "min_level", "gold", "min_gold", "deepest_dungeon_depth", "min_deepest_dungeon_depth", "label_contains"]:
		if step.has(key):
			return true
	return false


func _character_summary_expectation(step: Dictionary) -> Dictionary:
	var out := {}
	for key in ["character_id", "status", "level", "min_level", "gold", "min_gold", "deepest_dungeon_depth", "min_deepest_dungeon_depth", "label_contains"]:
		if step.has(key):
			out[key] = step[key]
	return out


func _character_row_matches_summary(row: Dictionary, step: Dictionary) -> bool:
	if step.has("character_id") and str(row.get("character_id", "")) != str(step.get("character_id", "")):
		return false
	if step.has("status") and str(row.get("status", "")) != str(step.get("status", "")):
		return false
	if step.has("level") and int(row.get("level", 0)) != int(step.get("level", 0)):
		return false
	if step.has("min_level") and int(row.get("level", 0)) < int(step.get("min_level", 0)):
		return false
	if step.has("gold") and int(row.get("gold", 0)) != int(step.get("gold", 0)):
		return false
	if step.has("min_gold") and int(row.get("gold", 0)) < int(step.get("min_gold", 0)):
		return false
	if step.has("deepest_dungeon_depth") and int(row.get("deepest_dungeon_depth", 0)) != int(step.get("deepest_dungeon_depth", 0)):
		return false
	if step.has("min_deepest_dungeon_depth") and int(row.get("deepest_dungeon_depth", 0)) < int(step.get("min_deepest_dungeon_depth", 0)):
		return false
	if step.has("label_contains") and str(row.get("label", "")).find(str(step.get("label_contains", ""))) < 0:
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


func _assert_multiplayer_filter(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("multiplayer_panel", {})
	var expected_search := OS.get_environment(str(step.get("search_text_env", ""))) if step.has("search_text_env") else str(step.get("search_text", ""))
	if (step.has("search_text") or step.has("search_text_env")) and str(panel.get("search_text", "")) != expected_search:
		_fail("assert_multiplayer_filter search_text failed: want=%s panel=%s step=%d scenario=%s" % [expected_search, str(panel), _step_index, str(scenario.get("id", "?"))])
		return false
	var checks := [
		["sort_mode", step.has("sort_mode"), str(panel.get("sort_mode", "")) == str(step.get("sort_mode", "")), str(step.get("sort_mode", ""))],
		["filtered_equals", step.has("filtered_equals"), int(panel.get("filtered_session_count", 0)) == int(step.get("filtered_equals", 0)), str(step.get("filtered_equals", 0))],
		["total_at_least", step.has("total_at_least"), int(panel.get("total_session_count", 0)) >= int(step.get("total_at_least", 0)), str(step.get("total_at_least", 0))],
	]
	for check in checks:
		if bool(check[1]) and not bool(check[2]):
			_fail("assert_multiplayer_filter %s failed: want=%s panel=%s step=%d scenario=%s" % [str(check[0]), str(check[3]), str(panel), _step_index, str(scenario.get("id", "?"))])
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
		"wait_ticks":
			return "ticks=%d" % int(step.get("ticks", 0))
		"wait_boss_health_bar", "assert_boss_health_bar":
			return "boss_health_bar=%s" % str(_boss_health_bar_expectation(step))
		"wait_quest_journal", "assert_quest_journal":
			return "quest_journal=%s" % str(step)
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
		"select_character_class":
			return "class_id=%s" % str(step.get("class_id", ""))
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
		"wait_shop_panel", "assert_shop_offer_count", "assert_shop_buy_button", "assert_shop_reroll_button", "assert_shop_sell_rows", \
		"assert_shop_offer_details", "assert_shop_sell_details", \
		"click_shop_buy_offer", "click_shop_reroll", "click_shop_sell_item":
			return "shop=%s" % str(step)
		"click_waypoint_level":
			return "target_level=%s" % str(step.get("target_level", ""))
		"wait_stash_panel", "assert_stash_item_count", "assert_stash_gold", "assert_stash_filter", "assert_stash_panel_visible", \
		"drag_bag_to_stash", "drag_stash_to_bag", \
		"click_stash_deposit_gold", "click_stash_withdraw_gold", "set_stash_search", "select_stash_sort":
			return "stash=%s" % str(step)
		"wait_market_panel", "assert_market_panel_visible", "assert_market_listing_rows", "assert_market_offer_rows", \
		"set_market_publish_price", "click_market_publish_item", "click_market_purchase_listing", "click_market_view_offers", "click_market_accept_offer":
			return "market=%s" % str(step)
		"wait_bishop_panel", "assert_bishop_panel", "assert_bishop_panel_visible", "click_bishop_respec":
			return "bishop=%s" % str(step)
		"wait_blacksmith_panel", "assert_blacksmith_panel", "assert_blacksmith_panel_visible", "click_blacksmith_upgrade":
			return "blacksmith=%s" % str(step)
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
	if stype == "wait_boss_health_bar":
		parts.append("boss_health_bar=%s" % str(state.get("boss_health_bar", {})))
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
	if stype == "wait_market_panel":
		parts.append("market_panel=%s" % str(state.get("market_panel", {})))
	if stype == "wait_blacksmith_panel":
		parts.append("blacksmith_panel=%s" % str(state.get("blacksmith_panel", {})))
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
	return BotStepCatalogScript.validate_scenario(data)


static func validate_step(step: Dictionary, index: int) -> String:
	return BotStepCatalogScript.validate_step(step, index)
