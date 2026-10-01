class_name CharacterStatsHeader
extends VBoxContainer
## Class-aware header for the character stats panel (v516 D1): class icon, name, class,
## level chip, XP bar and an unspent-points badge. Display-only; takes the same
## `character_progression` dictionary the panel already receives.

const CharacterPanelStyles := preload("res://scripts/character_panel_styles.gd")
const ClassIconScript := preload("res://scripts/class_icon.gd")
const TextCatalogScript := preload("res://scripts/text_catalog.gd")

var _icon: Control
var _name_label: Label
var _class_label: Label
var _level_chip: Label
var _xp_bar: ProgressBar
var _xp_label: Label
var _points_chip: Label
var _accent: Color = CharacterPanelStyles.neutral_accent()
var _state: Dictionary = {}


## 0..1 fill for the XP bar. A null/absent `experience_to_next_level` means max level (full).
static func xp_fraction(progression: Dictionary) -> float:
	var remaining = progression.get("experience_to_next_level", null)
	if remaining == null:
		return 1.0
	var xp := maxi(0, int(progression.get("experience", 0)))
	var total := xp + maxi(0, int(remaining))
	if total <= 0:
		return 0.0
	return clampf(float(xp) / float(total), 0.0, 1.0)


static func xp_text(progression: Dictionary) -> String:
	var xp := int(progression.get("experience", 0))
	var remaining = progression.get("experience_to_next_level", null)
	if remaining == null:
		return "XP %d  %s" % [xp, TextCatalogScript.get_text("character_screen.max_level", "Max level")]
	return "XP %d (+%d)" % [xp, int(remaining)]


## Class identity color from class_presentations (fallback class for unknown), neutral when no class.
static func class_accent(class_id: String) -> Color:
	if class_id.strip_edges() == "":
		return CharacterPanelStyles.neutral_accent()
	var icon := ClassIconScript.new()
	icon.configure(class_id)
	var color: Color = icon.fill_color
	icon.free()
	return color


func _init() -> void:
	add_theme_constant_override("separation", 6)
	var card := PanelContainer.new()
	card.name = "header_card"
	card.add_theme_stylebox_override("panel", CharacterPanelStyles.card_style())
	add_child(card)
	var col := VBoxContainer.new()
	col.add_theme_constant_override("separation", 4)
	card.add_child(col)

	var top := HBoxContainer.new()
	top.add_theme_constant_override("separation", 8)
	col.add_child(top)
	_icon = ClassIconScript.new()
	_icon.custom_minimum_size = Vector2(48, 48)
	top.add_child(_icon)
	var names := VBoxContainer.new()
	names.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	names.add_theme_constant_override("separation", 0)
	top.add_child(names)
	_name_label = _label(CharacterPanelStyles.font_title(), CharacterPanelStyles.text())
	names.add_child(_name_label)
	_class_label = _label(CharacterPanelStyles.font_caption(), CharacterPanelStyles.text_dim())
	names.add_child(_class_label)
	_level_chip = _label(CharacterPanelStyles.font_caption(), CharacterPanelStyles.text())
	_level_chip.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	top.add_child(_level_chip)

	_xp_bar = ProgressBar.new()
	_xp_bar.custom_minimum_size = Vector2(0, 10)
	_xp_bar.show_percentage = false
	_xp_bar.min_value = 0.0
	_xp_bar.max_value = 1.0
	_xp_bar.step = 0.0
	col.add_child(_xp_bar)
	var bottom := HBoxContainer.new()
	col.add_child(bottom)
	_xp_label = _label(CharacterPanelStyles.font_caption(), CharacterPanelStyles.text())
	_xp_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	bottom.add_child(_xp_label)
	_points_chip = _label(CharacterPanelStyles.font_caption(), CharacterPanelStyles.text())
	bottom.add_child(_points_chip)
	configure("", {})


func configure(hero_name: String, progression: Dictionary) -> void:
	var class_id := str(progression.get("character_class", "")).strip_edges()
	_accent = class_accent(class_id)
	if class_id != "":
		_icon.call("configure", class_id)
	_icon.visible = class_id != ""
	var level := int(progression.get("level", 1))
	var points := int(progression.get("unspent_stat_points", 0))
	var class_name_text := ""
	if class_id != "":
		class_name_text = TextCatalogScript.get_text("character.class.%s" % class_id, class_id.capitalize())
	var fraction := xp_fraction(progression)
	_name_label.text = hero_name
	_class_label.text = class_name_text
	_class_label.visible = class_name_text != ""
	_level_chip.text = "Level %d" % level
	_level_chip.add_theme_stylebox_override("normal", CharacterPanelStyles.chip_style(_accent, false))
	_xp_bar.value = fraction
	_xp_bar.add_theme_stylebox_override("background", CharacterPanelStyles.xp_bar_background_style())
	_xp_bar.add_theme_stylebox_override("fill", CharacterPanelStyles.xp_bar_fill_style(_accent))
	_xp_label.text = xp_text(progression)
	_points_chip.visible = points > 0
	_points_chip.text = TextCatalogScript.get_text("character_screen.points_available", "%d points to spend") % points
	_points_chip.add_theme_stylebox_override("normal", CharacterPanelStyles.chip_style(_accent, true))
	_state = {
		"class_id": class_id,
		"class_name": class_name_text,
		"level": level,
		"xp_fraction": fraction,
		"xp_text": _xp_label.text,
		"points": points,
		"points_visible": points > 0,
		"accent": _accent.to_html(false),
	}


func get_header_state() -> Dictionary:
	return _state.duplicate(true)


func accent() -> Color:
	return _accent


func _label(font_size: int, color: Color) -> Label:
	var label := Label.new()
	label.add_theme_font_size_override("font_size", font_size)
	label.add_theme_color_override("font_color", color)
	label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return label
