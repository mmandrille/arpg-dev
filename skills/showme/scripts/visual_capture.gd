extends SceneTree

const CharacterScene := preload("res://scenes/character.tscn")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const MultiplayerSessionsPanelScript := preload("res://scripts/multiplayer_sessions_panel.gd")
const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const MainScript := preload("res://scripts/main.gd")

const DEFAULT_GEAR_ITEMS := ["cave_blade", "cave_shield", "cave_helm", "cave_mail", "cave_boots"]
const ITEM_SLOT := {
	"rusty_sword": "main_hand",
	"cave_blade": "main_hand",
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


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)

	match _focus:
		"character-menu":
			await _setup_character_menu()
		"join-menu":
			await _setup_join_menu()
		"hud":
			await _setup_hud()
		"stairs":
			await _setup_stairs()
		"inventory":
			await _setup_inventory()
		"shop":
			await _setup_shop()
		_:
			await _setup_gear()

	for _i in range(8):
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
		"amulet": null,
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
	var tooltip = panel._make_item_tooltip(items[6])
	tooltip.position = Vector2(524, 28)
	get_root().add_child(tooltip)


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
		{"item_instance_id": "2005", "item_def_id": "cave_blade", "slot": "main_hand", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2006", "item_def_id": "cave_shield", "slot": "off_hand", "equipped": true, "rarity": "magic"},
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
