## CrosshairTargetSystem — ray pick, reach lock, highlight, and name tag for actionable targets.
##
## Perspective modes use the viewport center ray; isometric uses the mouse ray from context.
## Reachable targets get a rarity-colored highlight and name tag in every camera mode.
class_name CrosshairTargetSystem
extends RefCounted

const CameraPresentationsLoaderScript := preload("res://scripts/camera_presentations_loader.gd")
const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const CrosshairTargetNamesScript := preload("res://scripts/crosshair_target_names.gd")
const CrosshairTargetNameTagScript := preload("res://scripts/crosshair_target_name_tag.gd")
const PickTargetHighlightScript := preload("res://scripts/pick_target_highlight.gd")

const RAY_LENGTH := 200.0

var _ctx  # CrosshairTargetContext
var _locked_id := ""
var _highlighted_id := ""
var _highlight_kind := ""
var _name_tag  # CrosshairTargetNameTag
var _client_settings


func setup(ctx) -> void:
	_ctx = ctx


func sync_runtime(inv: Array, equip: Dictionary) -> void:
	if _ctx == null:
		return
	_ctx.inventory = inv
	_ctx.equipped = equip


static func perspective_reticle_active(client_settings) -> bool:
	if client_settings == null or client_settings.camera_mode == "isometric":
		return false
	CameraPresentationsLoaderScript.ensure_loaded()
	return bool(CameraPresentationsLoaderScript.mode(client_settings.camera_mode).get("reticle_enabled", false))


static func uses_center_ray_pick(client_settings) -> bool:
	return perspective_reticle_active(client_settings)


func tick_runtime(viewport: Viewport, world: World3D, inv: Array, equip: Dictionary, client_settings, input_locked: bool) -> void:
	sync_runtime(inv, equip)
	_client_settings = client_settings
	if input_locked or client_settings == null:
		clear()
		return
	tick(viewport, world)


func clear() -> void:
	_set_locked("")


func tick(viewport: Viewport, world: World3D) -> void:
	if _ctx == null or _ctx.camera == null or viewport == null:
		clear()
		return
	if not _ctx.ray_pick_entity.is_valid() and world == null:
		clear()
		return

	var candidate := _resolve_candidate(viewport, world)
	var next_locked := ""
	if candidate != "" and _target_in_reach(candidate):
		next_locked = candidate
	_set_locked(next_locked)
	_sync_name_tag()


func locked_target_id() -> String:
	return _locked_id


func build_click_pick() -> Dictionary:
	if _locked_id == "" or _ctx == null or not _ctx.entities.has(_locked_id):
		return {}

	var rec: Dictionary = _ctx.entities[_locked_id]
	var typ := str(rec.get("type", ""))
	if typ == "monster" and not _is_dead_monster(_locked_id):
		return {"kind": "monster", "target_id": _locked_id}

	return {"kind": "oneshot", "target_id": _locked_id}


func build_use_pick() -> Dictionary:
	if _locked_id == "" or _ctx == null or not _ctx.entities.has(_locked_id):
		return {}

	var rec: Dictionary = _ctx.entities[_locked_id]
	var typ := str(rec.get("type", ""))
	if typ == "monster" and not _is_dead_monster(_locked_id):
		return {}

	return {"kind": "oneshot", "target_id": _locked_id}


func _set_locked(target_id: String) -> void:
	if _locked_id == target_id:
		return

	_clear_highlight()
	_locked_id = target_id
	if _ctx != null and _ctx.aim_reticle != null:
		_ctx.aim_reticle.set_locked(target_id != "" and perspective_reticle_active(_client_settings))
	if target_id == "":
		_hide_name_tag()
		return
	_apply_highlight(target_id)


func _sync_name_tag() -> void:
	if _locked_id == "" or _ctx == null:
		_hide_name_tag()
		return
	if not _ctx.entities.has(_locked_id):
		_hide_name_tag()
		return
	var rec: Dictionary = _ctx.entities[_locked_id]
	var node := rec.get("node", null) as Node3D
	var text := CrosshairTargetNamesScript.display_name(rec)
	var accent := _highlight_color_for(rec)
	if node == null or text == "" or _ctx.camera == null or _ctx.name_tag_parent == null:
		_hide_name_tag()
		return
	_ensure_name_tag()
	_name_tag.attach_to(_ctx.name_tag_parent)
	_name_tag.show_for(_ctx.camera, node, text, accent)


