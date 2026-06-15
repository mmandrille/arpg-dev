class_name QuestEliteObjectiveState
extends RefCounted


static func quest_journal_objectives(entities: Dictionary) -> Array:
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
