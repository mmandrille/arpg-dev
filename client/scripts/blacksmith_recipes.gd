class_name BlacksmithRecipes
extends RefCounted

const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")
const RECIPE_ITEM_UPGRADE := "item_upgrade"
const RECIPE_ITEM_RENEW := "item_renew"


static func options(success_chance_percent: int, max_level: int) -> Array:
	return [
		_option(RECIPE_ITEM_UPGRADE, label(RECIPE_ITEM_UPGRADE), eligibility(RECIPE_ITEM_UPGRADE), "upgrade_shard", 1, success_chance_percent, max_level),
		_option(RECIPE_ITEM_RENEW, label(RECIPE_ITEM_RENEW), eligibility(RECIPE_ITEM_RENEW), "renew_stone", 1, 100, max_level),
	]


static func ids() -> Array:
	return [RECIPE_ITEM_UPGRADE, RECIPE_ITEM_RENEW]


static func label(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_ITEM_RENEW:
			return "Renew Item"
		_:
			return "Upgrade Item"


static func eligibility(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_ITEM_RENEW:
			return "Eligible: Equipment (reroll affixes)"
		_:
			return "Eligible: Equipment"


static func rejection_message(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_ITEM_RENEW:
			return "%s needs equipment" % label(recipe_id)
		_:
			return "Recipe cannot modify this item"


static func accepts_item(recipe_id: String, item: Dictionary) -> bool:
	return recipe_id == RECIPE_ITEM_UPGRADE or recipe_id == RECIPE_ITEM_RENEW


static func resource_item_def_id(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_ITEM_RENEW:
			return "renew_stone"
		_:
			return "upgrade_shard"


static func required_resource_level(recipe_id: String, item: Dictionary) -> int:
	return maxi(1, BlacksmithUpgradePreviewScript.item_level(item))


static func _option(id: String, option_label: String, option_eligibility: String, resource_item_def_id: String, resource_count: int, success_chance_percent: int, max_level: int) -> Dictionary:
	return {
		"id": id,
		"label": option_label,
		"eligibility": option_eligibility,
		"resource_item_def_id": resource_item_def_id,
		"resource_required_count": resource_count,
		"success_chance_percent": success_chance_percent,
		"max_level": max_level,
	}
