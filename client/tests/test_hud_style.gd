# v517: HudStyle resolves every color/frame from UiTheme hud_* tokens (configurability proof).
extends SceneTree

var _passed := 0
var _failed := 0
var _catalog_path := ""


func _initialize() -> void:
	_catalog_path = ProjectSettings.globalize_path("res://").path_join("../shared/assets/ui_theme.v0.json")
	UiTheme.invalidate()
	_test_every_hud_token_present()
	_test_temp_token_changes_styles()
	UiTheme.invalidate()
	print("[gdtest] PASS: test_hud_style (%d passed, %d failed)" % [_passed, _failed])
	quit(1 if _failed else 0)


func _check(ok: bool, label: String) -> void:
	if ok:
		_passed += 1
	else:
		_failed += 1
		push_error("FAIL: " + label)


func _read() -> Dictionary:
	return JSON.parse_string(FileAccess.get_file_as_string(_catalog_path))


func _test_every_hud_token_present() -> void:
	var catalog := _read()
	var script_text := FileAccess.get_file_as_string("res://scripts/hud_style.gd")
	var rx := RegEx.new()
	rx.compile("UiTheme\\.(?:color|spacing|frame)\\(\"(hud_[a-z0-9_]+)\"")
	var seen := 0
	for m in rx.search_all(script_text):
		var token: String = m.get_string(1)
		seen += 1
		var found: bool = catalog["colors"].has(token) or catalog["spacing"].has(token) or catalog["frames"].has(token)
		_check(found, "hud_style references existing token %s" % token)
	_check(seen > 20, "hud_style references many tokens")
	for key in catalog["colors"]:
		if str(key).begins_with("hud_"):
			_check(typeof(catalog["colors"][key]) in [TYPE_STRING, TYPE_ARRAY], "%s is a color value" % key)


func _test_temp_token_changes_styles() -> void:
	var catalog := _read()
	catalog["colors"]["hud_gold"] = "#123456"
	catalog["colors"]["hud_hp_high"] = "#654321"
	catalog["colors"]["hud_globe_ring"] = "#abcdef"
	catalog["spacing"]["hud_frame_border"] = 7
	var temp := "user://v517_temp_ui_theme.json"
	var file := FileAccess.open(temp, FileAccess.WRITE)
	file.store_string(JSON.stringify(catalog))
	file.close()
	_check(UiTheme.load_from_path(ProjectSettings.globalize_path(temp)), "temp catalog loads")
	_check(HudStyle.xp_fill().bg_color == Color("#123456"), "xp fill follows hud_gold token")
	_check(HudStyle.hp_color(1.0) == Color("#654321"), "hp color follows hud_hp_high token")
	_check(HudStyle.globe_ring() == Color("#abcdef"), "globe ring color follows token")
	_check(HudStyle.frame_panel().border_width_left == 7, "panel border follows hud_frame_border spacing")
	_check(HudStyle.minimap_frame(0.5).border_width_top != 7, "minimap border uses its own token")
	_check(is_equal_approx(HudStyle.minimap_frame(0.5).bg_color.a, 0.5), "minimap opacity override preserved")
	DirAccess.remove_absolute(ProjectSettings.globalize_path(temp))
