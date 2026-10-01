class_name LootNodeFactory
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const EquipmentDisplayLoaderScript := preload("res://scripts/equipment_display_loader.gd")
const PotionIconLabelScript := preload("res://scripts/potion_icon_label.gd")
const ModelTintScript := preload("res://scripts/model_tint.gd")
const LootQuestBadgeShapesScript := preload("res://scripts/loot_quest_badge_shapes.gd")
const RarityCueLoaderScript := preload("res://scripts/rarity_cue_loader.gd")
const RarityCuePresenterScript := preload("res://scripts/rarity_cue_presenter.gd")

var asset_manifest: Dictionary = {}
var item_presentations: Dictionary = {}

func _init(manifest: Dictionary = {}, presentations: Dictionary = {}) -> void:
	asset_manifest = manifest
	item_presentations = presentations

func configure(manifest: Dictionary, presentations: Dictionary) -> void:
	asset_manifest = manifest
	item_presentations = presentations

func make_loot_node(e: Dictionary) -> Node3D:
	var item_def_id := str(e.get("item_def_id", ""))
	var root := Node3D.new()
	root.name = "Loot_%s" % item_def_id
	var ground: Dictionary = item_presentations.get(item_def_id, {}).get("ground", {})
	var shape := str(ground.get("shape", "box"))
	var color := Color(str(ground.get("color", "#" + loot_color(item_def_id).to_html(false))))
	var accent := Color(str(ground.get("accent", "#f6e8b1")))
	var scale := float(ground.get("scale", 1.0))
	var model := make_ground_equipment_model(item_def_id, str(e.get("rarity", "common")), int(e.get("amount", 0)))
	if model != null:
		root.add_child(model)
	else:
		if LootQuestBadgeShapesScript.supports(shape):
			LootQuestBadgeShapesScript.add_shape(root, shape, color, accent, scale)
		else:
			add_loot_primitive(root, shape, color, accent, scale)
	add_loot_label(root, loot_label_text(e), scale, loot_label_color(e), not RarityCueLoaderScript.cue_for_item(e).is_empty())
	RarityCuePresenterScript.add_world_marker(root, e, scale)
	return root

