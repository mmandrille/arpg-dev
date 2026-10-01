class_name PlayerHealthBar
extends Control
## Player vitals HUD: a health globe and a mana globe at the bottom corners plus a name/level strip
## (v517). Public API and debug-state keys are unchanged from the former horizontal meters.

const HudGlobeScript := preload("res://scripts/hud_globe.gd")

var _hp_globe: HudGlobe
var _mana_globe: HudGlobe
var _identity_panel: PanelContainer
var _identity_label: Label
var _hp: int = 10
var _max_hp: int = 10
var _mana: int = 10
var _max_mana: int = 10
var _attack_recovery_remaining: float = 0.0
var _attack_recovery_total: float = 0.0
var _character_name := "Hero"
var _level := 1


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	_sync_position()
	get_viewport().size_changed.connect(_sync_position)
	set_process(false) # only ticks while an attack-recovery countdown runs


func _process(delta: float) -> void:
	_attack_recovery_remaining = maxf(0.0, _attack_recovery_remaining - maxf(0.0, delta))
	if _attack_recovery_remaining <= 0.0:
		set_process(false)


func update_hp(hp: int, max_hp: int, is_heal: bool = false) -> void:
	var was_hp := _hp
	_hp = hp
	_max_hp = max_hp
	_update_bars()
	if _hp_globe == null:
		return
	if is_heal and hp > was_hp:
		_hp_globe.flash(HudStyle.hp_flash_heal(), _hp_bar_color())
	elif hp < was_hp:
		_hp_globe.flash(HudStyle.hp_flash_damage(), _hp_bar_color())


func update_mana(mana: int, max_mana: int, is_restore: bool = false) -> void:
	var was_mana := _mana
	_mana = mana
	_max_mana = max_mana
	_update_bars()
	if _mana_globe != null and is_restore and mana > was_mana:
		_mana_globe.flash(HudStyle.mana_flash_restore(), _mana_bar_color())


func start_attack_recovery(duration_seconds: float) -> void:
	_attack_recovery_total = maxf(0.0, duration_seconds)
	_attack_recovery_remaining = _attack_recovery_total
	set_process(_attack_recovery_remaining > 0.0)


func set_identity(character_name: String, level: int) -> void:
	var next_name := character_name.strip_edges()
	_character_name = next_name if next_name != "" else "Hero"
	_level = maxi(1, level)
	_update_identity_label()


func get_debug_state() -> Dictionary:
	return {
		"character_name": _character_name,
		"level": _level,
		"identity_text": _identity_label.text if _identity_label != null else _identity_text(),
		"hp": _hp,
		"max_hp": _max_hp,
		"mana": _mana,
		"max_mana": _max_mana,
		"attack_recovery_remaining": _attack_recovery_remaining,
		"attack_recovery_total": _attack_recovery_total,
		"attack_recovery_fraction": _attack_recovery_fraction(),
	}


func _hp_bar_color() -> Color:
	return HudStyle.hp_color(float(_hp) / float(maxi(_max_hp, 1)))


func _mana_bar_color() -> Color:
	return HudStyle.mana()


func _update_bars() -> void:
	if _hp_globe == null or _mana_globe == null:
		return
	_hp_globe.set_ratio(float(_hp) / float(maxi(_max_hp, 1)))
	_mana_globe.set_ratio(float(_mana) / float(maxi(_max_mana, 1)))
	if not _hp_globe.is_flashing():
		_hp_globe.fill_color = _hp_bar_color()
	if not _mana_globe.is_flashing():
		_mana_globe.fill_color = _mana_bar_color()
	_hp_globe.set_label("%d / %d" % [_hp, _max_hp])
	_mana_globe.set_label("%d / %d" % [_mana, _max_mana])


func _identity_text() -> String:
	return "%s  Lv %d" % [_character_name, _level]


func _update_identity_label() -> void:
	if _identity_label != null:
		_identity_label.text = _identity_text()


func _build() -> void:
	if _hp_globe != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_hp_globe = HudGlobeScript.new()
	_mana_globe = HudGlobeScript.new()
	add_child(_hp_globe)
	add_child(_mana_globe)

	_identity_panel = PanelContainer.new()
	_identity_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_identity_panel.add_theme_stylebox_override("panel", HudStyle.frame_panel(0.84, 6.0, 2.0))
	add_child(_identity_panel)
	_identity_label = Label.new()
	_identity_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_identity_label.clip_text = true
	_identity_label.add_theme_color_override("font_color", HudStyle.text())
	_identity_label.add_theme_font_size_override("font_size", 15)
	_identity_panel.add_child(_identity_label)
	_update_identity_label()
	_update_bars()


func _attack_recovery_fraction() -> float:
	if _attack_recovery_total <= 0.0:
		return 0.0
	return clampf(_attack_recovery_remaining / _attack_recovery_total, 0.0, 1.0)


func _sync_position() -> void:
	if _hp_globe == null:
		return
	var vp := get_viewport_rect().size
	var health := HudLayout.health_globe_rect(vp)
	var mana := HudLayout.mana_globe_rect(vp)
	_hp_globe.position = health.position
	_hp_globe.size = health.size
	_mana_globe.position = mana.position
	_mana_globe.size = mana.size
	var identity := HudLayout.identity_rect(vp)
	_identity_panel.position = identity.position
	_identity_panel.size = identity.size
