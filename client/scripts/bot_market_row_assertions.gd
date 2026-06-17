class_name BotMarketRowAssertions
extends RefCounted


static func listing_rows_match(step: Dictionary, state: Dictionary) -> bool:
	if not _filter_state_matches(step, state):
		return false
	if not _has_listing_expectation(step):
		return true
	var rows := matching_listing_rows(step, state)
	if step.has("equals") and rows.size() != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and rows.size() < int(step.get("at_least", 0)):
		return false
	if step.has("first_item_def_id") and (rows.is_empty() or str((rows[0] as Dictionary).get("item_def_id", "")) != str(step.get("first_item_def_id", ""))):
		return false
	return true


static func offer_rows_match(step: Dictionary, state: Dictionary) -> bool:
	if not _has_offer_expectation(step):
		return true
	var rows := matching_offer_rows(step, state)
	if step.has("offer_equals") and rows.size() != int(step.get("offer_equals", 0)):
		return false
	if step.has("offer_at_least") and rows.size() < int(step.get("offer_at_least", 0)):
		return false
	return true


static func matching_listing_rows(step: Dictionary, state: Dictionary) -> Array:
	var panel: Dictionary = state.get("market_panel", {})
	var rows: Array = panel.get("owned_listing_rows", []) if bool(step.get("seller_owned", false)) else panel.get("listing_rows", [])
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
		if step.has("expiration_visible") and bool(rec.get("expiration_visible", false)) != bool(step.get("expiration_visible", false)):
			continue
		if step.has("expiration_contains") and not str(rec.get("expiration_label", "")).contains(str(step.get("expiration_contains", ""))):
			continue
		if step.has("seller_owned") and bool(step.get("seller_owned", false)) != (str(rec.get("seller_account_id", "")) == str(panel.get("account_id", ""))):
			continue
		out.append(rec)
	return out


static func matching_offer_rows(step: Dictionary, state: Dictionary) -> Array:
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
		if step.has("listing_item_def_id") and str(rec.get("listing_item_def_id", "")) != str(step.get("listing_item_def_id", "")):
			continue
		out.append(rec)
	return out


static func _filter_state_matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("market_panel", {})
	for key in ["market_search_text", "market_sort_mode"]:
		if step.has(key) and str(panel.get(key, "")) != str(step.get(key, "")):
			return false
	for key in ["filtered_listing_count", "filtered_owned_listing_count", "filtered_offer_count", "filtered_receipt_count"]:
		if step.has(key) and int(panel.get(key, -1)) != int(step.get(key, -1)):
			return false
	return true


static func _has_listing_expectation(step: Dictionary) -> bool:
	for key in ["equals", "at_least", "price_gold", "item_def_id", "rolled", "expiration_visible", "expiration_contains", "first_item_def_id", "listing_id", "seller_owned"]:
		if step.has(key):
			return true
	return false


static func _has_offer_expectation(step: Dictionary) -> bool:
	for key in ["offer_equals", "offer_at_least", "offer_item_def_id", "listing_item_def_id", "offer_status", "offer_id"]:
		if step.has(key):
			return true
	return false
