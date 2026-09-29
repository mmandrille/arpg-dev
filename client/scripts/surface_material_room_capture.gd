extends SceneTree

const DungeonDepthLightingScript := preload("res://scripts/dungeon_depth_lighting.gd")
const DungeonSurfaceDetailPresentationScript := preload("res://scripts/dungeon_surface_detail_presentation.gd")
const DungeonTorchLightsScript := preload("res://scripts/dungeon_torch_lights.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")

var _output: String = ""
var _level: int = -1


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(900, 620))
	get_root().size = Vector2i(900, 620)
	var world := Node3D.new()
	get_root().add_child(world)
	var factory = GroundWallFactoryScript.new()
	var ground := factory.make_ground_node(_level)
	world.add_child(ground)
	var walls_root := Node3D.new()
	walls_root.name = "WallsRoot"
	world.add_child(walls_root)
	var renderer = WallRendererScript.new(walls_root, factory)
	renderer.set_level(_level)
	var walls := renderer.render_wall_layout(_sample_material_layout())
	DungeonSurfaceDetailPresentationScript.sync(ground, walls_root, factory, _level, walls, {})
	world.add_child(_make_camera())
	# Runtime lighting path (ADR-0018 D9 baseline): biome depth profile + wall torches.
	# Key-light angle mirrors main.gd _build_scene until P1 moves it into render data.
	var key_light := DirectionalLight3D.new()
	key_light.rotation_degrees = Vector3(-50.0, -40.0, 0.0)
	world.add_child(key_light)
	var world_environment := WorldEnvironment.new()
	world.add_child(world_environment)
	DungeonDepthLightingScript.apply_for_level(_level, key_light, world_environment, factory)
	var torches = DungeonTorchLightsScript.new(world, null, factory, renderer)
	torches.sync(_level, walls, true)
	await process_frame
	await process_frame
	await process_frame
	var image := get_root().get_texture().get_image()
	if image == null:
		push_error("surface material capture failed: viewport image missing")
		quit(1)
		return
	var dir := _output.get_base_dir()
	if dir != "":
		DirAccess.make_dir_recursive_absolute(dir)
	var err := image.save_png(_output)
	if err != OK:
		push_error("surface material capture failed saving %s (err=%d)" % [_output, err])
		quit(1)
		return
	print("[surface-material-capture] screenshot: %s" % _output)
	quit(0)


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var index := 0
	while index < args.size():
		var arg := str(args[index])
		if arg == "--output" and index + 1 < args.size():
			_output = _resolve_output(str(args[index + 1]))
			index += 2
			continue
		if arg == "--level" and index + 1 < args.size():
			_level = int(str(args[index + 1]))
			index += 2
			continue
		index += 1
	if _output == "":
		var root := ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme")
		_output = root.path_join("surface-material-kit-room.png")


func _resolve_output(raw: String) -> String:
	if raw.is_absolute_path():
		return raw
	return ProjectSettings.globalize_path("res://").path_join("../").path_join(raw)


func _sample_material_layout() -> Array:
	return [
		{"id": "room_north", "position": {"x": 8.0, "y": 3.0}, "size": {"x": 14.0, "y": 1.0}, "source": "room_wall"},
		{"id": "room_west", "position": {"x": 1.5, "y": 8.0}, "size": {"x": 1.0, "y": 10.0}, "source": "room_wall"},
		{"id": "room_east", "position": {"x": 14.5, "y": 8.0}, "size": {"x": 1.0, "y": 10.0}, "source": "room_wall"},
		{"id": "room_south_left", "position": {"x": 4.0, "y": 13.0}, "size": {"x": 5.0, "y": 1.0}, "source": "room_wall"},
		{"id": "room_south_right", "position": {"x": 12.0, "y": 13.0}, "size": {"x": 5.0, "y": 1.0}, "source": "room_wall"},
		{"id": "center_column", "position": {"x": 5.0, "y": 7.0}, "size": {"x": 1.2, "y": 3.4}, "source": "generated", "kind": "column"},
		{"id": "water_pool", "position": {"x": 10.4, "y": 8.8}, "size": {"x": 4.2, "y": 2.4}, "source": "generated", "kind": "water"},
	]


func _make_camera() -> Camera3D:
	var camera := Camera3D.new()
	camera.current = true
	camera.fov = 38.0
	camera.look_at_from_position(Vector3(12.5, 9.8, 22.0), Vector3(8.0, 1.2, 8.1), Vector3.UP)
	return camera
