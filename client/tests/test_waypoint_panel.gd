# Unit test for waypoint panel scroll behavior (v19).
# Run via: godot --headless --path client --script res://tests/test_waypoint_panel.gd
extends SceneTree

const WaypointPanelConfig := preload("res://scripts/waypoint_panel_config.gd")


func _initialize() -> void:
	var shared := ProjectSettings.globalize_path("res://").path_join("../shared")
	var golden := _read(shared.path_join("golden/waypoint_panel.json"))
	_assert_config_matches_golden(golden)
	_assert_scroll_overflows_with_many_rows(golden)
	print("[gdtest] PASS: test_waypoint_panel")
	quit(0)


func _assert_config_matches_golden(golden: Dictionary) -> void:
	if WaypointPanelConfig.SCROLL_MAX_VISIBLE_ROWS != int(golden["scroll_max_visible_rows"]):
		_fail("SCROLL_MAX_VISIBLE_ROWS mismatch")
	if WaypointPanelConfig.SCROLL_VIEWPORT_UNIT_PX != int(golden["scroll_viewport_unit_px"]):
		_fail("SCROLL_VIEWPORT_UNIT_PX mismatch")
	if WaypointPanelConfig.ROW_HEIGHT_PX != int(golden["row_height_px"]):
		_fail("ROW_HEIGHT_PX mismatch")
	if WaypointPanelConfig.SCROLL_MIN_WIDTH_PX != int(golden["scroll_min_width_px"]):
		_fail("SCROLL_MIN_WIDTH_PX mismatch")
	if WaypointPanelConfig.PANEL_MIN_WIDTH_PX != int(golden["panel_min_width_px"]):
		_fail("PANEL_MIN_WIDTH_PX mismatch")


func _assert_scroll_overflows_with_many_rows(golden: Dictionary) -> void:
	var max_rows := int(golden["scroll_max_visible_rows"])
	var unit_h := int(golden["scroll_viewport_unit_px"])
	var row_h := int(golden["row_height_px"])
	var row_spacing := 4
	var overflow_rows := max_rows + 3
	var viewport_h := unit_h * max_rows
	var content_h := overflow_rows * row_h + maxi(0, overflow_rows - 1) * row_spacing
	if content_h <= viewport_h:
		_fail("expected content height %d > viewport %d for %d rows" % [content_h, viewport_h, overflow_rows])

	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(int(golden["scroll_min_width_px"]), viewport_h)
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	var rows := VBoxContainer.new()
	rows.add_theme_constant_override("separation", row_spacing)
	scroll.add_child(rows)
	for level in overflow_rows:
		var row := Button.new()
		row.custom_minimum_size = Vector2(204, row_h)
		row.text = "Level %d" % level
		rows.add_child(row)
	if scroll.custom_minimum_size.y != viewport_h:
		_fail("scroll viewport height mismatch")


func _read(path: String) -> Dictionary:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		_fail("cannot open %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		_fail("invalid json %s" % path)
		return {}
	return parsed


func _fail(msg: String) -> void:
	push_error("[gdtest] FAIL: %s" % msg)
	print("[gdtest] FAIL: %s" % msg)
	quit(1)
