extends SceneTree
## Single-hero silhouette capture using the same camera controller and zoom data as play.
## Example:
## godot --windowed --single-window --resolution 1120x720 --path client --script res://scripts/showme/class_silhouette_play_camera_capture.gd -- --class-id paladin --loadout class_appropriate --zoom normal --pose idle --output /tmp/paladin.png

const CharacterScene := preload("res://scenes/character.tscn")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const ShowmeGearMatrixCaptureScript := preload("res://scripts/showme/showme_gear_matrix_capture.gd")
const PlayerCameraControllerScript := preload("res://scripts/player_camera_controller.gd")
const PlayerCameraContextScript := preload("res://scripts/player_camera_context.gd")
const CameraPresentationsLoaderScript := preload("res://scripts/camera_presentations_loader.gd")

const CLASS_IDS := ["barbarian", "sorcerer", "paladin", "rogue", "ranger"]

var _class_id := ""
var _loadout := "empty"
var _zoom := "normal"
var _pose := "idle"
var _output := ""
var _grayscale := false


func _initialize() -> void:
	if not _parse_args():
		quit(2)
		return
	# SceneTree._initialize runs before its children can resolve global transforms.
	# Yield once so the play-camera controller can attach and frame live scene nodes.
	await process_frame
	DisplayServer.window_set_size(Vector2i(1120, 720))
	get_root().size = Vector2i(1120, 720)
	var world := Node3D.new()
	world.name = "ClassSilhouettePlayCameraCapture"
	get_root().add_child(world)
	_add_floor_and_light(world)
	var anchor := Node3D.new()
	anchor.name = "PlayerAnchor"
	world.add_child(anchor)
	var character := CharacterScene.instantiate() as Node3D
	character.name = "CharacterVisual"
	anchor.add_child(character)
	if not _apply_class_model(character):
		quit(1)
		return
	var resolver = EquipmentResolverScript.new(character)
	resolver.set_character_class(_class_id)
	resolver.apply_snapshot(_loadout_snapshot())
	var warnings: Array = resolver.get_debug_state().get("warnings", [])
	if not warnings.is_empty():
		push_error("silhouette capture equipment warnings for %s: %s" % [_class_id, warnings])
		quit(1)
		return
	var context = PlayerCameraContextScript.make(anchor, character, null, Callable())
	var camera_controller = PlayerCameraControllerScript.new()
	camera_controller.setup(context, world)
	var camera := camera_controller.get_gameplay_camera()
	var camera_cfg := CameraPresentationsLoaderScript.mode("isometric")
	var zoom_key := "zoom_default" if _zoom == "normal" else "zoom_min"
	if not camera_cfg.has(zoom_key):
		push_error("isometric camera catalog has no zoom value: %s" % zoom_key)
		quit(1)
		return
	camera.size = float(camera_cfg[zoom_key])
	var animation_player := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if animation_player == null or not animation_player.has_animation(_pose):
		push_error("silhouette capture animation is unavailable: %s" % _pose)
		quit(1)
		return
	animation_player.play(_pose)
	await process_frame
	await process_frame
	if _pose == "attack":
		quit(0 if await _save_attack_sequence(animation_player) else 1)
		return
	quit(0 if _save_image(_output) else 1)


func _parse_args() -> bool:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i < args.size():
		var key := str(args[i])
		if key == "--grayscale":
			_grayscale = true
			i += 1
			continue
		if i + 1 >= args.size():
			push_error("missing value for %s" % key)
			return false
		var value := str(args[i + 1])
		match key:
			"--class-id": _class_id = value
			"--loadout": _loadout = value
			"--zoom": _zoom = value
			"--pose": _pose = value
			"--output": _output = value
			_:
				push_error("unknown argument: %s" % key)
				return false
		i += 2
	if not CLASS_IDS.has(_class_id):
		push_error("--class-id must be one of %s" % [CLASS_IDS])
		return false
	if not ["empty", "class_appropriate"].has(_loadout):
		push_error("--loadout must be empty or class_appropriate")
		return false
	if not ["normal", "battle"].has(_zoom):
		push_error("--zoom must be normal or battle")
		return false
	if not ["idle", "walk", "attack", "hit", "death"].has(_pose):
		push_error("--pose must be idle, walk, attack, hit, or death")
		return false
	if not _output.is_absolute_path() or not _output.ends_with(".png"):
		push_error("--output must be an absolute .png path")
		return false
	return true


