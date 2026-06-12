extends SceneTree

const CharacterScene := preload("res://scenes/character.tscn")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const MarketPanelScript := preload("res://scripts/market_panel.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const MultiplayerSessionsPanelScript := preload("res://scripts/multiplayer_sessions_panel.gd")
const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const MainScript := preload("res://scripts/main.gd")
const HealRainEffectScript := preload("res://scripts/heal_rain_effect.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")

const DEFAULT_GEAR_ITEMS := ["cave_blade", "cave_shield", "cave_helm", "cave_mail", "cave_boots"]
const ITEM_SLOT := {
	"rusty_sword": "main_hand",
	"cave_blade": "main_hand",
	"starter_sorcerer_staff": "main_hand",
	"starter_barbarian_axe": "main_hand",
	"cave_greatsword": "main_hand",
	"cave_war_sword": "main_hand",
	"training_bow": "main_hand",
	"cave_bow": "main_hand",
	"cave_shield": "off_hand",
	"cave_helm": "head",
	"cave_mail": "chest",
	"cave_gloves": "gloves",
	"cave_belt": "belt",
	"cave_pack_belt": "belt",
	"cave_boots": "boots",
	"cave_ring": "ring_left",
	"cave_amulet": "amulet",
}

var _mode := "screenshot"
var _focus := "gear"
var _output := ""
var _width := 640
var _height := 480
var _duration := 0.0
var _items: Array = []
var _subject: Node3D
var _skills_panel: Control


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)

	match _focus:
		"floor-item":
			await _setup_floor_item()
		"character-menu":
			await _setup_character_menu()
		"join-menu":
			await _setup_join_menu()
		"hud":
			await _setup_hud()
		"stairs":
			await _setup_stairs()
		"chests":
			await _setup_chests()
		"vendors":
			await _setup_vendors()
		"monsters":
			await _setup_monsters()
		"classes":
			await _setup_classes()
		"heal-rain":
			await _setup_heal_rain()
		"town":
			await _setup_town()
		"inventory":
			await _setup_inventory()
		"skills":
			await _setup_skills()
		"shop":
			await _setup_shop()
		"market-board":
			await _setup_market_board()
		"market-publish":
			await _setup_market_publish()
		"market-offer":
			await _setup_market_offer()
		_:
			await _setup_gear()

	for _i in range(8):
		await process_frame
	if _focus == "skills" and _skills_panel != null:
		_skills_panel.bot_hover_skill("rage")
		await process_frame

	if _mode == "live":
		if _duration > 0.0:
			await create_timer(_duration).timeout
			quit(0)
		return

	var image := get_root().get_texture().get_image()
	var err := image.save_png(_output)
	if err != OK:
		printerr("[showme] failed to save screenshot: %s err=%d" % [_output, err])
		quit(1)
		return
	print("[showme] saved %s" % _output)
	quit(0)


func _process(delta: float) -> bool:
	if _mode == "live" and _subject != null:
		_subject.rotate_y(delta * 0.35)
	return false


func _parse_args() -> void:
	_items = DEFAULT_GEAR_ITEMS.duplicate()
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i < args.size():
		var key := str(args[i])
		match key:
			"--mode":
				i += 1
				_mode = str(args[i])
			"--focus":
				i += 1
				_focus = str(args[i])
			"--output":
				i += 1
				_output = str(args[i])
			"--width":
				i += 1
				_width = int(args[i])
			"--height":
				i += 1
				_height = int(args[i])
			"--duration":
				i += 1
				_duration = float(args[i])
			"--items":
				i += 1
				_items = _split_items(str(args[i]))
		i += 1
	if _output == "":
		_output = ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme/capture.png")


func _split_items(raw: String) -> Array:
	var out: Array = []
	for part in raw.split(",", false):
		var item := str(part).strip_edges()
		if item != "":
			out.append(item)
	return out if not out.is_empty() else DEFAULT_GEAR_ITEMS.duplicate()


func _setup_gear() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackGear"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(2.6, 2.4, 4.0), Vector3(0.0, 0.95, 0.0), 2.6)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(3.0, 0.04, 3.0)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.03, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#3f3f3c")
	floor.material_override = floor_mat
	root.add_child(floor)

	var character := CharacterScene.instantiate() as Node3D
	character.name = "FocusedCharacter"
	character.rotation.y = deg_to_rad(-18.0)
	root.add_child(character)
	_subject = character

	await process_frame
	var resolver = EquipmentResolverScript.new(character)
	resolver.apply_snapshot(_gear_snapshot(_items))


