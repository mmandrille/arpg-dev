# EquipmentVisualResolver (spec equip-and-see-it §4.7 / §5.1, ADR-0006).
#
# Resolves authoritative item state into a mounted runtime visual:
#   item_instance_id --(inventory cache)--> item_def_id
#   item_def_id      --(shared/assets/item_visuals.v0.json)--> asset_id + socket + transform
#   asset_id         --(assets/manifests/assets.v0.json)--> runtime .glb path
#   runtime_path     --(strip client/, prepend res://)--> Godot resource (ADR-0006 D6)
#
# The mount-root node is INJECTED (never an absolute /root/Main/... lookup) so the
# full interactive scene (main.gd) and the headless SceneTree smoke (smoke.gd)
# drive one code path. The resolver owns the inventory cache and currently
# equipped slot instances, so callers just forward protocol events.
extends RefCounted
class_name EquipmentVisualResolver

const EQUIPMENT_SLOTS := ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]
const FALLBACK_ASSET_BY_SLOT := {
	"head": "fallback_equipment_head_v0",
	"amulet": "fallback_equipment_amulet_v0",
	"chest": "fallback_equipment_chest_v0",
	"gloves": "fallback_equipment_gloves_v0",
	"belt": "fallback_equipment_belt_v0",
	"boots": "fallback_equipment_boots_v0",
	"ring_left": "fallback_equipment_ring_left_v0",
	"ring_right": "fallback_equipment_ring_right_v0",
	"main_hand": "weapon_rusty_sword_v0",
	"off_hand": "fallback_equipment_off_hand_v0",
}
const SOCKET_BY_SLOT := {
	"head": "head_socket",
	"amulet": "amulet_socket",
	"chest": "chest_socket",
	"gloves": "gloves_socket",
	"belt": "belt_socket",
	"boots": "boots_socket",
	"ring_left": "ring_left_socket",
	"ring_right": "ring_right_socket",
	"main_hand": "right_hand_socket",
	"off_hand": "off_hand_socket",
}
const RARITY_TINTS := {
	"common": Color("#f2f2ec"),
	"magic": Color("#5aa7ff"),
	"rare": Color("#ffd75e"),
	"unique": Color("#ff9b45"),
}

var _mount_root: Node3D
var _visuals: Dictionary = {}      # item_def_id -> visual metadata
var _assets: Dictionary = {}       # asset_id -> manifest entry
var _inventory: Dictionary = {}    # item_instance_id (String) -> inventory item Dictionary
var _equipped: Dictionary = {}      # slot -> item_instance_id
var _mounted_nodes: Dictionary = {} # slot -> Node3D
var _mounted_state: Dictionary = {} # slot -> debug state
var _warnings: Array = []


func _init(mount_root: Node3D) -> void:
	_mount_root = mount_root
	_load_data()


# --- protocol-event ingestion (called by main.gd / smoke.gd) ----------------

func apply_snapshot(payload: Dictionary) -> void:
	# Rebuild inventory cache + equipped state from an authoritative snapshot and
	# mount immediately (spec acceptance #8: resume restores from the snapshot,
	# not only from a live equipped_update delta).
	_inventory.clear()
	for item in payload.get("inventory", []):
		_record_item(item)
	_equipped.clear()
	var equipped: Dictionary = payload.get("equipped", {})
	for slot in EQUIPMENT_SLOTS:
		var item_id = equipped.get(str(slot), null)
		_equipped[str(slot)] = str(item_id) if item_id != null else ""
	_refresh_all()


func ingest_inventory_item(item: Dictionary) -> void:
	# Handles inventory_add / inventory_update. If it's the equipped item finally
	# arriving/resolving, (re)mount.
	_record_item(item)
	var item_id := str(item.get("item_instance_id", ""))
	for slot in _equipped.keys():
		if str(_equipped.get(slot, "")) == item_id:
			_refresh_slot(str(slot))


