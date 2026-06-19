extends Node3D
class_name ModelViewer

const CHARACTER_SCENE := "res://scenes/character.tscn"
const CLIP_ORDER := ["idle", "walk", "attack", "hit", "death"]

@onready var model_root: Node3D = $ModelRoot
@onready var camera: Camera3D = $Camera3D

var selected_asset_id: String = ""
var current_instance: Node3D
var current_animation_player: AnimationPlayer
var current_clips: Array[String] = []
var current_row: Dictionary = {}
var auto_cycle: bool = true
var _clip_index: int = 0


func _ready() -> void:
	selected_asset_id = OS.get_environment("MODEL_ASSET_ID")
	if selected_asset_id == "":
		selected_asset_id = ProjectSettings.get_setting("arpg/model_viewer/asset_id", "")
	var check_mode := OS.get_environment("MODEL_VIEWER_CHECK") in ["1", "true", "yes", "on"]
	if selected_asset_id == "":
		_fail("MODEL_ASSET_ID is required")
		return
	if not load_asset(selected_asset_id):
		return
	await get_tree().process_frame
	_frame_camera()
	if check_mode:
		if current_instance == null or current_animation_player == null or current_clips.is_empty():
			_fail("model viewer check failed for %s" % selected_asset_id)
			return
		print("[model-viewer] PASS %s clips=%s" % [selected_asset_id, ",".join(current_clips)])
		get_tree().quit(0)
		return
	if auto_cycle:
		_play_next_clip()


func load_asset(asset_id: String) -> bool:
	current_row = resolve(asset_id)
	if current_row.is_empty():
		_fail("unknown model asset_id: %s" % asset_id)
		return false
	_clear_model()
	var instance := _instantiate_for_row(current_row)
	if instance == null:
		_fail("could not instantiate %s" % asset_id)
		return false
	current_instance = instance
	model_root.add_child(current_instance)
	current_animation_player = current_instance.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if current_animation_player == null:
		_fail("%s has no AnimationPlayer" % asset_id)
		return false
	current_clips = _available_clips(current_animation_player)
	if current_clips.is_empty():
		_fail("%s has no previewable animation clips" % asset_id)
		return false
	print("[model-viewer] loaded %s used_by=%s runtime=%s" % [
		asset_id,
		",".join(current_row.get("used_by", [])),
		str(current_row.get("runtime_path", "")),
	])
	return true


static func resolve(asset_id: String) -> Dictionary:
	for row in catalog_rows():
		if str(row.get("asset_id", "")) == asset_id:
			return row
	return {}


static func catalog_rows() -> Array:
	var root := ProjectSettings.globalize_path("res://")
	var manifest := _read_json(root.path_join("../assets/manifests/assets.v0.json"))
	var class_presentations := _read_json(root.path_join("../shared/assets/class_presentations.v0.json"))
	var monster_visuals := _read_json(root.path_join("../shared/assets/monster_visuals.v0.json"))
	var assets: Dictionary = manifest.get("assets", {})
	var by_asset: Dictionary = {}

	for class_id in (class_presentations.get("classes", {}) as Dictionary).keys():
		var entry: Dictionary = class_presentations["classes"][class_id]
		var model: Dictionary = entry.get("model", {})
		var asset_id := str(model.get("asset_id", ""))
		if asset_id == "":
			continue
		var row := _ensure_row(by_asset, assets, asset_id, "character")
		if row.is_empty():
			continue
		row["used_by"].append(str(class_id))
		row["class_id"] = str(class_id)
		row["scale"] = _positive_float(model.get("scale", 1.0), 1.0)
		row["height_offset"] = float(model.get("height_offset", 0.0))

	for monster_def_id in (monster_visuals.get("monster_visuals", {}) as Dictionary).keys():
		var entry: Dictionary = monster_visuals["monster_visuals"][monster_def_id]
		var asset_id := str(entry.get("asset_id", ""))
		if asset_id == "":
			continue
		var row := _ensure_row(by_asset, assets, asset_id, "monster")
		if row.is_empty():
			continue
		row["used_by"].append(str(monster_def_id))
		row["scene"] = str(entry.get("scene", ""))
		row["scale"] = _positive_float(entry.get("scale", 1.0), 1.0)
		row["height_offset"] = float(entry.get("height_offset", 0.0))

	var rows: Array = by_asset.values()
	for row in rows:
		row["used_by"].sort()
	rows.sort_custom(func(a, b): return "%s:%s" % [a.get("type", ""), a.get("asset_id", "")] < "%s:%s" % [b.get("type", ""), b.get("asset_id", "")])
	return rows