func _setup_inventory() -> void:
	var panel = InventoryPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	var items := _inventory_items()
	panel.set_inventory_state(items, {
		"head": "2001",
		"amulet": "2011",
		"chest": "2002",
		"gloves": null,
		"belt": "2003",
		"boots": "2004",
		"ring_left": null,
		"ring_right": null,
		"main_hand": "2005",
		"off_hand": "2006",
	}, 4, 20, 145)
	panel.ensure_display_visible()
	var tooltip = panel._make_item_tooltip(_inventory_item_by_id(items, "2005"))
	tooltip.position = Vector2(524, 28)
	get_root().add_child(tooltip)


func _setup_floor_item() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackFloorItem"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(2.5, 2.3, 3.2), Vector3(0.0, 0.12, 0.0), 2.1)

	var floor := MeshInstance3D.new()
	floor.name = "reference_grass_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(3.2, 0.035, 2.5)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.03, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#496f3e")
	floor.material_override = floor_mat
	root.add_child(floor)

	var main: Node3D = MainScript.new()
	var base := ProjectSettings.globalize_path("res://")
	var manifest = _read_json(base.path_join("../assets/manifests/assets.v0.json"))
	if typeof(manifest) == TYPE_DICTIONARY:
		main.asset_manifest = manifest.get("assets", {})
	ItemRulesLoader.ensure_loaded()
	var item_def_id := str(_items[0]) if not _items.is_empty() else "cave_blade"
	var loot: Node3D = main._make_loot_node({
		"type": "loot",
		"item_def_id": item_def_id,
		"rarity": "magic",
		"amount": 1,
	})
	loot.name = "PreviewFloorItem_%s" % item_def_id
	root.add_child(loot)
	_subject = loot


func _setup_skills() -> void:
	var panel = SkillsPanelScript.new()
	_skills_panel = panel
	get_root().add_child(panel)
	await process_frame
	panel.set_character_progression({
		"level": 1,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 15},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": [
			{"skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": true},
			{"skill_id": "rage", "rank": 0, "max_rank": 5, "can_spend": false},
			{"skill_id": "heal", "rank": 0, "max_rank": 5, "can_spend": true},
		],
	})
	panel.ensure_display_visible()
	panel.bot_hover_skill("rage")


func _setup_shop() -> void:
	var offers := _shop_offers()
	var inventory_panel = InventoryPanelScript.new()
	get_root().add_child(inventory_panel)
	await process_frame
	inventory_panel.set_inventory_state(_shop_inventory(), {}, 3, 15, 550)
	inventory_panel.set_shop_sell_context("1004")
	inventory_panel.ensure_display_visible()

	var panel = ShopPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.show_shop("1004", "town_vendor", offers, 550, _shop_inventory(), {}, "Town Vendor", _shop_sell_appraisals())
	var tooltip = panel._make_offer_tooltip(offers[3])
	tooltip.position = Vector2(286, 142)
	get_root().add_child(tooltip)


func _setup_market_board() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackMarketBoard"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(2.8, 2.8, 4.5), Vector3(0.0, 0.82, 0.0), 2.4)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(3.0, 0.04, 2.2)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#353735")
	floor.material_override = floor_mat
	root.add_child(floor)

	var main: Node3D = MainScript.new()
	var board := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_market_board"}) as Node3D
	board.name = "PreviewMarketBoard"
	root.add_child(board)
	var incoming := board.find_child("IncomingBidCount", true, false) as Label3D
	var published := board.find_child("PublishedListingCount", true, false) as Label3D
	if incoming != null:
		incoming.text = "3"
		incoming.modulate = Color("#ffcf5a")
	if published != null:
		published.text = "2"
		published.modulate = Color("#9fd7ff")
	_subject = board


func _setup_market_publish() -> void:
	var panel = MarketPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.show_market("1009", _market_listings(), _market_stash_items(), "acct_player", "Publish from account stash")
	panel.bot_select_tab("publish")


func _setup_market_offer() -> void:
	var panel = MarketPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.show_market("1009", _market_listings(), _market_stash_items(), "acct_player", "Choose a stash item to offer")
	panel.bot_select_tab("offer")


