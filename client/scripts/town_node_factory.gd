class_name TownNodeFactory
extends RefCounted

const ChestPresentationScript := preload("res://scripts/chest_presentation.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")

static func make_door_node() -> Node3D:
	var root := Node3D.new()
	root.name = "InteractableDoor"
	var pivot := Node3D.new()
	pivot.name = "DoorPivot"
	pivot.position = Vector3(-0.5, 0.0, 0.0)
	root.add_child(pivot)
	var panel := MeshInstance3D.new()
	panel.name = "DoorPanel"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.0, 1.0, 0.25)
	panel.mesh = mesh
	panel.position = Vector3(0.5, 0.5, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.55, 0.32, 0.15)
	panel.material_override = mat
	pivot.add_child(panel)
	return root

static func make_chest_node(def_id: String, elite_objective: bool = false, quest_reward: bool = false) -> Node3D:
	var is_stash := def_id == "town_stash"
	var is_unique_test := def_id == "town_unique_chest"
	var root := Node3D.new()
	root.name = "UniqueTestChest" if is_unique_test else ("TownStashChest" if is_stash else "TreasureChest")
	var scale := 1.12 if is_stash else (1.08 if is_unique_test else 1.0)
	var wood := Color("#5a2d7a") if is_unique_test else (Color("#6f3b18") if is_stash else Color("#744018"))
	var dark_wood := Color("#30143f") if is_unique_test else (Color("#3c2111") if is_stash else Color("#4a2711"))
	var metal := Color("#c77dff") if is_unique_test else (Color("#d1b15d") if is_stash else Color("#8d8f8f"))
	var metal_dark := Color("#6f37a8") if is_unique_test else (Color("#6f5b2e") if is_stash else Color("#3d4143"))
	var cloth := Color("#f0b8ff") if is_unique_test else (Color("#244e66") if is_stash else Color("#5a2017"))
	var glow := Color("#d77dff") if is_unique_test else (Color("#f3d36b") if is_stash else Color("#f5b449"))
	ChestPresentationScript.add_part(root, "ChestShadow", Vector3(1.28, 0.035, 0.82) * scale, Vector3(0.0, 0.018, 0.0), Color("#181715"))
	ChestPresentationScript.add_part(root, "ChestBody", Vector3(1.08, 0.48, 0.70) * scale, Vector3(0.0, 0.29 * scale, 0.0), wood)
	ChestPresentationScript.add_part(root, "ChestFrontPanel", Vector3(0.92, 0.30, 0.045) * scale, Vector3(0.0, 0.30 * scale, 0.374 * scale), dark_wood)
	ChestPresentationScript.add_part(root, "ChestBackPanel", Vector3(0.92, 0.30, 0.045) * scale, Vector3(0.0, 0.30 * scale, -0.374 * scale), dark_wood)
	ChestPresentationScript.add_part(root, "ChestLeftPanel", Vector3(0.045, 0.30, 0.56) * scale, Vector3(-0.564 * scale, 0.30 * scale, 0.0), dark_wood)
	ChestPresentationScript.add_part(root, "ChestRightPanel", Vector3(0.045, 0.30, 0.56) * scale, Vector3(0.564 * scale, 0.30 * scale, 0.0), dark_wood)
	ChestPresentationScript.add_part(root, "ChestFrontBand", Vector3(1.18, 0.08, 0.055) * scale, Vector3(0.0, 0.48 * scale, 0.405 * scale), metal)
	ChestPresentationScript.add_part(root, "ChestBottomBand", Vector3(1.18, 0.075, 0.055) * scale, Vector3(0.0, 0.13 * scale, 0.405 * scale), metal_dark)
	for x in [-0.42, 0.42]:
		ChestPresentationScript.add_part(root, "ChestVerticalBand", Vector3(0.085, 0.54, 0.085) * scale, Vector3(x * scale, 0.33 * scale, 0.405 * scale), metal)
	for x in [-0.43, 0.43]:
		for z in [-0.28, 0.28]:
			ChestPresentationScript.add_part(root, "ChestFoot", Vector3(0.22, 0.10, 0.16) * scale, Vector3(x * scale, 0.055 * scale, z * scale), metal_dark)
	var lid_pivot := Node3D.new()
	lid_pivot.name = "ChestLidPivot"
	lid_pivot.position = Vector3(0.0, 0.56 * scale, -0.36 * scale)
	root.add_child(lid_pivot)
	ChestPresentationScript.add_part(lid_pivot, "ChestLid", Vector3(1.16, 0.30, 0.72) * scale, Vector3(0.0, 0.15 * scale, 0.36 * scale), wood)
	ChestPresentationScript.add_part(lid_pivot, "ChestLidCrown", Vector3(0.92, 0.12, 0.58) * scale, Vector3(0.0, 0.33 * scale, 0.36 * scale), Color("#7b35b0") if is_unique_test else (Color("#8a4f20") if is_stash else Color("#8b511f")))
	ChestPresentationScript.add_part(lid_pivot, "ChestLidFrontBand", Vector3(1.22, 0.08, 0.075) * scale, Vector3(0.0, 0.13 * scale, 0.74 * scale), metal)
	for x in [-0.42, 0.42]:
		ChestPresentationScript.add_part(lid_pivot, "ChestLidStrap", Vector3(0.075, 0.36, 0.80) * scale, Vector3(x * scale, 0.20 * scale, 0.36 * scale), metal)
	ChestPresentationScript.add_part(root, "ChestLockPlate", Vector3(0.22, 0.24, 0.075) * scale, Vector3(0.0, 0.42 * scale, 0.452 * scale), metal)
	ChestPresentationScript.add_part(root, "ChestLockSlot", Vector3(0.075, 0.11, 0.085) * scale, Vector3(0.0, 0.40 * scale, 0.502 * scale), metal_dark)
	ChestPresentationScript.add_part(root, "ChestLeftHandle", Vector3(0.075, 0.22, 0.30) * scale, Vector3(-0.64 * scale, 0.36 * scale, 0.0), metal)
	ChestPresentationScript.add_part(root, "ChestRightHandle", Vector3(0.075, 0.22, 0.30) * scale, Vector3(0.64 * scale, 0.36 * scale, 0.0), metal)
	if is_stash or is_unique_test:
		ChestPresentationScript.add_part(root, "ChestStashCrest", Vector3(0.36, 0.12, 0.082) * scale, Vector3(0.0, 0.61 * scale, 0.456 * scale), cloth)
	var inner := ChestPresentationScript.add_part(root, "ChestInnerGlow", Vector3(0.84, 0.045, 0.46) * scale, Vector3(0.0, 0.57 * scale, 0.02 * scale), glow)
	var glow_mat := inner.material_override as StandardMaterial3D
	glow_mat.emission_enabled = true
	glow_mat.emission = glow
	inner.visible = false
	ChestPresentationScript.sync_objective_marker(root, elite_objective, false)
	ChestPresentationScript.sync_quest_marker(root, quest_reward, false)
	return root