func _ensure_name_tag() -> void:
	if _name_tag != null:
		return
	_name_tag = CrosshairTargetNameTagScript.new()


func _hide_name_tag() -> void:
	if _name_tag != null:
		_name_tag.hide_tag()


func get_name_tag_text() -> String:
	return _name_tag.get_label_text() if _name_tag != null else ""


func _resolve_candidate(viewport: Viewport, world: World3D) -> String:
	if uses_center_ray_pick(_client_settings) and world != null:
		return _pick_entity_at_center(viewport, world)
	if _ctx != null and _ctx.ray_pick_entity.is_valid():
		return str(_ctx.ray_pick_entity.call(viewport, world))
	if world != null:
		return _pick_entity_at_center(viewport, world)
	return ""


func _pick_entity_at_center(viewport: Viewport, world: World3D) -> String:
	var center := viewport.get_visible_rect().get_center()
	var from: Vector3 = _ctx.camera.project_ray_origin(center)
	var normal: Vector3 = _ctx.camera.project_ray_normal(center)
	var query := PhysicsRayQueryParameters3D.create(from, from + normal * RAY_LENGTH)
	query.collide_with_areas = true
	query.collide_with_bodies = true
	var hit := world.direct_space_state.intersect_ray(query)
	if hit.is_empty():
		return ""

	var collider = hit.get("collider")
	if collider == null or not collider.has_meta("entity_id"):
		return ""

	var hit_entity_id: String = str(collider.get_meta("entity_id"))
	if _is_dead_monster(hit_entity_id) and not bool(_ctx.revive_hover_enabled.call()):
		var ground: Vector3 = hit.get("position", _ctx.center_ground_point.call())
		var loot_id: String = str(_ctx.nearest_loot_at_ground.call(ground))
		if loot_id != "":
			return loot_id

	return hit_entity_id


func _target_in_reach(target_id: String) -> bool:
	if _ctx == null or not _ctx.target_in_reach.is_valid():
		return false
	return bool(_ctx.target_in_reach.call(target_id))


func _is_dead_monster(entity_id: String) -> bool:
	if _ctx == null or not _ctx.entities.has(entity_id):
		return false
	var rec: Dictionary = _ctx.entities[entity_id]
	return str(rec.get("type", "")) == "monster" and int(rec.get("hp", 1)) <= 0


func _reaction_for(entity_id: String):
	if _ctx == null or not _ctx.entities.has(entity_id):
		return null
	return _ctx.entities[entity_id].get("reaction", null)


func _highlight_color_for(rec: Dictionary) -> Color:
	return ClientConstantsScript.target_highlight_color(str(rec.get("type", "")), str(rec.get("rarity", "common")))


func _apply_highlight(entity_id: String) -> void:
	if _ctx == null or not _ctx.entities.has(entity_id):
		return
	var rec: Dictionary = _ctx.entities[entity_id]
	var typ := str(rec.get("type", ""))
	var highlight_color := _highlight_color_for(rec)
	var reaction = _reaction_for(entity_id)
	if reaction != null:
		reaction.set_highlight(true, highlight_color)
		_highlight_kind = "reaction"
	elif typ in ["interactable", "loot"]:
		var node := rec.get("node", null) as Node3D
		if node != null:
			PickTargetHighlightScript.set_highlight(node, true, highlight_color)
			_highlight_kind = "node"
	else:
		return
	_highlighted_id = entity_id


func _clear_highlight() -> void:
	if _highlighted_id == "":
		return
	if _highlight_kind == "reaction":
		var reaction = _reaction_for(_highlighted_id)
		if reaction != null:
			reaction.set_highlight(false)
	elif _highlight_kind == "node":
		if _ctx != null and _ctx.entities.has(_highlighted_id):
			var node := _ctx.entities[_highlighted_id].get("node", null) as Node3D
			if node != null:
				PickTargetHighlightScript.set_highlight(node, false)
	_highlighted_id = ""
	_highlight_kind = ""
