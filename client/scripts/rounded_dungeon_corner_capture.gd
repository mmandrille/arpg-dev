extends SceneTree

const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const DungeonWallCornerPresentationScript := preload("res://scripts/dungeon_wall_corner_presentation.gd")

var _output: String = ""
var _style: String = DungeonWallCornerPresentationScript.STYLE_ROUNDED


func _initialize() -> void:
	_parse_args()
	var world := Node3D.new()
	get_root().add_child(world)
	var factory = GroundWallFactoryScript.new()
	var ground := factory.make_ground_node(-1)
	world.add_child(ground)
	var walls_root := Node3D.new()
	walls_root.name = "WallsRoot"
	world.add_child(walls_root)
	var renderer = WallRendererScript.new(walls_root, factory)
	renderer.set_level(-1)
	renderer.set_corner_review_style(_style)
	renderer.render_wall_layout(_sample_room_wall_layout())
	world.add_child(_make_camera())
	world.add_child(_make_key_light())
	world.add_child(_make_fill_light())
	await process_frame
	await process_frame
	var image := get_root().get_texture().get_image()
	if image == null:
		push_error("rounded corner capture failed: viewport image missing")
		quit(1)
		return
	var dir := _output.get_base_dir()
	if dir != "":
		DirAccess.make_dir_recursive_absolute(dir)
	var err := image.save_png(_output)
	if err != OK:
		push_error("rounded corner capture failed saving %s (err=%d)" % [_output, err])
		quit(1)
		return
	print("[rounded-corner-capture] screenshot: %s" % _output)
	quit(0)


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var index := 0
	while index < args.size():
		var arg := str(args[index])
		match arg:
			"--output":
				if index + 1 < args.size():
					_output = _resolve_output(str(args[index + 1]))
					index += 2
					continue
			"--style":
				if index + 1 < args.size():
					_style = DungeonWallCornerPresentationScript._validated_style(str(args[index + 1]))
					index += 2
					continue
		index += 1
	if _output == "":
		var root := ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme")
		_output = root.path_join("rounded-dungeon-corner-%s.png" % _style)


func _resolve_output(raw: String) -> String:
	if raw.is_absolute_path():
		return raw
	return ProjectSettings.globalize_path("res://").path_join("../").path_join(raw)


func _sample_room_wall_layout() -> Array:
	return [
		{"id": "room_north", "position": {"x": 6.0, "y": 2.0}, "size": {"x": 10.0, "y": 1.0}, "source": "room_wall"},
		{"id": "room_west", "position": {"x": 1.0, "y": 5.5}, "size": {"x": 1.0, "y": 7.0}, "source": "room_wall"},
		{"id": "room_east", "position": {"x": 11.0, "y": 5.5}, "size": {"x": 1.0, "y": 7.0}, "source": "room_wall"},
		{"id": "room_south_left", "position": {"x": 3.0, "y": 9.0}, "size": {"x": 4.0, "y": 1.0}, "source": "room_wall"},
		{"id": "room_south_right", "position": {"x": 9.0, "y": 9.0}, "size": {"x": 4.0, "y": 1.0}, "source": "room_wall"},
	]


func _make_camera() -> Camera3D:
	var camera := Camera3D.new()
	camera.current = true
	camera.fov = 38.0
	camera.look_at_from_position(Vector3(9.2, 8.4, 17.0), Vector3(6.0, 1.8, 6.2), Vector3.UP)
	return camera


func _make_key_light() -> DirectionalLight3D:
	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-48.0, 24.0, 0.0)
	light.light_energy = 1.85
	light.light_color = Color("#ffe0bd")
	light.shadow_enabled = true
	return light


func _make_fill_light() -> OmniLight3D:
	var light := OmniLight3D.new()
	light.position = Vector3(5.6, 3.2, 8.6)
	light.light_energy = 0.82
	light.omni_range = 30.0
	light.light_color = Color("#8aa0c5")
	return light
