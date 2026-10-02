class_name ItemModelThumbnailCache
extends Node
## Renders registered equipment models through one temporary viewport and keeps only small 2D
## thumbnails. Inventory slots never own a SubViewport or a live 3D model.

const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")
const ItemVisualsLoaderScript := preload("res://scripts/item_visuals_loader.gd")
const MAX_CACHED_THUMBNAILS := 32
const THUMBNAIL_SIZE := Vector2i(96, 96)

static var _service: ItemModelThumbnailCache
static var _thumbnail_cache: Dictionary = {}
static var _cache_order: Array[String] = []
static var _failed_assets: Dictionary = {}
static var _manifest: Dictionary = {}
static var _watchers: Dictionary = {}

var _viewport: SubViewport
var _camera: Camera3D
var _render_root: Node3D
var _queue: Array[String] = []
var _queued: Dictionary = {}
var _rendering := false


static func draw_if_available(canvas: Control, rect: Rect2, item: Dictionary, dimmed: bool) -> bool:
	var asset_id := asset_id_for_item(item)
	if asset_id == "":
		return false
	var texture := _thumbnail_cache.get(asset_id) as Texture2D
	if texture == null:
		_request_thumbnail(canvas, asset_id)
		return false
	var modulation := Color(0.66, 0.66, 0.66, 1.0) if dimmed else Color.WHITE
	canvas.draw_texture_rect(texture, rect.grow(-4.0), false, modulation)
	return true


static func asset_id_for_item(item: Dictionary) -> String:
	var def_id := str(item.get("item_def_id", item.get("item_template_id", "")))
	if def_id == "":
		return ""
	ItemRulesLoaderScript.ensure_loaded()
	var definition := ItemRulesLoaderScript.item_definition(def_id)
	if str(definition.get("category", "")).to_lower() != "equipment" or not bool(definition.get("equippable", false)):
		return ""
	var asset_id := ItemVisualsLoaderScript.hand_asset_id(def_id)
	if asset_id == "":
		asset_id = str(ItemRulesLoaderScript.item_presentation(def_id).get("3d_model", ""))
	if asset_id == "" or not _manifest_assets().has(asset_id):
		return ""
	var manifest_entry = _manifest_assets().get(asset_id, {})
	return asset_id if typeof(manifest_entry) == TYPE_DICTIONARY and str(manifest_entry.get("runtime_path", "")) != "" else ""


static func debug_cache_state() -> Dictionary:
	return {"cached_asset_ids": _cache_order.duplicate(), "failed_asset_ids": _failed_assets.keys(), "live_viewports": 1 if _service != null and is_instance_valid(_service) and _service._viewport != null else 0}


static func reset_for_tests() -> void:
	_thumbnail_cache.clear()
	_cache_order.clear()
	_failed_assets.clear()
	_watchers.clear()
	_manifest.clear()


static func wait_until_idle(scene_tree: SceneTree, max_frames: int = 180) -> bool:
	for _frame in range(max_frames):
		if _service == null or not is_instance_valid(_service) or not _service.is_busy():
			return true
		await scene_tree.process_frame
	return false


static func _manifest_assets() -> Dictionary:
	if _manifest.is_empty():
		var path := ProjectSettings.globalize_path("res://").path_join("../assets/manifests/assets.v0.json")
		var file := FileAccess.open(path, FileAccess.READ)
		if file == null:
			return {}
		var parsed = JSON.parse_string(file.get_as_text())
		if typeof(parsed) == TYPE_DICTIONARY:
			_manifest = parsed.get("assets", {})
	return _manifest


static func _request_thumbnail(canvas: Control, asset_id: String) -> void:
	if _failed_assets.has(asset_id):
		return
	if not _watchers.has(asset_id):
		_watchers[asset_id] = []
	(_watchers[asset_id] as Array).append(weakref(canvas))
	if _service == null or not is_instance_valid(_service):
		_service = load("res://scripts/item_model_thumbnail_cache.gd").new()
		canvas.get_tree().root.add_child(_service)
	if not _service._queued.has(asset_id):
		_service._queued[asset_id] = true
		_service._queue.append(asset_id)
		_service.call_deferred("_render_next")


