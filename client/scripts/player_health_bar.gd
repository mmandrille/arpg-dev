class_name PlayerHealthBar
extends Control

const BAR_W := 110.0
const BAR_H := 9.0

var _fill: ColorRect
var _label: Label
var _panel: PanelContainer
var _hp: int = 10
var _max_hp: int = 10
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
	_update_bar()
	if is_heal and hp > was_hp:
		_flash(Color(0.35, 1.0, 0.45))
	elif hp < was_hp:
		_flash(Color(1.0, 0.22, 0.18))


func _flash(color: Color) -> void:
	if _tween != null and _tween.is_valid():
		_tween.kill()
	_tween = create_tween()
	_tween.tween_property(_fill, "color", color, 0.05)
	_tween.tween_property(_fill, "color", _bar_color(), 0.45)


func _bar_color() -> Color:
	var pct := float(_hp) / float(maxi(_max_hp, 1))
	if pct > 0.6:
		return Color(0.22, 0.78, 0.28)
	if pct > 0.3:
		return Color(0.88, 0.68, 0.15)
	return Color(0.90, 0.18, 0.14)


func _update_bar() -> void:
	if _fill == null or _label == null:
		return
	var pct := float(_hp) / float(maxi(_max_hp, 1))
	_fill.size.x = BAR_W * clampf(pct, 0.0, 1.0)
	if _tween == null or not _tween.is_valid() or not _tween.is_running():
		_fill.color = _bar_color()
	_label.text = "%d / %d" % [_hp, _max_hp]


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

	var vbox := VBoxContainer.new()
	vbox.add_theme_constant_override("separation", 4)
	_panel.add_child(vbox)

	# Top row: heart + "HP: X / Y"
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 5)
	vbox.add_child(row)

	var heart := Label.new()
	heart.text = "♥"
	heart.add_theme_color_override("font_color", Color("#c0392b"))
	heart.add_theme_font_size_override("font_size", 13)
	row.add_child(heart)

	_label = Label.new()
	_label.add_theme_color_override("font_color", Color("#d8d0bd"))
	_label.add_theme_font_size_override("font_size", 11)
	_label.text = "10 / 10"
	row.add_child(_label)

	# Bar row: background + fill overlay
	var bar_bg := ColorRect.new()
	bar_bg.custom_minimum_size = Vector2(BAR_W, BAR_H)
	bar_bg.color = Color(0.13, 0.10, 0.08)
	vbox.add_child(bar_bg)

	_fill = ColorRect.new()
	_fill.size = Vector2(BAR_W, BAR_H)
	_fill.position = Vector2.ZERO
	_fill.color = _bar_color()
	_fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	bar_bg.add_child(_fill)


func _sync_position() -> void:
	if _panel != null:
		_panel.set_deferred("position", Vector2(12.0, get_viewport_rect().size.y - 62.0))
