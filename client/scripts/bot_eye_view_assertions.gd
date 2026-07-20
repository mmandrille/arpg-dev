class_name BotEyeViewAssertions
extends RefCounted


static func assert_weapon(runner, step: Dictionary, state: Dictionary) -> bool:
	var slot := str(step.get("slot", "main_hand"))
	var equipment: Dictionary = state.get("equipment_visuals", {})
	var eye_view: Dictionary = equipment.get("eye_view", {})
	var slot_state: Dictionary = eye_view.get(slot, {})
	for key in ["active", "visible"]:
		if step.has(key) and bool(slot_state.get(key, false)) != bool(step.get(key, true)):
			runner._fail("assert_eye_view_weapon failed: %s want=%s got=%s state=%s step=%d scenario=%s" % [
				key, str(step.get(key, true)), str(slot_state.get(key, false)), str(eye_view), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	if step.has("item_def_id") and str(slot_state.get("item_def_id", "")) != str(step.get("item_def_id", "")):
		runner._fail("assert_eye_view_weapon failed: item_def_id want=%s got=%s state=%s step=%d scenario=%s" % [
			str(step.get("item_def_id", "")), str(slot_state.get("item_def_id", "")), str(eye_view), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("attack_count_min") and int(slot_state.get("attack_count", 0)) < int(step.get("attack_count_min", 0)):
		runner._fail("assert_eye_view_weapon failed: attack_count_min want=%s got=%s state=%s step=%d scenario=%s" % [
			str(step.get("attack_count_min", 0)), str(slot_state.get("attack_count", 0)), str(eye_view), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("rig_active") and bool(eye_view.get("rig_active", false)) != bool(step.get("rig_active", true)):
		runner._fail("assert_eye_view_weapon failed: rig_active want=%s got=%s state=%s step=%d scenario=%s" % [
			str(step.get("rig_active", true)), str(eye_view.get("rig_active", false)), str(eye_view), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	for key in ["socket_parent", "node_parent", "last_attack_clip"]:
		if step.has(key) and str(slot_state.get(key, "")) != str(step.get(key, "")):
			runner._fail("assert_eye_view_weapon failed: %s want=%s got=%s state=%s step=%d scenario=%s" % [
				key, str(step.get(key, "")), str(slot_state.get(key, "")), str(eye_view), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	return true
