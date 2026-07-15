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

const WeaponSetTabsScript := preload("res://scripts/weapon_set_tabs.gd")
const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")
const EquipmentDisplayLoaderScript := preload("res://scripts/equipment_display_loader.gd")
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
var _character_class: String = ""
var _visuals: Dictionary = {}      # item_def_id -> visual metadata
var _assets: Dictionary = {}       # asset_id -> manifest entry
var _inventory: Dictionary = {}    # item_instance_id (String) -> inventory item Dictionary
var _equipped: Dictionary = {}      # slot -> item_instance_id
var _weapon_sets: Array = []
var _active_weapon_set: int = 0
var _mounted_nodes: Dictionary = {} # slot -> Node3D
var _mounted_mirror_nodes: Dictionary = {} # slot -> Node3D (e.g. right boot)
var _eye_view_camera: Camera3D
var _eye_view_enabled := false
var _eye_view_cfg: Dictionary = {}
var _eye_view_nodes: Dictionary = {}
var _eye_view_state: Dictionary = {}
var _eye_view_attack_count := 0
var _mounted_state: Dictionary = {} # slot -> debug state
var _warnings: Array = []


func set_character_class(class_id: String) -> void:
	_character_class = str(class_id)
	_refresh_all()


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
	_weapon_sets = []
	for set_data in payload.get("weapon_sets", []):
		_weapon_sets.append((set_data as Dictionary).duplicate(true))
	_active_weapon_set = int(payload.get("active_weapon_set", 0))
	_refresh_all()


func ingest_inventory_item(item: Dictionary) -> void:
	# Handles inventory_add / inventory_update. If it's the equipped item finally
	# arriving/resolving, (re)mount.
	_record_item(item)
	var item_id := str(item.get("item_instance_id", ""))
	for slot in _equipped.keys():
		if str(_equipped.get(slot, "")) == item_id:
			_refresh_slot(str(slot))


func apply_weapon_set_update(active_weapon_set: int, weapon_sets: Array) -> void:
	_active_weapon_set = clamp(active_weapon_set, 0, 1)
	_weapon_sets = []
	for set_data in weapon_sets:
		_weapon_sets.append((set_data as Dictionary).duplicate(true))
	_refresh_slot("main_hand")
	_refresh_slot("off_hand")


func apply_equipped_update(slot: String, item_instance_id) -> void:
	if not EQUIPMENT_SLOTS.has(slot):
		return
	_equipped[slot] = str(item_instance_id) if item_instance_id != null else ""
	_refresh_slot(slot)


func set_eye_view(camera: Camera3D, enabled: bool, cfg: Dictionary) -> void:
	_eye_view_camera = camera
	_eye_view_enabled = enabled and camera != null and is_instance_valid(camera)
	_eye_view_cfg = cfg.duplicate(true)
	_refresh_eye_view()


func present_eye_view_attack(slot: String = "main_hand") -> void:
	var node = _eye_view_nodes.get(slot, null)
	if node == null or not is_instance_valid(node):
		return
	var root := node as Node3D
	var base := _position_from_state(slot, root.position)
	root.position = base + Vector3(0.0, 0.04, -0.16)
	_eye_view_attack_count += 1
	if root.is_inside_tree():
		var tween := root.create_tween()
		tween.tween_property(root, "position", base, 0.16).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_OUT)
	_update_eye_view_motion_state(slot, root)


# --- debug surface (spec §4.4 / §4.7) ---------------------------------------