func _apply_class_model(character: Node3D) -> bool:
	var packed := ClassPresentationsLoaderScript.packed_scene_for_class(_class_id)
	if packed == null:
		push_error("class model is unavailable: %s" % _class_id)
		return false
	var old_model := character.find_child("ModelRoot", false, false) as Node
	if old_model != null:
		character.remove_child(old_model)
		old_model.free()
	var resolved := ClassPresentationsLoaderScript.resolve(_class_id)
	var model := packed.instantiate() as Node3D
	model.name = "ModelRoot"
	model.scale = Vector3.ONE * float(resolved.get("scale", 1.0))
	model.position.y = float(resolved.get("height_offset", 0.0))
	ClassIdleStanceScript.apply_to_model(model, _class_id)
	character.add_child(model)
	character.move_child(model, 0)
	var animation_player := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if animation_player != null:
		animation_player.root_node = NodePath("../ModelRoot")
	if "class_id" in character:
		character.set("class_id", _class_id)
	if character.has_method("refresh_gear_sockets"):
		character.call("refresh_gear_sockets")
	return true


func _loadout_snapshot() -> Dictionary:
	var items: Array = []
	if _loadout == "class_appropriate":
		for entry in ShowmeGearMatrixCaptureScript.DEFAULT_LOADOUTS:
			if str(entry.get("class_id", "")) == _class_id:
				items = entry.get("items", [])
				break
	var inventory: Array = []
	var equipped := {}
	var next_id := 8100
	for item in items:
		var instance_id := str(next_id)
		next_id += 1
		var entry: Dictionary = item.duplicate(true)
		entry["item_instance_id"] = instance_id
		entry["equipped"] = true
		entry["rarity"] = "common"
		inventory.append(entry)
		equipped[str(entry.get("slot", ""))] = instance_id
	return {"inventory": inventory, "equipped": equipped}


func _add_floor_and_light(world: Node3D) -> void:
	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(80.0, 0.08, 80.0)
	floor.mesh = floor_mesh
	floor.position.y = -0.04
	var floor_material := StandardMaterial3D.new()
	floor_material.albedo_color = Color("#383936")
	floor.material_override = floor_material
	world.add_child(floor)
	var light := DirectionalLight3D.new()
	light.name = "key_light"
	light.light_energy = 2.2
	light.rotation_degrees = Vector3(-55.0, -35.0, 0.0)
	world.add_child(light)


func _save_attack_sequence(animation_player: AnimationPlayer) -> bool:
	var length := animation_player.current_animation_length
	if length <= 0.0:
		push_error("attack animation has no duration")
		return false
	for mark in [0.2, 0.5, 0.8]:
		var time := length * float(mark)
		while animation_player.current_animation_position < time:
			await process_frame
		var suffix := "_attack_%d.png" % int(mark * 100.0)
		if not _save_image(_output.trim_suffix(".png") + suffix):
			return false
	return true


func _save_image(path: String) -> bool:
	var image := get_root().get_texture().get_image()
	if image == null:
		push_error("silhouette capture viewport image is missing")
		return false
	if _grayscale:
		for y in range(image.get_height()):
			for x in range(image.get_width()):
				var pixel := image.get_pixel(x, y)
				var value := pixel.r * 0.2126 + pixel.g * 0.7152 + pixel.b * 0.0722
				image.set_pixel(x, y, Color(value, value, value, pixel.a))
	var directory := path.get_base_dir()
	if not DirAccess.dir_exists_absolute(directory):
		var make_error := DirAccess.make_dir_recursive_absolute(directory)
		if make_error != OK:
			push_error("silhouette capture cannot create output directory: %s (%d)" % [directory, make_error])
			return false
	var err := image.save_png(path)
	if err != OK:
		push_error("silhouette capture save failed: %s (%d)" % [path, err])
		return false
	else:
		print("[class-silhouette-capture] saved %s" % path)
	return true