func apply_equipped_update(slot: String, item_instance_id) -> void:
	if not EQUIPMENT_SLOTS.has(slot):
		return
	_equipped[slot] = str(item_instance_id) if item_instance_id != null else ""
	_refresh_slot(slot)


# --- debug surface (spec §4.4 / §4.7) ---------------------------------------

func get_debug_state() -> Dictionary:
	var visuals := {}
	for slot in EQUIPMENT_SLOTS:
		if _mounted_state.has(str(slot)):
			visuals[str(slot)] = (_mounted_state[str(slot)] as Dictionary).duplicate(true)
	visuals["weapon"] = visuals.get("main_hand", null)
	return {
		"equipped_visuals": visuals,
		"warnings": _warnings,
	}


# --- internals --------------------------------------------------------------

func _record_item(item: Dictionary) -> void:
	var iid := str(item.get("item_instance_id", ""))
	if iid != "":
		_inventory[iid] = item.duplicate(true)


func _refresh_all() -> void:
	_warnings = []
	for slot in EQUIPMENT_SLOTS:
		_refresh_slot(str(slot), false)


func _refresh_slot(slot: String, reset_warnings: bool = true) -> void:
	# Each slot refresh recomputes from scratch: clear the prior mount (spec §7:
	# no duplicate stale nodes) and reset transient warnings for this attempt.
	if reset_warnings:
		_warnings = []
	_clear_mounted(slot)
	_mounted_state.erase(slot)

	var item_instance_id := str(_equipped.get(slot, ""))
	if item_instance_id == "":
		return

	var item: Dictionary = _inventory.get(item_instance_id, {})
	var def_id: String = str(item.get("item_def_id", ""))
	if def_id == "":
		# Equipped instance not (yet) in the inventory cache; a later
		# inventory_add/snapshot will resolve it. Surface it, render nothing.
		_warn({"code": "unknown_item_instance_id", "item_instance_id": item_instance_id, "slot": slot})
		return

	var vis: Dictionary = _visual_for(def_id, slot)
	if vis.is_empty():
		_warn({"code": "missing_fallback_visual", "item_def_id": def_id, "slot": slot})
		return

	var asset_id: String = str(vis["asset_id"])
	var entry = _assets.get(asset_id, null)
	if entry == null:
		_warn({"code": "unknown_asset_id", "asset_id": asset_id, "item_def_id": def_id, "slot": slot})
		return

	if _mount_root == null:
		_warn({"code": "missing_mount_socket", "mount_socket": str(vis["mount_socket"]), "slot": slot})
		return
	var mount_socket := _mount_socket_for_slot(slot, vis)
	var socket := _mount_root.find_child(mount_socket, true, false)
	if socket == null:
		_warn({"code": "missing_mount_socket", "mount_socket": mount_socket, "slot": slot})
		return

	var procedural_fallback := _procedural_fallback_visual(asset_id, slot)
	var inst: Node3D
	if procedural_fallback != null:
		inst = procedural_fallback
	else:
		var packed = load(_res_path(str(entry["runtime_path"])))
		if packed == null:
			_warn({"code": "unknown_asset_id", "asset_id": asset_id, "item_def_id": def_id, "slot": slot})
			return
		inst = (packed as PackedScene).instantiate()
	inst.name = asset_id
	_apply_transform(inst, _local_transform_for_slot(slot, vis))
	var rarity := str(item.get("rarity", "common")).to_lower()
	var tint: Color = RARITY_TINTS.get(rarity, RARITY_TINTS["common"])
	_apply_tint(inst, tint)
	socket.add_child(inst)
	_mounted_nodes[slot] = inst
	_mounted_state[slot] = {
		"slot": slot,
		"item_instance_id": item_instance_id,
		"item_def_id": def_id,
		"asset_id": asset_id,
		"mount_socket": mount_socket,
		"rarity": rarity,
		"tint": tint.to_html(false),
		"node_path": (str(inst.get_path()) if inst.is_inside_tree() else ""),
		"visible": inst.visible,
		"procedural_fallback": procedural_fallback != null,
	}