func _ensure_render_scene() -> void:
	if _viewport != null:
		return
	_viewport = SubViewport.new()
	_viewport.name = "CachedItemThumbnailViewport"
	_viewport.size = THUMBNAIL_SIZE
	_viewport.transparent_bg = true
	_viewport.own_world_3d = true
	_viewport.render_target_update_mode = SubViewport.UPDATE_DISABLED
	add_child(_viewport)
	_render_root = Node3D.new()
	_render_root.name = "ThumbnailRenderRoot"
	_viewport.add_child(_render_root)
	_camera = Camera3D.new()
	_camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	_camera.size = 1.0
	_camera.near = 0.01
	_camera.far = 50.0
	_camera.current = true
	_render_root.add_child(_camera)
	var key_light := DirectionalLight3D.new()
	key_light.name = "KeyLight"
	key_light.rotation_degrees = Vector3(-38.0, -28.0, 0.0)
	key_light.light_energy = 1.4
	_render_root.add_child(key_light)
	var fill_light := DirectionalLight3D.new()
	fill_light.name = "FillLight"
	fill_light.rotation_degrees = Vector3(-16.0, 144.0, 0.0)
	fill_light.light_energy = 0.55
	_render_root.add_child(fill_light)


func _render_next() -> void:
	if _rendering or _queue.is_empty():
		return
	_rendering = true
	_ensure_render_scene()
	while not _queue.is_empty():
		var asset_id: String = _queue.pop_front()
		_queued.erase(asset_id)
		await _render_asset(asset_id)
	_rendering = false


func is_busy() -> bool:
	return _rendering or not _queue.is_empty()


func _render_asset(asset_id: String) -> void:
	var entry = ItemModelThumbnailCache._manifest_assets().get(asset_id, {})
	var runtime_path := str(entry.get("runtime_path", "")) if typeof(entry) == TYPE_DICTIONARY else ""
	if runtime_path == "":
		ItemModelThumbnailCache._failed_assets[asset_id] = true
		_complete_request(asset_id)
		return
	var res_path := runtime_path
	if res_path.begins_with("client/"):
		res_path = res_path.substr("client/".length())
	res_path = "res://" + res_path
	var packed := ResourceLoader.load(res_path) as PackedScene
	if packed == null:
		ItemModelThumbnailCache._failed_assets[asset_id] = true
		_complete_request(asset_id)
		return
	var model := packed.instantiate() as Node3D
	if model == null:
		ItemModelThumbnailCache._failed_assets[asset_id] = true
		_complete_request(asset_id)
		return
	model.name = "ThumbnailModel"
	_render_root.add_child(model)
	await get_tree().process_frame
	var bounds := _model_bounds(model)
	if bounds.size.length_squared() <= 0.0001:
		model.queue_free()
		ItemModelThumbnailCache._failed_assets[asset_id] = true
		_complete_request(asset_id)
		return
	var center := bounds.get_center()
	model.position -= center
	var extent := maxf(bounds.size.x, maxf(bounds.size.y, bounds.size.z))
	_camera.size = maxf(extent * 1.35, 0.1)
	_camera.position = Vector3(1.0, 0.9, 1.45).normalized() * maxf(extent * 2.7, 0.5)
	_camera.look_at(Vector3.ZERO, Vector3.UP)
	_viewport.render_target_update_mode = SubViewport.UPDATE_ONCE
	await get_tree().process_frame
	await get_tree().process_frame
	var image := _viewport.get_texture().get_image()
	_viewport.render_target_update_mode = SubViewport.UPDATE_DISABLED
	model.queue_free()
	if image == null or image.is_empty():
		ItemModelThumbnailCache._failed_assets[asset_id] = true
	else:
		ItemModelThumbnailCache._cache_texture(asset_id, ImageTexture.create_from_image(image))
	_complete_request(asset_id)


func _model_bounds(root: Node3D) -> AABB:
	var bounds := AABB()
	var found := false
	for node in root.find_children("*", "VisualInstance3D", true, false):
		var visual := node as VisualInstance3D
		var local_box := visual.get_aabb()
		for x in [local_box.position.x, local_box.end.x]:
			for y in [local_box.position.y, local_box.end.y]:
				for z in [local_box.position.z, local_box.end.z]:
					var point := root.to_local(visual.global_transform * Vector3(x, y, z))
					if not found:
						bounds = AABB(point, Vector3.ZERO)
						found = true
					else:
						bounds = bounds.expand(point)
	return bounds


static func _cache_texture(asset_id: String, texture: Texture2D) -> void:
	_thumbnail_cache[asset_id] = texture
	_cache_order.append(asset_id)
	while _cache_order.size() > MAX_CACHED_THUMBNAILS:
		var expired: String = _cache_order.pop_front()
		_thumbnail_cache.erase(expired)


func _complete_request(asset_id: String) -> void:
	for weak_canvas in ItemModelThumbnailCache._watchers.get(asset_id, []):
		var canvas := (weak_canvas as WeakRef).get_ref() as Control
		if canvas != null and is_instance_valid(canvas):
			canvas.queue_redraw()
	ItemModelThumbnailCache._watchers.erase(asset_id)
