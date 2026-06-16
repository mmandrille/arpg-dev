class_name SetCollectionPanel
extends PanelContainer

const StatLabels := preload("res://scripts/stat_labels.gd")
const SET_RULES_REL := "../shared/rules/set_items.v0.json"
const TITLE_FONT_SIZE := 19
const BODY_FONT_SIZE := 15
const SET_COLOR := Color("#55e66f")
const ACTIVE_COLOR := Color("#7df095")
const INACTIVE_COLOR := Color("#7d8d7f")
const MUTED_COLOR := Color("#9c8f78")

var _sets: Dictionary = {}
var _root: VBoxContainer
var _state: Dictionary = {"sets": [], "visible": false}


func _ready() -> void:
	_ensure_built()
	set_items([], {})
	visible = false


func set_items(items: Array, equipped: Dictionary) -> void:
	_ensure_catalog_loaded()
	var equipped_ids := _equipped_ids(equipped)
	var by_set := _initial_state()
	for item in items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		_apply_item(by_set, item as Dictionary, equipped_ids)
	for set_id in by_set.keys():
		_finalize_set(by_set[set_id])
	_state = {"sets": _sorted_sets(by_set), "visible": visible}
	_render()


func get_debug_state() -> Dictionary:
	_state["visible"] = visible
	return _state.duplicate(true)


func has_any_progress() -> bool:
	for set_state in _state.get("sets", []):
		if int((set_state as Dictionary).get("owned_count", 0)) > 0:
			return true
	return false


func _ensure_built() -> void:
	if _root != null:
		return
	custom_minimum_size = Vector2(330, 150)
	add_theme_stylebox_override("panel", _panel_style(Color("#18251d"), Color("#31523a")))
	_root = VBoxContainer.new()
	_root.add_theme_constant_override("separation", 6)
	add_child(_root)


func _render() -> void:
	_ensure_built()
	for child in _root.get_children():
		child.queue_free()
	var title := _label("Set Collection", TITLE_FONT_SIZE, SET_COLOR)
	_root.add_child(title)
	for set_state in _state.get("sets", []):
		_root.add_child(_set_block(set_state as Dictionary))


func _set_block(set_state: Dictionary) -> Control:
	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 3)
	var name := str(set_state.get("set_name", "Set"))
	var owned := int(set_state.get("owned_count", 0))
	var equipped := int(set_state.get("equipped_count", 0))
	var total := int(set_state.get("total_count", 0))
	box.add_child(_label("%s  %d/%d owned, %d/%d equipped" % [name, owned, total, equipped, total], BODY_FONT_SIZE, SET_COLOR))
	for piece in set_state.get("pieces", []):
		var rec := piece as Dictionary
		var color := ACTIVE_COLOR if str(rec.get("state", "")) == "equipped" else MUTED_COLOR
		if str(rec.get("state", "")) == "owned":
			color = Color("#d5c9a8")
		box.add_child(_label("  %s - %s" % [str(rec.get("state", "missing")).capitalize(), str(rec.get("name", ""))], BODY_FONT_SIZE, color))
	for bonus in set_state.get("bonuses", []):
		var rec := bonus as Dictionary
		var color := ACTIVE_COLOR if bool(rec.get("active", false)) else INACTIVE_COLOR
		box.add_child(_label("  %s" % str(rec.get("text", "")), BODY_FONT_SIZE, color))
	return box


func _label(text: String, size: int, color: Color) -> Label:
	var label := Label.new()
	label.text = text
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	label.add_theme_font_size_override("font_size", size)
	label.add_theme_color_override("font_color", color)
	return label


func _initial_state() -> Dictionary:
	var out := {}
	for set_id in _sets.keys():
		var set_def: Dictionary = _sets[set_id]
		var pieces: Array = []
		for piece in set_def.get("items", []):
			var piece_rec := piece as Dictionary
			pieces.append({
				"id": str(piece_rec.get("id", "")),
				"name": str(piece_rec.get("display_name", "")),
				"state": "missing",
			})
		out[set_id] = {
			"set_id": set_id,
			"set_name": str(set_def.get("display_name", set_id)),
			"owned_count": 0,
			"equipped_count": 0,
			"total_count": pieces.size(),
			"pieces": pieces,
			"bonuses": [],
		}
	return out