func _setup_character_menu() -> void:
	var panel: CharacterSelectPanel = CharacterSelectPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.show_choose_or_create([
		{"character_id": "char_1", "name": "Astra", "created_at": "2026-06-09T00:00:00Z", "dead": false, "level": 4, "gold": 128, "deepest_dungeon_depth": 3},
		{"character_id": "char_2", "name": "Fallen", "created_at": "2026-06-08T00:00:00Z", "dead": true, "level": 2, "gold": 44, "deepest_dungeon_depth": 1},
	], "Choose Character")
	panel.submit_name()


func _setup_join_menu() -> void:
	var panel: MultiplayerSessionsPanel = MultiplayerSessionsPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.show_panel()
	panel.set_sessions([
		{"session_id": "sess_1", "host_display_name": "Astra", "connected_count": 1, "member_count": 4, "world_id": "dungeon_levels", "mode": "coop", "listed": true},
		{"session_id": "sess_2", "host_display_name": "Bram", "connected_count": 2, "member_count": 4, "world_id": "dungeon_levels", "mode": "coop", "listed": true},
	])


func _setup_hud() -> void:
	var panel: PlayerHealthBar = PlayerHealthBarScript.new()
	panel.set_identity("Astra", 4)
	panel.update_hp(9, 12)
	panel.update_mana(7, 14)
	get_root().add_child(panel)


func _setup_stairs() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackStairs"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(3.3, 3.0, 4.5), Vector3(0.0, 0.40, 0.0), 3.2)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(4.2, 0.04, 2.6)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#353735")
	floor.material_override = floor_mat
	root.add_child(floor)

	var main: Node3D = MainScript.new()
	var up := main._make_stair_node("stairs_up") as Node3D
	up.name = "PreviewStairsUp"
	up.position = Vector3(-1.0, 0.0, 0.0)
	root.add_child(up)
	var down := main._make_stair_node("stairs_down") as Node3D
	down.name = "PreviewStairsDown"
	down.position = Vector3(1.0, 0.0, 0.0)
	root.add_child(down)


func _setup_chests() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackChests"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(3.4, 2.6, 4.4), Vector3(0.0, 0.42, 0.0), 3.6)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(4.4, 0.04, 2.4)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#353735")
	floor.material_override = floor_mat
	root.add_child(floor)

	var main: Node3D = MainScript.new()
	var stash := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_stash"}) as Node3D
	stash.name = "PreviewTownStash"
	stash.position = Vector3(-1.0, 0.0, 0.0)
	root.add_child(stash)
	var chest := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest"}) as Node3D
	chest.name = "PreviewTreasureChest"
	chest.position = Vector3(1.0, 0.0, 0.0)
	root.add_child(chest)
	_subject = root


func _setup_vendors() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackVendors"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(3.4, 2.8, 4.6), Vector3(0.0, 0.70, 0.0), 3.7)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(4.4, 0.04, 2.4)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#353735")
	floor.material_override = floor_mat
	root.add_child(floor)

	var main: Node3D = MainScript.new()
	var vendor := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_vendor"}) as Node3D
	vendor.name = "PreviewTownVendor"
	vendor.position = Vector3(-1.0, 0.0, 0.0)
	root.add_child(vendor)
	var mystery := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_mystery_seller"}) as Node3D
	mystery.name = "PreviewMysterySeller"
	mystery.position = Vector3(1.0, 0.0, 0.0)
	root.add_child(mystery)
	_subject = root


func _setup_monsters() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackMonsters"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(4.2, 3.0, 5.3), Vector3(0.0, 0.70, 0.0), 4.8)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(6.2, 0.04, 2.6)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#353735")
	floor.material_override = floor_mat
	root.add_child(floor)

	var main: Node3D = MainScript.new()
	var entries := [
		{"name": "Dummy", "x": -2.25, "entity": {"type": "monster", "monster_def_id": "dungeon_mob", "rarity": "common"}},
		{"name": "Wolf", "x": -0.75, "entity": {"type": "monster", "monster_def_id": "dungeon_wolf", "rarity": "common"}},
		{"name": "Bat", "x": 0.75, "entity": {"type": "monster", "monster_def_id": "dungeon_bat", "rarity": "common"}},
		{"name": "Boss", "x": 2.25, "entity": {"type": "monster", "monster_def_id": "dungeon_mob", "rarity": "unique", "is_boss": true, "boss_template_id": "cave_warden", "visual_model": "monster_tiny_flyer", "visual_scale": 2.0, "visual_tint": "#b77cff"}},
	]
	for entry in entries:
		var monster := main._make_entity_node(entry["entity"] as Dictionary) as Node3D
		monster.name = "Preview%s" % str(entry["name"])
		monster.position = Vector3(float(entry["x"]), 0.0, 0.0)
		root.add_child(monster)
	_subject = root


