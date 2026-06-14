extends RefCounted

static func build_tabs(parent: Control, on_selected: Callable) -> Array:
	var tabs := []
	for index in range(2):
		var tab := Button.new()
		tab.text = "I" if index == 0 else "II"
		tab.position = Vector2(126 + index * 44, 330)
		tab.size = Vector2(40, 28)
		tab.custom_minimum_size = tab.size
		tab.focus_mode = Control.FOCUS_NONE
		tab.pressed.connect(func() -> void:
			on_selected.call(index)
		)
		parent.add_child(tab)
		tabs.append(tab)
	return tabs


static func render_tabs(tabs: Array, active_weapon_set: int, viewed_weapon_set: int) -> void:
	for i in range(tabs.size()):
		var tab := tabs[i] as Button
		if tab == null:
			continue
		var active := i == active_weapon_set
		var viewed := i == viewed_weapon_set
		tab.text = ("I" if i == 0 else "II") + ("*" if active else "")
		tab.add_theme_color_override("font_color", Color("#f4d58d") if active else Color("#c7b89a"))
		tab.add_theme_stylebox_override("normal", _tab_style(viewed, active))
		tab.add_theme_stylebox_override("hover", _tab_style(true, active))
		tab.add_theme_stylebox_override("pressed", _tab_style(true, active))


static func fallback_sets(equipped: Dictionary) -> Array:
	return [
		{"index": 0, "main_hand": equipped.get("main_hand", null), "off_hand": equipped.get("off_hand", null)},
		{"index": 1, "main_hand": null, "off_hand": null},
	]


static func hand_equipped_id(sets: Array, equipped: Dictionary, viewed_weapon_set: int, slot: String) -> Variant:
	if slot != "main_hand" and slot != "off_hand":
		return equipped.get(slot, null)
	for set_data in sets:
		var data := set_data as Dictionary
		if int(data.get("index", -1)) == viewed_weapon_set:
			return data.get(slot, null)
	return equipped.get(slot, null)


static func is_equipped_instance(equipped: Dictionary, sets: Array, slots: Array, item_instance_id: String) -> bool:
	if item_instance_id == "":
		return false
	for slot in slots:
		var equipped_id = equipped.get(str(slot), null)
		if equipped_id != null and str(equipped_id) == item_instance_id:
			return true
	for set_data in sets:
		var data := set_data as Dictionary
		for slot in ["main_hand", "off_hand"]:
			var equipped_id = data.get(slot, null)
			if equipped_id != null and str(equipped_id) == item_instance_id:
				return true
	return false


static func _tab_style(viewed: bool, active: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#4a3822") if active else Color("#24201c")
	if viewed:
		s.bg_color = Color("#5a452c") if active else Color("#36302a")
	s.border_color = Color("#d0a84f") if active else Color("#6f6252")
	s.set_border_width_all(1)
	s.corner_radius_top_left = 3
	s.corner_radius_top_right = 3
	s.corner_radius_bottom_left = 3
	s.corner_radius_bottom_right = 3
	return s
