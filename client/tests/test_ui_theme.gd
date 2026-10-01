## v514: shared UI theme loader and style builders derive from shared/assets/ui_theme.v0.json.
extends SceneTree

var _passed := 0
var _failed := 0


func _initialize() -> void:
	UiTheme.invalidate()
	var catalog := _read_catalog()
	_test_frames_follow_catalog(catalog)
	_test_state_overrides(catalog)
	_test_rarity_slots(catalog)
	_test_fonts(catalog)
	_test_independent_instances()
	_test_facades_use_theme()
	_test_fallbacks_and_reload()
	print("[gdtest] PASS: test_ui_theme (%d passed, %d failed)" % [_passed, _failed])
	quit(1 if _failed else 0)


func _check(ok: bool, label: String) -> void:
	if ok:
		_passed += 1
	else:
		_failed += 1
		push_error("FAIL: " + label)


func _read_catalog() -> Dictionary:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/ui_theme.v0.json")
	return JSON.parse_string(FileAccess.get_file_as_string(path))


func _test_frames_follow_catalog(catalog: Dictionary) -> void:
	for name in catalog["frames"]:
		var s := UiTheme.frame(name)
		var recipe: Dictionary = catalog["frames"][name]
		_check(s.bg_color == UiTheme.color(recipe["bg"]), "%s bg resolves via color token" % name)
		_check(s.border_color == UiTheme.color(recipe["border"]), "%s border resolves via color token" % name)
		var width = recipe["border_width"]
		var expected_left: int = UiTheme.spacing(width[0] if width is Array else width)
		_check(s.border_width_left == expected_left, "%s border width from spacing" % name)


func _test_state_overrides(catalog: Dictionary) -> void:
	for name in catalog["frames"]:
		var hover: Dictionary = catalog["frames"][name].get("states", {}).get("hover", {})
		if hover.is_empty():
			continue
		var s := UiTheme.frame(name, "hover")
		_check(s.bg_color == UiTheme.color(hover["bg"]), "%s hover bg" % name)
		_check(s.border_color == UiTheme.color(hover["border"]), "%s hover border" % name)
		_check(UiTheme.frame(name).bg_color != s.bg_color, "%s hover differs from base" % name)


func _test_rarity_slots(catalog: Dictionary) -> void:
	var mods: Dictionary = catalog["state_modifiers"]
	for rarity in catalog["rarity_slot_backgrounds"]:
		var base := UiTheme.rarity_background(rarity)
		_check(UiTheme.item_slot_frame(rarity).bg_color == base, "%s base bg" % rarity)
		_check(UiTheme.item_slot_frame(rarity, true).bg_color == base.lightened(float(mods["hover_lighten"])), "%s hover bg" % rarity)
		var dim := UiTheme.item_slot_frame(rarity, false, true)
		_check(dim.bg_color == base.darkened(float(mods["invalid_darken"])), "%s dimmed bg" % rarity)
		_check(dim.border_color == UiTheme.color(mods["invalid_border"]), "%s dimmed border" % rarity)
	_check(UiTheme.rarity_background("COMMON") == UiTheme.rarity_background("common"), "rarity lookup is case-insensitive")
	_check(UiTheme.rarity_background("not_a_rarity") == UiTheme.rarity_background("common"), "unknown rarity falls back to common")
	_check(UiTheme.rarity_background("") == UiTheme.rarity_background("common"), "empty rarity falls back to common")


func _test_fonts(catalog: Dictionary) -> void:
	for role in catalog["fonts"]:
		var font: Dictionary = catalog["fonts"][role]
		_check(UiTheme.font_size(role) == int(font["size"]), "%s font size" % role)
		_check(UiTheme.font_color(role) == UiTheme.color(font["color"]), "%s font color" % role)
	var label := Label.new()
	var role: String = catalog["fonts"].keys()[0]
	UiTheme.apply_font(label, role)
	_check(label.get_theme_font_size("font_size") == UiTheme.font_size(role), "apply_font sets size override")
	label.free()


func _test_independent_instances() -> void:
	var a := UiTheme.frame("slot")
	a.bg_color = Color.RED
	_check(UiTheme.frame("slot").bg_color != Color.RED, "frame() returns independent instances")
	var snapshot := UiTheme.catalog()
	snapshot["colors"].clear()
	_check(not UiTheme.catalog()["colors"].is_empty(), "catalog() does not expose mutable state")


func _test_facades_use_theme() -> void:
	_check(InventoryPanelStyles.panel_style().bg_color == UiTheme.frame("panel").bg_color, "inventory facade panel")
	_check(InventoryPanelStyles.slot_style(true).border_color == UiTheme.frame("slot", "hover").border_color, "inventory facade slot hover")
	_check(InventoryPanelStyles.item_slot_style("rare", false, true).bg_color == UiTheme.item_slot_frame("rare", false, true).bg_color, "inventory facade item slot")
	_check(CharacterPanelStyles.panel_style().bg_color == UiTheme.frame("panel_character").bg_color, "character facade panel")


func _test_fallbacks_and_reload() -> void:
	_check(UiTheme.has_token("color", "frame_border"), "has_token finds shipped color")
	_check(not UiTheme.has_token("color", "no_such_token"), "has_token rejects unknown")
	_check(UiTheme.color("no_such_token") == UiTheme.FALLBACK_COLOR, "unknown color returns magenta fallback")
	_check(UiTheme.spacing("no_such_token") == 0, "unknown spacing returns 0")
	_check(UiTheme.frame("no_such_frame").bg_color == UiTheme.FALLBACK_COLOR, "unknown frame returns fallback box")
	_check(UiTheme.font_size("no_such_role") == 0, "unknown font role returns 0")
	_check(not UiTheme.load_from_path("/nonexistent/ui_theme.json"), "missing catalog file does not crash")
	_check(UiTheme.frame("panel") != null, "frame() still returns a box with a failed load")
	UiTheme.invalidate()
	_check(UiTheme.has_token("frame", "panel"), "invalidate() reloads the shipped catalog")