static func make_merchant_node(def_id: String) -> Node3D:
	var is_mystery := def_id == "town_mystery_seller"
	var root := Node3D.new()
	root.name = "MysterySeller" if is_mystery else "ShopVendor"
	var cloth := Color("#2b124a") if is_mystery else Color("#e2b92e")
	var cloth_dark := Color("#150824") if is_mystery else Color("#8c6418")
	var accent := Color("#7c4dff") if is_mystery else Color("#ffe37a")
	var trim := Color("#c7a6ff") if is_mystery else Color("#f6d85c")
	var wood := Color("#4b2a16")
	var skin := Color("#c1845a") if is_mystery else Color("#c99666")
	var glow := Color("#6f40ff") if is_mystery else Color("#ffd85a")
	add_merchant_box(root, "MerchantShadow", Vector3(1.35, 0.035, 0.92), Vector3(0.0, 0.018, 0.0), Color("#181715"))
	add_merchant_box(root, "CounterTop", Vector3(1.12, 0.18, 0.48), Vector3(0.0, 0.39, 0.31), wood)
	add_merchant_box(root, "CounterFront", Vector3(1.18, 0.36, 0.09), Vector3(0.0, 0.22, 0.57), cloth_dark)
	add_merchant_box(root, "CounterTrim", Vector3(1.24, 0.08, 0.10), Vector3(0.0, 0.44, 0.60), trim)
	add_merchant_box(root, "CounterLeftLeg", Vector3(0.12, 0.34, 0.12), Vector3(-0.50, 0.17, 0.36), wood)
	add_merchant_box(root, "CounterRightLeg", Vector3(0.12, 0.34, 0.12), Vector3(0.50, 0.17, 0.36), wood)
	add_merchant_box(root, "Body", Vector3(0.42, 0.66, 0.30), Vector3(0.0, 0.78, -0.05), cloth)
	add_merchant_box(root, "Belt", Vector3(0.48, 0.09, 0.34), Vector3(0.0, 0.62, -0.04), Color("#2f2117"))
	add_merchant_box(root, "Head", Vector3(0.32, 0.30, 0.30), Vector3(0.0, 1.28, -0.03), skin)
	add_merchant_box(root, "Nose", Vector3(0.08, 0.08, 0.08), Vector3(0.0, 1.27, 0.16), skin.lerp(Color.WHITE, 0.12))
	add_merchant_box(root, "LeftArm", Vector3(0.12, 0.44, 0.16), Vector3(-0.32, 0.78, 0.02), cloth_dark)
	add_merchant_box(root, "RightArm", Vector3(0.12, 0.44, 0.16), Vector3(0.32, 0.78, 0.02), cloth_dark)
	add_merchant_box(root, "LeftHand", Vector3(0.13, 0.10, 0.13), Vector3(-0.32, 0.53, 0.08), skin)
	add_merchant_box(root, "RightHand", Vector3(0.13, 0.10, 0.13), Vector3(0.32, 0.53, 0.08), skin)
	if is_mystery:
		add_merchant_box(root, "Hood", Vector3(0.44, 0.22, 0.36), Vector3(0.0, 1.39, -0.02), cloth_dark)
		add_merchant_box(root, "HoodRim", Vector3(0.40, 0.08, 0.08), Vector3(0.0, 1.30, 0.18), accent)
		add_merchant_cylinder(root, "CrystalOrb", 0.16, 0.22, Vector3(0.34, 0.62, 0.43), glow, true)
		add_merchant_box(root, "MysterySign", Vector3(0.34, 0.24, 0.07), Vector3(-0.36, 0.67, 0.62), accent)
	else:
		add_merchant_box(root, "HatBrim", Vector3(0.54, 0.08, 0.42), Vector3(0.0, 1.43, -0.02), accent)
		add_merchant_box(root, "HatCrown", Vector3(0.34, 0.22, 0.30), Vector3(0.0, 1.56, -0.02), cloth)
		add_merchant_box(root, "CoinStackA", Vector3(0.16, 0.08, 0.16), Vector3(-0.30, 0.54, 0.43), glow)
		add_merchant_box(root, "CoinStackB", Vector3(0.14, 0.14, 0.14), Vector3(-0.12, 0.57, 0.42), glow)
		add_merchant_box(root, "VendorSign", Vector3(0.34, 0.24, 0.07), Vector3(0.36, 0.67, 0.62), accent)
	return root