func _setup_classes() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackClasses"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(4.5, 3.2, 5.4), Vector3(0.0, 1.05, 0.0), 4.7)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(5.8, 0.04, 2.8)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#383936")
	floor.material_override = floor_mat
	root.add_child(floor)

	var entries := [
		{"class_id": "barbarian", "label": "Barbarian", "x": -1.75},
		{"class_id": "sorcerer", "label": "Sorcerer", "x": 0.0},
		{"class_id": "paladin", "label": "Paladin", "x": 1.75},
	]
	for entry in entries:
		var packed := ClassPresentationsLoaderScript.packed_scene_for_class(str(entry["class_id"]))
		var model := packed.instantiate() as Node3D
		model.name = "Preview%s" % str(entry["label"])
		model.position = Vector3(float(entry["x"]), 0.0, 0.0)
		model.rotation.y = deg_to_rad(18.0)
		root.add_child(model)
		var label := Label3D.new()
		label.name = "%sLabel" % str(entry["label"])
		label.text = str(entry["label"])
		label.position = Vector3(float(entry["x"]), 2.85, 0.0)
		label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
		label.font_size = 42
		label.modulate = Color("#f3efe5")
		root.add_child(label)
	_subject = root


func _setup_heal_rain() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackHealRain"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(5.4, 5.0, 7.1), Vector3(0.0, 0.9, 0.0), 8.3)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(9.2, 0.04, 5.2)
	floor.mesh = floor_mesh
	floor.position = Vector3(0, -0.025, 0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#303631")
	floor.material_override = floor_mat
	root.add_child(floor)

	var character_mat := StandardMaterial3D.new()
	character_mat.albedo_color = Color(0.55, 0.72, 0.62, 0.58)
	character_mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	for i in range(5):
		var marker := _make_heal_target_marker(character_mat)
		marker.position = Vector3(-2.8 + float(i) * 1.4, 0.0, 0.0)
		root.add_child(marker)

	var effect = HealRainEffectScript.new()
	effect.setup(MainScript.HEAL_RAIN_RADIUS)
	effect.position = Vector3.ZERO
	root.add_child(effect)
	_subject = root


func _setup_town() -> void:
	var root := Node3D.new()
	root.name = "VisualFeedbackTown"
	get_root().add_child(root)

	_add_light(root)
	_add_camera(root, Vector3(20.0, 18.0, 25.0), Vector3(12.0, 0.8, 12.0), 17.5)

	var main: Node3D = MainScript.new()
	var town: Node3D = main.make_town_preview_scene()
	root.add_child(town)
	_subject = town


func _make_heal_target_marker(mat: StandardMaterial3D) -> Node3D:
	var root := Node3D.new()
	var body := MeshInstance3D.new()
	var body_mesh := CapsuleMesh.new()
	body_mesh.radius = 0.22
	body_mesh.height = 1.25
	body.mesh = body_mesh
	body.material_override = mat
	body.position.y = 0.68
	root.add_child(body)
	var head := MeshInstance3D.new()
	var head_mesh := SphereMesh.new()
	head_mesh.radius = 0.22
	head.mesh = head_mesh
	head.material_override = mat
	head.position.y = 1.45
	root.add_child(head)
	return root


func _add_light(root: Node3D) -> void:
	var light := DirectionalLight3D.new()
	light.name = "key_light"
	light.light_energy = 2.2
	light.rotation_degrees = Vector3(-55, -35, 0)
	root.add_child(light)


func _add_camera(root: Node3D, position: Vector3, target: Vector3, size: float) -> void:
	var camera := Camera3D.new()
	camera.name = "capture_camera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = size
	root.add_child(camera)
	camera.look_at_from_position(position, target, Vector3.UP)
	camera.current = true


func _tint_node(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		var mesh := root as MeshInstance3D
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		mat.roughness = 0.82
		mesh.material_override = mat
	for child in root.get_children():
		_tint_node(child, color)


func _read_json(path: String):
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return null
	return JSON.parse_string(f.get_as_text())


func _gear_snapshot(items: Array) -> Dictionary:
	var inventory := []
	var equipped := _empty_equipped()
	var next_id := 2001
	var ring_count := 0
	for item_def_id in items:
		var slot := str(ITEM_SLOT.get(item_def_id, ""))
		if slot == "":
			continue
		if item_def_id == "cave_ring":
			slot = "ring_right" if ring_count > 0 else "ring_left"
			ring_count += 1
		var iid := str(next_id)
		next_id += 1
		inventory.append({
			"item_instance_id": iid,
			"item_def_id": item_def_id,
			"slot": slot,
			"equipped": true,
			"rarity": "rare" if item_def_id in ["cave_helm", "cave_mail"] else "magic",
		})
		equipped[slot] = iid
	return {"inventory": inventory, "equipped": equipped}


func _inventory_items() -> Array:
	return [
		{"item_instance_id": "2001", "item_def_id": "cave_helm", "slot": "head", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2002", "item_def_id": "cave_mail", "slot": "chest", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2003", "item_def_id": "cave_pack_belt", "slot": "belt", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2004", "item_def_id": "cave_boots", "slot": "boots", "equipped": true, "rarity": "common"},
		{
			"item_instance_id": "2005",
			"item_def_id": "cave_blade",
			"item_template_id": "cave_blade",
			"display_name": "Magic Cave Blade",
			"slot": "main_hand",
			"equipped": true,
			"rarity": "magic",
			"rolled_stats": {"damage_min": 3, "damage_max": 4, "max_hp": 3},
			"requirements": {"level": 1},
			"requirement_status": [
				{"stat": "level", "required": 1, "current": 1, "met": true}
			],
			"requirements_met": true,
		},
		{"item_instance_id": "2006", "item_def_id": "cave_shield", "slot": "off_hand", "equipped": true, "rarity": "magic"},
		{
			"item_instance_id": "2011",
			"item_def_id": "cave_amulet",
			"item_template_id": "cave_amulet",
			"display_name": "Common Cave Amulet",
			"slot": "amulet",
			"equipped": true,
			"rarity": "common",
			"rolled_stats": {"max_hp": 4, "mana_regen_per_10_seconds": 2},
			"requirements": {"level": 1},
			"requirement_status": [
				{"stat": "level", "required": 1, "current": 1, "met": true}
			],
			"requirements_met": true,
		},
		{"item_instance_id": "2007", "item_def_id": "red_potion", "slot": "", "equipped": false, "rarity": "common"},
		{"item_instance_id": "2008", "item_def_id": "blue_potion", "slot": "", "equipped": false, "rarity": "common"},
		{"item_instance_id": "2009", "item_def_id": "cave_ring", "slot": "ring_left", "equipped": false, "rarity": "rare"},
		{
			"item_instance_id": "2010",
			"item_def_id": "cave_war_sword",
			"item_template_id": "cave_war_sword",
			"display_name": "Rare Cave War Sword",
			"slot": "main_hand",
			"equipped": false,
			"rarity": "rare",
			"rolled_stats": {"damage_min": 6, "damage_max": 11},
			"requirements": {"level": 2, "str": 15},
			"requirement_status": [
				{"stat": "level", "required": 2, "current": 2, "met": true},
				{"stat": "str", "required": 15, "current": 12, "met": false}
			],
			"requirements_met": false,
			"equip_preview": {
				"slot": "main_hand",
				"requirements_met": false,
				"deltas": [
					{"stat": "damage_max", "current": 7, "preview": 11, "delta": 4}
				]
			},
			"summary_lines": ["Slot: main hand", "Damage 6-11"],
		},
	]


func _inventory_item_by_id(items: Array, item_instance_id: String) -> Dictionary:
	for item in items:
		if typeof(item) == TYPE_DICTIONARY and str((item as Dictionary).get("item_instance_id", "")) == item_instance_id:
			return item as Dictionary
	return {}


func _shop_offers() -> Array:
	return [
		{
			"offer_id": "fixed:red_potion",
			"kind": "fixed",
			"item_def_id": "red_potion",
			"display_name": "Red Potion",
			"rarity": "common",
			"buy_price": 20,
			"category": "consumable",
			"summary_lines": ["Kind: consumable", "Restores 5 HP"],
		},
		{
			"offer_id": "fixed:blue_potion",
			"kind": "fixed",
			"item_def_id": "blue_potion",
			"display_name": "Blue Potion",
			"rarity": "common",
			"buy_price": 20,
			"category": "consumable",
			"summary_lines": ["Kind: consumable", "Restores 5 mana"],
		},
		{
			"offer_id": "generated:1:cave_helm",
			"kind": "generated",
			"item_def_id": "cave_helm",
			"item_template_id": "cave_helm",
			"display_name": "Magic Cave Helm",
			"rarity": "magic",
			"buy_price": 110,
			"slot": "head",
			"summary_lines": ["Slot: head", "Armor +5", "+0 Min damage vs equipped", "+1 Max damage vs equipped"],
		},
		{
			"offer_id": "generated:1:cave_blade",
			"kind": "generated",
			"item_def_id": "cave_blade",
			"item_template_id": "cave_blade",
			"display_name": "Rare Cave Blade",
			"rarity": "rare",
			"buy_price": 680,
			"slot": "main_hand",
			"summary_lines": ["Slot: main hand", "Damage 3-7", "Requires level 1", "+2 Min damage vs equipped", "+2 Max damage vs equipped", "-1 Armor vs equipped"],
		},
		{
			"offer_id": "generated:1:cave_shield",
			"kind": "generated",
			"item_def_id": "cave_shield",
			"item_template_id": "cave_shield",
			"display_name": "Guarding Cave Shield",
			"rarity": "magic",
			"buy_price": 95,
			"slot": "off_hand",
			"summary_lines": ["Slot: off hand", "Block +8", "+8 Block vs equipped"],
		},
	]


func _shop_inventory() -> Array:
	return [
		{"item_instance_id": "3001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Magic Cave Bow", "slot": "main_hand", "equipped": false, "rarity": "magic", "sell_price": 42},
		{"item_instance_id": "3002", "item_def_id": "cave_mail", "item_template_id": "cave_mail", "display_name": "Rare Cave Mail", "slot": "chest", "equipped": false, "rarity": "rare", "sell_price": 88},
	]


func _shop_sell_appraisals() -> Array:
	return [
		{
			"item_instance_id": "3001",
			"item_def_id": "cave_bow",
			"item_template_id": "cave_bow",
			"display_name": "Magic Cave Bow",
			"rarity": "magic",
			"slot": "main_hand",
			"sell_price": 42,
			"summary_lines": ["Slot: main hand", "Damage 2-5", "+1 Min damage vs equipped"],
		},
		{
			"item_instance_id": "3002",
			"item_def_id": "cave_mail",
			"item_template_id": "cave_mail",
			"display_name": "Rare Cave Mail",
			"rarity": "rare",
			"slot": "chest",
			"sell_price": 88,
			"summary_lines": ["Slot: chest", "Armor +9", "+9 Armor vs equipped"],
		},
	]


func _market_stash_items() -> Array:
	return [
		{"stash_item_id": "stash_9001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Magic Cave Bow", "rarity": "magic", "slot": "main_hand", "summary_lines": ["Slot: main hand", "Damage 2-5"]},
		{"stash_item_id": "stash_9002", "item_def_id": "cave_ring", "item_template_id": "cave_ring", "display_name": "Rare Cave Ring", "rarity": "rare", "slot": "ring", "summary_lines": ["Slot: ring", "Maximum Health +4"]},
		{"stash_item_id": "stash_9003", "item_def_id": "cave_mail", "item_template_id": "cave_mail", "display_name": "Common Cave Mail", "rarity": "common", "slot": "chest", "summary_lines": ["Slot: chest", "Armor +8"]},
	]


func _market_listings() -> Array:
	return [
		{"listing_id": "listing_1001", "seller_account_id": "acct_other", "stash_item_id": "seller_stash_1", "item_def_id": "cave_blade", "display_name": "Rare Cave Blade", "rarity": "rare", "rolled_stats": {"damage_min": 4, "damage_max": 8}},
		{"listing_id": "listing_1002", "seller_account_id": "acct_player", "stash_item_id": "my_stash_1", "item_def_id": "cave_shield", "display_name": "Magic Cave Shield", "rarity": "magic", "rolled_stats": {"block_percent": 8}},
	]


func _empty_equipped() -> Dictionary:
	return {
		"head": null,
		"amulet": null,
		"chest": null,
		"gloves": null,
		"belt": null,
		"boots": null,
		"ring_left": null,
		"ring_right": null,
		"main_hand": null,
		"off_hand": null,
	}
