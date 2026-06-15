class_name BotWaitProgressLogger
extends RefCounted


static func log_wait_progress(
	step: Dictionary,
	stype: String,
	state: Dictionary,
	step_elapsed: float,
	last_retry_at: float,
	scenario_id: String,
	step_index: int
) -> void:
	var parts: PackedStringArray = PackedStringArray([
		"waiting",
		stype,
		"elapsed=%.1fs" % step_elapsed,
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
	if stype == "click_entity_until_event" and step_elapsed - last_retry_at < float(step.get("retry_s", 0.25)):
		parts.append("next_attack_soon=true")
	print("[bot-client] %s scenario=%s step=%d" % [" ".join(parts), scenario_id, step_index])


static func _inventory_count(state: Dictionary, item_def_id: String) -> int:
	var total := 0
	for item in state.get("inventory", []):
		if str((item as Dictionary).get("item_def_id", "")) == item_def_id:
			total += 1
	return total
