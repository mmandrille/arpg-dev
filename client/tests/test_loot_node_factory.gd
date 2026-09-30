# Unit tests for ground-loot presentation polish.
extends SceneTree

const LootNodeFactoryScript := preload("res://scripts/loot_node_factory.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_common_gold_has_glow_marker_and_label()
	_test_rare_equipment_keeps_model_and_rarity_glow()
	_test_rare_equipment_has_pickup_beam()
	_test_every_hand_item_uses_its_item_visuals_model()
	_test_armor_keeps_family_fallback_model()
	_test_kit_ground_tint_keeps_texture()
	print("[gdtest] PASS: test_loot_node_factory (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_common_gold_has_glow_marker_and_label() -> void:
	var factory = LootNodeFactoryScript.new({}, {})
	var node := factory.make_loot_node({"id": "loot_gold", "type": "loot", "item_def_id": "gold", "rarity": "common", "amount": 12})
	_assert_true("common loot glow exists", node.find_child("RarityGlow", false, false) != null)
	_assert_true("common spawn pop exists", node.find_child("SpawnPopRing", false, false) != null)
	var label := node.find_child("LootLabel", false, false) as Label3D
	_assert_true("gold label exists", label != null)
	_assert_eq("gold label text", label.text if label != null else "", "12 gold")
	node.free()


func _test_rare_equipment_keeps_model_and_rarity_glow() -> void:
	var factory = LootNodeFactoryScript.new({}, {
		"long_sword": {
			"ground": {"shape": "blade", "color": "#b8c7d8", "accent": "#f6e8b1", "scale": 1.0},
			"3d_model": "fallback_equipment_main_hand_v0",
		},
	})
	var node := factory.make_loot_node({"id": "loot_sword", "type": "loot", "item_def_id": "long_sword", "rarity": "rare"})
	_assert_true("rare loot glow exists", node.find_child("RarityGlow", false, false) != null)
	_assert_true("rare spawn pop exists", node.find_child("SpawnPopRing", false, false) != null)
	_assert_true("rare primitive remains", node.find_child("Blade", false, false) != null)
	var glow := node.find_child("RarityGlow", false, false) as MeshInstance3D
	var mat := glow.material_override as StandardMaterial3D
	_assert_true("rare glow uses warm rarity color", mat != null and mat.albedo_color.r > mat.albedo_color.b)
	node.free()


func _test_rare_equipment_has_pickup_beam() -> void:
	var factory = LootNodeFactoryScript.new({}, {
		"long_sword": {
			"ground": {"shape": "blade", "color": "#b8c7d8", "accent": "#f6e8b1", "scale": 1.0},
			"3d_model": "fallback_equipment_main_hand_v0",
		},
	})
	var node := factory.make_loot_node({"id": "loot_sword", "type": "loot", "item_def_id": "long_sword", "rarity": "rare"})
	_assert_true("rare pickup beam exists", node.find_child("PickupBeam", false, false) != null)
	var common := factory.make_loot_node({"id": "loot_gold", "type": "loot", "item_def_id": "gold", "rarity": "common", "amount": 3})
	_assert_true("common loot has no pickup beam", common.find_child("PickupBeam", false, false) == null)
	node.free()
	common.free()


# --- v487 kit ground loot (catalog-derived; no pinned asset ids or tuning values) ---

func _json(rel: String) -> Dictionary:
	var path := ProjectSettings.globalize_path("res://").path_join("../" + rel)
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path))
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}


func _real_factory():
	ItemRulesLoader.ensure_loaded()
	return LootNodeFactoryScript.new(_json("assets/manifests/assets.v0.json").get("assets", {}), ItemRulesLoader.item_presentations)


