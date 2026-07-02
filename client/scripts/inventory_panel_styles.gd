class_name InventoryPanelStyles
extends RefCounted

const ITEM_RARITY_BACKGROUNDS := {
	"common": Color("#343432"),
	"magic": Color("#1b3458"),
	"rare": Color("#5a4520"),
	"unique": Color("#5a2f17"),
	"set": Color("#173f28"),
}


static func panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.92)
	s.border_color = Color("#6b5420")
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	s.content_margin_left = 14
	s.content_margin_top = 12
	s.content_margin_right = 14
	s.content_margin_bottom = 12
	return s


static func slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#3d2e10") if hover else Color("#0a0908")
	s.border_color = Color("#8b6914") if hover else Color("#5c4a1f")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 4
	s.content_margin_top = 4
	s.content_margin_right = 4
	s.content_margin_bottom = 4
	return s


static func item_slot_style(rarity: String, hover: bool) -> StyleBoxFlat:
	var s := slot_style(hover)
	var base: Color = ITEM_RARITY_BACKGROUNDS.get(rarity.to_lower(), ITEM_RARITY_BACKGROUNDS["common"])
	s.bg_color = base.lightened(0.12) if hover else base
	s.border_color = base.lightened(0.46) if hover else base.lightened(0.28)
	return s


static func empty_slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#3a3a37") if hover else Color("#242422")
	s.border_color = Color("#8a877d") if hover else Color("#5f5b52")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 4
	s.content_margin_top = 4
	s.content_margin_right = 4
	s.content_margin_bottom = 4
	return s


static func blocked_slot_style(hover: bool) -> StyleBoxFlat:
	var s := empty_slot_style(hover)
	s.bg_color = Color("#2f302f") if hover else Color("#202120")
	s.border_color = Color("#96928a") if hover else Color("#6f6b64")
	return s


static func paper_doll_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#171715")
	s.border_color = Color("#5f5b52")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 8
	s.corner_radius_top_right = 8
	s.corner_radius_bottom_right = 8
	s.corner_radius_bottom_left = 8
	return s