func add_loot_primitive(root: Node3D, shape: String, color: Color, accent: Color, scale: float) -> void:
	match shape:
		"blade":
			add_loot_box(root, "Blade", Vector3(0.12, 0.08, 0.78) * scale, Vector3(0.0, 0.20, 0.0), color)
			add_loot_box(root, "Grip", Vector3(0.34, 0.10, 0.10) * scale, Vector3(0.0, 0.16, 0.34 * scale), accent)
		"greatsword":
			add_loot_box(root, "GreatBlade", Vector3(0.14, 0.08, 1.02) * scale, Vector3(0.0, 0.24, 0.0), color)
			add_loot_box(root, "GreatGuard", Vector3(0.52, 0.08, 0.10) * scale, Vector3(0.0, 0.18, 0.38 * scale), accent)
			add_loot_box(root, "GreatGrip", Vector3(0.12, 0.10, 0.34) * scale, Vector3(0.0, 0.16, 0.62 * scale), accent)
		"staff":
			add_loot_box(root, "StaffShaft", Vector3(0.08, 0.08, 0.92) * scale, Vector3(0.0, 0.22, 0.0), color)
			add_loot_cylinder(root, "StaffOrb", 0.14 * scale, 0.12 * scale, Vector3(0.0, 0.58 * scale, 0.0), accent)
		"bow":
			add_loot_box(root, "BowTop", Vector3(0.10, 0.08, 0.42) * scale, Vector3(0.14 * scale, 0.20, -0.18 * scale), color)
			add_loot_box(root, "BowBottom", Vector3(0.10, 0.08, 0.42) * scale, Vector3(-0.14 * scale, 0.20, 0.18 * scale), color)
			add_loot_box(root, "String", Vector3(0.04, 0.06, 0.75) * scale, Vector3(0.0, 0.18, 0.0), accent)
		"shield":
			add_loot_cylinder(root, "ShieldFace", 0.30 * scale, 0.08 * scale, Vector3(0.0, 0.18, 0.0), color)
			add_loot_cylinder(root, "ShieldBoss", 0.11 * scale, 0.10 * scale, Vector3(0.0, 0.24, 0.0), accent)
		"helm":
			add_loot_cylinder(root, "HelmCap", 0.25 * scale, 0.22 * scale, Vector3(0.0, 0.22, 0.0), color)
			add_loot_box(root, "HelmBrow", Vector3(0.44, 0.08, 0.18) * scale, Vector3(0.0, 0.26, -0.12 * scale), accent)
		"chest":
			add_loot_box(root, "ChestPlate", Vector3(0.46, 0.16, 0.40) * scale, Vector3(0.0, 0.18, 0.0), color)
			add_loot_box(root, "ChestTrim", Vector3(0.34, 0.18, 0.08) * scale, Vector3(0.0, 0.26, -0.16 * scale), accent)
		"gloves":
			add_loot_box(root, "LeftGlove", Vector3(0.22, 0.12, 0.20) * scale, Vector3(-0.16 * scale, 0.18, 0.0), color)
			add_loot_box(root, "RightGlove", Vector3(0.22, 0.12, 0.20) * scale, Vector3(0.16 * scale, 0.18, 0.0), accent)
		"belt":
			add_loot_box(root, "BeltBand", Vector3(0.56, 0.10, 0.20) * scale, Vector3(0.0, 0.16, 0.0), color)
			add_loot_box(root, "BeltBuckle", Vector3(0.14, 0.12, 0.23) * scale, Vector3(0.0, 0.22, -0.02 * scale), accent)
		"boots":
			add_loot_box(root, "LeftBoot", Vector3(0.20, 0.16, 0.34) * scale, Vector3(-0.14 * scale, 0.18, 0.0), color)
			add_loot_box(root, "RightBoot", Vector3(0.20, 0.16, 0.34) * scale, Vector3(0.14 * scale, 0.18, 0.0), accent)
		"ring":
			add_loot_cylinder(root, "RingBand", 0.18 * scale, 0.05 * scale, Vector3(0.0, 0.17, 0.0), color)
			add_loot_box(root, "RingStone", Vector3(0.09, 0.08, 0.08) * scale, Vector3(0.0, 0.24, -0.14 * scale), accent)
		"amulet":
			add_loot_cylinder(root, "AmuletChain", 0.20 * scale, 0.04 * scale, Vector3(0.0, 0.17, 0.0), color)
			add_loot_box(root, "AmuletGem", Vector3(0.13, 0.12, 0.08) * scale, Vector3(0.0, 0.25, -0.15 * scale), accent)
		"coin":
			add_loot_cylinder(root, "Badge", 0.24 * scale, 0.08 * scale, Vector3(0.0, 0.16, 0.0), color)
			add_loot_cylinder(root, "BadgeMark", 0.12 * scale, 0.10 * scale, Vector3(0.0, 0.21, 0.0), accent)
		"potion":
			add_loot_cylinder(root, "Bottle", 0.17 * scale, 0.32 * scale, Vector3(0.0, 0.26, 0.0), color)
			add_loot_box(root, "Cork", Vector3(0.14, 0.10, 0.14) * scale, Vector3(0.0, 0.48 * scale, 0.0), accent)
		_:
			add_loot_box(root, "Box", Vector3(0.5, 0.5, 0.5) * scale, Vector3(0.0, 0.25 * scale, 0.0), color)

## v487: hand items (main_hand/off_hand) lie on the ground as the kit model the hero wields
## (item_visuals.v0.json). v490: amount-tiered families (gold) pick the highest
## `ground_model_tiers` entry whose min_amount <= amount (none below the first tier). Everything
## else falls back to its presentation family `3d_model`.
func ground_model_asset_id(item_def_id: String, amount: int = 0) -> String:
	var hand_asset := ItemVisualsLoader.hand_asset_id(item_def_id)
	if hand_asset != "":
		return hand_asset
	var presentation: Dictionary = item_presentations.get(item_def_id, {})
	var tiers = presentation.get("ground_model_tiers", [])
	if typeof(tiers) == TYPE_ARRAY and not (tiers as Array).is_empty():
		var chosen := ""
		for tier in tiers:
			if typeof(tier) == TYPE_DICTIONARY and amount >= int((tier as Dictionary).get("min_amount", 0)):
				chosen = str((tier as Dictionary).get("asset_id", ""))
		return chosen
	return str(presentation.get("3d_model", ""))

func make_ground_equipment_model(item_def_id: String, rarity: String, amount: int = 0) -> Node3D:
	var asset_id := ground_model_asset_id(item_def_id, amount)
	if asset_id == "":
		return null
	var entry = asset_manifest.get(asset_id, null)
	if typeof(entry) != TYPE_DICTIONARY:
		return null
	var runtime_path := str((entry as Dictionary).get("runtime_path", ""))
	var packed = load(res_path(runtime_path))
	if packed == null or not (packed is PackedScene):
		return null
	var inst := (packed as PackedScene).instantiate() as Node3D
	if inst == null:
		return null
	inst.name = "GroundModel_%s" % asset_id
	var rig_native := ItemVisualsLoader.hand_asset_id(item_def_id) != "" and ItemVisualsLoader.is_rig_native(item_def_id)
	var pose := EquipmentDisplayLoaderScript.ground_pose_for(asset_id, rig_native)
	var presentation: Dictionary = item_presentations.get(item_def_id, {})
	var ground: Dictionary = presentation.get("ground", {}) if typeof(presentation.get("ground", {})) == TYPE_DICTIONARY else {}
	var ground_scale := float(ground.get("scale", 1.0))
	var mesh_scale := ClientConstantsScript.GROUND_EQUIPMENT_MODEL_SCALE * ground_scale * EquipmentDisplayLoaderScript.ground_multiplier() * float(pose["scale"])
	inst.scale = Vector3.ONE * mesh_scale
	inst.position = Vector3(0.0, float(pose["height"]), 0.0)
	inst.rotation_degrees = pose["rotation_degrees"]
	_fit_ground_pose(inst, pose)
	var ground_tint = presentation.get("ground_tint", null)
	if typeof(ground_tint) == TYPE_DICTIONARY:
		apply_detail_tint(inst, Color(str(ground_tint.get("color", "#ffffff"))), float(ground_tint.get("strength", 0.0)))
	else:
		apply_model_tint(inst, ground_model_tint(rarity, rig_native))
	return inst

