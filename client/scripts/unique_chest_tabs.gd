class_name UniqueChestTabs
extends RefCounted

const UNIQUES := "uniques"
const SETS := "sets"
const TABS := [UNIQUES, SETS]


static func label(tab: String) -> String:
	return "Sets" if tab == SETS else "Uniques"


static func counts(items: Array) -> Dictionary:
	var out := {
		UNIQUES: 0,
		SETS: 0,
	}
	for item in items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rarity := str((item as Dictionary).get("rarity", "")).to_lower()
		if rarity == "set":
			out[SETS] += 1
		elif rarity == "unique":
			out[UNIQUES] += 1
	return out


static func item_matches_tab(item: Dictionary, tab: String) -> bool:
	var rarity := str(item.get("rarity", "")).to_lower()
	if tab == SETS:
		return rarity == "set"
	return rarity == "unique"


static func add_bar(parent: Control, font_size: int, select_tab: Callable) -> HBoxContainer:
	var bar := HBoxContainer.new()
	bar.add_theme_constant_override("separation", 6)
	bar.visible = false
	parent.add_child(bar)
	for tab in TABS:
		var btn := Button.new()
		btn.set_meta("unique_chest_tab", tab)
		btn.text = label(tab)
		btn.toggle_mode = true
		btn.focus_mode = Control.FOCUS_NONE
		btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		btn.custom_minimum_size = Vector2(0, 30)
		btn.add_theme_font_size_override("font_size", font_size)
		btn.pressed.connect(func() -> void: select_tab.call(tab))
		bar.add_child(btn)
	return bar


static func sync_bar(bar: HBoxContainer, visible: bool, active_tab: String, items: Array, interactive: bool) -> void:
	if bar == null:
		return
	bar.visible = visible
	var tab_counts := counts(items)
	for child in bar.get_children():
		if not child is Button:
			continue
		var btn := child as Button
		var tab := str(btn.get_meta("unique_chest_tab", UNIQUES))
		btn.text = "%s (%d)" % [label(tab), int(tab_counts.get(tab, 0))]
		btn.button_pressed = tab == active_tab
		btn.disabled = not interactive
