# Unit test for HudGlobe (v517): liquid geometry, clamping, redraw-on-change only, no per-frame process.
extends SceneTree

const HudGlobeScript := preload("res://scripts/hud_globe.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_liquid_polygon()
	await _test_state_and_redraw()
	print("[gdtest] PASS: test_hud_globe (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_liquid_polygon() -> void:
	var c := Vector2(50, 50)
	var r := 40.0
	_assert_eq("empty globe has no liquid", HudGlobeScript.liquid_polygon(c, r, 0.0).size(), 0)
	_assert_eq("negative ratio has no liquid", HudGlobeScript.liquid_polygon(c, r, -2.0).size(), 0)
	_assert_eq("zero radius has no liquid", HudGlobeScript.liquid_polygon(c, 0.0, 0.5).size(), 0)
	var full := HudGlobeScript.liquid_polygon(c, r, 1.0)
	_assert_true("full globe polygon exists", full.size() >= 3)
	for p in full:
		_assert_true("full polygon point on circle", absf(p.distance_to(c) - r) < 0.01)
	var half := HudGlobeScript.liquid_polygon(c, r, 0.5)
	var min_y := INF
	var max_y := -INF
	for p in half:
		min_y = minf(min_y, p.y)
		max_y = maxf(max_y, p.y)
	_assert_true("half level at center line", absf(min_y - c.y) < 0.01)
	_assert_true("half bottom at circle bottom", absf(max_y - (c.y + r)) < 0.01)
	var quarter := HudGlobeScript.liquid_polygon(c, r, 0.25)
	min_y = INF
	for p in quarter:
		min_y = minf(min_y, p.y)
	_assert_true("quarter level below center", absf(min_y - (c.y + r * 0.5)) < 0.01)
	var over := HudGlobeScript.liquid_polygon(c, r, 7.0)
	_assert_eq("ratio above one clamps to full", over.size(), full.size())


func _test_state_and_redraw() -> void:
	var globe = HudGlobeScript.new()
	globe.size = Vector2(100, 100)
	get_root().add_child(globe)
	await process_frame
	_assert_true("globe does not process per frame", not globe.is_processing())
	globe.set_ratio(0.5)
	globe.set_label("60 / 120")
	var requests: int = globe.redraw_requests
	globe.set_ratio(0.5)
	globe.set_label("60 / 120")
	globe.fill_color = globe.fill_color
	_assert_eq("unchanged values do not redraw", globe.redraw_requests, requests)
	globe.set_ratio(0.4)
	_assert_eq("changed ratio redraws once", globe.redraw_requests, requests + 1)
	globe.set_ratio(3.0)
	_assert_eq("ratio clamps high", float(globe.get_debug_state().get("ratio", -1.0)), 1.0)
	globe.set_ratio(-1.0)
	_assert_eq("ratio clamps low", float(globe.get_debug_state().get("ratio", -1.0)), 0.0)
	globe.flash(Color.RED, Color.GREEN)
	_assert_true("flash starts", globe.is_flashing())
	_assert_eq("flash sets color immediately", globe.fill_color, Color.RED)
	globe.free()


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
