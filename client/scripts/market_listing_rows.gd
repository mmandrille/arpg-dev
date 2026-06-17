class_name MarketListingRows
extends RefCounted

const UNKNOWN_ITEM := "Unknown item"


static func listing_title(listing: Dictionary) -> String:
	var display := str(listing.get("display_name", ""))
	return display if display != "" else str(listing.get("item_def_id", UNKNOWN_ITEM)).replace("_", " ").capitalize()


static func item_title(item: Dictionary) -> String:
	var display := str(item.get("display_name", ""))
	return display if display != "" else str(item.get("item_def_id", UNKNOWN_ITEM)).replace("_", " ").capitalize()


static func item_detail(item: Dictionary) -> String:
	var lines: Array = item.get("summary_lines", [])
	if not lines.is_empty():
		return str(lines[0])
	var slot := str(item.get("slot", ""))
	return "Slot: %s" % slot.replace("_", " ") if slot != "" else str(item.get("stash_item_id", ""))


static func short_label(def_id: String) -> String:
	if def_id == "":
		return "?"
	var out := ""
	for part in def_id.split("_"):
		if str(part).length() > 0:
			out += str(part).substr(0, 1).to_upper()
	return out.substr(0, 3)


static func seller_detail(listing: Dictionary) -> String:
	return "%d gold - seller %s" % [
		int(listing.get("price_gold", 0)),
		str(listing.get("seller_account_id", "")).substr(0, 10),
	]


static func expiration_label(listing: Dictionary, now_unix: float = -1.0) -> String:
	var expires_at := str(listing.get("expires_at", "")).strip_edges()
	if expires_at == "":
		return ""
	var expires_unix := _parse_rfc3339ish(expires_at)
	if expires_unix <= 0:
		return "Expires %s" % _compact_timestamp(expires_at)
	var now := now_unix if now_unix >= 0.0 else Time.get_unix_time_from_system()
	var remaining := int(ceil(expires_unix - now))
	if remaining <= 0:
		return "Expired"
	return "Expires in %s" % _duration_label(remaining)


static func debug_listing_rows(source: Array, stat_lines: Callable) -> Array:
	var rows: Array = []
	for listing in source:
		if typeof(listing) != TYPE_DICTIONARY:
			continue
		var rec := listing as Dictionary
		var expires_at := str(rec.get("expires_at", ""))
		var expiration := expiration_label(rec)
		rows.append({
			"listing_id": str(rec.get("listing_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
			"seller_account_id": str(rec.get("seller_account_id", "")),
			"price_gold": int(rec.get("price_gold", 0)),
			"visible_detail": seller_detail(rec),
			"expires_at": expires_at,
			"expiration_label": expiration,
			"expiration_visible": expiration != "",
			"has_icon": str(rec.get("item_def_id", "")) != "",
			"stat_lines": stat_lines.call(rec),
		})
	return rows


static func _parse_rfc3339ish(value: String) -> float:
	var text := value.strip_edges()
	if text.ends_with("Z"):
		text = text.substr(0, text.length() - 1)
	var dot := text.find(".")
	if dot >= 0:
		text = text.substr(0, dot)
	text = text.replace("T", " ")
	if text.length() < 19 or text.substr(4, 1) != "-" or text.substr(7, 1) != "-" or text.substr(10, 1) != " " or text.substr(13, 1) != ":" or text.substr(16, 1) != ":":
		return 0.0
	return Time.get_unix_time_from_datetime_string(text)


static func _compact_timestamp(value: String) -> String:
	return value.replace("T", " ").replace("Z", " UTC").substr(0, 16)


static func _duration_label(seconds: int) -> String:
	if seconds >= 86400:
		return "%dd %dh" % [seconds / 86400, (seconds % 86400) / 3600]
	if seconds >= 3600:
		return "%dh %dm" % [seconds / 3600, (seconds % 3600) / 60]
	if seconds >= 60:
		return "%dm" % [ceili(float(seconds) / 60.0)]
	return "<1m"
