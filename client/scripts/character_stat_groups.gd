class_name CharacterStatGroups
extends RefCounted
## Presentation grouping of derived stats for the character screen (v516 D2).
## Code-owned on purpose: this is display ordering, not balance tuning.

const TextCatalogScript := preload("res://scripts/text_catalog.gd")

const GROUPS := [
	{
		"id": "offense",
		"title_key": "character_screen.group.offense",
		"title_fallback": "Offense",
		"keys": ["damage_min", "damage_max", "ranged_damage_bonus_percent", "attack_speed", "attack_interval_ticks", "hit_chance", "crit_chance", "crit_damage"],
	},
	{
		"id": "defense",
		"title_key": "character_screen.group.defense",
		"title_fallback": "Defense",
		"keys": ["armor", "evade_chance", "block_percent"],
	},
	{
		"id": "vitals",
		"title_key": "character_screen.group.vitals",
		"title_fallback": "Vitals",
		"keys": ["max_hp", "max_mana", "health_regen_per_second", "mana_regen_per_second"],
	},
	{
		"id": "utility",
		"title_key": "character_screen.group.utility",
		"title_fallback": "Utility",
		"keys": ["movement_speed", "light_radius"],
	},
]


static func group_ids() -> Array:
	var out := []
	for group in GROUPS:
		out.append(str(group["id"]))
	return out


static func group_title(group: Dictionary) -> String:
	return TextCatalogScript.get_text(str(group.get("title_key", "")), str(group.get("title_fallback", "")))


static func group_for(stat_key: String) -> String:
	for group in GROUPS:
		if stat_key in group["keys"]:
			return str(group["id"])
	return ""


static func ordered_keys() -> Array:
	var out := []
	for group in GROUPS:
		out.append_array(group["keys"])
	return out
