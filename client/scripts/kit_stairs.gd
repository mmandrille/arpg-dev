## KayKit stairs for the stairs_up / stairs_down interactables (ADR-0018 P2, v489). Presentation
## only: position, footprint and pick collider stay owned by the entity. Data:
## shared/assets/dungeon_kit_presentation.v0.json -> stairs.
##
## stairs_up is a scaled kit flight. stairs_down is the kit open-grate hatch framing a dark pit face:
## the kit has no descending stair and the kit floor covers anything below floor level.
class_name KitStairs
extends RefCounted

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const ModelTintScript := preload("res://scripts/model_tint.gd")

const MODEL_NAME := "KitStairsModel"
const PIT_NAME := "KitStairsPit"


static func enabled() -> bool:
	return LoaderScript.prop_enabled("stairs")


## Kit stair node, or null (caller falls back to the procedural stairs) when disabled or unloadable.
static func make_node(def_id: String) -> Node3D:
	if not enabled():
		return null
	var cfg := LoaderScript.stairs_config()
	var is_down := def_id == "stairs_down"
	var asset_id := str(cfg.get("down_asset_id" if is_down else "up_asset_id", ""))
	var piece := LibraryScript.instantiate(asset_id)
	if piece == null:
		return null
	var root := Node3D.new()
	root.name = "Stairs_%s" % def_id
	root.set_meta("kit_stairs", true)
	var scale := float(cfg.get("down_scale" if is_down else "up_scale", 1.0))
	var box := LibraryScript.bounds(asset_id)
	piece.name = MODEL_NAME
	piece.scale = Vector3.ONE * scale
	if not is_down:
		piece.rotation.y = deg_to_rad(float(cfg.get("up_yaw_degrees", 0.0)))
	# Centre the piece footprint on the entity (kit stairs grow from one edge).
	var center := piece.transform.basis * box.get_center()
	var lift := float(cfg.get("down_lift", 0.0)) if is_down else 0.0
	piece.position = Vector3(-center.x, lift, -center.z)
	root.add_child(piece)
	if is_down:
		root.add_child(_make_pit(box, scale, lift, cfg))
	apply_state(root, def_id, "ready")
	return root


static func _make_pit(box: AABB, scale: float, lift: float, cfg: Dictionary) -> MeshInstance3D:
	var inset := float(cfg.get("down_pit_inset", 0.7))
	var pit := MeshInstance3D.new()
	pit.name = PIT_NAME
	var mesh := BoxMesh.new()
	mesh.size = Vector3(box.size.x * scale * inset, 0.01, box.size.z * scale * inset)
	pit.mesh = mesh
	# Between the floor surface and the hatch frame top, so the opening reads as a dark hole.
	pit.position = Vector3(0.0, lift + maxf(box.end.y * scale * 0.5, 0.005), 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(str(cfg.get("down_pit_color", "#050608")))
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	pit.material_override = mat
	return pit


## Ready vs locked/disabled: tint the kit meshes (texture kept) with the catalog state tint.
static func apply_state(node: Node3D, _def_id: String, state: String) -> void:
	var model := node.find_child(MODEL_NAME, true, false)
	if model == null:
		return
	var cfg := LoaderScript.stairs_config()
	var locked := state == "locked" or state == "disabled"
	var tint := Color(str(cfg.get("locked_tint" if locked else "ready_tint", "#ffffff")))
	for mesh in LibraryScript.mesh_instances(model):
		(mesh as MeshInstance3D).material_override = ModelTintScript.tinted_material(mesh as MeshInstance3D, tint)
