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
# drive one code path. The resolver owns the inventory->def cache and the
# currently-equipped weapon instance, so callers just forward protocol events.
extends RefCounted
class_name EquipmentVisualResolver

var _mount_root: Node3D
var _visuals: Dictionary = {}      # item_def_id -> visual metadata
var _assets: Dictionary = {}       # asset_id -> manifest entry
var _inventory: Dictionary = {}    # item_instance_id (String) -> item_def_id (String)
var _equipped_weapon: String = ""  # equipped weapon instance id ("" = none)
var _mounted_node: Node3D = null
var _mounted_state: Dictionary = {}
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
	var weapon = payload.get("equipped", {}).get("weapon", null)
	_equipped_weapon = str(weapon) if weapon != null else ""
	_refresh_weapon()


func ingest_inventory_item(item: Dictionary) -> void:
	# Handles inventory_add / inventory_update. If it's the equipped item finally
	# arriving/resolving, (re)mount.
	_record_item(item)
	if str(item.get("item_instance_id", "")) == _equipped_weapon:
		_refresh_weapon()


func apply_equipped_update(slot: String, item_instance_id) -> void:
	if slot != "weapon":
		return
	_equipped_weapon = str(item_instance_id) if item_instance_id != null else ""
	_refresh_weapon()


# --- debug surface (spec §4.4 / §4.7) ---------------------------------------

func get_debug_state() -> Dictionary:
	var weapon = _mounted_state if not _mounted_state.is_empty() else null
	return {
		"equipped_visuals": {"weapon": weapon},
		"warnings": _warnings,
	}


# --- internals --------------------------------------------------------------

func _record_item(item: Dictionary) -> void:
	var iid := str(item.get("item_instance_id", ""))
	if iid != "":
		_inventory[iid] = str(item.get("item_def_id", ""))


func _refresh_weapon() -> void:
	# Each refresh recomputes from scratch: clear the prior mount (spec §7: no
	# duplicate stale nodes) and reset transient warnings for this attempt.
	_warnings = []
	_clear_mounted()
	_mounted_state = {}

	if _equipped_weapon == "":
		return

	var def_id: String = str(_inventory.get(_equipped_weapon, ""))
	if def_id == "":
		# Equipped instance not (yet) in the inventory cache; a later
		# inventory_add/snapshot will resolve it. Surface it, render nothing.
		_warn({"code": "unknown_item_instance_id", "item_instance_id": _equipped_weapon})
		return

	var vis = _visuals.get(def_id, null)
	if vis == null:
		_warn({"code": "unknown_item_def_id", "item_def_id": def_id})
		return

	var asset_id: String = str(vis["asset_id"])
	var entry = _assets.get(asset_id, null)
	if entry == null:
		_warn({"code": "unknown_asset_id", "asset_id": asset_id})
		return

	if _mount_root == null:
		_warn({"code": "missing_mount_socket", "mount_socket": str(vis["mount_socket"])})
		return
	var socket := _mount_root.find_child(str(vis["mount_socket"]), true, false)
	if socket == null:
		_warn({"code": "missing_mount_socket", "mount_socket": str(vis["mount_socket"])})
		return

	var packed = load(_res_path(str(entry["runtime_path"])))
	if packed == null:
		_warn({"code": "unknown_asset_id", "asset_id": asset_id})
		return

	var inst := (packed as PackedScene).instantiate()
	inst.name = asset_id
	_apply_transform(inst, vis.get("local_transform", {}))
	socket.add_child(inst)
	_mounted_node = inst
	_mounted_state = {
		"item_instance_id": _equipped_weapon,
		"item_def_id": def_id,
		"asset_id": asset_id,
		"mount_socket": str(vis["mount_socket"]),
		"node_path": (str(inst.get_path()) if inst.is_inside_tree() else ""),
		"visible": inst.visible,
	}


func _clear_mounted() -> void:
	if _mounted_node != null and is_instance_valid(_mounted_node):
		_mounted_node.queue_free()
	_mounted_node = null


func _apply_transform(node: Node3D, t: Dictionary) -> void:
	if t == null or t.is_empty():
		return
	var p = t.get("position", {})
	node.position = Vector3(p.get("x", 0.0), p.get("y", 0.0), p.get("z", 0.0))
	var r = t.get("rotation_degrees", {})
	node.rotation_degrees = Vector3(r.get("x", 0.0), r.get("y", 0.0), r.get("z", 0.0))
	var s = t.get("scale", {})
	node.scale = Vector3(s.get("x", 1.0), s.get("y", 1.0), s.get("z", 1.0))


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