static func make_bishop_node() -> Node3D:
	var root := Node3D.new()
	root.name = "TownBishop"
	add_merchant_box(root, "BishopShadow", Vector3(0.92, 0.035, 0.78), Vector3(0.0, 0.018, 0.0), Color("#171313"))
	add_merchant_box(root, "RobeLower", Vector3(0.54, 0.62, 0.36), Vector3(0.0, 0.46, 0.0), Color("#621717"))
	add_merchant_box(root, "RobeUpper", Vector3(0.46, 0.58, 0.30), Vector3(0.0, 0.92, 0.0), Color("#b92d2d"))
	add_merchant_box(root, "Sash", Vector3(0.12, 0.68, 0.34), Vector3(0.0, 0.78, 0.02), Color("#f3d7a8"))
	add_merchant_box(root, "Shoulders", Vector3(0.62, 0.12, 0.34), Vector3(0.0, 1.16, 0.0), Color("#621717"))
	add_merchant_box(root, "Head", Vector3(0.30, 0.30, 0.28), Vector3(0.0, 1.42, 0.0), Color("#c99666"))
	add_merchant_box(root, "MitreBase", Vector3(0.38, 0.16, 0.28), Vector3(0.0, 1.64, 0.0), Color("#b92d2d"))
	add_merchant_box(root, "MitrePeak", Vector3(0.24, 0.24, 0.20), Vector3(0.0, 1.84, 0.0), Color("#b92d2d"))
	add_merchant_box(root, "MitreTrim", Vector3(0.42, 0.06, 0.30), Vector3(0.0, 1.56, 0.0), Color("#f3d7a8"))
	add_merchant_box(root, "LeftArm", Vector3(0.12, 0.50, 0.14), Vector3(-0.36, 0.86, 0.02), Color("#621717"))
	add_merchant_box(root, "RightArm", Vector3(0.12, 0.50, 0.14), Vector3(0.36, 0.86, 0.02), Color("#621717"))
	add_merchant_box(root, "LeftHand", Vector3(0.12, 0.10, 0.12), Vector3(-0.36, 0.56, 0.06), Color("#c99666"))
	add_merchant_box(root, "RightHand", Vector3(0.12, 0.10, 0.12), Vector3(0.36, 0.56, 0.06), Color("#c99666"))
	add_merchant_cylinder(root, "Staff", 0.035, 1.18, Vector3(0.55, 0.82, 0.02), Color("#5a351c"))
	add_merchant_cylinder(root, "StaffCrown", 0.11, 0.09, Vector3(0.55, 1.44, 0.02), Color("#d8a342"), true)
	add_merchant_box(root, "ServiceBook", Vector3(0.30, 0.08, 0.22), Vector3(-0.18, 0.58, 0.24), Color("#efe0bc"))
	return root

