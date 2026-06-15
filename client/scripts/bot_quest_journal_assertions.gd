class_name BotQuestJournalAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("quest_journal_panel", {})
	if step.has("visible") and bool(panel.get("visible", false)) != bool(step.get("visible", false)):
		return false
	if step.has("count") and int(panel.get("count", -1)) != int(step.get("count", 0)):
		return false
	if step.has("objective_id") or step.has("complete"):
		for objective in panel.get("objectives", []):
			var row: Dictionary = objective
			if step.has("objective_id") and str(row.get("id", "")) != str(step.get("objective_id", "")):
				continue
			if step.has("complete") and bool(row.get("complete", false)) != bool(step.get("complete", false)):
				continue
			return true
		return false
	return true