func _visual_for(def_id: String, slot: String) -> Dictionary:
	var vis = _visuals.get(def_id, null)
	if typeof(vis) == TYPE_DICTIONARY:
		return (vis as Dictionary).duplicate(true)
	var asset_id := str(FALLBACK_ASSET_BY_SLOT.get(slot, ""))
	var socket := str(SOCKET_BY_SLOT.get(slot, ""))
	if asset_id == "" or socket == "":
		return {}
	return {
		"asset_id": asset_id,
		"slot": slot,
		"mount_socket": socket,
		"local_transform": {
			"position": {"x": 0.0, "y": 0.0, "z": 0.0},
			"rotation_degrees": {"x": 0.0, "y": 0.0, "z": 0.0},
			"scale": {"x": 0.25, "y": 0.25, "z": 0.25},
		},
	}


func _mount_socket_for_slot(slot: String, vis: Dictionary) -> String:
	if str(vis.get("slot", slot)) != slot and SOCKET_BY_SLOT.has(slot):
		return str(SOCKET_BY_SLOT[slot])
	return str(vis.get("mount_socket", SOCKET_BY_SLOT.get(slot, "right_hand_socket")))


func _local_transform_for_slot(slot: String, vis: Dictionary) -> Dictionary:
	var transform: Dictionary = (vis.get("local_transform", {}) as Dictionary).duplicate(true)
	if slot != "off_hand" or str(vis.get("slot", slot)) == slot:
		return transform
	var position: Dictionary = (transform.get("position", {}) as Dictionary).duplicate(true)
	position["z"] = float(position.get("z", 0.0)) + 0.08
	transform["position"] = position
	var rotation: Dictionary = (transform.get("rotation_degrees", {}) as Dictionary).duplicate(true)
	rotation["z"] = float(rotation.get("z", 0.0)) + 180.0
	transform["rotation_degrees"] = rotation
	return transform


func _clear_mounted(slot: String) -> void:
	var mounted = _mounted_nodes.get(slot, null)
	if mounted != null and is_instance_valid(mounted):
		(mounted as Node3D).queue_free()
	_mounted_nodes.erase(slot)


func _apply_transform(node: Node3D, t: Dictionary) -> void:
	if t == null or t.is_empty():
		return
	var p = t.get("position", {})
	node.position = Vector3(p.get("x", 0.0), p.get("y", 0.0), p.get("z", 0.0))
	var r = t.get("rotation_degrees", {})
	node.rotation_degrees = Vector3(r.get("x", 0.0), r.get("y", 0.0), r.get("z", 0.0))
	var s = t.get("scale", {})
	node.scale = Vector3(s.get("x", 1.0), s.get("y", 1.0), s.get("z", 1.0))


