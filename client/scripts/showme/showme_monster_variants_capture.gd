## v511 monster variant matrix: families (columns) x rarities (rows) built through the real
## main.gd node path at one dungeon depth, on the playable isometric camera and client lighting.
## `--level -5` selects the depth palette; `--grayscale true` converts the saved PNG to luminance
## so rarity ordering can be judged without hue (v508 color-safe rule).
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const LightingRig := preload("res://scripts/scene_lighting_rig.gd")
const LoaderScript := preload("res://scripts/kit_monster_presentation_loader.gd")
const WARMUP_FRAMES := 30
const FAMILY_DEFS := [
	"dungeon_mob",
	"dungeon_archer",
	"dungeon_undead",
	"dungeon_wolf",
	"dungeon_bat",
]
const ROW_SPACING := 2.3
const COLUMN_SPACING := 2.2

var _output := ""
var _level := -1
var _grayscale := false
var _quality := "balanced"
var _width := 1280
var _height := 800


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)
	var world := Node3D.new()
	world.name = "MonsterVariantCapture"
	get_root().add_child(world)
	var factory := preload("res://scripts/ground_wall_factory.gd").new()
	_add_floor(world)
	_add_camera(world)
	var lighting = LightingRig.new()
	lighting.attach(world)
	lighting.sync(_level, factory, _quality)
	var main = MainScript.new()
	main.current_level = _level
	var rarities: Array = (LoaderScript.variants().get("rarities", {}) as Dictionary).keys()
	rarities.sort_custom(func(a, b): return _rarity_order(str(a)) < _rarity_order(str(b)))
	var row_origin := -(float(rarities.size()) - 1.0) * ROW_SPACING * 0.5
	var column_origin := -(float(FAMILY_DEFS.size()) - 1.0) * COLUMN_SPACING * 0.5
	for row in range(rarities.size()):
		for column in range(FAMILY_DEFS.size()):
			var entity := {"type": "monster", "monster_def_id": FAMILY_DEFS[column], "rarity": str(rarities[row])}
			var monster := main._make_entity_node(entity) as Node3D
			monster.name = "%s_%s" % [FAMILY_DEFS[column], rarities[row]]
			monster.position = Vector3(column_origin + column * COLUMN_SPACING, 0.0, row_origin + row * ROW_SPACING)
			monster.rotation.y = deg_to_rad(-30.0)
			world.add_child(monster)
			var player := monster.find_child("AnimationPlayer", true, false) as AnimationPlayer
			if player != null and player.has_animation("idle"):
				player.play("idle")
	for _i in range(WARMUP_FRAMES):
		await process_frame
	var image := get_root().get_texture().get_image()
	if _grayscale:
		image.convert(Image.FORMAT_L8)
	if image == null or image.save_png(_output) != OK:
		push_error("monster variant capture failed: %s" % _output)
		quit(1)
		return
	print("[monster-variant-capture] screenshot: %s level=%d rarities=%s" % [_output, _level, str(rarities)])
	quit(0)


func _rarity_order(rarity: String) -> int:
	var order := ["common", "champion", "rare", "unique"].find(rarity)
	return order if order >= 0 else 99


func _add_floor(world: Node3D) -> void:
	var floor := MeshInstance3D.new()
	floor.name = "ReferenceFloor"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(14.0, 0.08, 12.0)
	floor.mesh = mesh
	floor.position.y = -0.05
	var material := StandardMaterial3D.new()
	material.albedo_color = Color("#34322e")
	floor.material_override = material
	world.add_child(floor)


func _add_camera(world: Node3D) -> void:
	var camera := Camera3D.new()
	camera.name = "PlayerCamera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = 13.0
	camera.current = true
	world.add_child(camera)
	camera.look_at_from_position(Vector3(9.0, 20.0, 15.0), Vector3(0.0, 0.4, 0.0), Vector3.UP)


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i + 1 < args.size():
		var value := str(args[i + 1])
		match str(args[i]):
			"--output": _output = value
			"--level": _level = int(value)
			"--grayscale": _grayscale = value == "true"
			"--quality": _quality = value
			"--width": _width = int(value)
			"--height": _height = int(value)
		i += 2
