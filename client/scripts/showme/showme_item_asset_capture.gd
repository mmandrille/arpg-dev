class_name ShowmeItemAssetCapture
extends RefCounted


static func setup(capture: SceneTree, asset_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "VisualFeedbackItemAsset"
	capture.get_root().add_child(root)

	_add_light(root)
	_add_floor(root)

	var entry := _manifest_entry(asset_id)
	if entry.is_empty():
		push_warning("[item-asset] unknown asset_id: %s" % asset_id)
		return root

	var model := _instantiate_runtime_glb(entry)
	if model == null:
		push_warning("[item-asset] could not load asset: %s" % asset_id)
		return root

	model.name = "FocusedAsset"
	root.add_child(model)
	_add_camera(root, model)

	return model


static func _manifest_entry(asset_id: String) -> Dictionary:
	var base := ProjectSettings.globalize_path("res://")
	var manifest := _read_json(base.path_join("../assets/manifests/assets.v0.json"))
	var assets: Dictionary = manifest.get("assets", {}) if typeof(manifest) == TYPE_DICTIONARY else {}
	var entry = assets.get(asset_id, {})
	return entry if typeof(entry) == TYPE_DICTIONARY else {}


static func _read_json(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}


static func _instantiate_runtime_glb(entry: Dictionary) -> Node3D:
	var res_path := _res_path(str(entry.get("runtime_path", "")))
	if res_path == "" or not ResourceLoader.exists(res_path):
		return null
	var packed := load(res_path) as PackedScene
	if packed == null:
		return null
	return packed.instantiate() as Node3D


static func _res_path(runtime_path: String) -> String:
	if runtime_path == "":
		return ""
	var path := runtime_path
	if path.begins_with("client/"):
		path = path.substr("client/".length())
	return "res://" + path


static func _add_light(root: Node3D) -> void:
	var light := DirectionalLight3D.new()
	light.name = "key_light"
	light.light_energy = 2.2
	light.rotation_degrees = Vector3(-55, -35, 0)
	root.add_child(light)


static func _add_camera(root: Node3D, model: Node3D) -> void:
	var camera := Camera3D.new()
	camera.name = "capture_camera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = 1.8
	root.add_child(camera)
	var bounds := _node_bounds(model)
	var center := bounds.position + bounds.size * 0.5
	var radius := maxf(bounds.size.length() * 0.5, 0.35)
	camera.look_at_from_position(center + Vector3(radius * 1.1, radius * 0.9, radius * 1.4), center, Vector3.UP)
	camera.current = true


static func _node_bounds(node: Node) -> AABB:
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
		return AABB(Vector3(-0.25, 0.0, -0.25), Vector3(0.5, 0.5, 0.5))
	return bounds


static func _mesh_instances(node: Node) -> Array:
	var out := []
	if node is MeshInstance3D:
		out.append(node)
	for child in node.get_children():
		out.append_array(_mesh_instances(child))
	return out


static func _add_floor(root: Node3D) -> void:
	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(2.4, 0.04, 2.4)
	floor.mesh = mesh
	floor.position = Vector3(0, -0.03, 0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#3f3f3c")
	floor.material_override = mat
	root.add_child(floor)
