class_name BotMarketBadgeAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var badges: Dictionary = state.get("market_board_badges", {})
	if step.has("exists") and bool(badges.get("exists", false)) != bool(step.get("exists", false)):
		return false
	for key in ["incoming_bids", "published_listings"]:
		if step.has(key) and int(badges.get(key, -1)) != int(step.get(key, 0)):
			return false
	for key in ["incoming_visible", "published_visible"]:
		if step.has(key) and bool(badges.get(key, false)) != bool(step.get(key, false)):
			return false
	for key in ["incoming_text", "published_text", "incoming_color", "published_color"]:
		if step.has(key) and str(badges.get(key, "")) != str(step.get(key, "")):
			return false
	return true
