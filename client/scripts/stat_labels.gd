class_name StatLabels
extends RefCounted

const BASE_STATS := ["str", "dex", "vit", "magic"]

const DISPLAY_NAMES := {
	"level": "Level",
	"str": "STR",
	"dex": "DEX",
	"vit": "VIT",
	"magic": "Magic",
	"damage_min": "Min damage",
	"damage_max": "Max damage",
	"armor": "Armor",
	"block_percent": "Block",
	"max_hp": "Max HP",
	"hotbar_slots": "Hotbar slots",
	"inventory_rows": "Inventory rows",
}


static func display_name(stat: String) -> String:
	return str(DISPLAY_NAMES.get(stat, stat))
