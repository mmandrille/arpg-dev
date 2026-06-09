class_name PlayerHealthBar
extends Control

const BAR_W := 110.0
const BAR_H := 9.0

var _hp_fill: ColorRect
var _hp_label: Label
var _mana_fill: ColorRect
var _mana_label: Label
var _panel: PanelContainer
var _hp: int = 10
var _max_hp: int = 10
var _mana: int = 10
var _max_mana: int = 10
var _tween: Tween


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	_sync_position()
	get_viewport().size_changed.connect(_sync_position)


func update_hp(hp: int, max_hp: int, is_heal: bool = false) -> void:
	var was_hp := _hp
	_hp = hp
	_max_hp = max_hp
	_update_bars()
	if is_heal and hp > was_hp:
		_flash(_hp_fill, Color(0.35, 1.0, 0.45), _hp_bar_color())
	elif hp < was_hp:
		_flash(_hp_fill, Color(1.0, 0.22, 0.18), _hp_bar_color())


func update_mana(mana: int, max_mana: int, is_restore: bool = false) -> void:
	var was_mana := _mana
	_mana = mana
	_max_mana = max_mana
	_update_bars()
	if is_restore and mana > was_mana:
		_flash(_mana_fill, Color(0.45, 0.90, 1.0), _mana_bar_color())


func _flash(target: ColorRect, color: Color, return_color: Color) -> void:
	if target == null:
		return
	if _tween != null and _tween.is_valid():
		_tween.kill()
	_tween = create_tween()
	_tween.tween_property(target, "color", color, 0.05)
	_tween.tween_property(target, "color", return_color, 0.45)


func _hp_bar_color() -> Color:
	var pct := float(_hp) / float(maxi(_max_hp, 1))
	if pct > 0.6:
		return Color(0.22, 0.78, 0.28)
	if pct > 0.3:
		return Color(0.88, 0.68, 0.15)
	return Color(0.90, 0.18, 0.14)


func _mana_bar_color() -> Color:
	return Color("#48aeea")


func _update_bars() -> void:
	if _hp_fill == null or _hp_label == null or _mana_fill == null or _mana_label == null:
		return
	var hp_pct := float(_hp) / float(maxi(_max_hp, 1))
	_hp_fill.size.x = BAR_W * clampf(hp_pct, 0.0, 1.0)
	var mana_pct := float(_mana) / float(maxi(_max_mana, 1))
	_mana_fill.size.x = BAR_W * clampf(mana_pct, 0.0, 1.0)
	if _tween == null or not _tween.is_valid() or not _tween.is_running():
		_hp_fill.color = _hp_bar_color()
		_mana_fill.color = _mana_bar_color()
	_hp_label.text = "%d / %d" % [_hp, _max_hp]
	_mana_label.text = "%d / %d" % [_mana, _max_mana]


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)

	_panel = PanelContainer.new()
	_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.06, 0.05, 0.04, 0.84)
	style.border_color = Color("#5c4a1f")
	style.border_width_left = 1
	style.border_width_top = 1
	style.border_width_right = 1
	style.border_width_bottom = 1
	style.content_margin_left = 7
	style.content_margin_top = 5
	style.content_margin_right = 7
	style.content_margin_bottom = 5
	_panel.add_theme_stylebox_override("panel", style)
	add_child(_panel)

	var root := HBoxContainer.new()
	root.add_theme_constant_override("separation", 10)
	_panel.add_child(root)
	_build_meter(root, "♥", Color("#c0392b"), true)
	_build_meter(root, "✦", Color("#48aeea"), false)
	_update_bars()


func _build_meter(parent: HBoxContainer, icon_text: String, icon_color: Color, is_hp: bool) -> void:
	var vbox := VBoxContainer.new()
	vbox.add_theme_constant_override("separation", 4)
	parent.add_child(vbox)

	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 5)
	vbox.add_child(row)

	var icon := Label.new()
	icon.text = icon_text
	icon.add_theme_color_override("font_color", icon_color)
	icon.add_theme_font_size_override("font_size", 20)
	row.add_child(icon)

	var label := Label.new()
	label.add_theme_color_override("font_color", Color("#d8d0bd"))
	label.add_theme_font_size_override("font_size", 17)
	label.text = "10 / 10"
	row.add_child(label)

	var bar_bg := ColorRect.new()
	bar_bg.custom_minimum_size = Vector2(BAR_W, BAR_H)
	bar_bg.color = Color(0.13, 0.10, 0.08)
	vbox.add_child(bar_bg)

	var fill := ColorRect.new()
	fill.size = Vector2(BAR_W, BAR_H)
	fill.position = Vector2.ZERO
	fill.color = _hp_bar_color() if is_hp else _mana_bar_color()
	fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	bar_bg.add_child(fill)
	if is_hp:
		_hp_label = label
		_hp_fill = fill
	else:
		_mana_label = label
		_mana_fill = fill


func _sync_position() -> void:
	if _panel != null:
		_panel.set_deferred("position", Vector2(12.0, get_viewport_rect().size.y - 62.0))
