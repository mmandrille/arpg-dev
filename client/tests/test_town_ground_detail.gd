extends SceneTree

## v493 town ground detail: edge/rim/paths/scatter planning derived from
## town_presentation.v0.json -> dressing. No pinned coordinates or counts.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_loader_keeps_ground_detail_keys()
	_finish()


func _test_loader_keeps_ground_detail_keys() -> void:
	var dressing := TownPresentationLoader.dressing()
	for key in ["anchors", "service_paths", "edge", "scatter"]:
		_assert_true("dressing keeps %s (v491 merge-drop regression)" % key, dressing.has(key))
	_assert_true("plaza keeps rim", (dressing.get("plaza", {}) as Dictionary).has("rim"))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s" % label)


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_town_ground_detail (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_town_ground_detail (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
