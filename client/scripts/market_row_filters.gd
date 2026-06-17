class_name MarketRowFilters
extends RefCounted

const MarketListingRowsScript := preload("res://scripts/market_listing_rows.gd")
const MarketFilterControlsScript := preload("res://scripts/market_filter_controls.gd")
const SOURCE_INDEX := "_market_source_index"


static func filter_sort_listings(listings: Array, query: String, sort_mode: String) -> Array:
	return _filter_sort_records(listings, query, sort_mode, "listing")


static func filter_sort_offers(offers: Array, query: String, sort_mode: String) -> Array:
	return _filter_sort_records(offers, query, sort_mode, "offer")


static func filter_sort_receipts(receipts: Array, query: String, sort_mode: String) -> Array:
	return _filter_sort_records(receipts, query, sort_mode, "receipt")


static func matching_listing(source: Array, listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> Dictionary:
	var matches: Array = []
	for listing in source:
		if typeof(listing) != TYPE_DICTIONARY:
			continue
		var rec := listing as Dictionary
		if listing_id != "" and str(rec.get("listing_id", "")) != listing_id:
			continue
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if price_gold >= 0 and int(rec.get("price_gold", 0)) != price_gold:
			continue
		matches.append(rec)
	if matches.is_empty():
		return {}
	return (matches[clampi(listing_index, 0, matches.size() - 1)] as Dictionary).duplicate(true)


static func matching_offer(source: Array, offer_id: String = "", offer_index: int = 0) -> Dictionary:
	var matches: Array = []
	for offer in source:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		if offer_id != "" and str(rec.get("offer_id", "")) != offer_id:
			continue
		matches.append(rec)
	if matches.is_empty():
		return {}
	return (matches[clampi(offer_index, 0, matches.size() - 1)] as Dictionary).duplicate(true)


static func _filter_sort_records(source: Array, query: String, sort_mode: String, kind: String) -> Array:
	var out: Array = []
	var needle := query.strip_edges().to_lower()
	var source_index := 0
	for value in source:
		if typeof(value) != TYPE_DICTIONARY:
			source_index += 1
			continue
		var rec := (value as Dictionary).duplicate(true)
		rec[SOURCE_INDEX] = source_index
		if needle == "" or _record_matches(rec, needle, kind):
			out.append(rec)
		source_index += 1
	if sort_mode != MarketFilterControlsScript.SORT_DEFAULT:
		out.sort_custom(func(a, b) -> bool:
			return _record_less(a as Dictionary, b as Dictionary, sort_mode, kind)
		)
	for rec in out:
		if typeof(rec) == TYPE_DICTIONARY:
			(rec as Dictionary).erase(SOURCE_INDEX)
	return out


static func _record_matches(rec: Dictionary, needle: String, kind: String) -> bool:
	for value in _record_search_values(rec, kind):
		if str(value).to_lower().find(needle) >= 0:
			return true
	return false


static func _record_search_values(rec: Dictionary, kind: String) -> Array:
	match kind:
		"offer":
			var listing := _offer_listing(rec)
			var values := [
				str(rec.get("offer_id", "")),
				str(rec.get("listing_id", "")),
				str(rec.get("bidder_account_id", "")),
				str(rec.get("status", "")),
				MarketListingRowsScript.listing_title(listing),
				str(listing.get("item_def_id", "")),
				str(listing.get("item_template_id", "")),
				str(listing.get("price_gold", "")),
			]
			for item in _offer_items(rec):
				if typeof(item) == TYPE_DICTIONARY:
					values.append(MarketListingRowsScript.item_title(item as Dictionary))
					values.append(str((item as Dictionary).get("item_def_id", "")))
					values.append(str((item as Dictionary).get("item_template_id", "")))
			return values
		"receipt":
			return [
				str(rec.get("action", "")),
				str(rec.get("item_def_id", "")),
				str(rec.get("stash_item_id", "")),
				str(rec.get("listing_id", "")),
				str(rec.get("offer_id", "")),
				str(rec.get("created_at", "")),
			]
	return [
		MarketListingRowsScript.listing_title(rec),
		str(rec.get("item_def_id", "")),
		str(rec.get("item_template_id", "")),
		str(rec.get("seller_account_id", "")),
		str(rec.get("price_gold", "")),
		str(rec.get("status", "")),
		str(rec.get("expires_at", "")),
		MarketListingRowsScript.expiration_label(rec),
	]


static func _record_less(left: Dictionary, right: Dictionary, sort_mode: String, kind: String) -> bool:
	match sort_mode:
		MarketFilterControlsScript.SORT_NAME:
			return _text_sort_key(_record_name(left, kind), left) < _text_sort_key(_record_name(right, kind), right)
		MarketFilterControlsScript.SORT_PRICE_LOW, MarketFilterControlsScript.SORT_PRICE_HIGH:
			var left_price := _record_price(left, kind)
			var right_price := _record_price(right, kind)
			if left_price == right_price:
				return _tie_key(left, kind) < _tie_key(right, kind)
			return left_price > right_price if sort_mode == MarketFilterControlsScript.SORT_PRICE_HIGH else left_price < right_price
		MarketFilterControlsScript.SORT_STATUS:
			return _text_sort_key(_record_status(left, kind), left) < _text_sort_key(_record_status(right, kind), right)
	return int(left.get(SOURCE_INDEX, 0)) < int(right.get(SOURCE_INDEX, 0))


static func _record_name(rec: Dictionary, kind: String) -> String:
	match kind:
		"offer":
			var item_names: Array = []
			for item in _offer_items(rec):
				if typeof(item) == TYPE_DICTIONARY:
					item_names.append(MarketListingRowsScript.item_title(item as Dictionary))
			return "%s %s" % [MarketListingRowsScript.listing_title(_offer_listing(rec)), " ".join(item_names)]
		"receipt":
			return "%s %s" % [str(rec.get("item_def_id", "")), str(rec.get("action", ""))]
	return MarketListingRowsScript.listing_title(rec)


static func _record_price(rec: Dictionary, kind: String) -> int:
	return int(_offer_listing(rec).get("price_gold", 0)) if kind == "offer" else int(rec.get("price_gold", 0))


static func _record_status(rec: Dictionary, kind: String) -> String:
	return str(rec.get("action", "")) if kind == "receipt" else str(rec.get("status", "active"))


static func _text_sort_key(value: String, rec: Dictionary) -> String:
	return "%s|%s" % [value.to_lower(), _tie_key(rec, "")]


static func _tie_key(rec: Dictionary, kind: String) -> String:
	match kind:
		"offer":
			return "%s|%04d" % [str(rec.get("offer_id", "")), int(rec.get(SOURCE_INDEX, 0))]
		"receipt":
			return "%s|%s|%04d" % [str(rec.get("created_at", "")), str(rec.get("offer_id", "")), int(rec.get(SOURCE_INDEX, 0))]
	return "%s|%04d" % [str(rec.get("listing_id", "")), int(rec.get(SOURCE_INDEX, 0))]


static func _offer_listing(offer: Dictionary) -> Dictionary:
	var listing = offer.get("listing", {})
	return (listing as Dictionary) if typeof(listing) == TYPE_DICTIONARY else {}


static func _offer_items(offer: Dictionary) -> Array:
	var items: Variant = offer.get("items", [])
	return items as Array if typeof(items) == TYPE_ARRAY else []
