extends SceneTree

const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")
const CameraPresentationsLoaderScript := preload("res://scripts/camera_presentations_loader.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const SceneLightingRigScript := preload("res://scripts/scene_lighting_rig.gd")

const VIEWS := {
	"plaza": Vector2(11.0, 12.0),
	"vendor": Vector2(19.0, 12.0),
	"gate": Vector2(11.0, 24.0),
	"west": Vector2(-1.0, 12.0),
	"north": Vector2(11.0, 0.0),
}

var _output := ""
var _view := "plaza"
var _zoom := "normal"
var _quality := "balanced"
var _width := 1120
var _height := 720
var _mode := "screenshot"
var _duration := 45.0


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)
	var world := Node3D.new()
	world.name = "TownPlayCapture"
	get_root().add_child(world)
	var town := TownNodeFactoryScript.make_town_preview_scene()
	world.add_child(town)
	_use_live_ground(town)
	_add_palisade(world)
	_add_camera(world)
	var lighting = SceneLightingRigScript.new()
	lighting.attach(world)
	lighting.sync(0, GroundWallFactoryScript.new(), _quality)
	for _i in range(8):
		await process_frame
	if _mode == "live":
		if _duration > 0.0:
			await create_timer(_duration).timeout
			quit(0)
		return
	var image := get_root().get_texture().get_image()
	if image == null:
		push_error("town play capture: viewport image missing")
		quit(1)
		return
	var err := image.save_png(_output)
	if err != OK:
		push_error("town play capture: save failed %s (%d)" % [_output, err])
		quit(1)
		return
	print("[town-play-capture] screenshot: %s" % _output)
	quit(0)


func _add_palisade(world: Node3D) -> void:
	var rules_path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/worlds.v0.json")
	var data = JSON.parse_string(FileAccess.get_file_as_string(rules_path))
	if typeof(data) != TYPE_DICTIONARY:
		push_error("town play capture: world preset missing")
		return
	var preset: Dictionary = (data.get("worlds", {}) as Dictionary).get("dungeon_levels", {})
	var walls: Array = []
	for entity in preset.get("entities", []):
		if str(entity.get("type", "")) != "wall":
			continue
		walls.append({"position": entity.get("position", {}), "size": entity.get("size", {}), "kind": entity.get("kind", "wall"), "source": "town_perimeter"})
	var root := Node3D.new()
	root.name = "TownPalisade"
	world.add_child(root)
	var renderer = WallRendererScript.new(root, GroundWallFactoryScript.new())
	renderer.set_level(0)
	renderer.render_wall_layout(walls)


func _use_live_ground(town: Node3D) -> void:
	var preview := town.find_child("TownPreviewGround", false, false) as MeshInstance3D
	if preview == null:
		return
	var live := GroundWallFactoryScript.new().make_ground_node(0)
	preview.mesh = live.mesh
	preview.position = live.position
	preview.material_override = live.material_override
	live.free()


func _add_camera(world: Node3D) -> void:
	var point: Vector2 = VIEWS.get(_view, VIEWS["plaza"])
	var target := Vector3(point.x, 0.0, point.y)
	var cfg := CameraPresentationsLoaderScript.mode("isometric")
	var camera := Camera3D.new()
	camera.name = "PlayerCamera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = float(cfg.get("zoom_max" if _zoom == "max" else "zoom_default", 12.0))
	camera.current = true
	world.add_child(camera)
	camera.look_at_from_position(target + Vector3(9.0, 20.0, 15.0), target, Vector3.UP)


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i + 1 < args.size():
		var key := str(args[i])
		var value := str(args[i + 1])
		match key:
			"--output": _output = value
			"--town-view": _view = value
			"--town-zoom": _zoom = value
			"--quality": _quality = value
			"--width": _width = int(value)
			"--height": _height = int(value)
			"--mode": _mode = value
			"--duration": _duration = float(value)
		i += 2
	if _output == "":
		_output = ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme/town-play.png")
