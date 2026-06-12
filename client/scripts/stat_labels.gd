class_name StatLabels
extends RefCounted

const TextCatalogScript := preload("res://scripts/text_catalog.gd")

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
	"max_mana": "Max mana",
	"health_regen_per_second": "HP regen /s",
	"mana_regen_per_second": "Mana regen /s",
	"health_regen_per_10_seconds": "HP regen / 10s",
	"mana_regen_per_10_seconds": "Mana regen / 10s",
	"hotbar_slots": "Hotbar slots",
	"inventory_rows": "Inventory rows",
}


static func display_name(stat: String) -> String:
	var fallback := str(DISPLAY_NAMES.get(stat, stat))
	return TextCatalogScript.get_text("stat.%s" % stat, fallback)
