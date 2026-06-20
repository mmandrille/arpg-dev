class_name LootNodeFactory
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

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
	var rarity := str(e.get("rarity", "common"))
	add_loot_rarity_glow(root, rarity, scale)
	var model := make_ground_equipment_model(item_def_id, str(e.get("rarity", "common")))
	if model != null:
		root.add_child(model)
	else:
		add_loot_rarity_background(root, item_rarity_background(rarity), scale)
		add_loot_primitive(root, shape, color, accent, scale)
	add_loot_spawn_pop(root, rarity, scale)
	add_loot_label(root, loot_label_text(e), scale, loot_label_color(e))
	return root

func add_loot_primitive(root: Node3D, shape: String, color: Color, accent: Color, scale: float) -> void:
	match shape:
		"blade":
			add_loot_box(root, "Blade", Vector3(0.12, 0.08, 0.78) * scale, Vector3(0.0, 0.20, 0.0), color)
			add_loot_box(root, "Grip", Vector3(0.34, 0.10, 0.10) * scale, Vector3(0.0, 0.16, 0.34 * scale), accent)
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
		"badge", "coin":
			add_loot_cylinder(root, "Badge", 0.24 * scale, 0.08 * scale, Vector3(0.0, 0.16, 0.0), color)
			add_loot_cylinder(root, "BadgeMark", 0.12 * scale, 0.10 * scale, Vector3(0.0, 0.21, 0.0), accent)
		"leaf":
			add_loot_box(root, "Leaf", Vector3(0.42, 0.06, 0.24) * scale, Vector3(0.0, 0.16, 0.0), color)
			add_loot_box(root, "Stem", Vector3(0.06, 0.08, 0.46) * scale, Vector3(0.0, 0.18, 0.0), accent)
		"potion":
			add_loot_cylinder(root, "Bottle", 0.17 * scale, 0.32 * scale, Vector3(0.0, 0.26, 0.0), color)
			add_loot_box(root, "Cork", Vector3(0.14, 0.10, 0.14) * scale, Vector3(0.0, 0.48 * scale, 0.0), accent)
		_:
			add_loot_box(root, "Box", Vector3(0.5, 0.5, 0.5) * scale, Vector3(0.0, 0.25 * scale, 0.0), color)

func make_ground_equipment_model(item_def_id: String, rarity: String) -> Node3D:
	var presentation: Dictionary = item_presentations.get(item_def_id, {})
	var asset_id := str(presentation.get("3d_model", ""))
	if asset_id == "" or asset_id.begins_with("fallback_equipment_"):
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
	inst.scale = Vector3.ONE * ClientConstantsScript.GROUND_EQUIPMENT_MODEL_SCALE
	inst.position = Vector3(0.0, 0.12, 0.0)
	inst.rotation_degrees = Vector3(90.0, 35.0, 0.0)
	apply_model_tint(inst, ground_item_tint(rarity))
	return inst

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

func add_loot_rarity_background(parent: Node3D, color: Color, scale: float) -> void:
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.82, 0.04, 0.82) * maxf(scale, 0.85)
	add_loot_mesh(parent, "RarityBackground", mesh, Vector3(0.0, 0.045, 0.0), color)

func add_loot_rarity_glow(parent: Node3D, rarity: String, scale: float) -> void:
	var mesh := TorusMesh.new()
	mesh.inner_radius = 0.32 * maxf(scale, 0.85)
	mesh.outer_radius = 0.43 * maxf(scale, 0.85)
	mesh.ring_segments = 32
	var node := MeshInstance3D.new()
	node.name = "RarityGlow"
	node.mesh = mesh
	node.position = Vector3(0.0, 0.055, 0.0)
	node.rotation_degrees.x = 90.0
	node.material_override = _glow_material(item_rarity_background(rarity), 0.42)
	parent.add_child(node)

func add_loot_spawn_pop(parent: Node3D, rarity: String, scale: float) -> void:
	var mesh := TorusMesh.new()
	mesh.inner_radius = 0.48 * maxf(scale, 0.85)
	mesh.outer_radius = 0.52 * maxf(scale, 0.85)
	mesh.ring_segments = 28
	var node := MeshInstance3D.new()
	node.name = "SpawnPopRing"
	node.mesh = mesh
	node.position = Vector3(0.0, 0.095, 0.0)
	node.rotation_degrees.x = 90.0
	node.material_override = _glow_material(ground_item_tint(rarity), 0.28)
	parent.add_child(node)

func add_loot_label(parent: Node3D, text: String, scale: float, color: Color = Color("#f4ead8")) -> void:
	if text == "":
		return
	var label := Label3D.new()
	label.name = "LootLabel"
	label.text = text
	label.visible = false
	label.position = Vector3(0.0, 0.58 * maxf(scale, 0.8), 0.0)
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

func _glow_material(color: Color, alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.45
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	return mat

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
	var def := item_definition(item_def_id)
	var category := str(def.get("category", "")).to_lower()
	if item_def_id == "gold" or category == "currency":
		var amount := int(e.get("amount", 0))
		if amount > 0:
			return "%d gold" % amount
		return "gold"
	return generic_loot_name(item_def_id)

func item_rarity_background(rarity: String) -> Color:
	var key := rarity.to_lower()
	return ClientConstantsScript.ITEM_RARITY_BACKGROUNDS.get(key, ClientConstantsScript.ITEM_RARITY_BACKGROUNDS["common"])

func item_definition(item_def_id: String) -> Dictionary:
	return ItemRulesLoader.item_definition(item_def_id)

func generic_loot_name(item_def_id: String) -> String:
	var def := item_definition(item_def_id)
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

func res_path(runtime_path: String) -> String:
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p

func apply_model_tint(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		(root as MeshInstance3D).material_override = mat
	for child in root.get_children():
		apply_model_tint(child, color)
