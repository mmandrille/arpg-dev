class_name ShowmeGearMatrixCapture
extends RefCounted

const CharacterScene := preload("res://scenes/character.tscn")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const ClassBodyTintScript := preload("res://scripts/class_body_tint.gd")
const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")

# Default debug matrix: one class per column, distinct gear per class.
const DEFAULT_LOADOUTS := [
	{
		"class_id": "paladin",
		"title": "Paladin",
		"items": [
			{"item_def_id": "long_sword", "slot": "main_hand"},
			{"item_def_id": "shield", "slot": "off_hand"},
			{"item_def_id": "helm", "slot": "head"},
			{"item_def_id": "mail", "slot": "chest"},
			{"item_def_id": "boots", "slot": "boots"},
		],
	},
	{
		"class_id": "ranger",
		"title": "Ranger",
		"items": [
			{"item_def_id": "training_bow", "slot": "main_hand"},
			{"item_def_id": "leather_cap", "slot": "head"},
			{"item_def_id": "leather_vest", "slot": "chest"},
			{"item_def_id": "gloves", "slot": "gloves"},
		],
	},
	{
		"class_id": "rogue",
		"title": "Rogue",
		"items": [
			{"item_def_id": "dagger", "slot": "main_hand"},
			{"item_def_id": "dagger", "slot": "off_hand"},
			{"item_def_id": "cloth_wraps", "slot": "gloves"},
		],
	},
	{
		"class_id": "sorcerer",
		"title": "Sorcerer",
		"items": [
			{"item_def_id": "sorcerer_staff", "slot": "main_hand"},
			{"item_def_id": "tiara", "slot": "head"},
		],
	},
	{
		"class_id": "barbarian",
		"title": "Barbarian",
		"items": [
			{"item_def_id": "barbarian_axe", "slot": "main_hand"},
			{"item_def_id": "gauntlets", "slot": "gloves"},
			{"item_def_id": "war_girdle", "slot": "belt"},
		],
	},
]

const COLUMN_X := [-3.5, -1.75, 0.0, 1.75, 3.5]


static func setup(capture: SceneTree) -> Node3D:
	var root := Node3D.new()
	root.name = "VisualFeedbackGearMatrix"
	capture.get_root().add_child(root)

	_add_light(root)
	_add_camera(root)

	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(8.4, 0.04, 3.2)
	floor.mesh = floor_mesh
	floor.position = Vector3(0.0, -0.03, 0.0)
	var floor_mat := StandardMaterial3D.new()
	floor_mat.albedo_color = Color("#383936")
	floor.material_override = floor_mat
	root.add_child(floor)

	for i in range(DEFAULT_LOADOUTS.size()):
		var entry: Dictionary = DEFAULT_LOADOUTS[i]
		var class_id := str(entry.get("class_id", ""))
		var x := float(COLUMN_X[i]) if i < COLUMN_X.size() else float(i) * 1.75 - 3.5
		var character := CharacterScene.instantiate() as Node3D
		character.name = "GearMatrix_%s" % class_id
		character.position = Vector3(x, 0.0, 0.0)
		character.rotation.y = deg_to_rad(18.0)
		root.add_child(character)

		await capture.process_frame
		_apply_class_model(character, class_id)
		var resolver = EquipmentResolverScript.new(character)
		resolver.set_character_class(class_id)
		await capture.process_frame
		await capture.process_frame
		resolver.apply_snapshot(_snapshot_for_loadout(entry.get("items", [])))
		var ap := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
		if ap != null:
			ap.play("idle")

		var label := Label3D.new()
		label.name = "%sGearLabel" % class_id
		label.text = _label_text(entry)
		label.position = Vector3(x, 0.12, 0.72)
		label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
		label.font_size = 30
		label.modulate = Color("#f3efe5")
		label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		root.add_child(label)

	await capture.process_frame
	await capture.process_frame
	return root


static func _label_text(entry: Dictionary) -> String:
	var title := str(entry.get("title", entry.get("class_id", "Class")))
	var items: Array = entry.get("items", [])
	var names: PackedStringArray = []
	for item_entry in items:
		if typeof(item_entry) != TYPE_DICTIONARY:
			continue
		names.append(str(item_entry.get("item_def_id", "")))
	return "%s\n%s" % [title, ", ".join(names)]


static func _snapshot_for_loadout(items: Array) -> Dictionary:
	var inventory: Array = []
	var equipped := {
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
	var next_id := 3001
	for item_entry in items:
		if typeof(item_entry) != TYPE_DICTIONARY:
			continue
		var item_def_id := str(item_entry.get("item_def_id", ""))
		var slot := str(item_entry.get("slot", ""))
		if item_def_id == "" or slot == "":
			continue
		var iid := str(next_id)
		next_id += 1
		inventory.append({
			"item_instance_id": iid,
			"item_def_id": item_def_id,
			"slot": slot,
			"equipped": true,
			"rarity": "common",
		})
		equipped[slot] = iid
	return {"inventory": inventory, "equipped": equipped}


static func _apply_class_model(character: Node3D, class_id: String) -> void:
	var resolved := ClassPresentationsLoaderScript.resolve(class_id)
	var packed := ClassPresentationsLoaderScript.packed_scene_for_class(class_id)
	if packed == null:
		return
	var old_model := character.find_child("ModelRoot", false, false) as Node
	if old_model != null:
		character.remove_child(old_model)
		old_model.free()
	var model := packed.instantiate() as Node3D
	model.name = "ModelRoot"
	model.scale = Vector3.ONE * float(resolved.get("scale", 1.0))
	model.position.y = float(resolved.get("height_offset", 0.0))
	ClassIdleStanceScript.apply_to_model(model, class_id)
	character.add_child(model)
	character.move_child(model, 0)
	var ap := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.root_node = NodePath("../ModelRoot")
	if "class_id" in character:
		character.set("class_id", class_id)
	if character.has_method("refresh_gear_sockets"):
		character.call("refresh_gear_sockets")
	ClassBodyTintScript.apply_to_model(model, class_id)


static func _add_light(root: Node3D) -> void:
	var light := DirectionalLight3D.new()
	light.name = "key_light"
	light.light_energy = 2.2
	light.rotation_degrees = Vector3(-55, -35, 0)
	root.add_child(light)


static func _add_camera(root: Node3D) -> void:
	var camera := Camera3D.new()
	camera.name = "capture_camera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = 5.2
	root.add_child(camera)
	camera.look_at_from_position(Vector3(4.8, 3.4, 6.2), Vector3(0.0, 1.0, 0.35), Vector3.UP)
	camera.current = true