static func make_blacksmith_node() -> Node3D:
	var root := Node3D.new()
	root.name = "TownBlacksmith"
	add_merchant_box(root, "BlacksmithShadow", Vector3(1.42, 0.035, 0.82), Vector3(0.0, 0.018, 0.0), Color("#171513"))
	add_merchant_box(root, "AnvilBase", Vector3(0.72, 0.18, 0.46), Vector3(0.38, 0.18, 0.34), Color("#33383a"))
	add_merchant_box(root, "AnvilTop", Vector3(0.92, 0.16, 0.34), Vector3(0.38, 0.36, 0.34), Color("#9da3a6"))
	add_merchant_box(root, "ForgeBody", Vector3(0.58, 0.44, 0.50), Vector3(-0.46, 0.30, 0.28), Color("#4a2b17"))
	add_merchant_box(root, "ForgeMouth", Vector3(0.40, 0.22, 0.08), Vector3(-0.46, 0.32, 0.56), Color("#ff7a2f"))
	add_merchant_box(root, "Body", Vector3(0.46, 0.66, 0.30), Vector3(0.0, 0.78, -0.06), Color("#2f3f46"))
	add_merchant_box(root, "Apron", Vector3(0.34, 0.58, 0.34), Vector3(0.0, 0.68, 0.08), Color("#182429"))
	add_merchant_box(root, "Head", Vector3(0.32, 0.30, 0.30), Vector3(0.0, 1.28, -0.04), Color("#b87955"))
	add_merchant_box(root, "Cap", Vector3(0.42, 0.12, 0.34), Vector3(0.0, 1.48, -0.04), Color("#4b5154"))
	add_merchant_box(root, "LeftArm", Vector3(0.13, 0.48, 0.16), Vector3(-0.35, 0.80, 0.02), Color("#b87955"))
	add_merchant_box(root, "RightArm", Vector3(0.13, 0.48, 0.16), Vector3(0.35, 0.80, 0.02), Color("#b87955"))
	add_merchant_box(root, "HammerHandle", Vector3(0.07, 0.42, 0.07), Vector3(0.53, 0.58, 0.26), Color("#5b3218"))
	add_merchant_box(root, "HammerHead", Vector3(0.28, 0.12, 0.14), Vector3(0.53, 0.82, 0.26), Color("#9da3a6"))
	return root

