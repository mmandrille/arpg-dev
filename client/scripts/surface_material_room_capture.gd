extends SceneTree

const DungeonSurfaceDetailPresentationScript := preload("res://scripts/dungeon_surface_detail_presentation.gd")
const DungeonTorchLightsScript := preload("res://scripts/dungeon_torch_lights.gd")
const SceneLightingRigScript := preload("res://scripts/scene_lighting_rig.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const DressingScript := preload("res://scripts/dungeon_room_dressing.gd")
const KitLoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")

var _output: String = ""
var _level: int = -1
var _quality: String = "balanced"
var _dressing_enabled: bool = true


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
	var walls := renderer.render_wall_layout(_sample_material_layout(), "showme|%d" % _level if _dressing_enabled else "")
	if _dressing_enabled:
		var cfg := KitLoaderScript.dressing_config()
		var planned := DressingScript.plan(walls, "showme|%d" % _level, _level, [], cfg)
		print("[surface-material-capture] safe_candidates=%d prop_radius=%.2f reason=%s" % [int(planned["safe_candidates"]), DressingScript._max_prop_radius(cfg["props"]), str(planned["reason"])])
	DungeonSurfaceDetailPresentationScript.sync(ground, walls_root, factory, _level, walls, {})
	world.add_child(_make_camera())
	# Runtime lighting path (ADR-0018 D9): the same SceneLightingRig main.gd uses, plus wall torches.
	var lighting = SceneLightingRigScript.new()
	lighting.attach(world)
	lighting.sync(_level, factory, _quality)
	var torches = DungeonTorchLightsScript.new(world, null, factory, renderer)
	torches.sync(_level, walls, true)
	await process_frame
	await process_frame
	await process_frame
	var dressing := walls_root.get_node_or_null("DungeonRoomDressing")
	print("[surface-material-capture] dressing_instances=%d draw_calls=%d" % [int(dressing.get_meta("placement_count", 0)) if dressing != null else 0, int(Performance.get_monitor(Performance.RENDER_TOTAL_DRAW_CALLS_IN_FRAME))])
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
		if arg == "--quality" and index + 1 < args.size():
			_quality = str(args[index + 1])
			index += 2
			continue
		if arg == "--dressing" and index + 1 < args.size():
			_dressing_enabled = str(args[index + 1]) != "off"
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
		{"id": "room_north", "position": {"x": 14.0, "y": 3.0}, "size": {"x": 26.0, "y": 1.0}, "source": "room_wall"},
		{"id": "room_west", "position": {"x": 1.5, "y": 14.0}, "size": {"x": 1.0, "y": 22.0}, "source": "room_wall"},
		{"id": "room_east", "position": {"x": 26.5, "y": 14.0}, "size": {"x": 1.0, "y": 22.0}, "source": "room_wall"},
		{"id": "room_south_left", "position": {"x": 8.0, "y": 25.0}, "size": {"x": 13.0, "y": 1.0}, "source": "room_wall"},
		{"id": "room_south_right", "position": {"x": 22.0, "y": 25.0}, "size": {"x": 9.0, "y": 1.0}, "source": "room_wall"},
		{"id": "center_column", "position": {"x": 8.0, "y": 11.0}, "size": {"x": 1.2, "y": 3.4}, "source": "generated", "kind": "column"},
		{"id": "water_pool", "position": {"x": 18.0, "y": 17.0}, "size": {"x": 4.2, "y": 2.4}, "source": "generated", "kind": "water"},
	]


func _make_camera() -> Camera3D:
	var camera := Camera3D.new()
	camera.current = true
	camera.fov = 38.0
	camera.look_at_from_position(Vector3(22.0, 18.0, 38.0), Vector3(14.0, 1.2, 14.0), Vector3.UP)
	return camera