func get_debug_state() -> Dictionary:
	var visuals := {}
	for slot in EQUIPMENT_SLOTS:
		if _mounted_state.has(str(slot)):
			visuals[str(slot)] = (_mounted_state[str(slot)] as Dictionary).duplicate(true)
	visuals["weapon"] = visuals.get("main_hand", null)
	return {
		"equipped_visuals": visuals,
		"eye_view": _eye_view_state.duplicate(true),
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
	_refresh_eye_view()


func _refresh_slot(slot: String, reset_warnings: bool = true) -> void:
	# Each slot refresh recomputes from scratch: clear the prior mount (spec §7:
	# no duplicate stale nodes) and reset transient warnings for this attempt.
	if reset_warnings:
		_warnings = []
	_clear_mounted(slot)
	_mounted_state.erase(slot)

	var item_instance_id := _equipped_instance_id_for_slot(slot)
	if item_instance_id == "":
		return
	if slot == "off_hand" and _main_hand_blocks_off_hand_visual():
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

	var procedural_fallback := _procedural_fallback_visual(asset_id, slot, entry)
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
	_apply_transform(inst, _local_transform_for_slot(slot, vis), procedural_fallback == null)
	_apply_model_root_scale_compensation(inst, socket)
	var rarity := str(item.get("rarity", "common")).to_lower()
	var tint: Color = RARITY_TINTS.get(rarity, RARITY_TINTS["common"])
	_apply_tint(inst, tint)
	socket.add_child(inst)
	_mounted_nodes[slot] = inst
	if slot == "boots":
		_mount_boots_mirror(socket, _right_boot_transform(vis, slot), asset_id, entry, procedural_fallback == null)
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
	if slot == "main_hand" or slot == "off_hand":
		_refresh_eye_view_slot(slot)


func _equipped_instance_id_for_slot(slot: String) -> String:
	if slot == "main_hand" or slot == "off_hand":
		var item_id = WeaponSetTabsScript.hand_equipped_id(_weapon_sets, _equipped, _active_weapon_set, slot)
		return str(item_id) if item_id != null else ""
	return str(_equipped.get(slot, ""))


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
	if _character_class != "":
		var class_transforms: Dictionary = vis.get("class_transforms", {}) if typeof(vis.get("class_transforms", {})) == TYPE_DICTIONARY else {}
		var class_entry: Dictionary = class_transforms.get(_character_class, {}) if typeof(class_transforms.get(_character_class, {})) == TYPE_DICTIONARY else {}
		var class_transform: Dictionary = class_entry.get("local_transform", {}) if typeof(class_entry.get("local_transform", {})) == TYPE_DICTIONARY else {}
		if not class_transform.is_empty():
			transform = _merge_transform(transform, class_transform)
	if slot != "off_hand" or str(vis.get("slot", slot)) == slot:
		return transform
	var position: Dictionary = (transform.get("position", {}) as Dictionary).duplicate(true)
	position["z"] = float(position.get("z", 0.0)) + 0.08
	transform["position"] = position
	var rotation: Dictionary = (transform.get("rotation_degrees", {}) as Dictionary).duplicate(true)
	rotation["z"] = float(rotation.get("z", 0.0)) + 180.0
	transform["rotation_degrees"] = rotation
	return transform


func _merge_transform(base: Dictionary, override_transform: Dictionary) -> Dictionary:
	var out := base.duplicate(true)
	for key in ["position", "rotation_degrees", "scale"]:
		if not override_transform.has(key):
			continue
		var patch: Dictionary = override_transform.get(key, {}) if typeof(override_transform.get(key, {})) == TYPE_DICTIONARY else {}
		if patch.is_empty():
			continue
		var existing: Dictionary = out.get(key, {}) if typeof(out.get(key, {})) == TYPE_DICTIONARY else {}
		var merged_component := existing.duplicate(true)
		merged_component.merge(patch, true)
		out[key] = merged_component
	return out


func _main_hand_blocks_off_hand_visual() -> bool:
	var main_id := _equipped_instance_id_for_slot("main_hand")
	if main_id == "":
		return false
	var item: Dictionary = _inventory.get(main_id, {})
	var def_id := str(item.get("item_def_id", ""))
	if def_id == "":
		return false
	ItemRulesLoaderScript.ensure_loaded()
	var def: Dictionary = ItemRulesLoaderScript.item_definition(def_id)
	if str(def.get("handedness", "")) == "two_handed":
		return true
	var occupies = def.get("occupies_hands", [])
	return typeof(occupies) == TYPE_ARRAY and (occupies as Array).has("off_hand")


func _clear_mounted(slot: String) -> void:
	var mounted = _mounted_nodes.get(slot, null)
	if mounted != null and is_instance_valid(mounted):
		(mounted as Node3D).queue_free()
	_mounted_nodes.erase(slot)
	var mirror = _mounted_mirror_nodes.get(slot, null)
	if mirror != null and is_instance_valid(mirror):
		(mirror as Node3D).queue_free()
	_mounted_mirror_nodes.erase(slot)
	if slot == "main_hand" or slot == "off_hand":
		_clear_eye_view_slot(slot)


func _refresh_eye_view() -> void:
	for slot in ["main_hand", "off_hand"]:
		_refresh_eye_view_slot(slot)


func _refresh_eye_view_slot(slot: String) -> void:
	_clear_eye_view_slot(slot)
	_eye_view_state[slot] = {"active": false, "visible": false, "slot": slot}
	if not _eye_view_enabled or _eye_view_camera == null or not is_instance_valid(_eye_view_camera):
		return
	if slot == "off_hand" and _main_hand_blocks_off_hand_visual():
		return
	var item_instance_id := _equipped_instance_id_for_slot(slot)
	if item_instance_id == "":
		return
	var item: Dictionary = _inventory.get(item_instance_id, {})
	var def_id := str(item.get("item_def_id", ""))
	if def_id == "":
		return
	var vis: Dictionary = _visual_for(def_id, slot)
	if vis.is_empty():
		return
	var asset_id := str(vis.get("asset_id", ""))
	var entry = _assets.get(asset_id, null)
	if entry == null:
		return
	var node := _instantiate_visual(asset_id, slot, entry)
	if node == null:
		return
	node.name = "EyeView_%s_%s" % [slot, asset_id]
	_apply_transform(node, _eye_view_transform_for_slot(slot), false)
	var rarity := str(item.get("rarity", "common")).to_lower()
	_apply_tint(node, RARITY_TINTS.get(rarity, RARITY_TINTS["common"]))
	_eye_view_camera.add_child(node)
	_eye_view_nodes[slot] = node
	_eye_view_state[slot] = {
		"active": true,
		"visible": node.visible,
		"attack_count": _eye_view_attack_count,
		"slot": slot,
		"item_instance_id": item_instance_id,
		"item_def_id": def_id,
		"asset_id": asset_id,
		"node_path": str(node.get_path()) if node.is_inside_tree() else "",
		"position": {"x": node.position.x, "y": node.position.y, "z": node.position.z},
	}


func _instantiate_visual(asset_id: String, slot: String, entry) -> Node3D:
	var procedural_fallback := _procedural_fallback_visual(asset_id, slot, entry)
	if procedural_fallback != null:
		return procedural_fallback
	var packed = load(_res_path(str(entry["runtime_path"])))
	if packed == null:
		return null
	return (packed as PackedScene).instantiate()


func _eye_view_transform_for_slot(slot: String) -> Dictionary:
	var pos := _vec3_dict(_eye_view_cfg.get("view_model_position", [0.34, -0.32, -0.82]))
	if slot == "off_hand":
		pos["x"] = -float(pos.get("x", 0.34))
	var rot := _vec3_dict(_eye_view_cfg.get("view_model_rotation_degrees", [-10.0, -18.0, 8.0]))
	if slot == "off_hand":
		rot["y"] = -float(rot.get("y", -18.0))
		rot["z"] = -float(rot.get("z", 8.0))
	var scale := float(_eye_view_cfg.get("view_model_scale", 0.82))
	return {
		"position": pos,
		"rotation_degrees": rot,
		"scale": {"x": scale, "y": scale, "z": scale},
	}


func _vec3_dict(value) -> Dictionary:
	if value is Array and (value as Array).size() >= 3:
		return {"x": float(value[0]), "y": float(value[1]), "z": float(value[2])}
	return {"x": 0.0, "y": 0.0, "z": 0.0}


func _clear_eye_view_slot(slot: String) -> void:
	var mounted = _eye_view_nodes.get(slot, null)
	if mounted != null and is_instance_valid(mounted):
		(mounted as Node3D).queue_free()
	_eye_view_nodes.erase(slot)


func _position_from_state(slot: String, fallback: Vector3) -> Vector3:
	var slot_state: Dictionary = _eye_view_state.get(slot, {})
	var pos: Dictionary = slot_state.get("position", {})
	if pos.is_empty():
		return fallback
	return Vector3(float(pos.get("x", fallback.x)), float(pos.get("y", fallback.y)), float(pos.get("z", fallback.z)))


func _update_eye_view_motion_state(slot: String, node: Node3D) -> void:
	var slot_state: Dictionary = _eye_view_state.get(slot, {}).duplicate(true)
	slot_state["attack_count"] = _eye_view_attack_count
	slot_state["motion_active"] = true
	slot_state["motion_position"] = {"x": node.position.x, "y": node.position.y, "z": node.position.z}
	_eye_view_state[slot] = slot_state


func _right_boot_transform(vis: Dictionary, slot: String) -> Dictionary:
	var transform := _local_transform_for_slot(slot, vis).duplicate(true)
	var position: Dictionary = (transform.get("position", {}) as Dictionary).duplicate(true)
	position["x"] = -float(position.get("x", 0.0))
	transform["position"] = position
	return transform


func _mount_boots_mirror(
		left_socket: Node,
		transform: Dictionary,
		asset_id: String,
		entry,
		apply_glb_mesh_multiplier: bool,
	) -> void:
	var right_socket := _mount_root.find_child("boots_right_socket", true, false)
	if right_socket == null:
		return
	var procedural_fallback := _procedural_fallback_visual(asset_id, "boots", entry)
	var right_inst: Node3D
	if procedural_fallback != null:
		right_inst = procedural_fallback
	else:
		var packed = load(_res_path(str(entry["runtime_path"])))
		if packed == null:
			return
		right_inst = (packed as PackedScene).instantiate()
	right_inst.name = "%s_right" % asset_id
	_apply_transform(right_inst, transform, procedural_fallback == null)
	_apply_model_root_scale_compensation(right_inst, right_socket)
	right_socket.add_child(right_inst)
	_mounted_mirror_nodes["boots"] = right_inst


func _apply_transform(node: Node3D, t: Dictionary, apply_glb_mesh_multiplier: bool = false) -> void:
	if t == null or t.is_empty():
		return
	var p = t.get("position", {})
	node.position = Vector3(p.get("x", 0.0), p.get("y", 0.0), p.get("z", 0.0))
	var r = t.get("rotation_degrees", {})
	node.rotation_degrees = Vector3(r.get("x", 0.0), r.get("y", 0.0), r.get("z", 0.0))
	var s = t.get("scale", {})
	var mesh_mult := EquipmentDisplayLoaderScript.equipped_multiplier() if apply_glb_mesh_multiplier else 1.0
	node.scale = Vector3(
		float(s.get("x", 1.0)) * mesh_mult,
		float(s.get("y", 1.0)) * mesh_mult,
		float(s.get("z", 1.0)) * mesh_mult,
	)


func _apply_model_root_scale_compensation(node: Node3D, socket: Node) -> void:
	var model_scale := _model_root_scale_for(socket)
	node.position = Vector3(
		node.position.x * _inverse_scale_component(model_scale.x),
		node.position.y * _inverse_scale_component(model_scale.y),
		node.position.z * _inverse_scale_component(model_scale.z)
	)
	node.scale = Vector3(
		node.scale.x * _inverse_scale_component(model_scale.x),
		node.scale.y * _inverse_scale_component(model_scale.y),
		node.scale.z * _inverse_scale_component(model_scale.z)
	)


func _model_root_scale_for(node: Node) -> Vector3:
	var current := node
	while current != null and current != _mount_root:
		if current is Node3D and str(current.name) == "ModelRoot":
			return (current as Node3D).scale
		current = current.get_parent()
	return Vector3.ONE


func _inverse_scale_component(value: float) -> float:
	if absf(value) <= 0.0001:
		return 1.0
	return 1.0 / value


func _apply_tint(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		(root as MeshInstance3D).material_override = mat
	for child in root.get_children():
		_apply_tint(child, color)


func _procedural_fallback_visual(asset_id: String, slot: String, entry) -> Node3D:
	if not asset_id.begins_with("fallback_equipment_"):
		return null
	if typeof(entry) == TYPE_DICTIONARY:
		var runtime_path := str((entry as Dictionary).get("runtime_path", ""))
		if runtime_path.find("/equipment/weapons/rusty_sword/") == -1:
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


func reload_from_disk() -> void:
	_load_data()
	_refresh_all()


func reload_data_only() -> void:
	_load_data()


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