static func make_market_board_node() -> Node3D:
	var root := Node3D.new()
	root.name = "MarketBoard"
	add_merchant_box(root, "MarketBoardShadow", Vector3(1.70, 0.035, 0.52), Vector3(0.0, 0.018, 0.03), Color("#171513"))
	add_merchant_box(root, "MarketBoardLeftPost", Vector3(0.13, 1.24, 0.13), Vector3(-0.68, 0.62, 0.0), Color("#4a2b17"))
	add_merchant_box(root, "MarketBoardRightPost", Vector3(0.13, 1.24, 0.13), Vector3(0.68, 0.62, 0.0), Color("#4a2b17"))
	add_merchant_box(root, "MarketBoardPanel", Vector3(1.34, 0.82, 0.12), Vector3(0.0, 0.88, 0.02), Color("#6f431f"))
	add_merchant_box(root, "MarketBoardInset", Vector3(1.12, 0.58, 0.135), Vector3(0.0, 0.88, 0.09), Color("#2c2118"))
	add_merchant_box(root, "MarketBoardHeader", Vector3(1.44, 0.16, 0.14), Vector3(0.0, 1.38, 0.04), Color("#c99d4a"))
	add_merchant_box(root, "MarketBoardPaperA", Vector3(0.30, 0.36, 0.145), Vector3(-0.24, 0.83, 0.16), Color("#efe0bc"))
	add_merchant_box(root, "MarketBoardPaperB", Vector3(0.27, 0.30, 0.145), Vector3(0.22, 0.96, 0.16), Color("#d7c99e"))
	add_merchant_box(root, "MarketBoardSeal", Vector3(0.13, 0.13, 0.16), Vector3(0.48, 0.70, 0.17), Color("#b93131"))
	root.add_child(make_market_badge("IncomingBidBadge", "IncomingBidCount", Vector3(-0.58, 1.42, 0.20), Color("#4f2b12"), Color("#776d5e")))
	root.add_child(make_market_badge("PublishedListingBadge", "PublishedListingCount", Vector3(0.58, 1.42, 0.20), Color("#14324f"), Color("#776d5e")))
	return root

static func make_town_preview_scene() -> Node3D:
	var root := Node3D.new()
	root.name = "TownPreview"
	var ground := GroundWallFactoryScript.new().make_ground_node(0)
	ground.name = "TownPreviewGround"
	ground.position = Vector3(11.5, -0.02, 11.5)
	var ground_mesh := ground.mesh as PlaneMesh
	if ground_mesh != null:
		ground_mesh.size = Vector2(28.0, 22.0)
	root.add_child(ground)
	var service_entries := [
		{"def_id": "stairs_down", "position": Vector3(11.0, 0.0, 8.0)},
		{"def_id": "teleporter", "position": Vector3(2.0, 0.0, 12.0)},
		{"def_id": "town_vendor", "position": Vector3(17.0, 0.0, 10.0)},
		{"def_id": "town_mystery_seller", "position": Vector3(18.0, 0.0, 15.0)},
		{"def_id": "town_stash", "position": Vector3(7.0, 0.0, 14.0)},
		{"def_id": "town_bishop", "position": Vector3(15.0, 0.0, 6.0)},
		{"def_id": "town_market_board", "position": Vector3(10.0, 0.0, 18.0)},
		{"def_id": "town_blacksmith", "position": Vector3(6.0, 0.0, 10.0)},
	]
	for entry in service_entries:
		var service := make_interactable_node(str(entry["def_id"]))
		service.name = "TownService_%s" % str(entry["def_id"])
		service.position = entry["position"] as Vector3
		root.add_child(service)
	var cabin_a := make_town_cabin_node("west")
	cabin_a.name = "TownCabinWest"
	cabin_a.position = Vector3(5.0, 0.0, 7.0)
	cabin_a.rotation_degrees.y = -22.0
	root.add_child(cabin_a)
	var cabin_b := make_town_cabin_node("east")
	cabin_b.name = "TownCabinEast"
	cabin_b.position = Vector3(21.0, 0.0, 12.5)
	cabin_b.rotation_degrees.y = 18.0
	root.add_child(cabin_b)
	var fire := make_town_campfire_node()
	fire.position = Vector3(12.0, 0.0, 13.0)
	root.add_child(fire)
	return root

static func make_interactable_node(def_id: String, elite_objective: bool = false, quest_reward: bool = false) -> Node3D:
	match def_id:
		"stairs_down", "stairs_up":
			return make_stair_node(def_id)
		"teleporter":
			return make_teleporter_node()
		"treasure_chest", "town_stash", "town_unique_chest":
			return make_chest_node(def_id, elite_objective, quest_reward)
		"town_vendor", "town_mystery_seller":
			return make_merchant_node(def_id)
		"town_bishop":
			return make_bishop_node()
		"town_market_board":
			return make_market_board_node()
		"town_blacksmith":
			return make_blacksmith_node()
	return make_door_node()

