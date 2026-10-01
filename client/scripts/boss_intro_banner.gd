## One-shot boss intro name banner (v512). Non-blocking, mouse-transparent; timing/colors come from
## shared/assets/boss_presentation.v0.json. State is latched for bot assertions.
class_name BossIntroBanner
extends Control

const LoaderScript := preload("res://scripts/boss_presentation_loader.gd")

var played_count: int = 0
var last_template_id: String = ""
var last_title: String = ""
var last_epithet: String = ""
var banner_visible: bool = false

var _title: Label
var _epithet: Label
var _tween: Tween


func _init() -> void:
	name = "BossIntroBanner"
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_FULL_RECT)
	visible = false
	modulate.a = 0.0
	var box := VBoxContainer.new()
	box.mouse_filter = Control.MOUSE_FILTER_IGNORE
	box.alignment = BoxContainer.ALIGNMENT_CENTER
	box.set_anchors_preset(Control.PRESET_CENTER_TOP)
	box.anchor_left = 0.0
	box.anchor_right = 1.0
	box.offset_top = 150
	box.offset_bottom = 290
	add_child(box)
	_title = _make_label(46, 4)
	_epithet = _make_label(22, 3)
	box.add_child(_title)
	box.add_child(_epithet)


func _make_label(font_size: int, outline: int) -> Label:
	var label := Label.new()
	label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	label.add_theme_font_size_override("font_size", font_size)
	label.add_theme_constant_override("outline_size", outline)
	label.add_theme_color_override("font_outline_color", Color(0, 0, 0, 0.9))
	return label


## Returns false (and does nothing) when the boss has no presentation entry.
func play(template_id: String) -> bool:
	var cfg := LoaderScript.entry(template_id)
	if cfg.is_empty():
		return false
	var timing: Dictionary = cfg.get("banner", {})
	_title.text = str(cfg.get("display_name", template_id))
	_epithet.text = str(cfg.get("epithet", ""))
	_title.add_theme_color_override("font_color", Color(str(timing.get("title_color", "#ffffff"))))
	_epithet.add_theme_color_override("font_color", Color(str(timing.get("epithet_color", "#cccccc"))))
	played_count += 1
	last_template_id = template_id
	last_title = _title.text
	last_epithet = _epithet.text
	banner_visible = true
	visible = true
	modulate.a = 0.0
	if _tween != null and _tween.is_valid():
		_tween.kill()
	var fade_in := float(timing.get("fade_in_s", 0.5))
	var fade_out := float(timing.get("fade_out_s", 0.8))
	var hold := maxf(0.0, float(timing.get("duration_s", 3.0)) - fade_in - fade_out)
	if is_inside_tree():
		_tween = create_tween()
		_tween.tween_property(self, "modulate:a", 1.0, fade_in)
		_tween.tween_interval(hold)
		_tween.tween_property(self, "modulate:a", 0.0, fade_out)
		_tween.tween_callback(_finish)
	else:
		modulate.a = 1.0
	return true


func _finish() -> void:
	banner_visible = false
	visible = false


func get_debug_state() -> Dictionary:
	return {
		"visible": banner_visible,
		"played_count": played_count,
		"template_id": last_template_id,
		"title": last_title,
		"epithet": last_epithet,
		"alpha": modulate.a,
	}