## v490: family `ground_tint` colours textured kit props through the detail layer (keeps shading).
func apply_detail_tint(root: Node, color: Color, strength: float) -> void:
	if root is MeshInstance3D:
		ModelDetailTint.set_detail(root as MeshInstance3D, color, strength)
	for child in root.get_children():
		apply_detail_tint(child, color, strength)

## Applies the optional `max_extent` cap and `rest_on_floor` placement from the ground pose, using
## the posed model's bounds in loot-root space.
func _fit_ground_pose(inst: Node3D, pose: Dictionary) -> void:
	var max_extent := float(pose.get("max_extent", 0.0))
	var bounds := posed_bounds(inst)
	if bounds.size == Vector3.ZERO:
		return
	var longest := maxf(bounds.size.x, maxf(bounds.size.y, bounds.size.z))
	if max_extent > 0.0 and longest > max_extent:
		inst.scale *= max_extent / longest
		bounds = posed_bounds(inst)
	if bool(pose.get("rest_on_floor", false)):
		var center := bounds.get_center()
		inst.position += Vector3(-center.x, float(pose["height"]) - bounds.position.y, -center.z)

## Bounds of every mesh under `inst`, in the space of inst's parent (inst's own transform applied).
static func posed_bounds(inst: Node3D) -> AABB:
	var acc := {"box": AABB(), "any": false}
	_accumulate_bounds(inst, inst.transform, acc)
	return acc["box"]

static func _accumulate_bounds(node: Node, xf: Transform3D, acc: Dictionary) -> void:
	if node is MeshInstance3D and (node as MeshInstance3D).mesh != null:
		var box: AABB = xf * (node as MeshInstance3D).mesh.get_aabb()
		acc["box"] = box if not acc["any"] else (acc["box"] as AABB).merge(box)
		acc["any"] = true
	for child in node.get_children():
		if child is Node3D:
			_accumulate_bounds(child, xf * (child as Node3D).transform, acc)

## Textured kit models blend toward the rarity colour at the same strength as equipped kit
## weapons; untextured fallback GLBs take the full rarity tint.
func ground_model_tint(rarity: String, rig_native: bool) -> Color:
	var tint := ground_item_tint(rarity)
	if rig_native:
		return Color.WHITE.lerp(tint, EquipmentDisplayLoaderScript.rig_native_tint_strength())
	return tint

func ground_item_tint(rarity: String) -> Color:
	match rarity.to_lower():
		"magic":
			return Color("#5aa7ff")
		"rare":
			return Color("#ffd75e")
		"unique":
			return Color("#ff9f52")
		"set":
			return Color("#55e66f")
		_:
			return Color("#d8d0bd")

func add_loot_label(parent: Node3D, text: String, scale: float, color: Color = Color("#f4ead8"), has_rarity_cue: bool = false) -> void:
	if text == "":
		return
	var label := Label3D.new()
	label.name = "LootLabel"
	label.text = text
	label.visible = false
	var label_height := float(RarityCueLoaderScript.catalog().get("world", {}).get("revealed_label_height", 1.25)) if has_rarity_cue else 0.58
	label.position = Vector3(0.0, label_height * maxf(scale, 0.8), 0.0)
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.no_depth_test = true
	label.fixed_size = true
	label.pixel_size = 0.0018
	label.modulate = color
	label.outline_modulate = Color(0.06, 0.045, 0.035, 0.92)
	label.outline_size = 4
	parent.add_child(label)

func add_loot_box(parent: Node3D, node_name: String, size: Vector3, position: Vector3, color: Color) -> void:
	var mesh := BoxMesh.new()
	mesh.size = size
	add_loot_mesh(parent, node_name, mesh, position, color)

