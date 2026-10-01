# Unit tests for ground-loot presentation polish.
extends SceneTree

const LootNodeFactoryScript := preload("res://scripts/loot_node_factory.gd")
const LootQuestBadgeShapesScript := preload("res://scripts/loot_quest_badge_shapes.gd")
const ClientConstantsScript := preload("res://scripts/client_constants.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_common_gold_shows_item_and_label_without_aura()
	_test_rare_equipment_shows_item_without_aura()
	_test_every_hand_item_uses_its_item_visuals_model()
	_test_gear_keeps_family_fallback_models()
	_test_kit_ground_tint_keeps_texture()
	_test_gold_amount_picks_catalog_tier()
	_test_potions_use_detail_tint()
	_test_quest_and_badge_catalog_shapes()
	print("[gdtest] PASS: test_loot_node_factory (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_common_gold_shows_item_and_label_without_aura() -> void:
	var factory = LootNodeFactoryScript.new({}, {})
	var node := factory.make_loot_node({"id": "loot_gold", "type": "loot", "item_def_id": "gold", "rarity": "common", "amount": 12})
	_assert_true("gold primitive remains", node.find_child("Box", false, false) != null)
	_assert_no_aura(node, "common gold")
	var label := node.find_child("LootLabel", false, false) as Label3D
	_assert_true("gold label exists", label != null)
	_assert_eq("gold label text", label.text if label != null else "", "12 gold")
	node.free()


func _test_rare_equipment_shows_item_without_aura() -> void:
	var factory = LootNodeFactoryScript.new({}, {
		"long_sword": {
			"ground": {"shape": "blade", "color": "#b8c7d8", "accent": "#f6e8b1", "scale": 1.0},
			"3d_model": "fallback_equipment_main_hand_v0",
		},
	})
	var node := factory.make_loot_node({"id": "loot_sword", "type": "loot", "item_def_id": "long_sword", "rarity": "rare"})
	_assert_true("rare primitive remains", node.find_child("Blade", false, false) != null)
	_assert_no_aura(node, "rare equipment")
	node.free()


func _assert_no_aura(node: Node3D, item: String) -> void:
	for name in ["RarityGlow", "SpawnPopRing", "PickupBeam", "RarityBackground"]:
		_assert_true("%s has no %s" % [item, name], node.find_child(name, true, false) == null)


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


func _test_gear_keeps_family_fallback_models() -> void:
	var factory = _real_factory()
	var templates: Dictionary = _json("shared/rules/item_templates.v0.json").get("templates", {})
	var gear_slots := ["head", "chest", "gloves", "belt", "boots", "ring", "amulet"]
	var checked_slots := {}
	for def_id in templates.keys():
		var template: Dictionary = templates[def_id]
		var slot := str(template.get("slot", ""))
		if not bool(template.get("equippable", false)) or not gear_slots.has(slot):
			continue
		if not ItemRulesLoader.item_presentations.has(def_id):
			_assert_true("%s has an item presentation" % def_id, false)
			continue
		var presentation: Dictionary = ItemRulesLoader.item_presentations[def_id]
		var model_id := str(presentation.get("3d_model", ""))
		if model_id == "":
			_assert_true("%s has a family fallback model" % def_id, false)
			continue
		if checked_slots.has(slot):
			_assert_eq("%s shares its slot fallback" % def_id, model_id, checked_slots[slot])
		_assert_true("%s fallback %s is in the asset manifest" % [def_id, model_id], factory.asset_manifest.has(model_id))
		var node: Node3D = factory.make_loot_node({"item_def_id": def_id, "rarity": "rare"})
		var model := node.find_child("GroundModel_%s" % model_id, true, false) as Node3D
		_assert_true("%s keeps family fallback %s" % [def_id, model_id], model != null)
		_assert_true("%s fallback contains a renderable mesh" % def_id, model != null and _first_mesh(model) != null)
		if model != null:
			var pose := EquipmentDisplayLoader.ground_pose_for(model_id, false)
			_assert_true("%s ground pose rests and centers its model" % def_id, bool(pose.get("rest_on_floor", false)))
			_assert_ground_pose(def_id, model, pose)
			_assert_ground_scale(def_id, model, presentation, pose)
		_assert_true("%s GLB loot omits primitive rarity tile" % def_id, node.find_child("RarityBackground", true, false) == null)
		_assert_no_aura(node, str(def_id))
		node.free()
		checked_slots[slot] = model_id
	_assert_eq("all ground gear slots are covered", checked_slots.size(), gear_slots.size())


func _assert_ground_scale(def_id: String, model: Node3D, presentation: Dictionary, pose: Dictionary) -> void:
	var ground: Dictionary = presentation.get("ground", {})
	var expected := ClientConstantsScript.GROUND_EQUIPMENT_MODEL_SCALE \
		* float(ground.get("scale", 1.0)) \
		* EquipmentDisplayLoader.ground_multiplier() \
		* float(pose.get("scale", 1.0))
	var uniform := is_equal_approx(model.scale.x, model.scale.y) and is_equal_approx(model.scale.y, model.scale.z)
	_assert_true("%s ground model scale is uniform" % def_id, uniform)
	if float(pose.get("max_extent", 0.0)) > 0.0:
		_assert_true("%s ground pose cap does not enlarge its data scale" % def_id, model.scale.x <= expected + 0.001)
	else:
		_assert_true("%s ground model scale follows catalog data" % def_id, is_equal_approx(model.scale.x, expected))


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


# --- v490 kit coins and potions (catalog-derived) ---

func _test_gold_amount_picks_catalog_tier() -> void:
	var factory = _real_factory()
	var tiers: Array = ItemRulesLoader.item_presentations["gold"]["ground_model_tiers"]
	for i in range(tiers.size()):
		var tier: Dictionary = tiers[i]
		var amount := int(tier["min_amount"])
		var node: Node3D = factory.make_loot_node({"item_def_id": "gold", "rarity": "common", "amount": amount})
		_assert_true("gold x%d uses tier %s" % [amount, tier["asset_id"]], node.find_child("GroundModel_%s" % tier["asset_id"], true, false) != null)
		node.free()
		if i + 1 < tiers.size() and int(tiers[i + 1]["min_amount"]) - 1 > amount:
			var below_next: Node3D = factory.make_loot_node({"item_def_id": "gold", "rarity": "common", "amount": int(tiers[i + 1]["min_amount"]) - 1})
			_assert_true("gold just below the next tier stays on %s" % tier["asset_id"], below_next.find_child("GroundModel_%s" % tier["asset_id"], true, false) != null)
			below_next.free()
	var none: Node3D = factory.make_loot_node({"item_def_id": "gold", "rarity": "common", "amount": int(tiers[0]["min_amount"]) - 1})
	_assert_true("gold below the first tier keeps the primitive", none.find_child("GroundModel_*", true, false) == null)
	none.free()


func _test_potions_use_detail_tint() -> void:
	var factory = _real_factory()
	var checked := 0
	for def_id in ItemRulesLoader.item_presentations.keys():
		var presentation: Dictionary = ItemRulesLoader.item_presentations[def_id]
		if typeof(presentation.get("ground_tint", null)) != TYPE_DICTIONARY:
			continue
		var node: Node3D = factory.make_loot_node({"item_def_id": def_id, "rarity": "common"})
		var model := node.find_child("GroundModel_%s" % presentation["3d_model"], true, false)
		var mesh := _first_mesh(model) if model != null else null
		var mat := mesh.material_override as StandardMaterial3D if mesh != null else null
		var tint: Dictionary = presentation["ground_tint"]
		_assert_true("%s uses its kit bottle" % def_id, model != null)
		_assert_true("%s detail tint enabled with catalog colour" % def_id, mat != null and mat.detail_enabled and mat.detail_albedo != null \
			and _color_within_8bit(mat.detail_albedo.get_image().get_pixel(0, 0), Color(Color(str(tint["color"])), float(tint["strength"]))))
		_assert_true("%s keeps its albedo texture" % def_id, mat != null and mat.albedo_texture != null)
		node.free()
		checked += 1
	_assert_true("potion families with ground_tint were checked", checked > 0)


func _test_quest_and_badge_catalog_shapes() -> void:
	var factory = _real_factory()
	var items: Dictionary = _json("shared/assets/item_presentations.v0.json").get("items", {})
	var item_defs: Dictionary = _json("shared/rules/items.v0.json").get("items", {})
	var trophy_shapes := {}
	var trophy_count := 0
	var checked := 0
	for def_id in items.keys():
		var family := str((items[def_id] as Dictionary).get("family", ""))
		if not (family in ["quest", "badge"]):
			continue
		checked += 1
		_assert_true("%s has an authoritative item definition" % def_id, item_defs.has(def_id))
		var presentation: Dictionary = ItemRulesLoader.item_presentations.get(def_id, {})
		var ground: Dictionary = presentation.get("ground", {})
		var shape := str(ground.get("shape", ""))
		_assert_true("%s selects a known quest/badge shape" % def_id, LootQuestBadgeShapesScript.supports(shape))
		var node: Node3D = factory.make_loot_node({"item_def_id": def_id, "rarity": "rare"})
		var model := node.find_child("QuestBadgeShape_%s" % shape, false, false) as Node3D
		_assert_true("%s builds its catalog shape %s" % [def_id, shape], model != null)
		_assert_true("%s has renderable geometry" % def_id, model != null and _first_mesh(model) != null)
		if model != null:
			var bounds: AABB = LootNodeFactoryScript.posed_bounds(model)
			_assert_true("%s has nonzero visual bounds" % def_id, bounds.size.x > 0.0 and bounds.size.y > 0.0 and bounds.size.z > 0.0)
			_assert_true("%s rests above the floor" % def_id, bounds.position.y >= 0.0)
			var label := node.find_child("LootLabel", false, false) as Label3D
			_assert_true("%s model clears its label" % def_id, label != null and bounds.end.y < label.position.y)
		_assert_no_aura(node, str(def_id))
		_assert_true("%s uses no unrelated GLB" % def_id, node.find_child("GroundModel_*", true, false) == null)
		if str(def_id).begins_with("quest_trophy_"):
			trophy_count += 1
			trophy_shapes[shape] = true
		node.free()
	_assert_true("quest and badge catalog items were checked", checked > 0)
	_assert_true("quest trophies have distinct silhouettes", trophy_count > 1 and trophy_shapes.size() == trophy_count)
	_assert_true("unknown shape stays on generic fallback", not LootQuestBadgeShapesScript.supports("box"))
	var leaf_node: Node3D = factory.make_loot_node({"item_def_id": "quest_leaf", "rarity": "common"})
	var badge_node: Node3D = factory.make_loot_node({"item_def_id": "respec_badge", "rarity": "common"})
	var leaf := leaf_node.find_child("QuestBadgeShape_leaf", false, false) as Node3D
	var badge := badge_node.find_child("QuestBadgeShape_badge", false, false) as Node3D
	if leaf != null and badge != null:
		var leaf_size: Vector3 = LootNodeFactoryScript.posed_bounds(leaf).size
		var badge_size: Vector3 = LootNodeFactoryScript.posed_bounds(badge).size
		_assert_true("leaf is elongated, badge is round", leaf_size.x > leaf_size.z * 1.3 and absf(badge_size.x - badge_size.z) < badge_size.x * 0.2)
	else:
		_assert_true("representative quest and badge shapes exist", false)
	leaf_node.free()
	badge_node.free()


## The detail texture is RGBA8, so channels round to 1/255.
func _color_within_8bit(got: Color, want: Color) -> bool:
	var eps := 1.0 / 255.0
	return absf(got.r - want.r) <= eps and absf(got.g - want.g) <= eps and absf(got.b - want.b) <= eps and absf(got.a - want.a) <= eps


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