static func make_town_cabin_node(variant: String = "plain") -> Node3D:
	var root := Node3D.new()
	root.name = "TownCabin"
	var roof := Color("#5b2b1d") if variant == "west" else Color("#69401f")
	add_merchant_box(root, "CabinShadow", Vector3(2.45, 0.035, 1.85), Vector3(0.0, 0.018, 0.0), Color("#17130f"))
	add_merchant_box(root, "CabinBody", Vector3(1.78, 1.02, 1.28), Vector3(0.0, 0.54, 0.0), Color("#6d3f1f"))
	add_merchant_box(root, "CabinFront", Vector3(1.86, 0.74, 0.08), Vector3(0.0, 0.46, 0.68), Color("#3b2113"))
	add_merchant_box(root, "CabinDoor", Vector3(0.42, 0.62, 0.10), Vector3(-0.36, 0.35, 0.74), Color("#2b1710"))
	add_merchant_box(root, "CabinWindow", Vector3(0.34, 0.30, 0.11), Vector3(0.38, 0.58, 0.75), Color("#d7ad58"))
	add_merchant_box(root, "CabinRoofA", Vector3(2.15, 0.34, 1.58), Vector3(0.0, 1.15, -0.14), roof)
	add_merchant_box(root, "CabinRoofRidge", Vector3(2.28, 0.18, 0.22), Vector3(0.0, 1.42, 0.0), Color("#b18a4a"))
	for x in [-0.76, 0.0, 0.76]:
		add_merchant_box(root, "CabinWallLog", Vector3(0.08, 1.04, 1.36), Vector3(x, 0.56, 0.0), Color("#4f2d18"))
	return root

static func make_town_campfire_node() -> Node3D:
	var root := Node3D.new()
	root.name = "TownCampfire"
	add_merchant_cylinder(root, "FireStoneRing", 0.58, 0.08, Vector3(0.0, 0.04, 0.0), Color("#4e4d48"))
	for i in range(6):
		var angle := TAU * float(i) / 6.0
		var stone := add_merchant_box(root, "FireStone%d" % i, Vector3(0.20, 0.11, 0.16), Vector3(cos(angle) * 0.47, 0.10, sin(angle) * 0.47), Color("#777067"))
		stone.rotation_degrees.y = rad_to_deg(angle)
	for i in range(3):
		var log := add_merchant_box(root, "FireLog%d" % i, Vector3(0.72, 0.11, 0.16), Vector3(0.0, 0.17 + float(i) * 0.025, 0.0), Color("#4b2815"))
		log.rotation_degrees.y = 60.0 * float(i)
	var flame_outer := add_merchant_cylinder(root, "FireFlameOuter", 0.24, 0.62, Vector3(0.0, 0.52, 0.0), Color("#ff7a1a"), true)
	flame_outer.scale.x = 0.58
	flame_outer.scale.z = 0.58
	var flame_inner := add_merchant_cylinder(root, "FireFlameInner", 0.14, 0.44, Vector3(0.0, 0.58, 0.0), Color("#ffd45a"), true)
	flame_inner.scale.x = 0.52
	flame_inner.scale.z = 0.52
	var light := OmniLight3D.new()
	light.name = "CampfireLight"
	light.light_color = Color("#ff9b3d")
	light.light_energy = 1.8
	light.omni_range = 4.0
	light.position = Vector3(0.0, 0.78, 0.0)
	root.add_child(light)
	return root

static func make_market_badge(badge_name: String, count_name: String, position: Vector3, bg_color: Color, text_color: Color) -> Node3D:
	var badge := Node3D.new()
	badge.name = badge_name
	badge.position = position
	add_merchant_box(badge, "BadgeBack", Vector3(0.34, 0.24, 0.055), Vector3.ZERO, bg_color)
	var label := Label3D.new()
	label.name = count_name
	label.text = "0"
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 42
	label.modulate = text_color
	label.outline_size = 8
	label.outline_modulate = Color("#14110d")
	label.position = Vector3(0.0, -0.055, 0.04)
	label.pixel_size = 0.006
	badge.add_child(label)
	return badge