func _apply_item(by_set: Dictionary, item: Dictionary, equipped_ids: Dictionary) -> void:
	var display_name := str(item.get("display_name", ""))
	if display_name == "":
		return
	for set_id in by_set.keys():
		var set_state: Dictionary = by_set[set_id]
		var pieces: Array = set_state.get("pieces", [])
		for i in range(pieces.size()):
			var piece := pieces[i] as Dictionary
			if str(piece.get("name", "")) != display_name:
				continue
			var item_id := str(item.get("item_instance_id", ""))
			piece["state"] = "equipped" if equipped_ids.has(item_id) else "owned"
			pieces[i] = piece
			set_state["pieces"] = pieces
			by_set[set_id] = set_state
			return


func _finalize_set(set_state: Dictionary) -> void:
	var owned := 0
	var equipped := 0
	for piece in set_state.get("pieces", []):
		var state := str((piece as Dictionary).get("state", "missing"))
		if state == "owned" or state == "equipped":
			owned += 1
		if state == "equipped":
			equipped += 1
	set_state["owned_count"] = owned
	set_state["equipped_count"] = equipped
	set_state["bonuses"] = _bonus_rows(str(set_state.get("set_id", "")), equipped)


func _bonus_rows(set_id: String, equipped_count: int) -> Array:
	var set_def: Dictionary = _sets.get(set_id, {})
	var rows: Array = []
	for bonus in set_def.get("piece_bonuses", []):
		rows.append(_bonus_row(bonus as Dictionary, equipped_count))
	if set_def.has("full_set_bonus"):
		rows.append(_bonus_row(set_def.get("full_set_bonus", {}) as Dictionary, equipped_count))
	return rows


func _bonus_row(bonus: Dictionary, equipped_count: int) -> Dictionary:
	var required := int(bonus.get("required_pieces", 0))
	var active := equipped_count >= required
	return {
		"required_pieces": required,
		"active": active,
		"text": "%d-piece set bonus: %s (%s)" % [required, _stats_text(bonus.get("stats", {})), "active" if active else "inactive"],
	}


func _stats_text(stats_value: Variant) -> String:
	if typeof(stats_value) != TYPE_DICTIONARY:
		return "None"
	var stats := stats_value as Dictionary
	var parts: Array = []
	for key in stats.keys():
		parts.append("%s %s" % [StatLabels.display_name(str(key)), _format_stat(str(key), int(stats[key]))])
	parts.sort()
	return ", ".join(parts) if not parts.is_empty() else "None"


func _format_stat(stat: String, value: int) -> String:
	if stat.ends_with("_percent") or stat == "crit_chance" or stat == "evade_chance":
		return "+%d%%" % value
	return "+%d" % value


func _sorted_sets(by_set: Dictionary) -> Array:
	var ids := by_set.keys()
	ids.sort()
	var out: Array = []
	for set_id in ids:
		out.append(by_set[set_id])
	return out


func _equipped_ids(equipped: Dictionary) -> Dictionary:
	var ids := {}
	for slot in equipped.keys():
		var item_id := str(equipped.get(slot, ""))
		if item_id != "" and item_id != "0":
			ids[item_id] = true
	return ids


func _ensure_catalog_loaded() -> void:
	if not _sets.is_empty():
		return
	var path := ProjectSettings.globalize_path("res://").path_join(SET_RULES_REL)
	var text := FileAccess.get_file_as_string(path)
	var parsed = JSON.parse_string(text)
	if typeof(parsed) == TYPE_DICTIONARY:
		var all_sets: Dictionary = (parsed as Dictionary).get("sets", {})
		for set_id in all_sets.keys():
			var set_def := all_sets[set_id] as Dictionary
			if bool(set_def.get("enabled", false)) and str(set_def.get("status", "")) == "ready":
				_sets[str(set_id)] = set_def.duplicate(true)


func _panel_style(bg: Color, border: Color) -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = bg
	style.border_color = border
	style.set_border_width_all(1)
	style.corner_radius_top_left = 6
	style.corner_radius_top_right = 6
	style.corner_radius_bottom_left = 6
	style.corner_radius_bottom_right = 6
	style.content_margin_left = 8
	style.content_margin_right = 8
	style.content_margin_top = 8
	style.content_margin_bottom = 8
	return style
