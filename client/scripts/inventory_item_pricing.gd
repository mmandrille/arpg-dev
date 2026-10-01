class_name InventoryItemPricing
extends RefCounted
## Display-only gold/vendor price estimates for inventory item tooltips (v515).
## Reads shared shop rules through ItemRulesLoader; the server remains the
## authority for any real transaction price.


static func item_gold_value(item: Dictionary) -> int:
	for key in ["sell_price", "buy_price", "gold_value", "value"]:
		if item.has(key):
			return max(0, int(item.get(key, 0)))
	var buy_price := item_buy_price(item)
	if buy_price > 0:
		return max(1, int(floor(float(buy_price) * town_vendor_sell_multiplier())))
	return -1


static func item_buy_price(item: Dictionary) -> int:
	var def_id := str(item.get("item_def_id", ""))
	var shop := town_vendor_rules()
	if def_id == "" or shop.is_empty():
		return 0
	if is_generated_item(item):
		return generated_buy_price(item, shop)
	var fixed_offers = shop.get("fixed_offers", [])
	if typeof(fixed_offers) == TYPE_ARRAY:
		for offer in fixed_offers:
			if typeof(offer) == TYPE_DICTIONARY and str((offer as Dictionary).get("item_def_id", "")) == def_id:
				return int((offer as Dictionary).get("buy_price", 0))
	return 0


static func generated_buy_price(item: Dictionary, shop: Dictionary) -> int:
	var template_id := str(item.get("item_template_id", item.get("item_def_id", "")))
	var template: Dictionary = ItemRulesLoader.item_templates.get(template_id, {})
	if template.is_empty():
		return 0
	var pricing: Dictionary = shop.get("pricing", {})
	var round_to := int(pricing.get("round_buy_to", 0))
	if round_to <= 0:
		return 0
	var rarity := str(item.get("rarity", "common"))
	var rarity_multipliers: Dictionary = pricing.get("rarity_multipliers", {})
	var multiplier := float(rarity_multipliers.get(rarity, 0.0))
	if multiplier <= 0.0:
		return 0
	var base_stats: Dictionary = template.get("base_stats", {})
	var final_stats: Dictionary = item.get("rolled_stats", {})
	var stat_weights: Dictionary = pricing.get("stat_weights", {})
	var slot_base: Dictionary = pricing.get("slot_base", {})
	var score := int(slot_base.get(str(template.get("slot", "")), 0))
	for stat in stat_weights.keys():
		var key := str(stat)
		var weight := int(stat_weights.get(key, 0))
		score += int(base_stats.get(key, 0)) * weight
	for stat in stat_weights.keys():
		var key := str(stat)
		var weight := int(stat_weights.get(key, 0))
		var delta := int(final_stats.get(key, 0)) - int(base_stats.get(key, 0))
		if delta > 0:
			score += delta * weight
	var raw = max(1.0, float(score) * multiplier)
	return int(ceil(raw / float(round_to))) * round_to


static func is_generated_item(item: Dictionary) -> bool:
	if str(item.get("item_template_id", "")) != "":
		return true
	var def_id := str(item.get("item_def_id", ""))
	return ItemRulesLoader.item_templates.has(def_id) and str(item.get("rarity", "")) != ""


static func town_vendor_rules() -> Dictionary:
	var shops: Dictionary = ItemRulesLoader.shop_rules.get("shops", {})
	return shops.get("town_vendor", {})


static func town_vendor_sell_multiplier() -> float:
	var shop := town_vendor_rules()
	var pricing: Dictionary = shop.get("pricing", {})
	return float(pricing.get("sell_multiplier", 0.25))
