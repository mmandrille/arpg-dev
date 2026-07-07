class_name QuestEliteObjectiveState
extends RefCounted


static func quest_journal_objectives(entities: Dictionary, steward_hunt: Dictionary = {}) -> Array:
	var objectives: Array = quest_journal_objectives_from_entities(entities)
	var hunt_rows := steward_hunt_journal_objectives(steward_hunt)
	for row in hunt_rows:
		objectives.append(row)
	return objectives


static func quest_journal_objectives_from_entities(entities: Dictionary) -> Array:
	var reward_found := false
	var reward_complete := true
	for rec in entities.values():
		var row: Dictionary = rec
		if bool(row.get("quest_reward", false)):
			reward_found = true
			reward_complete = reward_complete and str(row.get("state", "")) == "open"
	if not reward_found:
		return []
	return [{
		"id": "reward_chest",
		"title": "Open the marked reward chest",
		"complete": reward_complete,
	}]


static func steward_hunt_journal_objectives(steward_hunt: Dictionary) -> Array:
	if steward_hunt.is_empty() or not bool(steward_hunt.get("active", false)):
		return []
	var banner_text := str(steward_hunt.get("banner_text", ""))
	if banner_text == "":
		return []
	return [{
		"id": "steward_hunt",
		"title": banner_text,
		"complete": bool(steward_hunt.get("complete", false)),
	}]


static func steward_hunt_banner_state(steward_hunt: Dictionary) -> Dictionary:
	if steward_hunt.is_empty() or not bool(steward_hunt.get("active", false)):
		return {"visible": false, "status": "hidden", "banner_text": ""}
	if bool(steward_hunt.get("complete", false)):
		return {
			"visible": true,
			"status": "complete",
			"banner_text": str(steward_hunt.get("banner_text", "")),
		}
	return {
		"visible": true,
		"status": "active",
		"banner_text": str(steward_hunt.get("banner_text", "")),
	}


static func elite_tracker_state(entities: Dictionary) -> Dictionary:
	var chest_found := false
	var chest_open := false
	var remaining := 0
	for rec in entities.values():
		var row: Dictionary = rec
		if bool(row.get("elite_objective", false)):
			chest_found = true
			chest_open = chest_open or str(row.get("state", "")) == "open"
		if bool(row.get("monster_pack_leader", false)) and int(row.get("hp", 1)) > 0:
			remaining += 1
	if not chest_found:
		return {"visible": false, "status": "hidden", "remaining_leaders": 0}
	if chest_open:
		return {"visible": true, "status": "complete", "remaining_leaders": 0}
	if remaining > 0:
		return {"visible": true, "status": "active", "remaining_leaders": remaining}
	return {"visible": true, "status": "claim", "remaining_leaders": 0}