func _test_every_hand_item_uses_its_item_visuals_model() -> void:
	var factory = _real_factory()
	var visuals: Dictionary = _json("shared/assets/item_visuals.v0.json").get("item_visuals", {})
	var checked := 0
	for def_id in visuals.keys():
		var visual: Dictionary = visuals[def_id]
		if not (str(visual.get("slot", "")) in ["main_hand", "off_hand"]) or not ItemRulesLoader.item_presentations.has(def_id):
			continue
		var node: Node3D = factory.make_loot_node({"item_def_id": def_id, "rarity": "magic"})
		var asset_id := str(visual.get("asset_id", ""))
		var model := node.find_child("GroundModel_%s" % asset_id, true, false) as Node3D
		_assert_true("%s ground model is its item_visuals asset %s" % [def_id, asset_id], model != null)
		if model != null:
			_assert_ground_pose(def_id, model, EquipmentDisplayLoader.ground_pose_for(asset_id, bool(visual.get("rig_native", false))))
		node.free()
		checked += 1
	_assert_true("hand items were checked", checked > 0)


## The ground pose contract: rotation from data; with rest_on_floor the posed bounds sit on the
## floor at `height`, centred on the loot root; otherwise the origin sits at `height`;
## max_extent caps the longest side.
func _assert_ground_pose(def_id: String, model: Node3D, pose: Dictionary) -> void:
	_assert_true("%s ground rotation from ground_pose" % def_id, model.rotation_degrees.is_equal_approx(pose["rotation_degrees"]))
	var bounds: AABB = LootNodeFactoryScript.posed_bounds(model)
	if bool(pose["rest_on_floor"]):
		_assert_true("%s rests on the floor at height" % def_id, absf(bounds.position.y - float(pose["height"])) < 0.001)
		_assert_true("%s is centred on the loot root" % def_id, absf(bounds.get_center().x) < 0.001 and absf(bounds.get_center().z) < 0.001)
	else:
		_assert_true("%s origin at height" % def_id, is_equal_approx(model.position.y, float(pose["height"])))
	var max_extent := float(pose["max_extent"])
	if max_extent > 0.0:
		var longest := maxf(bounds.size.x, maxf(bounds.size.y, bounds.size.z))
		_assert_true("%s longest side %.3f within max_extent %.3f" % [def_id, longest, max_extent], longest <= max_extent + 0.001)


func _test_armor_keeps_family_fallback_model() -> void:
	var factory = _real_factory()
	for def_id in ItemRulesLoader.item_presentations.keys():
		var model_id := str((ItemRulesLoader.item_presentations[def_id] as Dictionary).get("3d_model", ""))
		if model_id == "" or ItemVisualsLoader.hand_asset_id(def_id) != "":
			continue
		var node: Node3D = factory.make_loot_node({"item_def_id": def_id, "rarity": "rare"})
		_assert_true("%s keeps family fallback %s" % [def_id, model_id], node.find_child("GroundModel_%s" % model_id, true, false) != null)
		node.free()
		return
	_assert_true("an armor item with a family fallback exists", false)


func _first_mesh(root: Node) -> MeshInstance3D:
	if root is MeshInstance3D:
		return root as MeshInstance3D
	for child in root.get_children():
		var found := _first_mesh(child)
		if found != null:
			return found
	return null


func _test_kit_ground_tint_keeps_texture() -> void:
	var factory = _real_factory()
	var visuals: Dictionary = _json("shared/assets/item_visuals.v0.json").get("item_visuals", {})
	for def_id in visuals.keys():
		if not bool((visuals[def_id] as Dictionary).get("rig_native", false)) or ItemVisualsLoader.hand_asset_id(def_id) == "":
			continue
		var node: Node3D = factory.make_loot_node({"item_def_id": def_id, "rarity": "rare"})
		var mesh := _first_mesh(node.find_child("GroundModel_*", true, false))
		_assert_true("%s kit ground model has a mesh" % def_id, mesh != null)
		if mesh != null:
			var source := mesh.mesh.surface_get_material(0) as StandardMaterial3D if mesh.mesh != null else null
			var tinted := mesh.material_override as StandardMaterial3D
			var want := Color.WHITE.lerp(factory.ground_item_tint("rare"), EquipmentDisplayLoader.rig_native_tint_strength())
			_assert_true("%s tint keeps the kit albedo texture" % def_id, tinted != null and source != null and tinted.albedo_texture == source.albedo_texture and tinted.albedo_texture != null)
			_assert_true("%s tint blends at rig_native strength" % def_id, tinted != null and tinted.albedo_color.is_equal_approx(want))
		node.free()
		return
	_assert_true("a rig-native hand item exists", false)


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