static func _ensure_row(by_asset: Dictionary, assets: Dictionary, asset_id: String, expected_type: String) -> Dictionary:
	var asset: Dictionary = assets.get(asset_id, {})
	if str(asset.get("type", "")) != expected_type:
		return {}
	if not by_asset.has(asset_id):
		by_asset[asset_id] = {
			"asset_id": asset_id,
			"type": expected_type,
			"runtime_path": str(asset.get("runtime_path", "")),
			"used_by": [],
			"scene": "",
			"class_id": "",
			"scale": 1.0,
			"height_offset": 0.0,
		}
	return by_asset[asset_id]


func _instantiate_for_row(row: Dictionary) -> Node3D:
	if str(row.get("type", "")) == "character":
		return _instantiate_character(row)
	var scene := str(row.get("scene", ""))
	if scene != "":
		var scene_path := "res://scenes/%s.tscn" % scene
		if ResourceLoader.exists(scene_path):
			return (load(scene_path) as PackedScene).instantiate() as Node3D
	return _instantiate_runtime_glb(row)


func _instantiate_character(row: Dictionary) -> Node3D:
	var character := (load(CHARACTER_SCENE) as PackedScene).instantiate() as Node3D
	var old_model := character.find_child("ModelRoot", false, false) as Node
	if old_model != null:
		character.remove_child(old_model)
		old_model.free()
	var model := _instantiate_runtime_glb(row)
	if model == null:
		return null
	model.name = "ModelRoot"
	model.scale = Vector3.ONE * float(row.get("scale", 1.0))
	model.position.y = float(row.get("height_offset", 0.0))
	character.add_child(model)
	character.move_child(model, 0)
	var ap := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.root_node = NodePath("../ModelRoot")
	return character


func _instantiate_runtime_glb(row: Dictionary) -> Node3D:
	var res_path := _res_path(str(row.get("runtime_path", "")))
	if res_path == "" or not ResourceLoader.exists(res_path):
		return null
	var packed := load(res_path) as PackedScene
	if packed == null:
		return null
	return packed.instantiate() as Node3D


func _available_clips(player: AnimationPlayer) -> Array[String]:
	var clips: Array[String] = []
	for clip in CLIP_ORDER:
		if player.has_animation(clip):
			clips.append(clip)
	for clip in player.get_animation_list():
		var name := str(clip)
		if not clips.has(name):
			clips.append(name)
	return clips


func _play_next_clip() -> void:
	if current_animation_player == null or current_clips.is_empty():
		return
	var clip := current_clips[_clip_index % current_clips.size()]
	_clip_index += 1
	current_animation_player.play(clip)
	var anim := current_animation_player.get_animation(clip)
	var delay = maxf(0.45, minf(1.4, anim.length if anim != null else 0.8))
	await get_tree().create_timer(delay).timeout
	_play_next_clip()


func _clear_model() -> void:
	for child in model_root.get_children():
		model_root.remove_child(child)
		child.free()
	current_instance = null
	current_animation_player = null
	current_clips = []
	_clip_index = 0


func _frame_camera() -> void:
	if camera == null or current_instance == null:
		return
	var bounds := _node_bounds(current_instance)
	var center := bounds.position + bounds.size * 0.5
	var radius = maxf(bounds.size.length() * 0.5, 1.0)
	camera.look_at_from_position(center + Vector3(radius * 1.2, radius * 1.1, radius * 1.7), center, Vector3.UP)


func _node_bounds(node: Node) -> AABB:
	var found := false
	var bounds := AABB(Vector3.ZERO, Vector3.ONE)
	for mesh in _mesh_instances(node):
		var mi := mesh as MeshInstance3D
		var local := mi.get_aabb()
		var global_aabb := AABB(mi.global_transform * local.position, Vector3.ZERO)
		for i in range(8):
			global_aabb = global_aabb.expand(mi.global_transform * local.get_endpoint(i))
		if not found:
			bounds = global_aabb
			found = true
		else:
			bounds = bounds.merge(global_aabb)
	if not found:
		return AABB(Vector3(-0.5, 0.0, -0.5), Vector3.ONE)
	return bounds


func _mesh_instances(node: Node) -> Array:
	var out := []
	if node is MeshInstance3D:
		out.append(node)
	for child in node.get_children():
		out.append_array(_mesh_instances(child))
	return out


static func _read_json(path: String) -> Dictionary:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("[model-viewer] missing json: %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("[model-viewer] malformed json: %s" % path)
		return {}
	return parsed


static func _res_path(runtime_path: String) -> String:
	if runtime_path == "":
		return ""
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p


static func _positive_float(value, fallback: float) -> float:
	var parsed := float(value)
	if parsed <= 0.0:
		return fallback
	return parsed


func _fail(message: String) -> void:
	printerr("[model-viewer] FAIL: %s" % message)
	if OS.get_environment("MODEL_VIEWER_CHECK") in ["1", "true", "yes", "on"]:
		get_tree().quit(1)
