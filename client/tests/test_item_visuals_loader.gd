extends SceneTree

## v487: ItemVisualsLoader is the single reader of item_visuals.v0.json, and
## EquipmentDisplayLoader.ground_pose_for layers the ground pose from equipment_display.v0.json.
## Expectations derive from the shared JSON, not from pinned asset ids or tuning values.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_loader_matches_catalog()
	_test_hand_asset_id_only_for_hand_slots()
	_test_ground_pose_layers()
	_finish()


func _read(rel: String) -> Dictionary:
	var path := ProjectSettings.globalize_path("res://").path_join("../" + rel)
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path))
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}


func _test_loader_matches_catalog() -> void:
	ItemVisualsLoader.invalidate()
	var catalog: Dictionary = _read("shared/assets/item_visuals.v0.json").get("item_visuals", {})
	_assert_eq("loader exposes every catalog entry", ItemVisualsLoader.all().size(), catalog.size())
	for def_id in catalog.keys():
		_assert_eq("visual_for(%s) asset" % def_id, str(ItemVisualsLoader.visual_for(def_id).get("asset_id", "")), str(catalog[def_id].get("asset_id", "")))
	_assert_eq("unknown item has no visual", ItemVisualsLoader.visual_for("not_an_item").is_empty(), true)


func _test_hand_asset_id_only_for_hand_slots() -> void:
	var catalog: Dictionary = _read("shared/assets/item_visuals.v0.json").get("item_visuals", {})
	var hand_seen := 0
	for def_id in catalog.keys():
		var entry: Dictionary = catalog[def_id]
		var is_hand := str(entry.get("slot", "")) in ItemVisualsLoader.HAND_SLOTS
		var want := str(entry.get("asset_id", "")) if is_hand else ""
		_assert_eq("hand_asset_id(%s)" % def_id, ItemVisualsLoader.hand_asset_id(def_id), want)
		if is_hand:
			hand_seen += 1
	_assert_eq("catalog has hand items", hand_seen > 0, true)


func _test_ground_pose_layers() -> void:
	EquipmentDisplayLoader.invalidate()
	var pose_data: Dictionary = _read("shared/assets/equipment_display.v0.json").get("ground_pose", {})
	var default_pose: Dictionary = pose_data.get("default", {})
	var kit_pose: Dictionary = pose_data.get("rig_native", {})
	var legacy := EquipmentDisplayLoader.ground_pose_for("some_legacy_asset", false)
	_assert_eq("legacy height from default", legacy["height"], float(default_pose.get("height", 0.0)))
	_assert_eq("legacy rotation from default", legacy["rotation_degrees"], _vec(default_pose.get("rotation_degrees", {})))
	var kit := EquipmentDisplayLoader.ground_pose_for("some_kit_asset", true)
	_assert_eq("kit height from rig_native", kit["height"], float(kit_pose.get("height", default_pose.get("height", 0.0))))
	_assert_eq("kit scale from rig_native", kit["scale"], float(kit_pose.get("scale", default_pose.get("scale", 1.0))))
	var overrides: Dictionary = pose_data.get("assets", {})
	for asset_id in overrides.keys():
		var entry: Dictionary = overrides[asset_id]
		if entry.has("scale"):
			_assert_eq("asset override %s scale" % asset_id, EquipmentDisplayLoader.ground_pose_for(asset_id, true)["scale"], float(entry["scale"]))


func _vec(d: Dictionary) -> Vector3:
	return Vector3(float(d.get("x", 0.0)), float(d.get("y", 0.0)), float(d.get("z", 0.0)))


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_item_visuals_loader (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_item_visuals_loader (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
