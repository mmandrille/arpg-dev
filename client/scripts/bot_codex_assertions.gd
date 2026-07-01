class_name BotCodexAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("codex_panel", {})
	if step.has("visible"):
		if bool(panel.get("visible", false)) != bool(step.get("visible", true)):
			return false
	if step.has("chapter_id"):
		if str(panel.get("chapter_id", "")) != str(step.get("chapter_id", "")):
			return false
	if step.has("page_id"):
		if str(panel.get("page_id", "")) != str(step.get("page_id", "")):
			return false
	if step.has("page_title_contains"):
		var title := str(panel.get("page_title", ""))
		if str(step.get("page_title_contains", "")) not in title:
			return false
	if step.has("min_chapter_count"):
		if int(panel.get("chapter_count", 0)) < int(step.get("min_chapter_count", 0)):
			return false
	return true
