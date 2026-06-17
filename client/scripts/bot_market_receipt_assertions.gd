class_name BotMarketReceiptAssertions
extends RefCounted

static func matches(step: Dictionary, state: Dictionary) -> bool:
	if not (step.has("receipt_equals") or step.has("receipt_at_least") or step.has("receipt_action") or step.has("receipt_item_def_id")):
		return true
	var rows := _matching_rows(step, state)
	if step.has("receipt_equals") and rows.size() != int(step.get("receipt_equals", 0)):
		return false
	if step.has("receipt_at_least") and rows.size() < int(step.get("receipt_at_least", 0)):
		return false
	return true

static func _matching_rows(step: Dictionary, state: Dictionary) -> Array:
	var rows: Array = (state.get("market_panel", {}) as Dictionary).get("receipt_rows", [])
	var out: Array = []
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if step.has("receipt_action") and str(rec.get("action", "")) != str(step.get("receipt_action", "")):
			continue
		if step.has("receipt_item_def_id") and str(rec.get("item_def_id", "")) != str(step.get("receipt_item_def_id", "")):
			continue
		out.append(rec)
	return out
