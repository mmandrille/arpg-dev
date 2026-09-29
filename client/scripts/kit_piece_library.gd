## Resolves manifest asset ids to imported kit scenes, meshes and bounds (ADR-0018 P2).
## The asset manifest (assets/manifests/assets.v0.json) is the only asset_id -> file link;
## runtime_path (client/...) maps to res:// by stripping the leading client/ (ADR-0006 D6).
class_name KitPieceLibrary
extends RefCounted

const MANIFEST_PATH := "../assets/manifests/assets.v0.json"

static var _assets: Dictionary = {}
static var _loaded: bool = false
static var _scenes: Dictionary = {}
static var _meshes: Dictionary = {}
static var _bounds: Dictionary = {}


static func _ensure_manifest() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join(MANIFEST_PATH)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("KitPieceLibrary: cannot open manifest %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_assets = (parsed as Dictionary).get("assets", {})


static func res_path(asset_id: String) -> String:
	_ensure_manifest()
	var entry: Dictionary = _assets.get(asset_id, {})
	var runtime_path := str(entry.get("runtime_path", ""))
	if runtime_path.begins_with("client/"):
		runtime_path = runtime_path.substr("client/".length())
	return "res://" + runtime_path if runtime_path != "" else ""


static func scene(asset_id: String) -> PackedScene:
	if _scenes.has(asset_id):
		return _scenes[asset_id]
	var path := res_path(asset_id)
	var packed: PackedScene = null
	if path != "" and ResourceLoader.exists(path):
		packed = load(path) as PackedScene
	if packed == null:
		push_warning("KitPieceLibrary: no scene for asset %s (%s)" % [asset_id, path])
	_scenes[asset_id] = packed
	return packed


static func instantiate(asset_id: String) -> Node3D:
	var packed := scene(asset_id)
	return packed.instantiate() as Node3D if packed != null else null


## First mesh of the piece (kit pieces are single-mesh), for MultiMesh floors.
static func mesh(asset_id: String) -> Mesh:
	if _meshes.has(asset_id):
		return _meshes[asset_id]
	var found: Mesh = null
	var node := instantiate(asset_id)
	if node != null:
		for child in mesh_instances(node):
			found = (child as MeshInstance3D).mesh
			break
		node.free()
	_meshes[asset_id] = found
	return found


## Piece bounds in the piece's own space (all mesh AABBs, with their node transforms).
static func bounds(asset_id: String) -> AABB:
	if _bounds.has(asset_id):
		return _bounds[asset_id]
	var box := AABB()
	var first := true
	var node := instantiate(asset_id)
	if node != null:
		for child in mesh_instances(node):
			var mi := child as MeshInstance3D
			var local := _transform_to(node, mi) * mi.get_aabb()
			box = local if first else box.merge(local)
			first = false
		node.free()
	_bounds[asset_id] = box
	return box


static func mesh_instances(root: Node) -> Array:
	var out: Array = []
	if root is MeshInstance3D:
		out.append(root)
	for child in root.get_children():
		out.append_array(mesh_instances(child))
	return out


static func _transform_to(root: Node3D, node: Node3D) -> Transform3D:
	var xf := Transform3D.IDENTITY
	var cursor: Node = node
	while cursor != null and cursor != root:
		if cursor is Node3D:
			xf = (cursor as Node3D).transform * xf
		cursor = cursor.get_parent()
	return xf
