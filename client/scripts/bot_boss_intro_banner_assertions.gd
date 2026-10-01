class_name BotBossIntroBannerAssertions
extends RefCounted
## wait_boss_intro_banner (v512): matches the latched boss intro banner debug state.
## Keys: template_id, title, visible, played_min (default 1).


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var banner: Dictionary = state.get("boss_intro_banner", {})
	if int(banner.get("played_count", 0)) < int(step.get("played_min", 1)):
		return false
	if step.has("visible") and bool(banner.get("visible", false)) != bool(step.get("visible", true)):
		return false
	if step.has("template_id") and str(banner.get("template_id", "")) != str(step.get("template_id", "")):
		return false
	if step.has("title") and str(banner.get("title", "")) != str(step.get("title", "")):
		return false
	return true