func _apply_tint(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		(root as MeshInstance3D).material_override = mat
	for child in root.get_children():
		_apply_tint(child, color)


func _procedural_fallback_visual(asset_id: String, slot: String) -> Node3D:
	if not asset_id.begins_with("fallback_equipment_"):
		return null
	var root := Node3D.new()
	match slot:
		"off_hand":
			root.add_child(_mesh_part("round_shield_face", _cylinder_mesh(0.48, 0.08, 32), Vector3.ZERO, Vector3(90, 0, 0)))
			root.add_child(_mesh_part("round_shield_boss", _cylinder_mesh(0.16, 0.10, 24), Vector3(0, 0, 0.05), Vector3(90, 0, 0)))
			root.add_child(_mesh_part("round_shield_grip", _box_mesh(Vector3(0.12, 0.62, 0.07)), Vector3(0, 0, -0.07)))
		"head":
			root.add_child(_mesh_part("helmet_cap", _cylinder_mesh(0.62, 0.56, 24), Vector3.ZERO))
			root.add_child(_mesh_part("helmet_brow", _box_mesh(Vector3(1.0, 0.12, 0.62)), Vector3(0, -0.18, -0.16)))
		"chest":
			root.add_child(_mesh_part("chest_plate", _box_mesh(Vector3(0.86, 1.0, 0.28)), Vector3.ZERO))
			root.add_child(_mesh_part("left_pauldron", _box_mesh(Vector3(0.32, 0.18, 0.34)), Vector3(-0.58, 0.34, 0)))
			root.add_child(_mesh_part("right_pauldron", _box_mesh(Vector3(0.32, 0.18, 0.34)), Vector3(0.58, 0.34, 0)))
		"boots":
			root.add_child(_mesh_part("left_boot", _box_mesh(Vector3(0.48, 0.62, 0.78)), Vector3(-0.52, 0, -0.08)))
			root.add_child(_mesh_part("right_boot", _box_mesh(Vector3(0.48, 0.62, 0.78)), Vector3(0.52, 0, -0.08)))
		"gloves":
			root.add_child(_mesh_part("left_glove", _box_mesh(Vector3(0.42, 0.42, 0.36)), Vector3(-0.36, 0, 0)))
			root.add_child(_mesh_part("right_glove", _box_mesh(Vector3(0.42, 0.42, 0.36)), Vector3(0.36, 0, 0)))
		"belt":
			root.add_child(_mesh_part("belt_band", _box_mesh(Vector3(1.05, 0.24, 0.34)), Vector3.ZERO))
			root.add_child(_mesh_part("belt_buckle", _box_mesh(Vector3(0.24, 0.28, 0.40)), Vector3(0, 0, -0.04)))
		"amulet":
			root.add_child(_mesh_part("amulet_chain", _cylinder_mesh(0.34, 0.04, 24), Vector3.ZERO, Vector3(90, 0, 0)))
			root.add_child(_mesh_part("amulet_gem", _box_mesh(Vector3(0.20, 0.24, 0.12)), Vector3(0, -0.32, 0)))
		"ring_left", "ring_right":
			root.add_child(_mesh_part("ring_band", _cylinder_mesh(0.32, 0.06, 24), Vector3.ZERO, Vector3(90, 0, 0)))
			root.add_child(_mesh_part("ring_stone", _box_mesh(Vector3(0.14, 0.12, 0.10)), Vector3(0, -0.30, 0)))
		_:
			return null
	return root


func _mesh_part(name: String, mesh: Mesh, position: Vector3, rotation_degrees: Vector3 = Vector3.ZERO) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = name
	part.mesh = mesh
	part.position = position
	part.rotation_degrees = rotation_degrees
	return part


func _box_mesh(size: Vector3) -> BoxMesh:
	var mesh := BoxMesh.new()
	mesh.size = size
	return mesh


func _cylinder_mesh(radius: float, height: float, radial_segments: int) -> CylinderMesh:
	var mesh := CylinderMesh.new()
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = height
	mesh.radial_segments = radial_segments
	return mesh


func _res_path(runtime_path: String) -> String:
	# Manifest runtime_path is repo-root-relative (client/assets/...); the Godot
	# project root IS client/, so strip the leading client/ and prepend res://.
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p


func _warn(entry: Dictionary) -> void:
	push_warning("[equip-visual] %s" % JSON.stringify(entry))
	_warnings.append(entry)


func _load_data() -> void:
	# Repo-root shared/manifest JSON via the v0 cross-language pattern
	# (test_golden.gd): project root res:// is client/, so shared/ and assets/
	# sit one level up.
	var base := ProjectSettings.globalize_path("res://")
	var iv = _read_json(base.path_join("../shared/assets/item_visuals.v0.json"))
	_visuals = iv.get("item_visuals", {}) if iv != null else {}
	var mf = _read_json(base.path_join("../assets/manifests/assets.v0.json"))
	_assets = mf.get("assets", {}) if mf != null else {}


func _read_json(path: String):
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("[equip-visual] cannot open %s" % path)
		return null
	return JSON.parse_string(f.get_as_text())