static func add_merchant_box(parent: Node3D, part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	part.material_override = merchant_material(color)
	parent.add_child(part)
	return part

static func add_merchant_cylinder(parent: Node3D, part_name: String, radius: float, height: float, position: Vector3, color: Color, emit: bool = false) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := CylinderMesh.new()
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = height
	mesh.radial_segments = 24
	part.mesh = mesh
	part.position = position
	part.material_override = merchant_material(color, emit)
	parent.add_child(part)
	return part

static func merchant_material(color: Color, emit: bool = false) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	if emit:
		mat.emission_enabled = true
		mat.emission = color
	return mat

static func make_stair_node(def_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "Stairs_%s" % def_id
	var is_down := def_id == "stairs_down"
	if is_down:
		add_stair_box(root, "DownStoneFrame", Vector3(1.46, 0.16, 1.20), Vector3(0.0, 0.08, 0.0), stair_base_color(def_id, "ready"))
		add_stair_box(root, "DownPitOpening", Vector3(1.02, 0.06, 0.76), Vector3(0.0, 0.185, -0.04), Color("#030507"))
		add_stair_box(root, "DownBackWall", Vector3(1.02, 0.38, 0.12), Vector3(0.0, 0.34, -0.46), Color("#10151c"))
		add_stair_box(root, "DownLeftWall", Vector3(0.12, 0.30, 0.76), Vector3(-0.57, 0.30, -0.04), Color("#18202a"))
		for i in range(5):
			var t := float(i) / 4.0
			add_stair_box(root, "DownStep%d" % i, Vector3(0.84 - t * 0.12, 0.09, 0.16), Vector3(0.0, 0.34 - t * 0.22, 0.28 - t * 0.54), Color("#66707a").lerp(Color("#10151d"), t))
		add_stair_box(root, "DownThreshold", Vector3(1.20, 0.10, 0.14), Vector3(0.0, 0.24, 0.42), Color("#89909a"))
	else:
		add_stair_box(root, "UpGroundPad", Vector3(1.42, 0.14, 1.14), Vector3(0.0, 0.07, 0.0), stair_base_color(def_id, "ready"))
		add_stair_box(root, "UpHighLanding", Vector3(0.64, 0.16, 0.46), Vector3(0.0, 0.72, -0.45), Color("#d6cfaa"))
		add_stair_box(root, "UpBackWall", Vector3(0.78, 0.42, 0.12), Vector3(0.0, 0.54, -0.72), Color("#aaa78c"))
		for i in range(5):
			var t := float(i) / 4.0
			add_stair_box(root, "UpStep%d" % i, Vector3(1.12 - t * 0.42, 0.12, 0.20), Vector3(0.0, 0.19 + t * 0.45, 0.36 - t * 0.66), Color("#7c817a").lerp(Color("#d3cda9"), t))
	return root

static func add_stair_box(parent: Node3D, part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	part.material_override = stair_material(color)
	parent.add_child(part)
	return part

static func stair_base_color(def_id: String, state: String) -> Color:
	if state == "locked" or state == "disabled":
		return Color("#6b2e2e")
	if def_id == "stairs_down":
		return Color("#111821")
	return Color("#666d68")

static func stair_material(color: Color, glow: bool = false) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	if glow:
		mat.emission_enabled = true
		mat.emission = color
	return mat

static func make_teleporter_node() -> Node3D:
	var root := Node3D.new()
	root.name = "Teleporter"
	var base := MeshInstance3D.new()
	var base_mesh := CylinderMesh.new()
	base_mesh.top_radius = 0.62
	base_mesh.bottom_radius = 0.72
	base_mesh.height = 0.16
	base.mesh = base_mesh
	base.position = Vector3(0.0, 0.08, 0.0)
	var base_mat := StandardMaterial3D.new()
	base_mat.albedo_color = Color(0.16, 0.19, 0.22)
	base.material_override = base_mat
	root.add_child(base)
	var core := MeshInstance3D.new()
	var core_mesh := CylinderMesh.new()
	core_mesh.top_radius = 0.34
	core_mesh.bottom_radius = 0.34
	core_mesh.height = 0.42
	core.mesh = core_mesh
	core.position = Vector3(0.0, 0.32, 0.0)
	var core_mat := StandardMaterial3D.new()
	core_mat.albedo_color = Color(0.15, 0.62, 0.70)
	core_mat.emission_enabled = true
	core_mat.emission = Color(0.05, 0.55, 0.68)
	core.material_override = core_mat
	root.add_child(core)
	return root
