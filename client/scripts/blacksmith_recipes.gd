class_name BlacksmithRecipes
extends RefCounted

const RECIPE_ITEM_UPGRADE := "item_upgrade"
const RECIPE_WEAPON_HONING := "weapon_honing"
const RECIPE_ARMOR_REINFORCEMENT := "armor_reinforcement"
const ARMOR_SLOTS := ["off_hand", "head", "chest", "gloves", "belt", "boots"]


static func options(resource_item_def_id: String, resource_count: int, success_chance_percent: int, max_level: int) -> Array:
	return [
		_option(RECIPE_ITEM_UPGRADE, label(RECIPE_ITEM_UPGRADE), eligibility(RECIPE_ITEM_UPGRADE), resource_item_def_id, resource_count, success_chance_percent, max_level),
		_option(RECIPE_WEAPON_HONING, label(RECIPE_WEAPON_HONING), eligibility(RECIPE_WEAPON_HONING), resource_item_def_id, resource_count, success_chance_percent, max_level),
		_option(RECIPE_ARMOR_REINFORCEMENT, label(RECIPE_ARMOR_REINFORCEMENT), eligibility(RECIPE_ARMOR_REINFORCEMENT), resource_item_def_id, resource_count, success_chance_percent, max_level),
	]


static func ids() -> Array:
	return [RECIPE_ITEM_UPGRADE, RECIPE_WEAPON_HONING, RECIPE_ARMOR_REINFORCEMENT]


static func label(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_WEAPON_HONING:
			return "Hone Weapon"
		RECIPE_ARMOR_REINFORCEMENT:
			return "Reinforce Armor"
		_:
			return "Upgrade Item"


static func eligibility(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_WEAPON_HONING:
			return "Eligible: Weapons only"
		RECIPE_ARMOR_REINFORCEMENT:
			return "Eligible: Armor pieces only"
		_:
			return "Eligible: Equipment"


static func rejection_message(recipe_id: String) -> String:
	match recipe_id:
		RECIPE_WEAPON_HONING:
			return "%s needs a weapon" % label(recipe_id)
		RECIPE_ARMOR_REINFORCEMENT:
			return "%s needs armor" % label(recipe_id)
		_:
			return "Recipe cannot modify this item"


static func accepts_item(recipe_id: String, item: Dictionary) -> bool:
	if recipe_id == RECIPE_ITEM_UPGRADE:
		return true
	var def := _item_definition(item)
	if recipe_id == RECIPE_WEAPON_HONING:
		return _can_be_weapon_honed(def)
	if recipe_id == RECIPE_ARMOR_REINFORCEMENT:
		return _can_be_armor_reinforced(def)
	return false


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


static func _item_definition(item: Dictionary) -> Dictionary:
	var def_id := str(item.get("item_template_id", item.get("item_def_id", "")))
	var def := ItemRulesLoader.item_definition(def_id)
	if def.is_empty():
		def = ItemRulesLoader.item_definition(str(item.get("item_def_id", "")))
	return def


static func _can_be_weapon_honed(def: Dictionary) -> bool:
	var stats: Dictionary = def.get("base_stats", {})
	return str(def.get("slot", "")) == "main_hand" and int(stats.get("damage_min", 0)) > 0 and int(stats.get("damage_max", 0)) > 0


static func _can_be_armor_reinforced(def: Dictionary) -> bool:
	var stats: Dictionary = def.get("base_stats", {})
	return ARMOR_SLOTS.has(str(def.get("slot", ""))) and int(stats.get("armor", 0)) > 0
