## Fixed five-drop fixture using the playable isometric camera settings and client lighting.
extends SceneTree

const CameraLoader := preload("res://scripts/camera_presentations_loader.gd")
const LightingRig := preload("res://scripts/scene_lighting_rig.gd")
const LootFactory := preload("res://scripts/loot_node_factory.gd")
const RarityPresenter := preload("res://scripts/rarity_cue_presenter.gd")
const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")
const WARMUP_FRAMES := 60
const SAMPLE_FRAMES := 120

var _output := ""
var _zoom := "normal"
var _ground_tone := "dark"
var _reveal := false
var _quality := "balanced"
var _baseline := false
var _width := 1120
var _height := 720


func _initialize() -> void:
	_parse_args()
	RarityPresenter.capture_baseline = _baseline
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)
	var world := Node3D.new()
	world.name = "RarityCueCapture"
	get_root().add_child(world)
	_add_ground(world)
	_add_camera(world)
	_add_loot(world)
	var lighting = LightingRig.new()
	lighting.attach(world)
	lighting.sync(0, preload("res://scripts/ground_wall_factory.gd").new(), _quality)
	for _i in range(WARMUP_FRAMES):
		await process_frame
	var samples: Array = []
	var previous_us := Time.get_ticks_usec()
	for _i in range(SAMPLE_FRAMES):
		await process_frame
		var now_us := Time.get_ticks_usec()
		samples.append({
			"frame_ms": float(now_us - previous_us) / 1000.0,
			"process_ms": float(Performance.get_monitor(Performance.TIME_PROCESS)) * 1000.0,
			"draw_calls": int(Performance.get_monitor(Performance.RENDER_TOTAL_DRAW_CALLS_IN_FRAME)),
		})
		previous_us = now_us
	var image := get_root().get_texture().get_image()
	if image == null or image.save_png(_output) != OK:
		push_error("rarity cue capture failed: %s" % _output)
		quit(1)
		return
	var evidence := {"samples": samples, "warmup_frames": WARMUP_FRAMES, "zoom": _zoom, "ground_tone": _ground_tone, "reveal": _reveal, "baseline": _baseline, "quality": _quality, "renderer": ProjectSettings.get_setting("rendering/renderer/rendering_method"), "viewport": [_width, _height], "item_def_id": "long_sword", "rarities": ["common", "magic", "rare", "unique", "set"]}
	var metadata := FileAccess.open(_output.get_basename() + ".json", FileAccess.WRITE)
	if metadata != null:
		metadata.store_string(JSON.stringify(evidence))
	print("[rarity-cue-capture] screenshot: %s" % _output)
	quit(0)


func _add_ground(world: Node3D) -> void:
	var ground := MeshInstance3D.new()
	ground.name = "ContrastGround_%s" % _ground_tone
	var mesh := BoxMesh.new()
	mesh.size = Vector3(15.0, 0.08, 8.0)
	ground.mesh = mesh
	ground.position.y = -0.05
	var material := StandardMaterial3D.new()
	material.albedo_color = Color("#b8ae94" if _ground_tone == "light" else "#252926")
	ground.material_override = material
	world.add_child(ground)


func _add_camera(world: Node3D) -> void:
	var cfg := CameraLoader.mode("isometric")
	var camera := Camera3D.new()
	camera.name = "PlayerCamera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = float(cfg.get("zoom_max" if _zoom == "max" else "zoom_default", 12.0))
	camera.current = true
	world.add_child(camera)
	camera.look_at_from_position(Vector3(9.0, 20.0, 15.0), Vector3.ZERO, Vector3.UP)


func _add_loot(world: Node3D) -> void:
	var root := ProjectSettings.globalize_path("res://")
	var manifest_data = JSON.parse_string(FileAccess.get_file_as_string(root.path_join("../assets/manifests/assets.v0.json")))
	var manifest: Dictionary = manifest_data if typeof(manifest_data) == TYPE_DICTIONARY else {}
	ItemRulesLoaderScript.ensure_loaded()
	var factory = LootFactory.new(manifest.get("assets", {}), ItemRulesLoaderScript.item_presentations)
	var rarity_data = JSON.parse_string(FileAccess.get_file_as_string(root.path_join("../shared/rules/item_templates.v0.json")))
	var rarities: Dictionary = rarity_data.get("rarities", {}) if typeof(rarity_data) == TYPE_DICTIONARY else {}
	var keys := rarities.keys()
	for index in range(keys.size()):
		var rarity := str(keys[index])
		var loot: Node3D = factory.make_loot_node({"id": "rarity_%s" % rarity, "type": "loot", "item_def_id": "long_sword", "rarity": rarity, "display_name": "Long Sword"})
		loot.position = Vector3((float(index) - 2.0) * 2.0, 0.0, 0.0)
		loot.name = "Rarity_%s" % rarity
		world.add_child(loot)
		var label := loot.find_child("LootLabel", false, false) as Label3D
		if label != null:
			label.visible = _reveal
			if _baseline:
				label.text = "Long Sword"


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i + 1 < args.size():
		var key := str(args[i])
		var value := str(args[i + 1])
		match key:
			"--output": _output = value
			"--town-zoom": _zoom = value
			"--ground-tone": _ground_tone = value
			"--reveal": _reveal = value == "true"
			"--quality": _quality = value
			"--baseline": _baseline = value == "true"
			"--width": _width = int(value)
			"--height": _height = int(value)
		i += 2
	if _output == "":
		_output = ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme/rarity-cues.png")
