class_name StatusEffectsBar
extends Control

const TICK_DURATION_S := 1.0 / 15.0
const ICON_SIZE := Vector2(36, 36)

var _effects: Dictionary = {}
var _row: HBoxContainer


class StatusEffectIcon:
	extends Control

	var label: String = ""
	var color: Color = Color("#9aa0a6")
	var accent: Color = Color.WHITE
	var fraction: float = 1.0

	func _ready() -> void:
		custom_minimum_size = Vector2(36, 36)
		mouse_filter = Control.MOUSE_FILTER_IGNORE

	func _draw() -> void:
		var rect := Rect2(Vector2.ZERO, size)
		draw_rect(rect, Color(0.045, 0.04, 0.035, 0.92), true)
		draw_rect(Rect2(Vector2(3, 3), size - Vector2(6, 6)), color.darkened(0.45), true)
		var fill_h := maxf(0.0, size.y - 6.0) * clampf(fraction, 0.0, 1.0)
		var fill_rect := Rect2(Vector2(3, size.y - 3.0 - fill_h), Vector2(size.x - 6.0, fill_h))
		draw_rect(fill_rect, color, true)
		draw_rect(rect, accent, false, 1.0)
		var font := get_theme_default_font()
		var font_size := 15
		var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_CENTER, -1, font_size)
		var text_pos := Vector2((size.x - text_size.x) * 0.5, (size.y + text_size.y * 0.45) * 0.5)
		draw_string(font, text_pos + Vector2(1, 1), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color(0, 0, 0, 0.7))
		draw_string(font, text_pos, label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, accent)


func _ready() -> void:
	SkillRulesLoader.ensure_loaded()
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_process(true)
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	_build()


func _process(delta: float) -> void:
	if _effects.is_empty():
		return
	var changed := false
	for skill_id in _effects.keys():
		var rec: Dictionary = _effects[skill_id]
		rec["remaining_ticks"] = maxf(0.0, float(rec.get("remaining_ticks", 0.0)) - (delta / TICK_DURATION_S))
		if float(rec.get("remaining_ticks", 0.0)) <= 0.0:
			_effects.erase(skill_id)
		changed = true
	if changed:
		_render()


func start_effect(event: Dictionary) -> void:
	var skill_id := str(event.get("skill_id", ""))
	if skill_id == "":
		return
	var total_ticks := int(event.get("total_ticks", event.get("remaining_ticks", 0)))
	var remaining_ticks := float(event.get("remaining_ticks", total_ticks))
	if total_ticks <= 0 or remaining_ticks <= 0.0:
		return
	var presentation := SkillRulesLoader.skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	_effects[skill_id] = {
		"skill_id": skill_id,
		"label": str(icon.get("label", skill_id.substr(0, 1).to_upper())),
		"color": str(icon.get("color", "#9aa0a6")),
		"accent": str(icon.get("accent", "#ffffff")),
		"remaining_ticks": remaining_ticks,
		"total_ticks": total_ticks,
		"name": SkillRulesLoader.skill_display_name(skill_id),
	}
	_render()


func end_effect(skill_id: String) -> void:
	if skill_id == "":
		return
	_effects.erase(skill_id)
	_render()


func clear_effects() -> void:
	_effects.clear()
	_render()


func get_debug_state() -> Dictionary:
	var rows: Array = []
	var ids := _effects.keys()
	ids.sort()
	for skill_id in ids:
		var rec: Dictionary = _effects[skill_id]
		var total_ticks := int(rec.get("total_ticks", 0))
		var remaining_ticks := float(rec.get("remaining_ticks", 0.0))
		rows.append({
			"skill_id": str(skill_id),
			"label": str(rec.get("label", "")),
			"remaining_ticks": int(ceil(remaining_ticks)),
			"total_ticks": total_ticks,
			"fraction": _effect_fraction(rec),
		})
	return {"effects": rows, "visible": visible}


func _build() -> void:
	_row = HBoxContainer.new()
	_row.add_theme_constant_override("separation", 5)
	add_child(_row)
	_render()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_position_row()


func _position_row() -> void:
	if _row == null:
		return
	var vp := get_viewport_rect().size
	var width := 10.0 * 41.0
	_row.position = Vector2((vp.x - width) * 0.5, vp.y - 122.0)
	_row.size = Vector2(width, 40.0)


func _render() -> void:
	if _row == null:
		return
	for child in _row.get_children():
		child.queue_free()
	var ids := _effects.keys()
	ids.sort()
	for skill_id in ids:
		var rec: Dictionary = _effects[skill_id]
		var icon := StatusEffectIcon.new()
		icon.label = str(rec.get("label", ""))
		icon.color = Color(str(rec.get("color", "#9aa0a6")))
		icon.accent = Color(str(rec.get("accent", "#ffffff")))
		icon.fraction = _effect_fraction(rec)
		icon.tooltip_text = "%s %ds" % [str(rec.get("name", skill_id)), int(ceil(float(rec.get("remaining_ticks", 0.0)) * TICK_DURATION_S))]
		_row.add_child(icon)
	visible = not _effects.is_empty()
	_position_row()


func _effect_fraction(rec: Dictionary) -> float:
	var total_ticks := float(rec.get("total_ticks", 0))
	if total_ticks <= 0.0:
		return 0.0
	return clampf(float(rec.get("remaining_ticks", 0.0)) / total_ticks, 0.0, 1.0)