func add_loot_cylinder(parent: Node3D, node_name: String, radius: float, height: float, position: Vector3, color: Color) -> void:
	var mesh := CylinderMesh.new()
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = height
	mesh.radial_segments = 16
	add_loot_mesh(parent, node_name, mesh, position, color)

func add_loot_mesh(parent: Node3D, node_name: String, mesh: Mesh, position: Vector3, color: Color) -> void:
	var node := MeshInstance3D.new()
	node.name = node_name
	node.mesh = mesh
	node.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	node.material_override = mat
	parent.add_child(node)

func loot_color(item_def_id: String) -> Color:
	var def: Dictionary = ItemRulesLoader.item_definition(item_def_id)
	var category := str(def.get("category", "equipment" if bool(def.get("equippable", false)) else "currency"))
	match category:
		"equipment":
			return Color(0.62, 0.62, 0.62)
		"quest":
			return Color(0.2, 0.85, 0.35)
		"consumable":
			return Color(0.95, 0.15, 0.12)
		_:
			return Color(1.0, 0.85, 0.2)

func loot_label_color(e: Dictionary) -> Color:
	var item_def_id := str(e.get("item_def_id", ""))
	var def := item_definition(item_def_id)
	var category := str(def.get("category", "")).to_lower()
	if item_def_id == "gold" or category == "currency":
		return ClientConstantsScript.LOOT_LABEL_CATEGORY_COLORS["currency"]
	if ClientConstantsScript.LOOT_LABEL_CATEGORY_COLORS.has(category):
		return ClientConstantsScript.LOOT_LABEL_CATEGORY_COLORS[category]
	var rarity := str(e.get("rarity", "common")).to_lower()
	return ClientConstantsScript.LOOT_LABEL_RARITY_COLORS.get(rarity, ClientConstantsScript.LOOT_LABEL_RARITY_COLORS["common"])

func loot_label_text(e: Dictionary) -> String:
	var item_def_id := str(e.get("item_def_id", ""))
	if PotionIconLabelScript.is_leveled_potion(item_def_id):
		return _potion_ground_label(item_def_id)
	var def := item_definition(item_def_id)
	var category := str(def.get("category", "")).to_lower()
	if item_def_id == "gold" or category == "currency":
		var amount := int(e.get("amount", 0))
		if amount > 0:
			return "%d gold" % amount
		return "gold"
	var display_name := str(e.get("display_name", "")).strip_edges()
	if display_name != "":
		return RarityCueLoaderScript.revealed_label(display_name, RarityCueLoaderScript.cue_for_item(e))
	var rule_name := str(def.get("name", "")).strip_edges()
	if rule_name != "":
		return RarityCueLoaderScript.revealed_label(rule_name, RarityCueLoaderScript.cue_for_item(e))
	return RarityCueLoaderScript.revealed_label(generic_loot_name(item_def_id), RarityCueLoaderScript.cue_for_item(e))

func item_definition(item_def_id: String) -> Dictionary:
	return ItemRulesLoader.item_definition(item_def_id)

func generic_loot_name(item_def_id: String) -> String:
	var def := item_definition(item_def_id)
	var item_type := str(def.get("item_type", "")).to_lower()
	match item_type:
		"sword":
			return "Sword"
		"greatsword":
			return "Greatsword"
		"staff":
			return "Staff"
		"axe":
			return "Axe"
		"bow":
			return "Bow"
		"shield":
			return "Shield"
		"helm":
			return "Helm"
		"chest":
			return "Armor"
		"gloves":
			return "Gloves"
		"belt":
			return "Belt"
		"boots":
			return "Boots"
		"ring":
			return "Ring"
		"amulet":
			return "Amulet"
	var slot := str(def.get("slot", ""))
	match slot:
		"main_hand":
			return "Bow" if str(def.get("attack_mode", "melee")) == "ranged" else "Sword"
		"off_hand":
			return "Shield"
		"head":
			return "Helm"
		"chest":
			return "Armor"
		"gloves":
			return "Gloves"
		"belt":
			return "Belt"
		"boots":
			return "Boots"
		"amulet":
			return "Amulet"
		"ring":
			return "Ring"
	match str(def.get("category", "")):
		"consumable":
			return "Potion"
		"currency":
			return "Badge"
		"quest":
			return "Item"
	return "Item"

func _potion_ground_label(item_def_id: String) -> String:
	match item_def_id:
		"red_potion":
			return "Health Potion"
		"blue_potion":
			return "Mana Potion"
		"rejuv_potion":
			return "Rejuv Potion"
		_:
			return "Potion"

func res_path(runtime_path: String) -> String:
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p

## ModelTint duplicates the mesh's own material, so kit albedo textures survive the tint.
func apply_model_tint(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		(root as MeshInstance3D).material_override = ModelTintScript.tinted_material(root as MeshInstance3D, color)
	for child in root.get_children():
		apply_model_tint(child, color)
