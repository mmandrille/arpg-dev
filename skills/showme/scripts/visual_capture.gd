extends SceneTree

const CharacterScene := preload("res://scenes/character.tscn")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")

const DEFAULT_GEAR_ITEMS := ["cave_blade", "cave_shield", "cave_helm", "cave_mail", "cave_boots"]
const ITEM_SLOT := {
	"rusty_sword": "main_hand",
	"cave_blade": "main_hand",
	"cave_greatsword": "main_hand",
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

	if _focus == "inventory":
		await _setup_inventory()
	else:
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
	panel.set_inventory_state(_inventory_items(), {
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
